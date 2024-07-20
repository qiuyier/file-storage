package file_storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/qiuyier/file-storage/pkg/util"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type UploaderLocalConfig struct {
	LocalPath string
	Domain    string
}

type UploaderLocal struct {
	localPath string
	domain    string
}

func NewUploaderLocal(config UploaderLocalConfig) (uploader *UploaderLocal, err error) {
	uploader = &UploaderLocal{
		localPath: util.TrimRight(config.LocalPath, string(os.PathSeparator)),
		domain:    config.Domain,
	}
	return
}

func (u *UploaderLocal) Upload(ctx context.Context, file *multipart.FileHeader, randomly bool) (path, fileUrl string, err error) {
	nowDate := time.Now().Format(time.DateOnly)

	// 文件保存路径
	dirPath := fmt.Sprintf("%s/%s", u.localPath, nowDate)

	// 判断路径是否存在且为文件夹
	if !exists(dirPath) {
		// 不存在则创建文件夹
		if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
			err = errors.New("create dir " + dirPath + ", err: " + err.Error())
			return "", "", err
		}
	} else if !isDir(dirPath) {
		// 路径存在但不为文件夹时
		return "", "", NotDirErr
	}

	f, err := file.Open()
	defer f.Close()

	if err != nil {
		return "", "", errors.New("open file " + file.Filename + ", err: " + err.Error())
	}

	name := filepath.Base(file.Filename)

	// 如果设置随机名，则重新命名
	if randomly {
		random := util.RandomlyName(6)
		name = strings.ToLower(strconv.FormatInt(time.Now().UnixNano(), 36) + random)
		name = fmt.Sprintf("%s%s", name, util.Ext(file.Filename))
	}

	filePath := util.Join(dirPath, name)
	newFile, err := create(filePath)
	if err != nil {
		return "", "", err
	}
	defer newFile.Close()

	if _, err = io.Copy(newFile, f); err != nil {
		err = errors.New("copy file " + filePath + "err: " + err.Error())
		return "", "", err
	}

	return filePath, util.Join(u.domain, filePath), nil
}

func (u *UploaderLocal) GetUploaderType() string {
	return Local
}

// 代码源于 gf 框架
func exists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

func mkdir(path string) (err error) {
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		err = errors.New("mkdir " + path + "err: " + err.Error())
		return err
	}
	return nil
}

func create(path string) (*os.File, error) {
	dir := dir(path)
	if !exists(dir) {
		if err := mkdir(dir); err != nil {
			return nil, err
		}
	}
	file, err := os.Create(path)
	if err != nil {
		err = errors.New("create file " + path + ", err: " + err.Error())
		return nil, err
	}

	return file, nil
}

func dir(path string) string {
	if path == "." {
		return filepath.Dir(realPath(path))
	}
	return filepath.Dir(path)
}

func realPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !exists(p) {
		return ""
	}
	return p
}
