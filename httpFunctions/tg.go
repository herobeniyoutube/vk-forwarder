package httpservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/herobeniyoutube/vk-forwarder/config"
	u "github.com/herobeniyoutube/vk-forwarder/utils"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type TgService struct {
	url       string
	chatId    int
	VkGroupId int64
}

type Media struct {
	Caption string `json:"caption,omitempty"`
	Type    string `json:"type"`
	Media   string `json:"media"`
}

type Caption struct {
	UserMessage string
	GroupName   string
	Text        string
}

func NewTgService(config config.Config) *TgService {
	url := fmt.Sprintf("https://api.telegram.org/bot%s", config.TgBotToken)
	chatId, err := strconv.Atoi(config.TgChatID)
	if err != nil {
		panic("error creating tg service" + err.Error())
	}

	vkGroupId, err := strconv.Atoi(config.VkGroupID)
	if err != nil {
		panic("error creating tg service" + err.Error())
	}

	return &TgService{url, chatId, int64(vkGroupId)}
}

func (tg *TgService) SendClip(downloadId string, caption *string) error {

	body, contentType, err := tg.createMultipleFilesBody([]string{downloadId})
	if err != nil {
		log.Printf("Error creating request body: %s", err.Error())
		return err
	}

	locations := map[string]string{
		downloadId: "video",
	}

	media := createBatchMediaType(locations, caption)
	err = tg.sendMediaGroup(string(media), body, contentType)
	if err != nil {
		return err
	}

	log.Printf("Sended video %s from vk groupId: %d for telegramm groupId %d", downloadId, tg.VkGroupId, tg.chatId)
	return nil
}

func (tg *TgService) SendPhoto(url string, caption *string) error {

	locations := map[string]string{
		url: "photo",
	}

	media := createBatchMediaType(locations, caption)

	tg.sendMediaGroup(string(media), nil, nil)

	log.Printf("Sended photo from vk groupId: %d for telegramm groupId %d", tg.VkGroupId, tg.chatId)
	return nil
}

func (tg *TgService) SendBatch(locations map[string]string, downloadIds []string, caption Caption) error {

	text := fmt.Sprintf("От: %s\n%s", caption.GroupName, caption.Text)
	if caption.UserMessage != "" {
		text = fmt.Sprintf("Сообщение пользователя: %s\n%s", caption.UserMessage, text)
	}

	isTooLong := 1024 < len(text)
	if isTooLong {
		text = fmt.Sprintf("От: %s", caption.GroupName)
	}

	media := createBatchMediaType(locations, &text)
	var err error
	var b *bytes.Buffer
	var contentType *string

	if downloadIds != nil || len(downloadIds) > 0 {
		b, contentType, err = tg.createMultipleFilesBody(downloadIds)
		if err != nil {
			log.Println(err.Error())
		}
	}

	err = tg.sendMediaGroup(string(media), b, contentType)
	if err != nil {
		return err
	}

	if isTooLong {
		err = tg.sendMessage(caption.Text)
		if err != nil {
			log.Println(err.Error())
		}
	}

	// err = s.sendMessage("")

	log.Printf("Sended photo from vk groupId: %d for telegramm groupId %d", tg.VkGroupId, tg.chatId)
	return nil
}

func (tg *TgService) sendMediaGroup(media string, body *bytes.Buffer, contentType *string) error {
	var b io.Reader

	if body != nil {
		b = body
	}

	url := fmt.Sprintf("%s/sendMediaGroup", tg.url)
	queries := map[string]string{
		"media": media,
	}

	headers := map[string]string{}
	if contentType != nil {
		headers["Content-Type"] = *contentType
	}

	err := tg.send(b, url, headers, queries)
	if err != nil {
		return err
	}

	return nil
}

func (tg *TgService) sendMessage(text string) error {
	url := fmt.Sprintf("%s/sendMessage", tg.url)

	queries := map[string]string{
		"text": text,
	}

	err := tg.send(nil, url, nil, queries)
	if err != nil {
		return err
	}

	return nil
}

func (tg *TgService) send(body io.Reader, url string, headhers map[string]string, queries map[string]string) error {
	var err error
	var b io.Reader

	if body != nil {
		b = body
	}

	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("chat_id", strconv.Itoa(tg.chatId))
	for key, value := range queries {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	for key, value := range headhers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("telegram response: status=%d body=%s", resp.StatusCode, string(respBody))

	return nil
}

func createBatchMediaType(locations map[string]string, caption *string) []byte {
	mediaInputData := make([]Media, 0)

	for key, value := range locations {
		location := key

		if value == "video" {
			location = fmt.Sprintf("attach://%s", key)
		}

		mediaInputData = append(mediaInputData, Media{"", value, location})
	}

	mediaInputData[0].Caption = *caption

	mediaJSON, err := json.Marshal(mediaInputData)
	if err != nil {
		panic(err.Error())
	}

	return mediaJSON
}

func (tg *TgService) createMultipleFilesBody(downloadIds []string) (*bytes.Buffer, *string, error) {
	var buffer bytes.Buffer

	if len(downloadIds) == 0 {
		return nil, nil, nil
	}

	writer := multipart.NewWriter(&buffer)
	defer writer.Close()

	for _, downloadId := range downloadIds {
		pathToVideo := u.GetVideoPath(tg.VkGroupId, downloadId)
		file, err := os.Open(pathToVideo)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer file.Close()

		name := fmt.Sprintf("%s.mp4", downloadId)
		part, err := writer.CreateFormFile(downloadId, name)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		io.Copy(part, file)
	}

	contentType := writer.FormDataContentType()
	return &buffer, &contentType, nil
}
