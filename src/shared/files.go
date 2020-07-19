package gos3Shared

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3SharedFiles struct {
	s3sdk *s3.S3
}

func (f S3SharedFiles) deleteFiles(bucket string, files []*s3.Object) ([]*s3.DeletedObject, error) {
	var items []*s3.ObjectIdentifier
	for _, file := range files {
		items = append(items, &s3.ObjectIdentifier{
			Key: aws.String(*file.Key),
		})
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: items,
			Quiet:   aws.Bool(true),
		},
	}

	result, err := f.s3sdk.DeleteObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result.Deleted, err
}
