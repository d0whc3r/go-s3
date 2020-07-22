package gos3_test

import (
  "fmt"
  "log"
  "os"
  "path"
  "path/filepath"
  "time"

  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
  "github.com/joho/godotenv"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  . "github.com/onsi/gomega/gstruct"

  "s3/src/config"
  "s3/src/gos3"
  "s3/tests"
)

const sampleFile1 = "../../tests/sample/sample1.txt"
const sampleFile2 = "../../tests/sample/sample2.jpg"
const envFile = "../../test.env"
const sampleFolder = "sample-folder"

var baseSampleFile1 = path.Base(sampleFile1)
var baseSampleFile2 = path.Base(sampleFile2)

var _ = Describe("Gos3", func() {
  var s3Manager gos3.S3Manager
  var bucketName string
  var s3sdk *s3.S3

  clearBucket := func(bucketName string) {
    _, _ = s3Manager.RemoveBucket(bucketName, true)
  }

  BeforeSuite(func() {
    err := godotenv.Load(envFile)
    if err != nil {
      fmt.Println("Error loading .env file: ", err)
    }
  })

  BeforeEach(func() {
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
    s3Manager = gos3.New(s3sdk)
    _, _ = s3Manager.CreateBucket(bucketName)
  })

  AfterEach(func() {
    clearBucket(bucketName)
  })

  It("New instance", func() {
    m := gos3.New(s3sdk)
    Expect(m).ToNot(BeNil())
    Expect(*m.DefaultUploadOptions.Create).To(BeTrue())
    Expect(*m.DefaultUploadOptions.Replace).To(BeFalse())
  })

  It("Bucket created exist", func() {
    b := s3Manager.BucketExist(bucketName)
    Expect(b).To(BeTrue())
  })

  It("Bucket random do not exist", func() {
    b := s3Manager.BucketExist(tests.GetRandomBucketName())
    Expect(b).To(BeFalse())
  })

  It("Create bucket good", func() {
    randomName := tests.GetRandomBucketName()
    result, err := s3Manager.CreateBucket(randomName)
    defer s3Manager.RemoveBucket(randomName, true)
    Expect(result).NotTo(BeNil())
    Expect(err).To(BeNil())
  })

  It("Create bucket bad", func() {
    result, err := s3Manager.CreateBucket(bucketName)
    Expect(result).To(BeNil())
    Expect(err).NotTo(BeNil())
  })

  It("List buckets", func() {
    result, err := s3Manager.GetBuckets()
    Expect(err).To(BeNil())
    Expect(result).NotTo(BeNil())
    Expect(result).Should(ContainElement(
      PointTo(
        MatchFields(IgnoreExtras, Fields{
          "CreationDate": PointTo(BeTemporally("~", time.Now(), time.Second)),
          "Name":         PointTo(Equal(bucketName)),
        }))))
  })

  Describe("Upload Single file", func() {
    It("Upload file", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile1))
    })
    It("Upload file outside folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal(baseSampleFile1))
    })
    It("Upload file with sub route folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder+"/subfolder/other", nil)
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder+"/subfolder/other", baseSampleFile1))
    })
    It("Upload file existing with replace", func() {
      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
      r := true
      result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, &gos3.UploadOptions{Replace: &r})
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())
    })
    It("Upload file existing with no replace", func() {
      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
      r := false
      result, err := s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, &gos3.UploadOptions{Replace: &r})
      Expect(result).To(BeNil())
      Expect(err).ToNot(BeNil())
    })
    It("Upload file with create bucket", func() {
      otherBucketName := tests.GetRandomBucketName()
      defer clearBucket(otherBucketName)

      c := true
      result, err := s3Manager.UploadFile(otherBucketName, sampleFile1, sampleFolder, &gos3.UploadOptions{Create: &c})
      Expect(result).ToNot(BeNil())
      Expect(err).To(BeNil())

      exist := s3Manager.BucketExist(otherBucketName)
      Expect(exist).To(BeTrue())
    })
    It("Upload file with no create bucket", func() {
      otherBucketName := tests.GetRandomBucketName()
      defer clearBucket(otherBucketName)

      c := false
      result, err := s3Manager.UploadFile(otherBucketName, sampleFile1, sampleFolder, &gos3.UploadOptions{Create: &c})
      Expect(result).To(BeNil())
      Expect(err).ToNot(BeNil())

      exist := s3Manager.BucketExist(otherBucketName)
      Expect(exist).To(BeFalse())
    })
  })
  Describe("Upload Multiple file", func() {
    It("Upload files", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, nil)
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
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
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = true
      err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, &gos3.UploadOptions{Compress: &c})
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
    })

    It("Upload files with custom zip name", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = "zipfile.zip"
      err = s3Manager.UploadFiles(bucketName, []string{sampleFile1, sampleFile2}, sampleFolder, &gos3.UploadOptions{Compress: &c})
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, "zipfile.zip"))
    })

    It("Upload files in folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = false
      err = s3Manager.UploadFiles(bucketName, []string{filepath.Dir(sampleFile1)}, sampleFolder, &gos3.UploadOptions{Compress: &c})
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
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
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = true
      err = s3Manager.UploadFiles(bucketName, []string{filepath.Dir(sampleFile1)}, sampleFolder, &gos3.UploadOptions{Compress: &c})
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
    })

    It("Upload files in folder with asterisk", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      var c interface{} = false
      err = s3Manager.UploadFiles(
        bucketName,
        []string{fmt.Sprintf("%s%s*", filepath.Dir(sampleFile1), string(os.PathSeparator))},
        sampleFolder,
        &gos3.UploadOptions{Compress: &c},
      )
      Expect(err).To(BeNil())

      files, err = s3Manager.GetFiles(bucketName)
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

  Describe("Clean older files", func() {
    It("Clean older simple", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      _, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2))
    })

    It("Clean older in folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, sampleFolder, nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Manager.UploadFile(bucketName, sampleFile2, "", nil)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      _, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(Equal(baseSampleFile2))
    })

    It("Clean older outside folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(err).To(BeNil())
      Expect(files).To(HaveLen(2))

      _, _ = s3Manager.CleanOlder(bucketName, "1s", "")
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).To(HaveLen(1))
      Expect(*files[0].Key).To(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2))
    })

    It("Clean older inside folder", func() {
      files, err := s3Manager.GetFiles(bucketName)
      Expect(files).To(BeNil())
      Expect(err).To(BeNil())

      _, _ = s3Manager.UploadFile(bucketName, sampleFile1, "", nil)
      time.Sleep(time.Second * 3)
      _, _ = s3Manager.UploadFile(bucketName, sampleFile2, sampleFolder, nil)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).ToNot(BeNil())
      Expect(files).To(HaveLen(2))

      _, _ = s3Manager.CleanOlder(bucketName, "1s", sampleFolder)
      files, _ = s3Manager.GetFiles(bucketName)
      Expect(files).To(HaveLen(2))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Minute)),
            "Key":          PointTo(Equal(baseSampleFile1)),
          }))))
      Expect(files).Should(ContainElement(
        PointTo(
          MatchFields(IgnoreExtras, Fields{
            "LastModified": PointTo(BeTemporally("~", time.Now(), time.Second)),
            "Key":          PointTo(MatchRegexp("%s/%s", sampleFolder, baseSampleFile2)),
          }))))
    })
  })

  It("Upload mysql file dump", func() {
    files, err := s3Manager.GetFiles(bucketName)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    _ = s3Manager.UploadMysql(bucketName, sampleFolder, nil)
    files, _ = s3Manager.GetFiles(bucketName)
    Expect(files).To(HaveLen(1))
    Expect(*files[0].Key).To(MatchRegexp("%s/mysqldump-.*%s", sampleFolder, ".sql"))
  })

  It("Upload mysql file dump zip", func() {
    files, err := s3Manager.GetFiles(bucketName)
    Expect(files).To(BeNil())
    Expect(err).To(BeNil())

    var c interface{} = true
    _ = s3Manager.UploadMysql(bucketName, sampleFolder, &gos3.UploadOptions{Compress: &c})
    files, _ = s3Manager.GetFiles(bucketName)
    Expect(files).To(HaveLen(1))
    Expect(*files[0].Key).To(MatchRegexp("%s/.*%s", sampleFolder, ".zip"))
  })
})
