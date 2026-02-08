package application

import "fmt"

type VideoTypeError struct {
	receivedType string
}

type YtDlpError struct {
	stderr    string
	exitError string
}

type HandlerError struct {
	err string
}

func (err VideoTypeError) Error() string {
	return fmt.Sprintf("Received type %s not found", err.receivedType)
}

func (err YtDlpError) Error() string {
	return fmt.Sprintf("Error while executing yt-dlp: %s\nExit error: %s", err.stderr, err.exitError)
}

func (h HandlerError) Error() string {
	return h.err
}
