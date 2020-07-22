package config_test

import (
  "os"
  "strconv"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"

  "s3/src/config"
)

const (
  endpoint = "none"
  bucket   = "the-bucket"
  region   = "the-region"
  retries  = 5
  force    = true
  ssl      = false
  access   = "access-key"
  secret   = "secret-key"
  host     = "localhost"
  port     = 3306
  user     = "username"
  pass     = "password"
  db       = "database"
)

var _ = Describe("Config", func() {
  BeforeSuite(func() {
    f := "false"
    if force {
      f = "true"
    }
    s := "false"
    if ssl {
      s = "true"
    }
    _ = os.Setenv("ENDPOINT", endpoint)
    _ = os.Setenv("BUCKET", bucket)
    _ = os.Setenv("REGION", region)
    _ = os.Setenv("MAX_RETRIES", strconv.Itoa(retries))
    _ = os.Setenv("FORCE_PATH_STYLE", f)
    _ = os.Setenv("SSL_ENABLED", s)
    _ = os.Setenv("ACCESS_KEY", access)
    _ = os.Setenv("SECRET_KEY", secret)
    _ = os.Setenv("MYSQL_HOST", host)
    _ = os.Setenv("MYSQL_PORT", strconv.Itoa(port))
    _ = os.Setenv("MYSQL_USER", user)
    _ = os.Setenv("MYSQL_PASSWORD", pass)
    _ = os.Setenv("MYSQL_DATABASE", db)
  })

  It("Config values", func() {
    cfg := config.Config()
    Expect(cfg.Endpoint).To(BeIdenticalTo(endpoint))
    Expect(cfg.Bucket).To(BeIdenticalTo(bucket))
    Expect(cfg.Region).To(BeIdenticalTo(region))
    Expect(cfg.MaxRetries).To(BeIdenticalTo(retries))
    Expect(cfg.ForcePathStyle).To(BeIdenticalTo(force))
    Expect(cfg.SslEnabled).To(BeIdenticalTo(ssl))
    Expect(cfg.MysqlHost).To(BeIdenticalTo(host))
    Expect(cfg.MysqlPort).To(BeIdenticalTo(port))
    Expect(cfg.MysqlUser).To(BeIdenticalTo(user))
    Expect(cfg.MysqlPassword).To(BeIdenticalTo(pass))
    Expect(cfg.MysqlDatabase).To(BeIdenticalTo(db))
  })

  It("AWS Config default values", func() {
    cfg := config.AwsConfig(nil)
    e := endpoint
    r := region
    m := retries
    f := force
    s := !ssl
    Expect(cfg.Endpoint).To(Equal(&e))
    Expect(cfg.Region).To(Equal(&r))
    Expect(cfg.MaxRetries).To(Equal(&m))
    Expect(cfg.S3ForcePathStyle).To(Equal(&f))
    Expect(cfg.DisableSSL).To(Equal(&s))
  })

  It("AWS Config with values", func() {
    e := "other-endpoint"
    r := "other-region"
    m := 99
    f := true
    s := false
    cfg := config.AwsConfig(&config.S3Config{
      Endpoint:       &e,
      Region:         &r,
      MaxRetries:     &m,
      ForcePathStyle: &f,
      SslEnabled:     &s,
    })
    Expect(cfg.Endpoint).To(Equal(&e))
    Expect(cfg.Region).To(Equal(&r))
    Expect(cfg.MaxRetries).To(Equal(&m))
    Expect(cfg.S3ForcePathStyle).To(Equal(&f))
    Expect(cfg.DisableSSL).ToNot(Equal(&s))
  })
})
