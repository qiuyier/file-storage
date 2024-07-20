package file_storage

import (
	"context"
	"github.com/qiuyier/file-storage/pkg/log"
	"github.com/qiuyier/file-storage/pkg/util"
	"go.uber.org/zap/zapcore"
	"mime/multipart"
)

type Uploader struct {
	uploader IUpload
	logger   *log.Logger
}

type UploadResult struct {
	Driver   string
	FileName string
	Path     string
	Size     string
	FileUrl  string
	Ext      string
}

type IUpload interface {
	Upload(ctx context.Context, file *multipart.FileHeader, randomName bool) (path, fileUrl string, err error)
	GetUploaderType() string
}

func NewFileUploader() *Uploader {
	// 注册日志
	logger := log.NewLogger()

	return &Uploader{
		logger: logger,
	}
}

func (u *Uploader) Upload(ctx context.Context, file *multipart.FileHeader, randomName bool) (res UploadResult, err error) {
	path, fileUrl, err := u.uploader.Upload(ctx, file, randomName)
	if err != nil {
		u.logger.Errorf("upload err: %v", err)
	}

	res = UploadResult{
		Driver:   u.uploader.GetUploaderType(),
		FileName: file.Filename,
		Path:     path,
		Size:     util.FileSize(file.Size),
		FileUrl:  fileUrl,
		Ext:      util.Ext(file.Filename),
	}

	return
}

func (u *Uploader) RegisterUploader(uploader IUpload) *Uploader {
	u.uploader = uploader
	return u
}

func (u *Uploader) SetLogName(appName string) *Uploader {
	u.logger.SetLogName(appName)
	return u
}

func (u *Uploader) SetLevel(level zapcore.Level) *Uploader {
	u.logger.SetLevel(level)
	return u
}

func (u *Uploader) SetOutputPath(path string) *Uploader {
	u.logger.SetOutputPath(path)
	return u
}
