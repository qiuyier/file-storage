package file_storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func TestUpload(t *testing.T) {
	//localUploader, _ := NewUploaderLocal(UploaderLocalConfig{
	//	LocalPath: "attachment/",
	//	Domain:    "http://localhost/",
	//})

	minioUploader, err := NewUploaderMinio(UploaderMinioConfig{
		AccessKeyID:     "0tuG0FGHGjCCHLzGge5O",
		SecretAccessKey: "ffx1ipSKMH9W1adY8dTwsDolAD7xhV8YFQJEFlmE",
		EndPoint:        "127.0.0.1:9000",
		BucketName:      "e-code",
		Path:            "test",
		UseSSL:          false,
		Domain:          "http://127.0.0.1:9000",
	})

	if err != nil {
		fmt.Println(err)
	}

	uploader := NewFileUploader().SetLogName("upload").SetOutputPath(fmt.Sprintf("./%s.log", "upload")).RegisterUploader(minioUploader)

	fileHeader := createMultipartFileHeader("default.png")

	res, err := uploader.Upload(context.TODO(), fileHeader, true)
	if err != nil {
		uploader.logger.Errorf("upload err: %v", err)
		return
	}
	uploader.logger.Infof("upload res: %v", res)
}

func createMultipartFileHeader(filePath string) *multipart.FileHeader {
	// open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	// create a buffer to hold the file in memory
	var buff bytes.Buffer
	buffWriter := io.Writer(&buff)

	// create a new form and create a new file field
	formWriter := multipart.NewWriter(buffWriter)
	formPart, err := formWriter.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// copy the content of the file to the form's file field
	if _, err := io.Copy(formPart, file); err != nil {
		fmt.Println(err)
		return nil
	}

	// close the form writer after the copying process is finished
	// I don't use defer in here to avoid unexpected EOF error
	formWriter.Close()

	// transform the bytes buffer into a form reader
	buffReader := bytes.NewReader(buff.Bytes())
	formReader := multipart.NewReader(buffReader, formWriter.Boundary())

	// read the form components with max stored memory of 1MB
	multipartForm, err := formReader.ReadForm(1 << 20)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// return the multipart file header
	files, exists := multipartForm.File["file"]
	if !exists || len(files) == 0 {
		fmt.Println("multipart file not exists")
		return nil
	}

	return files[0]
}
