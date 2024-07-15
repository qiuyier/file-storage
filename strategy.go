package file_storage

import (
	ALiYun2 "github.com/qiuyier/file-storage/ALiYun"
	Local2 "github.com/qiuyier/file-storage/Local"
	Minio2 "github.com/qiuyier/file-storage/Minio"
)

type StrategyFunc func() IUpload

var strategies = map[DriverType]StrategyFunc{
	Local: func() IUpload {
		return &Local2.UploaderLocal{}
	},
	ALiYun: func() IUpload {
		return &ALiYun2.UploaderOss{}
	},
	Minio: func() IUpload {
		return &Minio2.UploaderMinio{}
	},
}
