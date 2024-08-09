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
	path = util.GenName(u.path, file.Filename, randomly)

	if err = s3utils.CheckValidObjectName(path); err != nil {
		return "", "", err
	}

	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	err = u.bucket.PutObject(path, fd)
	if err != nil {
		return "", "", err
	}
	fileUrl = util.Join(u.domain, path)

	return
}

func (u *UploaderOss) GetUploaderType() string {
	return AliYun
}

func (u *UploaderOss) MultipartUpload(ctx context.Context, file *multipart.FileHeader, randomly bool, chunkSize int) (path, fileUrl string, err error) {
	// 上传路径
	path = util.GenName(u.path, file.Filename, randomly)

	// 指定过期时间。
	expires := time.Now().Add(time.Minute * 3)
	// 如果需要在初始化分片时设置请求头，请参考以下示例代码。
	options := []oss.Option{
		oss.MetadataDirective(oss.MetaReplace),
		oss.Expires(expires),
	}

	// 初始化一个分片上传事件。
	v, err := u.bucket.InitiateMultipartUpload(path, options...)
	if err != nil {
		return "", "", err
	}

	// 获取文件信息
	fd, err := file.Open()
	defer fd.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	chunkSize = chunkSize * 1024 * 1024
	chunks, err := util.SplitFileByPartSize(fd, file.Size, int64(chunkSize))
	if err != nil {
		return "", "", err
	}

	// 上传分片。
	var parts []oss.UploadPart
	for _, chunk := range chunks {
		part, err := u.bucket.UploadPart(v, chunk.Buf, chunk.Size, chunk.Number)
		if err != nil {
			_ = u.bucket.AbortMultipartUpload(v)
			return "", "", err
		}
		parts = append(parts, part)

	}

	// 步骤3：完成分片上传。
	_, _ = u.bucket.CompleteMultipartUpload(v, parts)
	fileUrl = util.Join(u.domain, path)

	return
}
