package application

type MessageNewEvent struct {
	GroupID int64  `json:"group_id"`
	Type    string `json:"type"`
	EventID string `json:"event_id"`
	V       string `json:"v"`
	Object  struct {
		Message Message `json:"message"`
	} `json:"object"`
}

type Video struct {
	VideoId int    `json:"id"`
	OwnerId int    `json:"owner_id"`
	Type    string `json:"type"`
}

type Photo struct {
	OrigPhoto struct {
		Url string `json:"url"`
	} `json:"orig_photo"`
}

type Wall struct {
	From struct {
		Name string `json:"name"`
	} `json:"from"`
	Attachments []Attachment `json:"attachments"`
	Text        string       `json:"text"`
}

type Attachment struct {
	Type  string `json:"type"`
	Video Video  `json:"video"`
	Photo Photo  `json:"photo"`
	Wall  Wall   `json:"wall"`
}

type Message struct {
	Attachments []Attachment `json:"attachments"`
	MessageId   int64        `json:"conversation_message_id"`
	PeerId      int64        `json:"peer_id"`
	Text        string       `json:"text"`
}

type Caption struct {
	UserMessage string
	GroupName   string
	Text        string
}
