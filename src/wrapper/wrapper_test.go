package wrapper_test

import (
  "fmt"
  "os"
  "path"
  "path/filepath"
  "time"

  "github.com/joho/godotenv"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  . "github.com/onsi/gomega/gstruct"

  "s3/src/config"
  "s3/src/gos3"
  "s3/src/wrapper"
  "s3/tests"
)

const sampleFile1 = "../../tests/sample/sample1.txt"
const sampleFile2 = "../../tests/sample/sample2.jpg"
const envFile = "../../test.env"
const sampleFolder = "sample-folder"

var baseSampleFile1 = path.Base(sampleFile1)
var baseSampleFile2 = path.Base(sampleFile2)

var _ = Describe("Wrapper", func() {
  var s3Wrapper wrapper.S3Wrapper

  clearBucket := func(bucketName string) {
    _, _ = s3Wrapper.RemoveBucket(true, nil)
  }

  BeforeSuite(func() {
    err := godotenv.Load(envFile)
    if err != nil {
      fmt.Println("Error loading .env file: ", err)
    }
  })

  BeforeEach(func() {
    b := tests.GetRandomBucketName()
    s3Wrapper = wrapper.New(&config.S3Config{Bucket: &b})
    _, _ = s3Wrapper.CreateBucket(nil)
  })

  AfterEach(func() {
    clearBucket(s3Wrapper.Bucket)
  })

  It("Bucket created exist", func() {
    b := s3Wrapper.BucketExist(nil)
    Expect(b).To(BeTrue())
  })

  It("Bucket created exist with name", func() {
    b := s3Wrapper.BucketExist(&s3Wrapper.Bucket)
    Expect(b).To(BeTrue())
  })

  It("Bucket random do not exist", func() {
    bucket := tests.GetRandomBucketName()
    b := s3Wrapper.BucketExist(&bucket)
    Expect(b).To(BeFalse())
  })

  It("Create bucket good", func() {
    randomName := tests.GetRandomBucketName()
    result, err := s3Wrapper.CreateBucket(&randomName)
    defer s3Wrapper.RemoveBucket(true, &randomName)
    Expect(result).NotTo(BeNil())
    Expect(err).To(BeNil())
  })

  It("Create bucket bad", func() {
    result, err := s3Wrapper.CreateBucket(nil)
    Expect(result).To(BeNil())
    Expect(err).NotTo(BeNil())
  })

  It("List buckets", func() {
    result, err := s3Wrapper.GetBuckets()
    Expect(err).To(BeNil())
    Expect(result).NotTo(BeNil())
    Expect(result).Should(ContainElement(
      PointTo(
        MatchFields(IgnoreExtras, Fields{
          "CreationDate": PointTo(BeTemporally("~", time.Now(), time.Second)),
          "Name":         PointTo(Equal(s3Wrapper.Bucket)),
        }))))
  })

  Describe("Upload Single file", func() {
    It("Upload file", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile1))
    })
    It("Upload file outside folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Wrapper.UploadFile(sampleFile1, "", nil, nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal(baseSampleFile1))
    })
    It("Upload file with sub route folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder+"/subfolder/other", nil, nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder+"/subfolder/other", baseSampleFile1))
    })
    It("Upload file existing with replace", func() {
      _, _ = s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
      r := true
      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder, &gos3.UploadOptions{Replace: &r}, nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())
    })
    It("Upload file existing with no replace", func() {
      _, _ = s3Wrapper.UploadFile(sampleFile1, sampleFolder, nil, nil)
      r := false
      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder, &gos3.UploadOptions{Replace: &r}, nil)
      Expect(result).To(BeNil())
      Expect(err).ToNot(BeNil())
    })
    It("Upload file with create bucket", func() {
      otherBucketName := tests.GetRandomBucketName()
      defer clearBucket(otherBucketName)

      c := true
      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder, &gos3.UploadOptions{Create: &c}, &otherBucketName)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      exist := s3Wrapper.BucketExist(&otherBucketName)
      Expect(exist).To(BeTrue())
    })
    It("Upload file with no create bucket", func() {
      otherBucketName := tests.GetRandomBucketName()
      defer clearBucket(otherBucketName)

      c := false
      result, err := s3Wrapper.UploadFile(sampleFile1, sampleFolder, &gos3.UploadOptions{Create: &c}, &otherBucketName)
      Expect(result).To(BeNil())
      Expect(err).ToNot(BeNil())

      exist := s3Wrapper.BucketExist(&otherBucketName)
      Expect(exist).To(BeFalse())
    })
  })

  Describe("Upload Multiple file", func() {
    It("Upload files", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      err = s3Wrapper.UploadFiles([]string{sampleFile1, sampleFile2}, sampleFolder, nil, nil)
      Expect(err).To(BeNil())

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

    It("Upload files with zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = true
      err = s3Wrapper.UploadFiles([]string{sampleFile1, sampleFile2}, sampleFolder, &gos3.UploadOptions{Compress: &c}, nil)
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
    })

    It("Upload files with custom zip name", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = "zipfile.zip"
      err = s3Wrapper.UploadFiles([]string{sampleFile1, sampleFile2}, sampleFolder, &gos3.UploadOptions{Compress: &c}, nil)
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, "zipfile.zip"))
    })

    It("Upload files in folder", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = false
      err = s3Wrapper.UploadFiles([]string{filepath.Dir(sampleFile1)}, sampleFolder, &gos3.UploadOptions{Compress: &c}, nil)
      Expect(err).To(BeNil())

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

    It("Upload files in folder with zip", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = true
      err = s3Wrapper.UploadFiles([]string{filepath.Dir(sampleFile1)}, sampleFolder, &gos3.UploadOptions{Compress: &c}, nil)
      Expect(err).To(BeNil())

      files, err = s3Wrapper.GetFiles(nil)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
    })

    It("Upload files in folder with asterisk", func() {
      files, err := s3Wrapper.GetFiles(nil)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = false
      err = s3Wrapper.UploadFiles(
        []string{fmt.Sprintf("%s%s*", filepath.Dir(sampleFile1), string(os.PathSeparator))},
        sampleFolder,
        &gos3.UploadOptions{Compress: &c},
        nil,
      )
      Expect(err).To(BeNil())

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
  })

  It("Mysql backup", func() {
    files, err := s3Wrapper.GetFiles(nil)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    _ = s3Wrapper.UploadMysql(sampleFolder, nil, nil)
    files, _ = s3Wrapper.GetFiles(nil)
    Expect(files).To(HaveLen(1))
    Expect(*files[0].Key).To(MatchRegexp("%s/mysqldump-.*%s", sampleFolder, ".sql"))
  })

  It("Mysql backup with zip", func() {
    files, err := s3Wrapper.GetFiles(nil)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    var c interface{} = true
    _ = s3Wrapper.UploadMysql(sampleFolder, &gos3.UploadOptions{Compress: &c}, nil)
    files, _ = s3Wrapper.GetFiles(nil)
    Expect(files).To(HaveLen(1))
    Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
  })

  It("Clean older files", func() {
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

    _, _ = s3Wrapper.CleanOlder("1s", sampleFolder, nil)
    files, _ = s3Wrapper.GetFiles(nil)
    Expect(files).To(HaveLen(1))
    Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2))
  })
})
