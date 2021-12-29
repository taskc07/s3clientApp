package dto

type CreateDirsDto struct {
	BucketName         string `json:"bucketName"`
	EndPoint           string `json:"endPoint"`
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	ParentDir          string `json:"parentDir"`
	Instance           int    `json:"instance"`
	DirPrefix          string `json:"dirPrefix"`
	DirSuffix          string `json:"dirSuffix"`
	Verbose            string `json:"verbose"`
	EnableCommonClient string `json:"enableCommonClient"`
}
