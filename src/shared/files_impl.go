package gos3Shared

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewFiles(s3sdk *s3.S3) S3SharedFiles {
	var s3SharedFiles S3SharedFiles
	s3SharedFiles.s3sdk = s3sdk
	return s3SharedFiles
}

func (f S3SharedFiles) DeleteAllContent(bucket string) ([]*s3.DeletedObject, error) {
	files, err := f.GetFiles(bucket)
	if err == nil {
		return f.deleteFiles(bucket, files)
	}
	return nil, err
}

func (f S3SharedFiles) GetFiles(bucket string) ([]*s3.Object, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	result, err := f.s3sdk.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result.Contents, err
}
