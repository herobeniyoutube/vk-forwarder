package application

import (
	"fmt"
	"log"

	config "github.com/herobeniyoutube/vk-forwarder/config"
	httpservice "github.com/herobeniyoutube/vk-forwarder/httpFunctions"
	postgresql "github.com/herobeniyoutube/vk-forwarder/storage/postresql"
	u "github.com/herobeniyoutube/vk-forwarder/utils"
)

type VkEventHandler struct {
	groupId              int64
	chatId               int64
	messageId            int64
	eventType            string
	retryCountHeader     int
	ignoreIdempotencyKey bool
	idempotencyKey       string
	text                 string
	callbackCode         string
	//services:
	sender         ISender
	callbackGetter ICallbackCodeGetter
	downloader     IVideoDownloader
	db             *postgresql.PostgresDb
	config         *config.Config

	payload MessageNewEvent
}

func NewVkEventHandler(s ISender, r ICallbackCodeGetter, d IVideoDownloader, db *postgresql.PostgresDb) IHandler {
	h := &VkEventHandler{
		sender:         s,
		callbackGetter: r,
		downloader:     d,
		db:             db}

	callbackCode, err := h.getCallbackConfirmationCode()
	if err != nil {
		panic("couldn't start service" + err.Error())
	}

	h.callbackCode = *callbackCode

	return h
}

func (h *VkEventHandler) Setup(event MessageNewEvent, retryCountHeader int, ignoreIdempotencyKey bool) IHandler {

	msg := event.Object.Message
	messageId := msg.MessageId
	chatId := msg.PeerId

	h.groupId = event.GroupID
	h.chatId = chatId
	h.messageId = messageId
	h.eventType = event.Type
	h.text = msg.Text
	h.idempotencyKey = u.BuildIdempotencyKey(chatId, messageId)
	h.ignoreIdempotencyKey = ignoreIdempotencyKey

	return h
}

func (h *VkEventHandler) Handle() (*string, error) {

	if h.retryCountHeader > 1 {
		log.Printf("Retry header value: %d", h.retryCountHeader)
	}

	switch h.eventType {
	case "confirmation":
		h.getCallbackConfirmationCode()
	case "message_new":
		var (
			res *string
			err error
		)
		found, err := h.db.HasIdempotencyKey(h.idempotencyKey)
		if err != nil {
			return nil, createError("error checking idempotency key", h.eventType, err)
		} else if found && !h.ignoreIdempotencyKey {
			r := "message_new processed already"
			return &r, nil
		}

		attachments := h.payload.Object.Message.Attachments
		attachmentType := attachments[0].Type

		if len(attachments) == 1 {

			switch attachmentType {
			case "video":
				res, err = h.downloadVideoHandler()
			case "photo":
				res, err = h.sendPhotoHandler()
			case "wall":
				res, err = h.sendWallPostHandler()
			default:

			}
		} else {
			switch attachmentType {

			default:
				return h.sendBatchHandler()
			}
		}
		h.addIdempotencyKey(err)

		return res, err

	}
	return nil, createError("event type not found", h.eventType, nil)
}

func (h *VkEventHandler) addIdempotencyKey(err error) {
	if !h.ignoreIdempotencyKey && err == nil {
		err = h.db.AddIdempotencyKey(h.idempotencyKey)
		if err != nil {
			log.Printf("Error creating idempotency key: %s", err.Error())
		}
	}
	// log.Printf("Idempotency key ignored: %s", err.Error())
}

func (h *VkEventHandler) getCallbackConfirmationCode() (*string, error) {
	code, err := h.callbackGetter.GetCallbackConfirmation()
	if code == nil {
		return code, createError("confirmation code empty"+err.Error(), h.eventType, nil)
	}

	return code, nil
}

func (h *VkEventHandler) downloadVideoHandler() (*string, error) {
	m := h.payload.Object.Message
	v := m.Attachments[0].Video

	downloadId, err := h.downloader.Download(v.Type, v.VideoId, v.OwnerId)
	if err != nil {
		return nil, createError("error downloading video", h.eventType, err)
	}
	defer h.downloader.DisposeSendedVideo(*downloadId)

	err = h.sender.SendClip(*downloadId, &m.Text)
	if err != nil {
		return nil, err
	}

	r := "video sended"

	return &r, nil
}

func (h *VkEventHandler) sendPhotoHandler() (*string, error) {
	photoUrl := h.payload.Object.Message.Attachments[0].Photo.OrigPhoto.Url

	err := h.sender.SendPhoto(photoUrl, &h.text)
	if err != nil {
		return nil, err
	}

	r := "photo sended"

	return &r, nil
}

func (h *VkEventHandler) sendBatchHandler() (*string, error) {
	a := h.payload.Object.Message.Attachments
	m := h.payload.Object.Message

	downloadIds := make([]string, 0)
	locations := make(map[string]string, 0)

	h.formMediaContent(a, downloadIds, locations)

	err := h.sender.SendBatch(locations, downloadIds, httpservice.Caption{Text: m.Text})
	if err != nil {
		return nil, err
	}

	r := "batch sended"

	d := VideoDownloader{h.groupId}
	for _, val := range downloadIds {
		d.DisposeSendedVideo(val)
	}

	return &r, nil
}

func (h *VkEventHandler) sendWallPostHandler() (*string, error) {
	w := h.payload.Object.Message.Attachments[0].Wall

	a := w.Attachments
	userMessage := h.payload.Object.Message
	groupName := w.From.Name
	wallText := w.Text

	text := httpservice.Caption{GroupName: groupName, UserMessage: userMessage.Text, Text: wallText}

	downloadIds := make([]string, 0)
	locations := make(map[string]string, 0)

	h.formMediaContent(a, downloadIds, locations)

	err := h.sender.SendBatch(locations, downloadIds, text)
	if err != nil {
		return nil, err
	}

	r := "batch sended"

	return &r, nil
}

func (h *VkEventHandler) formMediaContent(a []Attachment, downloadIds []string, locations map[string]string) {
	for _, value := range a {
		if value.Type == "video" {

			downloadId, err := h.downloader.Download(value.Type, value.Video.VideoId, value.Video.OwnerId)
			if err != nil {
				log.Println(createError("error downloading video", h.eventType, err).Error())
				continue
			}

			downloadIds = append(downloadIds, *downloadId)
			locations[*downloadId] = "video"
		} else {
			url := value.Photo.OrigPhoto.Url
			locations[url] = "photo"
		}
	}
}

// probably could use this as something like generic in utils
func createError(info string, eventType string, err error) HandlerError {
	i := fmt.Sprintf("Error happened while handling event %s: %s", eventType, info)

	if err != nil {
		i += fmt.Sprintf(" Inner error: %s", err.Error())
	}

	return HandlerError{i}
}
