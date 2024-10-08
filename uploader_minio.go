package file_storage

import (
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/qiuyier/file-storage/pkg/util"
	"mime/multipart"
)

type UploaderMinioConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	EndPoint        string
	BucketName      string
	Path            string
	UseSSL          bool
	Domain          string
}

type UploaderMinio struct {
	client     *minio.Client
	bucketName string
	path       string
	domain     string
}

func NewUploaderMinio(config UploaderMinioConfig) (uploader *UploaderMinio, err error) {
	client, err := minio.New(config.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	uploader = &UploaderMinio{
		client:     client,
		bucketName: config.BucketName,
		path:       config.Path,
		domain:     config.Domain,
	}
	return
}

func (u *UploaderMinio) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
	if err = s3utils.CheckValidBucketName(u.bucketName); err != nil {
		return "", "", err
	}

	path = util.GenName(u.path, file.Filename, randomly)

	if err = s3utils.CheckValidObjectName(path); err != nil {
		return "", "", err
	}

	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	contentType := util.GetContentType(util.Ext(file.Filename))
	fileUrl = util.Join(u.domain, u.bucketName, path)

	_, err = u.client.PutObject(ctx, u.bucketName, path, fd, file.Size, minio.PutObjectOptions{ContentType: contentType})

	return
}

func (u *UploaderMinio) GetUploaderType() string {
	return Minio
}

func (u *UploaderMinio) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error) {
	return "", "", errors.New("minio driver does not support multipart upload")
}

func (u *UploaderMinio) DeleteObjects(ctx context.Context, path []string) error {
	var err error

	for _, v := range path {
		err = u.client.RemoveObject(ctx, u.bucketName, v, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
