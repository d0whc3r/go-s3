package cmd_test

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "path"
  "time"

  "github.com/joho/godotenv"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  . "github.com/onsi/gomega/gstruct"

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
    if len(c) > 0 {
      c = append(c, "--bucket", bucketName)
    }
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

  It("No args", func() {
    out := runCommand([]string{})
    Expect(out).To(MatchRegexp(fmt.Sprintf(`^Help for go s3 v%s`, version.Gos3Version)))
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

    It("Simple backup one file to folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-f", sampleFolder})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile1))
    })

    It("Simple backup two files", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-b", sampleFile2})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Second)),
            "Key":          PointTo(MatchRegexp("%s", baseSampleFile1)),
          }))))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Second)),
            "Key":          PointTo(MatchRegexp("%s", baseSampleFile2)),
          }))))
    })

    It("Simple backup two files to folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-b", sampleFile2, "-f", sampleFolder})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Second)),
            "Key":          PointTo(MatchRegexp("%s/%s", sampleFolder, baseSampleFile1)),
          }))))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Second)),
            "Key":          PointTo(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2)),
          }))))
    })

    It("Simple backup one file with zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-z"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp(".*%s", ".zip"))
    })

    It("Simple backup one file with zip name", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-n", "zipname.zip"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal("zipname.zip"))
    })

    It("Simple backup two files with zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-b", sampleFile1, "-b", sampleFile2, "-z"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp(".*%s", ".zip"))
    })
  })

  Describe("Mysql dump", func() {
    It("Backup mysql dump", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-m"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] MySql dump backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("mysqldump-.*%s$", ".sql"))
    })

    It("Backup mysql dump to folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-m", "-f", sampleFolder})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] MySql dump backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/mysqldump-.*%s$", sampleFolder, ".sql"))
    })

    It("Backup mysql dump to zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-m", "-z"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] MySql dump backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp(".*%s$", ".zip"))
    })

    It("Backup mysql dump to folder and zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      out := runCommand([]string{"-m", "-f", sampleFolder, "-z"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] MySql dump backup success in bucket '%s'\n",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/.*%s$", sampleFolder, ".zip"))
    })
  })

  Describe("Clean older files", func() {
    It("Clear files simple", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Wrapper.UploadFile(sampleFile1, "", nil, nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Wrapper.UploadFile(sampleFile2, "", nil, nil)
      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      out := runCommand([]string{"-d", "1s"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Deleted files older than '%s' in bucket '%s'\n",
        "1s",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal(baseSampleFile2))
    })

    It("Clear files simple in folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Wrapper.UploadFile(sampleFile2, sampleFolder, nil, nil)
      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      out := runCommand([]string{"-d", "1s", "-f", sampleFolder})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Deleted files in folder '%s' older than '%s' in bucket '%s'\n",
        sampleFolder,
        "1s",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2))
    })

    It("Clear files complex in folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Wrapper.UploadFile(sampleFile2, sampleFolder, nil, nil)
      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      out := runCommand([]string{"-d", sampleFolder + "=1s"})
      Expect(out).To(Equal(fmt.Sprintf(
        "[go-s3] Deleted files in folder '%s' older than '%s' in bucket '%s'\n",
        sampleFolder,
        "1s",
        s3Wrapper.Bucket,
      )))

      files, _ = s3Wrapper.GetFiles(nil)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2))
    })
  })
})
