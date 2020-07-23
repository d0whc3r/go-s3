package wrapper

import (
  "github.com/aws/aws-sdk-go/service/s3"

  "s3/src/gos3"
)

type S3Wrapper struct {
  Bucket    string
  Endpoint  string
  s3        *s3.S3
  s3Manager gos3.S3Manager
}

func (w S3Wrapper) getBucketName(bucket *string) string {
  var bucketName string
  if bucket != nil {
    bucketName = *bucket
  } else {
    bucketName = w.Bucket
  }
  return bucketName
}
