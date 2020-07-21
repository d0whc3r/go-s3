package gos3

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"s3/src/config"
	"s3/tests"
)

var s3Manager S3Manager
var bucketName string
var s3sdk *s3.S3
var previousBuckets int

const sampleFile1 = "../../tests/sample/sample1.txt"
const sampleFile2 = "../../tests/sample/sample2.jpg"
const envFile = "../../test.env"
const sampleFolder = "sample-folder"

func clearBucket(bucketName string) {
	_, _ = s3Manager.RemoveBucket(bucketName, true)
}

func restartBucket(bucketName string) {
	_, _ = s3Manager.RemoveBucket(bucketName, true)
	_, _ = s3Manager.CreateBucket(bucketName)
}

func TestMain(m *testing.M) {
	fmt.Println("[BEFORE ALL] Tests")
	err := godotenv.Load(envFile)
	if err != nil {
		fmt.Println("Error loading .env file: ", err)
	}
	bucketName = tests.GetRandomBucketName()
	awsConfig := config.AwsConfig(&config.S3Config{
		Bucket:         &bucketName,
		Endpoint:       nil,
		Region:         nil,
		MaxRetries:     nil,
		ForcePathStyle: nil,
		SslEnabled:     nil,
	})
	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		log.Fatal(err)
	}
	s3sdk = s3.New(sess)
	s3Manager = New(s3sdk)
	eb, err := s3Manager.GetBuckets()
	if err != nil {
		log.Fatal(err)
	}
	previousBuckets = len(eb)
	_, _ = s3Manager.CreateBucket(bucketName)
	exit := m.Run()
	defer mainTearDown(exit)
}

func mainTearDown(exit int) {
	fmt.Println("[AFTER ALL] Tests")
	clearBucket(bucketName)
	os.Exit(exit)
}

func TestNew(t *testing.T) {
	m := New(s3sdk)
	assert.NotNil(t, m)
	assert.True(t, *m.defaultUploadOptions.Create)
	assert.False(t, *m.defaultUploadOptions.Replace)
}

func TestBucketExist(t *testing.T) {
	b := s3Manager.BucketExist(bucketName)
	assert.True(t, b)
}

func TestBucketNotExist(t *testing.T) {
	b := s3Manager.BucketExist(tests.GetRandomBucketName())
	assert.False(t, b)
}

func TestCreateBucketGood(t *testing.T) {
	randomName := tests.GetRandomBucketName()
	result, err := s3Manager.CreateBucket(randomName)
	defer s3Manager.RemoveBucket(randomName, true)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestCreateBucketBad(t *testing.T) {
	result, err := s3Manager.CreateBucket(bucketName)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestListBuckets(t *testing.T) {
	result, err := s3Manager.GetBuckets()
	assert.Nil(t, err)
	assert.Len(t, result, previousBuckets+1)
	if previousBuckets == 0 {
		assert.Equal(t, bucketName, *result[0].Name)
	}
}

func TestUploadFile(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
	assert.NotNil(t, result)
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, sampleFolder+"/sample1.txt", *files[0].Key)
}

func TestUploadFileWithoutFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	result, err := s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
	assert.NotNil(t, result)
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "sample1.txt", *files[0].Key)
}

func TestUploadFileWithMultiFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder+"/subfolder/other", nil)
	assert.NotNil(t, result)
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, sampleFolder+"/subfolder/other/sample1.txt", *files[0].Key)
}

func TestUploadFileExistingWithReplace(t *testing.T) {
	defer restartBucket(bucketName)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
	r := true
	result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, &UploadOptions{
		Replace: &r,
	})
	assert.NotNil(t, result)
	assert.Nil(t, err)
}

func TestUploadFileExistingWithNoReplace(t *testing.T) {
	defer restartBucket(bucketName)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
	r := false
	result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, &UploadOptions{
		Replace: &r,
	})
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestUploadFileNotExistingBucketWithCreate(t *testing.T) {
	otherBucketName := tests.GetRandomBucketName()
	defer clearBucket(otherBucketName)

	c := true
	result, err := s3Manager.UploadFile(otherBucketName, sampleFile1, sampleFolder, &UploadOptions{
		Create: &c,
	})
	assert.NotNil(t, result)
	assert.Nil(t, err)

	exist := s3Manager.BucketExist(otherBucketName)
	assert.True(t, exist)
}

func TestUploadFileNotExistingBucketWithNoCreate(t *testing.T) {
	otherBucketName := tests.GetRandomBucketName()
	defer clearBucket(otherBucketName)

	c := false
	result, err := s3Manager.UploadFile(otherBucketName, sampleFile1, sampleFolder, &UploadOptions{
		Create: &c,
	})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	exist := s3Manager.BucketExist(otherBucketName)
	assert.False(t, exist)
}

func TestUploadFiles(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, nil)
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 2)
}

func TestUploadFilesZip(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	var c interface{} = true
	err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, &UploadOptions{Compress: &c})
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, *files[0].Key, ".zip")
	assert.Contains(t, *files[0].Key, sampleFolder+"/")
}

func TestUploadFilesZipName(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	var c interface{} = "zipfile.zip"
	err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, &UploadOptions{Compress: &c})
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, sampleFolder+"/zipfile.zip", *files[0].Key)
}

func TestUploadFilesFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	var c interface{} = false
	err = s3Manager.UploadFiles(bucketName, []string{filepath.Dir(sampleFile1)}, sampleFolder, &UploadOptions{Compress: &c})
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, *files[0].Key, sampleFolder+"/sample1.txt")
	assert.Contains(t, *files[1].Key, sampleFolder+"/sample2.jpg")
}

func TestUploadFilesZipFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	var c interface{} = true
	err = s3Manager.UploadFiles(bucketName, []string{filepath.Dir(sampleFile1)}, sampleFolder, &UploadOptions{Compress: &c})
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, *files[0].Key, ".zip")
	assert.Contains(t, *files[0].Key, sampleFolder+"/")
}

func TestUploadFilesFolderAsterisk(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	var c interface{} = false
	err = s3Manager.UploadFiles(bucketName, []string{filepath.Dir(sampleFile1) + string(os.PathSeparator) + "*"}, sampleFolder, &UploadOptions{Compress: &c})
	assert.Nil(t, err)

	files, err = s3Manager.GetFiles(bucketName)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, sampleFolder+"/sample1.txt", *files[0].Key)
	assert.Equal(t, sampleFolder+"/sample2.jpg", *files[1].Key)
}

func TestCleanOlderSimple(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
	time.Sleep(time.Second * 3)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 2)

	_, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 1)
	assert.Equal(t, sampleFolder+"/sample2.jpg", *files[0].Key)
}

func TestCleanOlderInFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
	time.Sleep(time.Second * 3)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile2, "", nil)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 2)

	_, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 1)
	assert.Equal(t, "sample2.jpg", *files[0].Key)
}

func TestCleanOlderOutFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
	time.Sleep(time.Second * 3)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 2)

	_, _ = s3Manager.CleanOlder(bucketName, "1s", "")
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 1)
	assert.Equal(t, sampleFolder+"/sample2.jpg", *files[0].Key)
}

func TestCleanOlderWithFolder(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)

	_, _ = s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
	time.Sleep(time.Second * 3)
	_, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 2)

	_, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 2)
}

func TestUploadMysql(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)
	_ = s3Manager.UploadMysql(bucketName, sampleFolder, nil)
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 1)
	assert.Contains(t, *files[0].Key, sampleFolder+"/")
	assert.Contains(t, *files[0].Key, "mysqldump-")
	assert.Contains(t, *files[0].Key, ".sql")
}

func TestUploadMysqlZip(t *testing.T) {
	defer restartBucket(bucketName)
	files, err := s3Manager.GetFiles(bucketName)
	assert.Nil(t, files)
	assert.Nil(t, err)
	var c interface{} = true
	_ = s3Manager.UploadMysql(bucketName, sampleFolder, &UploadOptions{Compress: &c})
	files, _ = s3Manager.GetFiles(bucketName)
	assert.Len(t, files, 1)
	assert.Contains(t, *files[0].Key, sampleFolder+"/")
	assert.Contains(t, *files[0].Key, ".zip")
}
