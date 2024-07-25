package file_storage

import (
	"context"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/qiuyier/file-storage/pkg/util"
	"mime/multipart"
	"time"
)

type UploaderObsConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	EndPoint        string
	BucketName      string
	Path            string
	Domain          string
}

type UploaderObs struct {
	client *obs.ObsClient
	path   string
	domain string
	bucket string
}

func NewUploaderObs(config UploaderObsConfig) (uploader *UploaderObs, err error) {
	obsClient, err := obs.New(config.AccessKeyID, config.SecretAccessKey, config.EndPoint, obs.WithSignature(obs.SignatureObs))
	if err != nil {
		return nil, err
	}

	uploader = &UploaderObs{
		client: obsClient,
		path:   config.Path,
		domain: config.Domain,
		bucket: config.BucketName,
	}

	return
}

func (u *UploaderObs) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
	name := util.GenName(file.Filename, randomly)
	nowDate := time.Now().Format(time.DateOnly)
	path = util.Join(u.path, nowDate, name)

	fd, err := file.Open()
	defer fd.Close()

	input := &obs.PutObjectInput{}

	input.Bucket = u.bucket

	input.Key = path

	input.Body = fd

	_, err = u.client.PutObject(input)
	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderObs) GetUploaderType() string {
	return HuaWei
}
