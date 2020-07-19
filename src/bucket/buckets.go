package gos3Buckets

import (
	"github.com/aws/aws-sdk-go/service/s3"

	gos3Shared "s3/src/shared"
)

type S3WrapperBuckets struct {
	s3sdk   *s3.S3
	s3Files gos3Shared.S3SharedFiles
}
