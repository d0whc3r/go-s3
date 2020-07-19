package gos3Shared

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3SharedBuckets struct {
	s3sdk *s3.S3
}
