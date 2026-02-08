package application

type ISender interface {
	SendClip(downloadId string, caption *string) error
	SendPhoto(url string, caption *string) error
	SendBatch(locations map[string]string, downloadIds []string, caption Caption) error
}

type ICallbackCodeGetter interface {
	GetCallbackConfirmation() (*string, error)
}

type IVideoDownloader interface {
	Download(groupId int64, videoType string, videoId int, ownerId int) (*string, error)
	DisposeSendedVideo(groupId int64, downloadId string) error
}

type IHandler interface {
	Handle(event MessageNewEvent, retryCount int, ignoreIdempotencyKey bool) (*string, error)
}
