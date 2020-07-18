package gos3

import (
	"time"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type UploadOptionsBasic struct {
	create     bool
	replace    bool
	expire     string
	expireDate time.Time
}

type S3WrapperFiles struct {
	s3 S3Wrapper
	defaultUploadOptions UploadOptionsBasic
}

func (files *S3WrapperFiles) New() {
	var sess = session.Must(session.NewSession())
	files.s3 =
}
