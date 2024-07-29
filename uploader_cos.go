package file_storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/qiuyier/file-storage/pkg/util"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
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

	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	_, err = u.client.Object.Put(ctx, path, fd, nil)
	if err != nil {
		return "", "", err
	}
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderCos) GetUploaderType() string {
	return Tencent
}

func (u *UploaderCos) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error) {
	// 上传路径
	name := util.GenName(file.Filename, randomly)
	nowDate := time.Now().Format(time.DateOnly)
	path = util.Join(u.path, nowDate, name)

	v, _, err := u.client.Object.InitiateMultipartUpload(ctx, path, nil)
	if err != nil {
		return "", "", err
	}

	// 获取文件信息
	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	uploadId := v.UploadID

	// 计算分块大小和分块数量
	if chunkSize < 1 {
		return "", "", errors.New("chunk size must be greater than 1")
	}
	chunkSize = chunkSize * 1024 * 1024
	totalChunks := (file.Size + int64(chunkSize) - 1) / int64(chunkSize)

	// 分块上传
	opt := &cos.CompleteMultipartUploadOptions{}
	for i := int64(1); i <= totalChunks; i++ {
		partNumber := int(i)
		offset := (i - 1) * int64(chunkSize)
		size := int64(chunkSize)
		if i == totalChunks {
			size = file.Size - offset
		}

		buf := make([]byte, size)
		_, err := fd.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return "", "", errors.New("Error reading file chunk: " + err.Error())
		}

		resp, err := u.client.Object.UploadPart(ctx, path, uploadId, partNumber, strings.NewReader(string(buf)), nil)
		if err != nil {
			return "", "", errors.New("Error uploading part:" + err.Error())
		}

		PartETag := resp.Header.Get("ETag")
		opt.Parts = append(opt.Parts, cos.Object{
			PartNumber: partNumber, ETag: PartETag},
		)

	}

	// 完成分片上传
	_, _, err = u.client.Object.CompleteMultipartUpload(
		ctx, path, uploadId, opt,
	)

	if err != nil {
		return "", "", errors.New("Error completing multipart upload: " + err.Error())
	}

	fileUrl = util.Join(u.domain, path)

	return
}
