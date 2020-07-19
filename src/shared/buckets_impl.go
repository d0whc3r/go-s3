package gos3Shared

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewBuckets(s3sdk *s3.S3) S3SharedBuckets {
	var s3SharedBuckets S3SharedBuckets
	s3SharedBuckets.s3sdk = s3sdk
	return s3SharedBuckets
}

func (b S3SharedBuckets) CreateBucket(bucket string) (*s3.CreateBucketOutput, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	result, err := b.s3sdk.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
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

func (b S3SharedBuckets) BucketExist(bucket string) bool {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	_, err := b.s3sdk.HeadBucket(input)
	return err == nil
}
