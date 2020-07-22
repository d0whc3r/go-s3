package cmd

import (
  "fmt"
  "io"
  "os"
  "regexp"

  "github.com/aws/aws-sdk-go/service/s3"
  "github.com/spf13/cobra"

  "s3/src/config"
  "s3/src/gos3"
  "s3/src/version"
  "s3/src/wrapper"
)

const tag = "[go-s3]"

var s3Wrapper wrapper.S3Wrapper
var out io.Writer

func startCli(cmd *cobra.Command, o OptionsCli) {
  out = cmd.OutOrStdout()
  if cap(o.args) == 0 {
    _ = cmd.Help()
    return
  }
  s3Wrapper = wrapper.New(&config.S3Config{
    Bucket:   &o.bucket,
    Endpoint: &o.endpoint,
  })
  switch {
  case o.version:
    cliVersion()
  case len(o.backup) > 0:
    cliBackup(o.backup, o.folder, o.replace, o.create, o.zip, o.zipname)
    fallthrough
  case len(o.delete) > 0:
    cliDelete(o.delete, o.folder)
    fallthrough
  case o.mysql:
    cliMysql(o.folder, o.replace, o.create, o.zip, o.zipname)
    fallthrough
  case o.list:
    cliList()
  }
}

func cliVersion() {
  fmt.Fprintf(out, "v%s\n", version.Gos3Version)
}

func cliMysql(folder string, replace bool, create bool, zip bool, zipname string) {
  var z interface{}
  if zipname != "" {
    z = zipname
  } else {
    z = zip
  }
  err := s3Wrapper.UploadMysql(folder, &gos3.UploadOptions{
    Create:   &create,
    Replace:  &replace,
    Compress: &z,
  }, nil)
  if err != nil {
    os.Exit(1)
  }
}

func cliDelete(deletes []string, folder string) {
  r := regexp.MustCompile(`(\w+)=(.*)`)
  for _, d := range deletes {
    var err error
    if r.MatchString(d) {
      match := r.FindStringSubmatch(d)[1:]
      _, err = s3Wrapper.CleanOlder(match[0], match[1], nil)
    } else {
      _, err = s3Wrapper.CleanOlder(deletes[0], folder, nil)
    }
    if err != nil {
      os.Exit(1)
    }
  }

}

func cliBackup(files []string, folder string, replace bool, create bool, zip bool, zipname string) {
  var z interface{}
  if zipname != "" {
    z = zipname
  } else {
    z = zip
  }
  err := s3Wrapper.UploadFiles(files, folder, &gos3.UploadOptions{Create: &create, Replace: &replace, Compress: &z}, nil)
  if err != nil {
    os.Exit(1)
  }
}

func cliList() {
  f, err := s3Wrapper.GetFiles(nil)
  if err != nil {
    os.Exit(1)
  }
  if len(f) > 0 {
    fmt.Fprintf(out, "%s File list in bucket '%s': %d\n", tag, s3Wrapper.Bucket, len(f))
    showFiles(f)
  } else {
    fmt.Fprintf(out, "%s No files found in bucket '%s'\n", tag, s3Wrapper.Bucket)
  }
}

func showFiles(f []*s3.Object) {
  for _, fo := range f {
    fmt.Fprintf(out, *fo.Key)
  }
}
