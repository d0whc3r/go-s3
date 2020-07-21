package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRandomBucketName(t *testing.T) {
	b := GetRandomBucketName()
	assert.NotNil(t, b)
	assert.Contains(t, b, "s3-test-")
}

func TestGetRandomBucketNameDistinct(t *testing.T) {
	b1 := GetRandomBucketName()
	assert.NotNil(t, b1)
	b2 := GetRandomBucketName()
	assert.NotNil(t, b2)
	assert.NotEqual(t, b1, b2)
}
