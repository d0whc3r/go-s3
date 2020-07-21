package gos3

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

