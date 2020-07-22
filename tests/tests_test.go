package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"s3/tests"
)

var _ = Describe("Tests", func() {
	It("Random name", func() {
		b := tests.GetRandomBucketName()
		Expect(b).To(HavePrefix("s3-test-"))
	})
	It("Random name twice", func() {
		b1 := tests.GetRandomBucketName()
		Expect(b1).To(HavePrefix("s3-test-"))
		b2 := tests.GetRandomBucketName()
		Expect(b2).To(HavePrefix("s3-test-"))
		Expect(b1).ToNot(Equal(b2))
	})
})
