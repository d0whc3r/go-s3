package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	gos3Files "s3/src/file"
	"s3/src/gos3"
)

var MainWrapper gos3.S3Wrapper
var BucketName string

func TestMain(m *testing.M) {
	fmt.Println("[BEFORE ALL] Tests")
	err := godotenv.Load("../test.env")
	if err != nil {
		fmt.Println("Error loading .env file: ", err)
	}
	BucketName = initBucket()
	MainWrapper = gos3.New(&gos3.S3Config{
		Bucket:         &BucketName,
		Endpoint:       nil,
		Region:         nil,
		MaxRetries:     nil,
		ForcePathStyle: nil,
		SslEnabled:     nil,
	})
	exit := m.Run()
	defer mainTearDown(exit)
}

func mainTearDown(exit int) {
	fmt.Println("[AFTER ALL] Tests")
	clearBucket(BucketName)
	os.Exit(exit)
}

func TestCreateBucketGood(t *testing.T) {
	randomName := getRandomBucketName()
	result, err := MainWrapper.CreateBucket(randomName)
	defer MainWrapper.RemoveBucket(true, randomName)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestCreateBucketBad(t *testing.T) {
	result, err := MainWrapper.CreateBucket(&BucketName)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestListBuckets(t *testing.T) {
	result, err := MainWrapper.GetBuckets()
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, BucketName, *result[0].Name)
}

func TestBucketExistGood(t *testing.T) {
	result := MainWrapper.BucketExist(BucketName)
	assert.True(t, result)
}

func TestBucketExistBad(t *testing.T) {
	notExisting := getRandomBucketName()
	result := MainWrapper.BucketExist(*notExisting)
	assert.False(t, result)
}

func TestUploadFile(t *testing.T) {
	defer restartBucket(MainWrapper)
	files, err := MainWrapper.GetFiles(nil)
	assert.Nil(t, files)
	assert.Nil(t, err)

	result, err := MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", nil, nil)
	assert.NotNil(t, result)
	assert.Nil(t, err)

	files, err = MainWrapper.GetFiles(nil)
	assert.NotNil(t, files)
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "sample-folder/sample1.txt", *files[0].Key)
}

func TestUploadFileExistingWithReplace(t *testing.T) {
	defer restartBucket(MainWrapper)
	_, _ = MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", nil, nil)
	r := true
	result, err := MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", &gos3Files.UploadOptions{
		Replace: &r,
	}, nil)
	assert.NotNil(t, result)
	assert.Nil(t, err)
}

func TestUploadFileExistingWithNoReplace(t *testing.T) {
	defer restartBucket(MainWrapper)
	_, _ = MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", nil, nil)
	r := false
	result, err := MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", &gos3Files.UploadOptions{
		Replace: &r,
	}, nil)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestUploadFileNotExistingBucketWithCreate(t *testing.T) {
	bucketName := getRandomBucketName()
	defer clearBucket(*bucketName)

	c := true
	result, err := MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", &gos3Files.UploadOptions{
		Create: &c,
	}, bucketName)
	assert.NotNil(t, result)
	assert.Nil(t, err)

	exist := MainWrapper.BucketExist(*bucketName)
	assert.True(t, exist)
}

func TestUploadFileNotExistingBucketWithNoCreate(t *testing.T) {
	bucketName := getRandomBucketName()
	defer clearBucket(*bucketName)

	c := false
	result, err := MainWrapper.UploadFile("./sample/sample1.txt", "sample-folder", &gos3Files.UploadOptions{
		Create: &c,
	}, bucketName)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	exist := MainWrapper.BucketExist(*bucketName)
	assert.False(t, exist)
}
