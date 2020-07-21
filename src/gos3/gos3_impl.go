package gos3

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mholt/archiver/v3"
)

func New(s3sdk *s3.S3) (s3Manager S3Manager) {
	s3Manager.s3sdk = s3sdk
	c := true
	r := false
	var cp interface{} = false
	s3Manager.defaultUploadOptions = UploadOptions{
		Create:   &c,
		Replace:  &r,
		Compress: &cp,
	}
	return
}

func (m S3Manager) GetBuckets() ([]*s3.Bucket, error) {
	result, err := m.s3sdk.ListBuckets(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return []*s3.Bucket{}, err
	}
	return result.Buckets, err
}

func (m S3Manager) RemoveBucket(bucket string, force bool) (*s3.DeleteBucketOutput, error) {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}
	result, err := m.s3sdk.DeleteBucket(input)
	if err != nil {
		if force {
			_, dErr := m.DeleteAllContent(bucket)
			if dErr == nil {
				return m.RemoveBucket(bucket, false)
			} else {
				return nil, dErr
			}
		}
		return nil, err
	}

	return result, err
}

func (m S3Manager) CreateBucket(bucket string) (*s3.CreateBucketOutput, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	result, err := m.s3sdk.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result, err
}

func (m S3Manager) BucketExist(bucket string) bool {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	_, err := m.s3sdk.HeadBucket(input)
	return err == nil
}

// Files

func (m S3Manager) FileInfo(bucket, fileName string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	}

	result, err := m.s3sdk.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result, err
}

func (m S3Manager) DeleteAllContent(bucket string) ([]*s3.DeletedObject, error) {
	files, err := m.GetFiles(bucket)
	if err == nil {
		return m.deleteFiles(bucket, files)
	}
	return nil, err
}

func (m S3Manager) GetFiles(bucket string) ([]*s3.Object, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	result, err := m.s3sdk.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result.Contents, err
}

func (m S3Manager) FileExist(bucket, fileName string) bool {
	_, err := m.FileInfo(bucket, fileName)
	return err == nil
}

func (m S3Manager) UploadFile(bucket string, file string, folder string, options *UploadOptions) (*s3.PutObjectOutput, error) {
	o := m.getOptions(options)
	osFile, err := os.Open(file)
	if err != nil {
		log.Fatalf("os.Open - filename: %v, err: %v", file, err)
	}
	defer osFile.Close()

	name := filepath.Base(file)
	destination := strings.Join([]string{folder, name}, string(filepath.Separator))
	mime, _ := contentType(osFile)

	if *o.Create {
		m.createIfNeeded(bucket)
	}

	if !m.canBeReplaced(bucket, destination, o) {
		return nil, errors.New("file '" + destination + "' already exists in bucket '" + bucket + "' and will not be replaced")
	}

	input := &s3.PutObjectInput{
		Key:         aws.String(destination),
		Bucket:      aws.String(bucket),
		Body:        aws.ReadSeekCloser(strings.NewReader(file)),
		Expires:     nil, // TODO: Parse Expire to send when upload
		ContentType: aws.String(mime),
	}

	result, err := m.s3sdk.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result, err
}

func (m S3Manager) UploadFiles(bucket string, files []string, folder string, options *UploadOptions) error {
	o := m.getOptions(options)
	compress := ""
	if o.Compress != nil {
		c := *(o.Compress)
		switch c.(type) {
		case string:
			compress = c.(string)
		case bool:
			if c.(bool) {
				formatDate := time.Now().Format("2006-01-02.150405")
				compress = "zipped_" + formatDate + ".zip"
			}
		}
	}
	uploadFiles := files[:]
	if compress != "" {
		dir, err := ioutil.TempDir("", "s3zip")
		if err != nil {
			return err
		}
		dest := path.Join(dir, compress)
		err = archiver.Archive(files, dest)
		if err != nil {
			log.Fatal("Error in compress files: ", err)
		}
		uploadFiles = []string{dest}
	}

	err := m.uploadMultipleFiles(bucket, uploadFiles, folder, options)
	if err != nil {
		return err
	}
	return nil
}
