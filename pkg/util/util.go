package util

import (
	"fmt"
	"math/rand"
	"mime"
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

func GenName(fileName string, randomly bool) string {
	name := filepath.Base(fileName)

	// 如果设置随机名，则重新命名
	if randomly {
		random := RandomlyName(6)
		name = strings.ToLower(strconv.FormatInt(time.Now().UnixNano(), 36) + random)
		name = fmt.Sprintf("%s%s", name, Ext(fileName))
	}
	return name
}

func GetContentType(ext string) string {
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	} else {
		return "application/octet-stream"
	}
}
