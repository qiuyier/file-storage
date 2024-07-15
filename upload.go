package file_storage

import (
	"context"
	"mime/multipart"
	"sync"
)

var (
	mutex               sync.RWMutex
	fileUploaderFactory map[DriverType]IUpload
)

type IUpload interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (path string, err error)
}

func GetFileUploaderFactory(uploaderType DriverType) (IUpload, error) {
	if uploader, ok := fileUploaderFactory[uploaderType]; ok {
		return uploader, nil
	} else {
		uploader, err := newFileUploaderFactory(uploaderType)
		if err != nil {
			return nil, err
		}
		return uploader, nil
	}
}

func newFileUploaderFactory(uploaderType DriverType) (IUpload, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var uploader IUpload
	if strategy, ok := strategies[uploaderType]; ok {
		uploader = strategy()
		fileUploaderFactory[uploaderType] = uploader

		return uploader, nil
	}

	return nil, NilFileUploader
}
