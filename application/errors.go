package application

import "fmt"

type VideoTypeError struct {
	ReceivedType string
}

type DownloaderError struct {
	Stderr    string
	ExitError string
}

type HandlerError struct {
	err string
}

func (err VideoTypeError) Error() string {
	return fmt.Sprintf("Received type %s not found", err.ReceivedType)
}

func (err DownloaderError) Error() string {
	return fmt.Sprintf("Error while executing yt-dlp: %s\nExit error: %s", err.Stderr, err.ExitError)
}

func (h HandlerError) Error() string {
	return h.err
}
