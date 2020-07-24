package wrapper

import (
  "log"

  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"

  "s3/src/config"
  "s3/src/gos3"
)

func New(options *config.S3Config) S3Wrapper {
  cfg := config.Config()

  var s3Wrapper S3Wrapper
  s3Wrapper.Bucket = cfg.Bucket
  if options != nil && options.Bucket != nil && *options.Bucket != "" {
    s3Wrapper.Bucket = *options.Bucket
  }
  if s3Wrapper.Bucket == "" {
    log.Fatal("no bucket configured")
  }

  awsConfig := config.AwsConfig(options)
  s3Wrapper.Endpoint = *awsConfig.Endpoint
  if s3Wrapper.Endpoint == "" {
    log.Fatal("no endpoint configured")
  }
  sess, err := session.NewSession(&awsConfig)
  if err != nil {
    log.Fatal(err)
  }
  s3Wrapper.s3 = s3.New(sess)
  s3Wrapper.s3Manager = gos3.New(s3Wrapper.s3)
  return s3Wrapper
}

func (w S3Wrapper) CreateBucket(bucket *string) (*s3.CreateBucketOutput, error) {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.CreateBucket(bucketName)
}

func (w S3Wrapper) GetBuckets() ([]*s3.Bucket, error) {
  return w.s3Manager.GetBuckets()
}

func (w S3Wrapper) BucketExist(bucket *string) bool {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.BucketExist(bucketName)
}

func (w S3Wrapper) RemoveBucket(force bool, bucket *string) (*s3.DeleteBucketOutput, error) {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.RemoveBucket(bucketName, force)
}

func (w S3Wrapper) GetFiles(bucket *string) ([]*s3.Object, error) {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.GetFiles(bucketName)
}

func (w S3Wrapper) UploadFile(file string, folder string, options *gos3.UploadOptions, bucket *string) (*s3.PutObjectOutput, error) {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.UploadFile(bucketName, file, folder, options)
}

func (w S3Wrapper) UploadFiles(files []string, folder string, options *gos3.UploadOptions, bucket *string) error {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.UploadFiles(bucketName, files, folder, options)
}

func (w S3Wrapper) CleanOlder(time string, folder string, bucket *string) ([]*s3.DeletedObject, error) {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.CleanOlder(bucketName, time, folder)
}

func (w S3Wrapper) UploadMysql(folder string, options *gos3.UploadOptions, bucket *string) error {
  bucketName := w.getBucketName(bucket)
  return w.s3Manager.UploadMysql(bucketName, folder, options)
}
