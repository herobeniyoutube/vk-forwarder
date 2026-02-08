package infrastructure

import (
	"encoding/json"
	"github.com/herobeniyoutube/vk-forwarder/config"
	"io"
	"log"
	"net/http"
)

type VkSerivce struct {
	callbackCode *string
	token        string
	vkVersion    string
	vkGroupId    string
}

type CodeResponse struct {
	Response Code `json:"response"`
}

type Code struct {
	Code string `json:"code"`
}

func (s *VkSerivce) GetCallbackConfirmation() (*string, error) {
	if s.callbackCode != nil {
		return s.callbackCode, nil
	}

	t := "Bearer " + s.token

	var err error

	req, err := http.NewRequest("POST", "https://api.vk.com/method/groups.getCallbackConfirmationCode", nil)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Add("Authorization", t)
	q := req.URL.Query()

	q.Add("v", s.vkVersion)
	q.Add("group_id", s.vkGroupId)

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer r.Body.Close()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	log.Print(string(raw))

	response := &CodeResponse{}
	err = json.Unmarshal(raw, response)
	if err != nil {
		log.Println("error getting callback api code" + err.Error())
		return nil, err
	}

	return &response.Response.Code, nil
}

func NewVkService(config config.Config) *VkSerivce {
	s := &VkSerivce{nil, config.VkGroupToken, config.VkVersion, config.VkGroupID}

	s.callbackCode, _ = s.GetCallbackConfirmation()

	return s
}
