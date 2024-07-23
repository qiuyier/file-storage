package file_storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/qiuyier/file-storage/pkg/util"
	"github.com/tencentyun/cos-go-sdk-v5"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type UploaderCosConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	EndPoint        string
	BucketName      string
	Path            string
	Domain          string
	Region          string
}

type UploaderCos struct {
	client *cos.Client
	path   string
	domain string
}

func NewUploaderCos(config UploaderCosConfig) (uploader *UploaderCos, err error) {
	u, _ := url.Parse(config.EndPoint)

	su, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", config.Region))

	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKeyID,
			SecretKey: config.SecretAccessKey,
		},
	})

	uploader = &UploaderCos{
		client: client,
		path:   config.Path,
		domain: config.Domain,
	}

	return
}

func (u *UploaderCos) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
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

	_, err = u.client.Object.Put(ctx, path, f, nil)
	if err != nil {
		return "", "", err
	}
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderCos) GetUploaderType() string {
	return Tencent
}
