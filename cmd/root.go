/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
  "fmt"
  "os"

  "github.com/spf13/cobra"

  "s3/src/version"
)

type StringBool struct {
  S string
}

type OptionsCli struct {
  endpoint string
  bucket   string
  list     bool
  backup   []string
  zip      bool
  zipname  string
  replace  bool
  create   bool
  folder   string
  delete   []string
  mysql    bool
  version  bool
  args     []string
}

var options OptionsCli

var RootCmd = &cobra.Command{
  Use:   "gos3",
  Short: fmt.Sprintf("Usage of go s3 [v%s] in command line", version.Gos3Version),
  Long:  fmt.Sprintf("Help for go s3 v%s", version.Gos3Version),
  Run: func(cmd *cobra.Command, args []string) {
    options.args = args
    startCli(cmd, options)
  },
  Example: `  1. List files in "sample" bucket.                                                                             $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -l
  2. Backup multiple files to "backupFolder" folder.                                                            $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/index.ts -b images/logo.png -f backupFolder
  3. Backup files using wildcard to "backup" folder.                                                            $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -f backup
  4. Backup files using wildcard and zip into "zipped" folder, bucket will be created if it doesn't exists.     $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z -f zipped -c
  5. Backup files using wildcard and zip using "allfiles.zip" as filename into "zipped" folder, bucket will     $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -n allfiles.zip -f zipped -c -r
  be created if it doesn't exists and zipfile will be replaced if it exists
  6. Delete files in "uploads" folder older than 2days and files in "monthly" folder older than 1month          $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -d uploads=2d -d monthly=1M
  7. Delete files in "uploads" folder older than 1minute                                                        $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f uploads -d 1m
  8. Generate mysql dump file zip it and upload to "mysql-backup" folder                                        $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f mysql-backup -m -z`,
}

func Execute() {
  options = OptionsCli{}
  if err := RootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
  RootCmd.PersistentFlags().StringVarP(&options.endpoint, "endpoint", "e", "", "Destination url (can be defined by $ENDPOINT env variable)")
  RootCmd.PersistentFlags().StringVar(&options.bucket, "bucket", "", "Destination bucket (can be defined by $BUCKET env variable)")
  RootCmd.PersistentFlags().BoolVarP(&options.list, "list", "l", false, "List all files")
  RootCmd.PersistentFlags().StringArrayVarP(&options.backup, "backup", "b", []string{}, "Backup files")
  RootCmd.PersistentFlags().BoolVarP(&options.zip, "zip", "z", false, "Zip backup files")
  RootCmd.PersistentFlags().StringVarP(&options.zipname, "zipname", "n", "", "Zip name for backup files")
  RootCmd.PersistentFlags().BoolVarP(&options.replace, "replace", "r", false, "Replace files if already exists when backup upload")
  RootCmd.PersistentFlags().BoolVarP(&options.create, "create", "c", false, "Create destination upload bucket")
  RootCmd.PersistentFlags().StringVarP(&options.folder, "folder", "f", "", "Folder name to upload file/s")
  RootCmd.PersistentFlags().StringArrayVarP(&options.delete, "delete", "d", []string{}, "Clean files older than duration in foldername")
  RootCmd.PersistentFlags().BoolVarP(&options.mysql, "mysql", "m", false, "Mysql backup using environment variables to connect mysql server ($MYSQL_USER, $MYSQL_PASSWORD, $MYSQL_DATABASE, $MYSQL_HOST, $MYSQL_PORT)")
  RootCmd.PersistentFlags().BoolVarP(&options.version, "version", "v", false, "Go s3 version")
}
