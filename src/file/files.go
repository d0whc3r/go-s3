package gos3Files

import (
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	gos3Shared "s3/src/shared"
)

type UploadOptions struct {
	Create     *bool
	Replace    *bool
	Expire     *string
	ExpireDate *time.Time
	Compress   *interface{}
}

type S3WrapperFiles struct {
	s3sdk                *s3.S3
	s3Buckets            gos3Shared.S3SharedBuckets
	defaultUploadOptions UploadOptions
}

func contentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func (f S3WrapperFiles) getOptions(o *UploadOptions) UploadOptions {
	op := f.defaultUploadOptions
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
	}
	return op
}

func (f S3WrapperFiles) createIfNeeded(bucket string) {
	if !f.s3Buckets.BucketExist(bucket) {
		_, _ = f.s3Buckets.CreateBucket(bucket)
	}
}

func (f S3WrapperFiles) canBeReplaced(bucket string, destination string, o UploadOptions) bool {
	return !(!*o.Replace && f.FileExist(bucket, destination))
}
