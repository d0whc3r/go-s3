package gos3_test

import (
  "testing"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

func TestGos3(t *testing.T) {
  RegisterFailHandler(Fail)
  RunSpecs(t, "Gos3 Suite")
}
