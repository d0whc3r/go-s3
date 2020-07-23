package gos3

import (
  "net/http"
  "os"
  "time"

  "github.com/aws/aws-sdk-go/service/s3"
)

type UploadOptions struct {
  Create     *bool
  Replace    *bool
  Expire     *string
  ExpireDate *time.Time
  Compress   *interface{}
}

type S3Manager struct {
  s3sdk                *s3.S3
  DefaultUploadOptions UploadOptions
}

func contentType(out *os.File) (string, error) {
  buf := make([]byte, 512)

  _, err := out.Read(buf)
  if err != nil {
    return "", err
  }

  // Use the net/http package's handy DectectContentType function. Always returns a valid
  // content-type by returning "application/octet-stream" if no others seemed to match.
  contentType := http.DetectContentType(buf)

  return contentType, nil
}

func (m S3Manager) getOptions(o *UploadOptions) UploadOptions {
  op := m.DefaultUploadOptions
  if o != nil {
    oo := *o
    if oo.Create != nil {
      op.Create = oo.Create
    }
    if oo.Replace != nil {
      op.Replace = oo.Replace
    }
    if oo.Expire != nil {
      op.Expire = oo.Expire
    }
    if oo.ExpireDate != nil {
      op.ExpireDate = oo.ExpireDate
    }
    if oo.Compress != nil {
      op.Compress = oo.Compress
    }
  }
  return op
}

func (m S3Manager) createIfNeeded(bucket string) {
  if !m.BucketExist(bucket) {
    _, _ = m.CreateBucket(bucket)
  }
}

func (m S3Manager) canBeReplaced(bucket string, destination string, o UploadOptions) bool {
  return !(!*o.Replace && m.FileExist(bucket, destination))
}
