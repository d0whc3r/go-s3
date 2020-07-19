package tests

import (
	"fmt"
	"math/rand"
	"time"

	"s3/src/gos3"
)

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getRandomBucketName() *string {
	name := "s3-test-" + randSeq(6)
	return &name
}

func initBucket() string {
	bucketName := getRandomBucketName()
	_, _ = gos3.New(&gos3.S3Config{Bucket: bucketName}).CreateBucket(nil)
	return *bucketName
}

func clearBucket(bucketName string) {
	_, err := gos3.New(&gos3.S3Config{Bucket: &bucketName}).RemoveBucket(true, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func restartBucket(mainWrapper gos3.S3Wrapper) {
	_, _ = mainWrapper.RemoveBucket(true, nil)
	_, _ = mainWrapper.CreateBucket(nil)
}
