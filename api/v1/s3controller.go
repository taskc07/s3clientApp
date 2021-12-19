package v1

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/khan1507017/s3clientApp/config"
	"github.com/khan1507017/s3clientApp/dto"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
	"sync"
)

type S3controllerInf interface {
	CreateFiles(e echo.Context) error
	CreateDirs(e echo.Context) error
}

type S3controllerInstance struct {
}

func (s S3controllerInstance) CreateFiles(e echo.Context) error {
	var input dto.CreateFilesDto
	err := e.Bind(&input)
	if err != nil {
		log.Printf("invalid input payload: " + err.Error())
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	if input.AccessKey == "" || input.SecretKey == "" {
		return e.JSON(http.StatusBadRequest, "access key and/or secret key cannot be null")
	}

	creds := credentials.NewStaticCredentials(input.AccessKey, input.SecretKey, "")
	pathStyle := true
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Credentials:      creds,
		Endpoint:         &input.EndPoint,
		S3ForcePathStyle: &pathStyle,
	})
	if err != nil {
		return err
	}
	var wg = &sync.WaitGroup{}
	for i := 0; i < input.Instance; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, count int) {
			defer wg.Done()
			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:        aws.String(input.BucketName),
				Key:           aws.String(input.ParentDir + input.KeyPrefix + "(" + strconv.Itoa(count) + ")" + input.KeySuffix),
				ACL:           aws.String("private"),
				Body:          bytes.NewReader(config.PdfBytes),
				ContentLength: aws.Int64(int64(len(config.PdfBytes))),
				ContentType:   aws.String(http.DetectContentType(config.PdfBytes)),
			})

			if err != nil {
				log.Printf("goroutine error: " + err.Error())
			}
		}(wg, i)
	}

	if err != nil {
		log.Printf("invalid input payload: " + err.Error())
		return e.JSON(http.StatusBadRequest, err.Error())
	}
	wg.Wait()
	return e.JSON(http.StatusOK, nil)
}

func (s S3controllerInstance) CreateDirs(e echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func S3Controller() S3controllerInf {
	return new(S3controllerInstance)
}
