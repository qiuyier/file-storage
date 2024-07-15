package file_storage

type DriverType string

const (
	Local  DriverType = "Local"
	ALiYun DriverType = "ALiYun"
	Minio  DriverType = "Minio"
)
