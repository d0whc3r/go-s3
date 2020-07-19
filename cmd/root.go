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
)

var (
	endpoint string
	bucket   string
	list     bool
	backup   []string
	zip      string
	replace  bool
	create   bool
	folder   string
	del      []string
	mysql    bool
)

type OptionsCli struct {
	endpoint string
	bucket   string
	list     bool
	backup   []string
	zip      string
	replace  bool
	create   bool
	folder   string
	delete   []string
	mysql    bool
	args     []string
}

var rootCmd = &cobra.Command{
	Use:   "gos3",
	Short: "Usage of go s3 in command line",
	Long:  "Help for go s3",
	Run: func(cmd *cobra.Command, args []string) {
		parsedArgs := OptionsCli{
			endpoint: endpoint,
			bucket:   bucket,
			list:     list,
			backup:   backup,
			zip:      zip,
			replace:  replace,
			create:   create,
			folder:   folder,
			delete:   del,
			mysql:    mysql,
			args:     args,
		}
		startCli(cmd, parsedArgs)
	},
	Example: `  1. List files in "sample" bucket.                                                                             $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -l
  2. Backup multiple files to "backupFolder" folder.                                                            $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/index.ts -b images/logo.png -f backupFolder
  3. Backup files using wildcard to "backup" folder.                                                            $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -f backup
  4. Backup files using wildcard and zip into "zipped" folder, bucket will be created if it doesn't exists.     $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z -f zipped.zip -c
  5. Backup files using wildcard and zip using "allfiles.zip" as filename into "zipped" folder, bucket will     $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z allfiles.zip -f zipped -c -r
  be created if it doesn't exists and zipfile will be replaced if it exists                                                                                                                                     
  6. Delete files in "uploads" folder older than 2days and files in "monthly" folder older than 1month          $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -d uploads=2d -d monthly=1M
  7. Delete files in "uploads" folder older than 1minute                                                        $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f uploads -d 1m
  8. Generate mysql dump file zip it and upload to "mysql-backup" folder                                        $ gos3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f mysql-backup -m -z`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Destination url (can be defined by $ENDPOINT env variable)")
	rootCmd.PersistentFlags().StringVar(&bucket, "bucket", "", "Destination bucket (can be defined by $BUCKET env variable)")
	rootCmd.PersistentFlags().BoolVarP(&list, "list", "l", false, "List all files")
	rootCmd.PersistentFlags().StringArrayVarP(&backup, "backup", "b", []string{}, "Backup files")
	rootCmd.PersistentFlags().StringVarP(&zip, "zip", "z", "", "Zip backup files")
	rootCmd.PersistentFlags().BoolVarP(&replace, "replace", "r", false, "Replace files if already exists when backup upload")
	rootCmd.PersistentFlags().BoolVarP(&create, "create", "c", false, "Create destination upload bucket")
	rootCmd.PersistentFlags().StringVarP(&folder, "folder", "f", "", "Folder name to upload file/s")
	rootCmd.PersistentFlags().StringArrayVarP(&del, "delete", "d", []string{}, "Clean files older than duration in foldername")
	rootCmd.PersistentFlags().BoolVarP(&mysql, "mysql", "m", false, "Mysql backup using environment variables to connect mysql server ($MYSQL_USER, $MYSQL_PASSWORD, $MYSQL_DATABASE, $MYSQL_HOST, $MYSQL_PORT)")
}
