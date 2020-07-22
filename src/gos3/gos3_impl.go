package gos3

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

func New(s3sdk *s3.S3) (s3Manager S3Manager) {
	s3Manager.s3sdk = s3sdk
	c := true
	r := false
	var cp interface{} = false
	s3Manager.DefaultUploadOptions = UploadOptions{
		Create:   &c,
		Replace:  &r,
		Compress: &cp,
	}
	return
}
