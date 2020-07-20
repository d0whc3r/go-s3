package gos3Buckets

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	gos3Shared "s3/src/shared"
)

func New(s3sdk *s3.S3) S3WrapperBuckets {
	var s3WrapperBuckets S3WrapperBuckets
	s3WrapperBuckets.s3sdk = s3sdk
	s3WrapperBuckets.s3Files = gos3Shared.NewFiles(s3sdk)
	return s3WrapperBuckets
}

func (b S3WrapperBuckets) GetBuckets() ([]*s3.Bucket, error) {
	result, err := b.s3sdk.ListBuckets(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return []*s3.Bucket{}, err
	}
	return result.Buckets, err
}

func (b S3WrapperBuckets) RemoveBucket(bucket string, force bool) (*s3.DeleteBucketOutput, error) {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}
	result, err := b.s3sdk.DeleteBucket(input)
	if err != nil {
		if force {
			_, dErr := b.s3Files.DeleteAllContent(bucket)
			if dErr == nil {
				return b.RemoveBucket(bucket, false)
			} else {
				return nil, dErr
			}
		}
		return nil, err
	}

	return result, err
}
