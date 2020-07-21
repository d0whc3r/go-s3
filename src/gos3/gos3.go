package gos3

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	defaultUploadOptions UploadOptions
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
	op := m.defaultUploadOptions
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

func (m S3Manager) deleteFiles(bucket string, files []*s3.Object) ([]*s3.DeletedObject, error) {
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

	result, err := m.s3sdk.DeleteObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result.Deleted, err
}

func (m S3Manager) uploadMultipleFiles(bucket string, uploadFiles []string, folder string, options *UploadOptions) error {
	var err error
	for _, f := range uploadFiles {
		i, ie := os.Stat(f)
		isDir := ie == nil && i.IsDir()
		if isDir || strings.Contains(f, "*") {
			err = m.recursiveUpload(bucket, f, err, folder, options)
		} else {
			_, err = m.UploadFile(bucket, f, folder, options)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m S3Manager) recursiveUpload(bucket string, f string, err error, folder string, options *UploadOptions) error {
	var newFiles []string

	if strings.Contains(f, "*") {
		newFiles, err = filepath.Glob(f)
	} else {
		newFiles, err = filepath.Glob(f + string(os.PathSeparator) + "*")
	}
	if err != nil {
		return err
	}
	err = m.UploadFiles(bucket, newFiles, folder, options)
	return err
}
