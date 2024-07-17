package file_storage

import "errors"

var (
	NotSupportedFileUploader = errors.New("file uploader not supported")
	NilFileUploader          = errors.New("file uploader is not exists")
	NotDirErr                = errors.New(`"dirPath\" should be a directory path`)
	CreateDirErr             = errors.New("create directory fail")
)
