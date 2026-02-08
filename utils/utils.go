package utils

import "fmt"

func GetVideoPath(groupId int64, downloadId string) string {
	return fmt.Sprintf("./tmp/vk-forwarder/%d/%s.mp4", groupId, downloadId)
}

func BuildIdempotencyKey(chatId int64, messageId int64) string {
	return fmt.Sprintf("%d:%d", chatId, messageId)
}
