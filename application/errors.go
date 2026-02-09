package application

import (
	"fmt"

	"github.com/herobeniyoutube/vk-forwarder/domain/statuses"
)

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

type ProcessStatusRestrictions struct {
	Status statuses.IdempotencyStatus
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

func (h ProcessStatusRestrictions) Error() string {
	switch h.Status {
	case statuses.Done:
		return "Process already done"
	case statuses.Processing:
		return "This request is being processed"
	case statuses.Error:
		return "This request can be reprocessed due to previous error"
	default:
		return "unknown status"
	}
}
