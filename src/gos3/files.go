package gos3

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

func (m S3Manager) deleteFiles(bucket string, files []*s3.Object) ([]*s3.DeletedObject, error) {
	if files == nil || len(files) == 0 {
		return nil, nil
	}
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

func getLimit(timeSpace string) (*time.Time, error) {
	d, err := time.ParseDuration(timeSpace)
	if err != nil {
		return nil, err
	}
	r := time.Now().Add(d * -1)
	return &r, nil
}

func (m S3Manager) getFilesInFolder(bucket, folder string) (files []*s3.Object) {
	f, err := m.GetFiles(bucket)
	if err != nil {
		return
	}
	if folder == "" {
		return f
	}
	test := func(e *s3.Object) bool {
		r := regexp.MustCompile(fmt.Sprintf("^%s/", folder))
		return r.MatchString(*e.Key)
	}
	for _, s := range f {
		if test(s) {
			files = append(files, s)
		}
	}
	return
}
