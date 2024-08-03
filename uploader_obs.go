package file_storage

import (
	"context"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/qiuyier/file-storage/pkg/util"
	"mime/multipart"
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
	path = util.GenName(u.path, file.Filename, randomly)

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

func (u *UploaderObs) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error) {
	// 上传路径
	path = util.GenName(u.path, file.Filename, randomly)

	inputInit := &obs.InitiateMultipartUploadInput{}
	// 指定存储桶名称
	inputInit.Bucket = u.bucket
	// 指定对象名
	inputInit.Key = path
	// 初始化上传段任务
	outputInit, err := u.client.InitiateMultipartUpload(inputInit)
	if err != nil {
		return "", "", errors.New("init multipart upload err: " + err.Error())
	}

	// 获取文件信息
	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	uploadId := outputInit.UploadId

	// 计算分块大小和分块数量
	chunkSize = chunkSize * 1024 * 1024
	chunks, err := util.SplitFileByPartSize(fd, file.Size, int64(chunkSize))
	if err != nil {
		return "", "", err
	}

	var opt []obs.Part
	for _, chunk := range chunks {
		inputUploadPart := &obs.UploadPartInput{}
		inputUploadPart.Bucket = u.bucket
		inputUploadPart.Key = path
		inputUploadPart.UploadId = uploadId
		inputUploadPart.PartNumber = chunk.Number

		inputUploadPart.Body = chunk.Buf
		outputUploadPart, err := u.client.UploadPart(inputUploadPart)
		if err != nil {
			abortInput := &obs.AbortMultipartUploadInput{}
			// 指定存储桶名称
			abortInput.Bucket = u.bucket
			// 指定上传对象名
			abortInput.Key = path
			// 指定多段上传任务号
			abortInput.UploadId = uploadId
			// 取消分段上传任务
			_, _ = u.client.AbortMultipartUpload(abortInput)
			return "", "", err
		}

		PartETag := outputUploadPart.ETag
		opt = append(opt, obs.Part{PartNumber: chunk.Number, ETag: PartETag})
	}

	// 上传完成
	inputCompleteMultipart := &obs.CompleteMultipartUploadInput{}
	inputCompleteMultipart.Bucket = u.bucket
	inputCompleteMultipart.Key = path
	inputCompleteMultipart.UploadId = uploadId
	inputCompleteMultipart.Parts = opt
	_, _ = u.client.CompleteMultipartUpload(inputCompleteMultipart)

	fileUrl = util.Join(u.domain, path)

	return
}
