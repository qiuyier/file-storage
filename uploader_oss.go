package file_storage

import (
	"context"
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/qiuyier/file-storage/pkg/util"
	"mime/multipart"
	"time"
)

type UploaderOssConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	EndPoint        string
	BucketName      string
	Path            string
	Domain          string
}

type UploaderOss struct {
	bucket *oss.Bucket
	path   string
	domain string
}

func NewUploaderOss(config UploaderOssConfig) (uploader *UploaderOss, err error) {
	client, err := oss.New(config.EndPoint, config.AccessKeyID, config.SecretAccessKey)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return nil, err
	}

	uploader = &UploaderOss{
		bucket: bucket,
		path:   config.Path,
		domain: config.Domain,
	}

	return
}

func (u *UploaderOss) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
	name := util.GenName(file.Filename, randomly)
	nowDate := time.Now().Format(time.DateOnly)
	path = util.Join(u.path, nowDate, name)

	if err = s3utils.CheckValidObjectName(path); err != nil {
		return "", "", err
	}

	f, err := file.Open()
	defer f.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	err = u.bucket.PutObject(path, f)
	if err != nil {
		return "", "", err
	}
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderOss) GetUploaderType() string {
	return AliYun
}
