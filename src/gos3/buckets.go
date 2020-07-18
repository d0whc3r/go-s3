package gos3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3WrapperBuckets struct {
	s3sdk   s3.S3
	s3Files S3WrapperFiles
}

func (b S3WrapperBuckets) New(s3 s3.S3) {
	b.s3sdk = s3
}

func (b S3WrapperBuckets) GetBuckets() {
	var input = &s3.ListBucketsInput{}
	var result, err = b.s3sdk.ListBuckets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println(result)
}
