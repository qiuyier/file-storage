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
	Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error)
	// MultipartUpload
	//chunkSize 单位byte
	MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error)
	GetUploaderType() string
	DeleteObjects(ctx context.Context, path []string) error
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

func (u *Uploader) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomName bool, chunkSize int) (res UploadResult, err error) {
	path, fileUrl, err := u.uploader.MultipartUpload(ctx, file, randomName, chunkSize)
	if err != nil {
		u.logger.Errorf("multipart upload err: %v", err)
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

func (u *Uploader) DeleteObjects(ctx context.Context, path []string) error {
	err := u.uploader.DeleteObjects(ctx, path)
	if err != nil {
		u.logger.Errorf("delete err: %v", err)
	}

	return err
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
