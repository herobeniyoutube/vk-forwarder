package application

import (
	"fmt"
	"log"

	"github.com/herobeniyoutube/vk-forwarder/domain/statuses"
	u "github.com/herobeniyoutube/vk-forwarder/utils"
)

type VkEventHandler struct {
	callbackCode string
	//services:
	sender         ISender
	callbackGetter ICallbackCodeGetter
	downloader     IVideoDownloader
	repo           IIdempotencyRepo
}

func NewVkEventHandler(s ISender, r ICallbackCodeGetter, d IVideoDownloader, repo IIdempotencyRepo) IHandler {
	h := &VkEventHandler{
		sender:         s,
		callbackGetter: r,
		downloader:     d,
		repo:           repo}

	callbackCode, err := h.getCallbackConfirmationCode()
	if err != nil {
		panic("couldn't start service" + err.Error())
	}

	h.callbackCode = *callbackCode

	return h
}

func (h *VkEventHandler) Handle(event MessageNewEvent, retryCountHeader int, ignoreIdempotencyKey bool) (*string, error) {
	if retryCountHeader > 1 {
		log.Printf("Retry header value: %d", retryCountHeader)
	}

	msg := event.Object.Message
	messageId := msg.MessageId
	chatId := msg.PeerId
	eventType := event.Type
	idempotencyKey := u.BuildIdempotencyKey(chatId, messageId)

	switch eventType {
	case "confirmation":
		return h.getCallbackConfirmationCode()
	case "message_new":
		var (
			res *string
			err error
		)

		if !ignoreIdempotencyKey {
			_, err := h.repo.AddOrUpdateIdempotencyKey(idempotencyKey, statuses.Processing)
			if err != nil {
				if err, ok := err.(ProcessStatusRestrictions); ok {
					switch err.Status {
					case statuses.Processing:
						s := string(statuses.Processing)
						return &s, nil
					case statuses.Done:
						s := err.Error()
						return &s, nil
					}
				}
				return nil, err
			}
		}

		// if err != nil {
		// 	return nil, createError("error checking idempotency key", eventType, err)
		// } else if found && !ignoreIdempotencyKey {
		// 	r := "message_new processed already"
		// 	return &r, nil
		// }

		attachments := msg.Attachments
		if len(attachments) == 0 {
			return nil, createError("attachments are empty", eventType, nil)
		}
		attachmentType := attachments[0].Type

		if len(attachments) == 1 {

			switch attachmentType {
			case "video":
				res, err = h.downloadVideoHandler(msg, eventType, event.GroupID)
			case "photo":
				res, err = h.sendPhotoHandler(msg)
			case "wall":
				res, err = h.sendWallPostHandler(msg, eventType, event.GroupID)
			default:

			}
		} else {
			switch attachmentType {

			default:
				return h.sendBatchHandler(msg, eventType, event.GroupID)
			}
		}

		if !ignoreIdempotencyKey {
			var repoErr error
			if err != nil {
				repoErr = h.repo.UpdateStatus(idempotencyKey, statuses.Error)
			} else {
				repoErr = h.repo.UpdateStatus(idempotencyKey, statuses.Done)

			}
			if repoErr != nil {
				log.Print(repoErr.Error())
			}
		}

		return res, err

	}
	return nil, createError("event type not found", eventType, nil)
}

// func (h *VkEventHandler) addIdempotencyKey(idempotencyKey string, ignoreIdempotencyKey bool, err error) {
// 	if !ignoreIdempotencyKey && err == nil {
// 		err = h.repo.AddIdempotencyKey(idempotencyKey)
// 		if err != nil {
// 			log.Printf("Error creating idempotency key: %s", err.Error())
// 		}
// 	}
// 	// log.Printf("Idempotency key ignored: %s", err.Error())
// }

func (h *VkEventHandler) getCallbackConfirmationCode() (*string, error) {
	code, err := h.callbackGetter.GetCallbackConfirmation()
	if code == nil {
		return code, createError("confirmation code empty"+err.Error(), "confirmation", nil)
	}

	return code, nil
}

func (h *VkEventHandler) downloadVideoHandler(m Message, eventType string, groupId int64) (*string, error) {
	v := m.Attachments[0].Video

	downloadId, err := h.downloader.Download(groupId, v.Type, v.VideoId, v.OwnerId)
	if err != nil {
		return nil, createError("error downloading video", eventType, err)
	}
	defer h.downloader.DisposeSendedVideo(groupId, *downloadId)

	err = h.sender.SendClip(*downloadId, &m.Text)
	if err != nil {
		return nil, err
	}

	r := "video sended"

	return &r, nil
}

func (h *VkEventHandler) sendPhotoHandler(m Message) (*string, error) {
	photoUrl := m.Attachments[0].Photo.OrigPhoto.Url

	err := h.sender.SendPhoto(photoUrl, &m.Text)
	if err != nil {
		return nil, err
	}

	r := "photo sended"

	return &r, nil
}

func (h *VkEventHandler) sendBatchHandler(m Message, eventType string, groupId int64) (*string, error) {
	downloadIds, locations := h.formMediaContent(m.Attachments, eventType, groupId)

	err := h.sender.SendBatch(locations, downloadIds, Caption{Text: m.Text})
	if err != nil {
		return nil, err
	}

	r := "batch sended"

	for _, val := range downloadIds {
		h.downloader.DisposeSendedVideo(groupId, val)
	}

	return &r, nil
}

func (h *VkEventHandler) sendWallPostHandler(m Message, eventType string, groupId int64) (*string, error) {
	w := m.Attachments[0].Wall

	a := w.Attachments
	groupName := w.From.Name
	wallText := w.Text

	text := Caption{GroupName: groupName, UserMessage: m.Text, Text: wallText}

	downloadIds, locations := h.formMediaContent(a, eventType, groupId)

	err := h.sender.SendBatch(locations, downloadIds, text)
	if err != nil {
		return nil, err
	}

	r := "batch sended"

	return &r, nil
}

func (h *VkEventHandler) formMediaContent(a []Attachment, eventType string, groupId int64) ([]string, map[string]string) {
	downloadIds := make([]string, 0)
	locations := make(map[string]string, 0)

	for _, value := range a {
		if value.Type == "video" {

			downloadId, err := h.downloader.Download(groupId, value.Type, value.Video.VideoId, value.Video.OwnerId)
			if err != nil {
				log.Println(createError("error downloading video", eventType, err).Error())
				continue
			}

			downloadIds = append(downloadIds, *downloadId)
			locations[*downloadId] = "video"
		} else {
			url := value.Photo.OrigPhoto.Url
			locations[url] = "photo"
		}
	}

	return downloadIds, locations
}

// probably could use this as something like generic in utils
func createError(info string, eventType string, err error) HandlerError {
	i := fmt.Sprintf("Error happened while handling event %s: %s", eventType, info)

	if err != nil {
		i += fmt.Sprintf(" Inner error: %s", err.Error())
	}

	return HandlerError{i}
}
