package gos3Files

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	gos3Shared "s3/src/shared"
)

func New(s3sdk *s3.S3) S3WrapperFiles {
	var s3WrapperFiles S3WrapperFiles
	s3WrapperFiles.s3sdk = s3sdk
	c := true
	r := false
	s3WrapperFiles.defaultUploadOptions = UploadOptions{
		Create:  &c,
		Replace: &r,
	}
	s3WrapperFiles.s3Buckets = gos3Shared.NewBuckets(s3sdk)
	return s3WrapperFiles
}

func (f S3WrapperFiles) FileInfo(bucket, fileName string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	}

	result, err := f.s3sdk.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result, err
}

func (f S3WrapperFiles) FileExist(bucket, fileName string) bool {
	_, err := f.FileInfo(bucket, fileName)
	return err == nil
}

func (f S3WrapperFiles) UploadFile(bucket string, file string, folder string, options *UploadOptions) (*s3.PutObjectOutput, error) {
	o := f.getOptions(options)
	osFile, err := os.Open(file)
	if err != nil {
		log.Fatalf("os.Open - filename: %v, err: %v", file, err)
	}
	defer osFile.Close()

	name := filepath.Base(file)
	destination := strings.Join([]string{folder, name}, "/")
	mime, _ := contentType(osFile)

	if *o.Create {
		f.createIfNeeded(bucket)
	}

	if !f.canBeReplaced(bucket, destination, o) {
		return nil, errors.New("file '" + destination + "' already exists in bucket '" + bucket + "' and will not be replaced")
	}

	input := &s3.PutObjectInput{
		Key:         aws.String(destination),
		Bucket:      aws.String(bucket),
		Body:        aws.ReadSeekCloser(strings.NewReader(file)),
		Expires:     nil, // TODO: Parse Expire to send when upload
		ContentType: aws.String(mime),
	}

	result, err := f.s3sdk.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result, err

}
