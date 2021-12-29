package helper

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func CheckCredentials(client *s3.S3) error {

	listBucketInput := s3.ListBucketsInput{}
	_, err := client.ListBuckets(&listBucketInput)
	return err
}
func CheckBucketName(client *s3.S3, bucket string) error {
	listObjectsInput := s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}
	_, err := client.ListObjects(&listObjectsInput)
	return err
}
func PadNumberWithZero(value int) string {
	return fmt.Sprintf("%07d", value)
}
