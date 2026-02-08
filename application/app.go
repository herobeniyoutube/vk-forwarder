package application

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	u "github.com/herobeniyoutube/vk-forwarder/utils"
)

type VideoDownloader struct{}

func NewVideoDownloader() *VideoDownloader {
	return &VideoDownloader{}
}

func (d *VideoDownloader) Download(groupId int64, videoType string, videoId int, ownerId int) (*string, error) {
	var err error

	downloadLink, err := createLink(videoType, videoId, ownerId)

	if err != nil {
		log.Printf("Wrong video type: %s", videoType)
		return nil, err
	}

	downloadId := uuid.NewString()
	filePath := u.GetVideoPath(groupId, downloadId)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		log.Printf("Couldn't create tmp dir: %s", err.Error())
		return nil, err
	}
	formatFlag := "worstvideo[height>=480]+worstaudio/worst[height>=480]"

	cmd := exec.Command(
		"yt-dlp",
		"-f", formatFlag,
		"--merge-output-format", "mp4",
		"-o", filePath, *downloadLink)

	var bufferOut, bufferErr bytes.Buffer
	cmd.Stdout = &bufferOut
	cmd.Stderr = &bufferErr

	err = cmd.Run()
	if err != nil {
		err = YtDlpError{bufferErr.String(), err.Error()}
		log.Println(err.Error())
		return nil, err
	}

	//log.Printf("File downloaded: %s", bufferOut.String())
	log.Printf("Downloaded video %s from groupId %d", downloadId, groupId)

	return &downloadId, nil
}

func (d *VideoDownloader) DisposeSendedVideo(groupId int64, downloadId string) error {
	err := os.Remove(u.GetVideoPath(groupId, downloadId))
	if err != nil {
		return err
	}
	return nil
}

func createLink(videoType string, ownerId int, videoId int) (*string, error) {
	var factualType string

	switch videoType {
	case "video":
		factualType = "video"
	case "short_video":
		factualType = "clip"
	default:
		return nil, VideoTypeError{videoType}
	}

	result := fmt.Sprintf("https://vk.com/%s%d_%d", factualType, videoId, ownerId)

	return &result, nil
}
