package gos3

import (
  "fmt"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/awserr"
  "github.com/aws/aws-sdk-go/service/s3"
)

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
