package dto

type CreateFilesDto struct {
	BucketName string `json:"bucketName"`
	EndPoint   string `json:"endPoint"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	ParentDir  string `json:"parentDir"`
	Instance   int    `json:"instance"`
	KeyPrefix  string `json:"KeyPrefix"`
	KeySuffix  string `json:"keySuffix"`
	NullData   string `json:"nullData"`
}
