package gos3

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Get the file information
	info, err := f.Stat()
	if err != nil {
		return err
	}

	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	h.Name = filename
	h.Method = zip.Deflate

	w, err := zipWriter.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func compressFiles(files []string, output string) error {
	dir, err := ioutil.TempDir("zip", "s3zip")
	if err != nil {
		return err
	}
	dest := path.Join(dir, output)
	z, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer z.Close()

	w := zip.NewWriter(z)
	defer w.Close()

	for _, file := range files {
		if err = addFileToZip(w, file); err != nil {
			return err
		}
	}
	return nil
}
