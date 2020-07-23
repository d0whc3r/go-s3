package cmd_test

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "path"

  "github.com/joho/godotenv"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"

  "s3/cmd"
  "s3/src/config"
  "s3/src/version"
  "s3/src/wrapper"
  "s3/tests"
)

const sampleFile1 = "../tests/sample/sample1.txt"
const sampleFile2 = "../tests/sample/sample2.jpg"
const envFile = "../test.env"
const sampleFolder = "sample-folder"

var baseSampleFile1 = path.Base(sampleFile1)
var baseSampleFile2 = path.Base(sampleFile2)

var _ = Describe("Cmd", func() {
  var s3Wrapper wrapper.S3Wrapper
  var bucketName string

  clearBucket := func(bucketName string) {
    _, _ = s3Wrapper.RemoveBucket(true, nil)
  }

  runCommand := func(c []string) string {
    c = append(c, "--bucket", bucketName)
    rootCmd := cmd.RootCmd
    buffer := bytes.NewBufferString("")
    rootCmd.SetOut(buffer)
    rootCmd.SetArgs(c)
    cmd.Execute()
    out, err := ioutil.ReadAll(buffer)
    if err != nil {
      Fail(err.Error())
    }
    return string(out)
  }

  BeforeSuite(func() {
    err := godotenv.Load(envFile)
    if err != nil {
      fmt.Println("Error loading .env file: ", err)
    }
  })

  BeforeEach(func() {
    bucketName = tests.GetRandomBucketName()
    s3Wrapper = wrapper.New(&config.S3Config{Bucket: &bucketName})
    _, _ = s3Wrapper.CreateBucket(nil)
  })

  AfterEach(func() {
    clearBucket(s3Wrapper.Bucket)
  })

  It("Version", func() {
    out := runCommand([]string{"-v"})
    Expect(out).To(Equal(fmt.Sprintf("v%s\n", version.Gos3Version)))
  })

  It("List no files", func() {
    files, err := s3Wrapper.GetFiles(nil)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    out := runCommand([]string{"-l"})
    Expect(out).To(Equal(fmt.Sprintf("[go-s3] No files found in bucket '%s'\n", s3Wrapper.Bucket)))
  })

  It("List files", func() {
    files, err := s3Wrapper.GetFiles(nil)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    _, _ = s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
    _, _ = s3Wrapper.UploadFile(sampleFile2, "", nil, nil)
    out := runCommand([]string{"-l"})
    Expect(out).To(Equal(fmt.Sprintf(
      "[go-s3] File list in bucket '%s': 2\n%s/%s\n%s\n",
      s3Wrapper.Bucket,
      sampleFolder,
      baseSampleFile1,
      baseSampleFile2,
    )))
  })

  Describe("Backup files", func() {
    It("Simple backup one file", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal(baseSampleFile1))
    })
  })
})
