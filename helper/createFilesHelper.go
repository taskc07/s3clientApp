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

type cfhs struct {
	input        dto.CreateFilesDto
	bucketName   string
	endPoint     string
	accessKey    string
	secretKey    string
	parentDir    string
	instance     int
	keyPrefix    string
	keySuffix    string
	sess         *session.Session
	bytesStream  *[]byte
	mtx          *sync.Mutex
	waitGrout    *sync.WaitGroup
	commonClient *s3.S3
}

type createFilesHelper struct {
}

var singletonCfhs *createFilesHelper
var onceCfhs sync.Once

func CreateFiles() *createFilesHelper {
	onceCfhs.Do(func() {
		singletonCfhs = new(createFilesHelper)
	})
	return singletonCfhs
}

func (t *createFilesHelper) Execute(e echo.Context) error {

	this := new(cfhs)
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

func (createFilesHelper) init(e echo.Context, this *cfhs) error {
	var temp = new(dto.CreateFilesDto)
	err := e.Bind(temp)
	if err != nil {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "", err.Error())
		return errors.New("data bind error")
	} else {
		this.input = *temp
	}
	return nil
}
func (createFilesHelper) checkInput(e echo.Context, this *cfhs) error {

	//checking accessKey
	if this.input.AccessKey == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.EMPTY_ACCESS_KEY, "access key cannot be empty", nil)
		return errors.New("access key cannot be empty")
	}
	this.accessKey = this.input.AccessKey

	//checking secretKey
	if this.input.SecretKey == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.EMPTY_SECRET, "secret key cannot be empty", nil)
		return errors.New("secret key cannot be empty")
	}
	this.secretKey = this.input.SecretKey

	//validating instanceCount
	if this.input.Instance < 1 || this.input.Instance > 500000 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "instance should be between 1~500000", nil)
		return errors.New("invalid instance")
	}
	this.instance = this.input.Instance

	//checking nullData field
	if this.input.NullData != constant.TRUE && this.input.NullData != constant.FALSE {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "nullData should be true or false", nil)
		return errors.New("invalid null data param")
	}
	if this.input.NullData == constant.TRUE {
		this.bytesStream = &config.NullBytes
	} else {
		this.bytesStream = &config.PdfBytes
	}

	//checking bucketName
	if this.input.BucketName == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_BUCKET, "bucketName cannot be empty", nil)
		return errors.New("empty bucket name")
	}
	this.bucketName = this.input.BucketName

	//checking endPoint
	if this.input.EndPoint == "" {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_ENDPOINT, "endPoint cannot be empty", nil)
		return errors.New("null endPoint")
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
	if len(this.input.KeyPrefix) > 10 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "keyPrefix len cannot be greater than 10", nil)
		return errors.New("invalid keyPrefix")
	}
	if strings.Contains(this.input.KeyPrefix, constant.SLASH) {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "keyPrefix cannot contain \"/\" ", nil)
		return errors.New("invalid kePrefix")
	}
	this.keyPrefix = this.input.KeyPrefix

	//checking keySuffix
	if len(this.input.KeySuffix) > 10 {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "keySuffix len cannot be greater than 10", nil)
		return errors.New("invalid keySuffix")
	}
	if strings.Contains(this.input.KeySuffix, constant.SLASH) {
		response.Helper().ErrorResponse(e, http.StatusBadRequest, constant.INVALID_INPUT, "keySuffix cannot contain \"/\" ", nil)
		return errors.New("invalid keySuffix")
	}
	this.keySuffix = this.input.KeySuffix

	return nil
}
func (createFilesHelper) connectingS3(e echo.Context, this *cfhs) error {
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
	this.commonClient = s3.New(sess)
	return nil
}
func (createFilesHelper) doPerform(e echo.Context, this *cfhs) error {
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

			var client *s3.S3
			if this.input.EnableCommonClient == constant.TRUE {
				client = this.commonClient
			} else {
				client = s3.New(sess)
			}
			_, err := client.PutObject(&s3.PutObjectInput{
				Bucket:        aws.String(this.bucketName),
				Key:           aws.String(this.parentDir + this.keyPrefix + "(" + PadNumberWithZero(count) + ")" + this.keySuffix),
				ACL:           aws.String("private"),
				Body:          bytes.NewReader(*this.bytesStream),
				ContentLength: aws.Int64(int64(len(*this.bytesStream))),
				ContentType:   aws.String(http.DetectContentType(*this.bytesStream)),
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
func (createFilesHelper) onRollback(e echo.Context, this *cfhs) error {
	return nil

}
