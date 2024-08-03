package util

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	DefaultTrimChars = string([]byte{
		'\t', // Tab.
		'\v', // Vertical tab.
		'\n', // New line (line feed).
		'\r', // Carriage return.
		'\f', // New page.
		' ',  // Ordinary space.
		0x00, // NUL-byte.
		0x85, // Delete.
		0xA0, // Non-breaking space.
	})
)

type FileChunk struct {
	Number int   // Chunk number
	Offset int64 // Chunk offset
	Size   int64 // Chunk size.
	Buf    *strings.Reader
}

// RandomlyName 生成随机字符串
func RandomlyName(length int) string {
	if length <= 0 {
		return ""
	}

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func TrimRight(str string, characterMask ...string) string {
	trimChars := DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.TrimRight(str, trimChars)
}

func FileSize(data int64) string {
	var factor float64 = 1024
	res := float64(data)
	for _, unit := range []string{"", "K", "M", "G", "T", "P"} {
		if res < factor {
			return fmt.Sprintf("%.2f%sB", res, unit)
		}
		res /= factor
	}
	return fmt.Sprintf("%.2f%sB", res, "P")
}

func Ext(path string) string {
	ext := filepath.Ext(path)
	if p := strings.IndexByte(ext, '?'); p != -1 {
		ext = ext[0:p]
	}
	return ext
}

func Join(paths ...string) string {
	var (
		s         string
		Separator = string(os.PathSeparator)
	)
	for _, path := range paths {
		if s != "" {
			s += Separator
		}
		s += TrimRight(path, Separator)
	}

	return s
}

func GenName(path, fileName string, randomly bool) string {
	name := filepath.Base(fileName)

	// 如果设置随机名，则重新命名
	if randomly {
		random := RandomlyName(6)
		name = strings.ToLower(strconv.FormatInt(time.Now().UnixNano(), 36) + random)
		name = fmt.Sprintf("%s%s", name, Ext(fileName))
	}

	nowDate := time.Now().Format(time.DateOnly)

	return Join(path, nowDate, name)
}

func GetContentType(ext string) string {
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	} else {
		return "application/octet-stream"
	}
}

// SplitFileByPartSize 来自oss SplitFileByPartSize，修改用于文件流分片
func SplitFileByPartSize(fd multipart.File, fileSize, chunkSize int64) ([]FileChunk, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize invalid")
	}

	var chunkN = fileSize / chunkSize
	if chunkN >= 10000 {
		return nil, errors.New("too many parts, please increase part size")
	}

	var chunks []FileChunk
	var chunk = FileChunk{}
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * chunkSize
		chunk.Size = chunkSize

		buf := make([]byte, chunk.Size)
		_, err := fd.ReadAt(buf, chunk.Offset)
		if err != nil && err != io.EOF {
			return nil, errors.New("Error reading file chunk: " + err.Error())
		}
		chunk.Buf = strings.NewReader(string(buf))

		chunks = append(chunks, chunk)
	}

	if fileSize%chunkSize > 0 {
		chunk.Number = len(chunks) + 1
		chunk.Offset = int64(len(chunks)) * chunkSize
		chunk.Size = fileSize % chunkSize

		buf := make([]byte, chunk.Size)
		_, err := fd.ReadAt(buf, chunk.Offset)
		if err != nil && err != io.EOF {
			return nil, errors.New("Error reading file chunk: " + err.Error())
		}
		chunk.Buf = strings.NewReader(string(buf))

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
