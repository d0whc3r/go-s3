package tests

import (
	"math/rand"
	"time"
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

func GetRandomBucketName() string {
	name := "s3-test-" + randSeq(6)
	return name
}
