package file_storage

import (
	"context"
	"errors"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/qiuyier/file-storage/pkg/util"
	"mime/multipart"
)

type UploaderQiNiuConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Path            string
	Domain          string
	UseSSL          bool
	UseCdn          bool
}

type UploaderQiNiu struct {
	client    *storage.ResumeUploaderV2
	putPolicy storage.PutPolicy
	mac       *auth.Credentials
	bucket    string
	path      string
	domain    string
}

func NewUploaderQiNiu(config UploaderQiNiuConfig) (uploader *UploaderQiNiu, err error) {
	cfg := storage.Config{
		UseHTTPS:      config.UseSSL,
		UseCdnDomains: config.UseCdn,
	}

	// 空间对应的机房
	cfg.Region, err = storage.GetRegion(config.AccessKeyID, config.BucketName)
	if err != nil {
		return
	}

	client := storage.NewResumeUploaderV2(&cfg)

	uploader = &UploaderQiNiu{
		client: client,
		putPolicy: storage.PutPolicy{
			Scope: config.BucketName,
		},
		bucket: config.BucketName,
		mac:    auth.New(config.AccessKeyID, config.SecretAccessKey),
		path:   config.Path,
		domain: config.Domain,
	}

	return
}

func (u *UploaderQiNiu) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
	path = util.GenName(u.path, file.Filename, randomly)

	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	upToken := u.putPolicy.UploadToken(u.mac)

	err = u.client.Put(ctx, storage.PutRet{}, upToken, path, fd, file.Size, &storage.RputV2Extra{})
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderQiNiu) GetUploaderType() string {
	return QiNiu
}

func (u *UploaderQiNiu) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error) {
	path = util.GenName(u.path, file.Filename, randomly)

	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	upToken := u.putPolicy.UploadToken(u.mac)

	err = u.client.Put(ctx, storage.PutRet{}, upToken, path, fd, file.Size, &storage.RputV2Extra{
		PartSize: int64(chunkSize * 1024 * 1024),
	})
	fileUrl = util.Join(u.domain, path)

	return
}
