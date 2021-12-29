package helper

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/khan1507017/s3clientApp/config"
	"github.com/khan1507017/s3clientApp/constant"
	"github.com/khan1507017/s3clientApp/dto"
	"github.com/khan1507017/s3clientApp/response"
	"github.com/labstack/echo/v4"
	"github.com/schollz/progressbar/v3"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type cdhs struct {
	input      dto.CreateDirsDto
	bucketName string
	endPoint   string
	accessKey  string
	secretKey  string
	parentDir  string
	instance   int
	dirPrefix  string
	dirSuffix  string
	sess       *session.Session
	mtx        *sync.Mutex
	waitGrout  *sync.WaitGroup
}

type createDirsHelper struct {
}

var singletonCdhs *createDirsHelper
var onceCdhs sync.Once

func CreateDirs() *createDirsHelper {
	onceCdhs.Do(func() {
		singletonCdhs = new(createDirsHelper)
	})
	return singletonCdhs
}

func (t *createDirsHelper) Execute(e echo.Context) error {

	this := new(cdhs)
	this.mtx = &sync.Mutex{}
	this.waitGrout = &sync.WaitGroup{}

	log.Println("[INFO]Initializing Properties...")
	err := t.init(e, this)
	if err != nil {
		return err
	}

	log.Println("[INFO]Validating Inputs...")
	err = t.checkInput(e, this)
	if err != nil {
		return err
	}

	log.Println("[INFO]Connecting To S3...")
	err = t.connectingS3(e, this)
	if err != nil {
		return err
	}

	log.Println("[INFO]Performing Execution...")
	err = t.doPerform(e, this)
	if err != nil {
		log.Println("Error Occurred. Performing a Rollback...")
		_ = t.onRollback(e, this)
		return err
	}

	response.Helper().SuccessResponse(e, http.StatusOK, constant.STATUS_SUCCESS, "request on process")
	return nil
}

func (createDirsHelper) init(e echo.Context, this *cdhs) error {
	var temp = new(dto.CreateDirsDto)
	err := e.Bind(temp)
	if err != nil {
		log.Println("[ERROR] ", err.Error())
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "", err.Error())
		return err
	} else {
		this.input = *temp
	}
	return nil
}
func (createDirsHelper) checkInput(e echo.Context, this *cdhs) error {

	//checking accessKey
	if this.input.AccessKey == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.EMPTY_ACCESS_KEY, "access key cannot be empty", nil)
		return errors.New("empty access key")
	}
	this.accessKey = this.input.AccessKey

	//checking secretKey
	if this.input.SecretKey == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.EMPTY_SECRET, "secret key cannot be empty", nil)
		return errors.New("empty secret key")
	}
	this.secretKey = this.input.SecretKey

	//validating instanceCount
	if this.input.Instance < 1 || this.input.Instance > 500000 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "instance should be between 1~500000", nil)
		return errors.New("invalid instance")
	}
	this.instance = this.input.Instance

	//checking bucketName
	if this.input.BucketName == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_BUCKET, "bucketName cannot be empty", nil)
		return errors.New("empty bucketName")
	}
	this.bucketName = this.input.BucketName

	//checking endPoint
	if this.input.EndPoint == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_ENDPOINT, "endPoint cannot be empty", nil)
		return errors.New("invalid endPoint")
	}
	this.endPoint = this.input.EndPoint

	//checking parentDir
	if len(this.input.ParentDir) == 0 {
		this.parentDir = this.input.ParentDir
	} else {
		if this.input.ParentDir[len(this.input.ParentDir)-1] != '/' {
			response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "parentDir must end with '/' ", nil)
			return errors.New("invalid parentDir")
		}
		if this.input.ParentDir[0] == '/' {
			response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "parentDir cannot start with '/' ", nil)
			return errors.New("invalid parentDir")
		}
		this.parentDir = this.input.ParentDir
	}

	//checking keyPrefix
	if len(this.input.DirPrefix) > 10 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "dirPrefix len cannot be greater than 10", nil)
		return errors.New("invalid dirPrefix")
	}
	if strings.Contains(this.input.DirPrefix, constant.SLASH) {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "dirPrefix cannot contain \"/\" ", nil)
		return errors.New("invalid dirPrefix")
	}
	this.dirPrefix = this.input.DirPrefix

	//checking keySuffix
	if len(this.input.DirSuffix) > 10 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "dirSuffix len cannot be greater than 10", nil)
		return errors.New("invalid dirSuffix")
	}
	if strings.Contains(this.input.DirSuffix, constant.SLASH) {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "dirSuffix cannot contain \"/\" ", nil)
		return errors.New("invalid dirSuffix")
	}
	this.dirSuffix = this.input.DirSuffix

	return nil
}
func (createDirsHelper) connectingS3(e echo.Context, this *cdhs) error {
	staticCredentials := credentials.NewStaticCredentials(this.accessKey, this.secretKey, "")
	pathStyle := true
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(constant.KCS3_DEFAULT_REGION),
		Credentials:      staticCredentials,
		Endpoint:         aws.String(this.endPoint),
		S3ForcePathStyle: &pathStyle,
	})
	if err != nil {
		log.Printf("[ERROR]unable to create s3 session: " + err.Error())
		response.Helper().ErrorResponse(e, http.StatusInternalServerError, constant.INTERNAL_SERVER_ERROR, "unable to create s3 session", err.Error())
		return err
	}
	this.sess = sess
	err = CheckCredentials(s3.New(sess))
	if err != nil {
		log.Printf("[ERROR]unable to conect to s3: " + err.Error())
		response.Helper().ErrorResponse(e, http.StatusUnauthorized, constant.UNAUTHORIZED, "unable to connect with s3", err.Error())
		return err
	}
	err = CheckBucketName(s3.New(sess), this.bucketName)
	if err != nil {
		log.Printf("[ERROR]unable to conect with bucket: " + err.Error())
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_BUCKET, "unable to connect with to bucket", err.Error())
		return err
	}
	return nil
}
func (createDirsHelper) doPerform(e echo.Context, this *cdhs) error {
	concurrentRoutinesCount := 0
	var v *progressbar.ProgressBar
	if this.input.Verbose == constant.FALSE {
		v = progressbar.Default(int64(this.instance))
	} else {
		v = nil
	}
	for i := 0; i < this.instance; i++ {
		this.waitGrout.Add(1)

		go func(sess *session.Session, wg *sync.WaitGroup, count int, concurrentRoutines *int, lock *sync.Mutex) {
			defer wg.Done()
			defer func() {
				lock.Lock()
				*concurrentRoutines--
				lock.Unlock()
			}()
		label:
			for true {
				lock.Lock()
				if *concurrentRoutines > config.MaxGoRoutinesExecution {
					lock.Unlock()
					time.Sleep(time.Millisecond * 10)
					continue
				} else {
					*concurrentRoutines++
					lock.Unlock()
					break
				}
			}

			client := s3.New(sess)
			_, err := client.PutObject(&s3.PutObjectInput{
				Bucket:        aws.String(this.bucketName),
				Key:           aws.String(this.parentDir + this.dirPrefix + "(" + PadNumberWithZero(count) + ")" + this.dirPrefix + constant.SLASH),
				ACL:           aws.String("private"),
				Body:          bytes.NewReader(config.NullBytes),
				ContentLength: aws.Int64(int64(len(config.NullBytes))),
				ContentType:   aws.String(http.DetectContentType(config.NullBytes)),
			})

			if err != nil {
				if this.input.Verbose == constant.TRUE {
					log.Println("[ERROR] id:", count, " - ", err.Error())
				}
				lock.Lock()
				*concurrentRoutines--
				lock.Unlock()
				time.Sleep(time.Millisecond * 50)
				goto label
			} else {
				lock.Lock()
				if v != nil {
					_ = v.Add(1)
				}
				lock.Unlock()
			}
		}(this.sess, this.waitGrout, i, &concurrentRoutinesCount, this.mtx)

	}
	go func(group *sync.WaitGroup) {
		group.Wait()
		log.Printf("[DONE]upload task complete")
	}(this.waitGrout)
	return nil
}
func (createDirsHelper) onRollback(e echo.Context, this *cdhs) error {
	return nil

}
