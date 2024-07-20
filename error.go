package file_storage

import "errors"

var (
	NotDirErr = errors.New(`"dirPath\" should be a directory path`)
)
