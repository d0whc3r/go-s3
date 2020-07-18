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
	// cfgFile  string
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gos3",
	Short: "Usage of go s3 in command line",
	Long:  "Help for go s3",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: startCli,
	Example: `  1. List files in "sample" bucket.                                                                             $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -l
  2. Backup multiple files to "backupFolder" folder.                                                            $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/index.ts -b images/logo.png -f backupFolder
  3. Backup files using wildcard to "backup" folder.                                                            $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -f backup
  4. Backup files using wildcard and zip into "zipped" folder, bucket will be created if it doesn't exists.     $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z -f zipped.zip -c
  5. Backup files using wildcard and zip using "allfiles.zip" as filename into "zipped" folder, bucket will     $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z allfiles.zip -f zipped -c -r
  be created if it doesn't exists and zipfile will be replaced if it exists                                                                                                                                     
  6. Delete files in "uploads" folder older than 2days and files in "monthly" folder older than 1month          $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -d uploads=2d -d monthly=1M
  7. Delete files in "uploads" folder older than 1minute                                                        $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f uploads -d 1m
  8. Generate mysql dump file zip it and upload to "mysql-backup" folder                                        $ node-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f mysql-backup -m -z`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.s3.yaml)")
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Find home directory.
// 		home, err := homedir.Dir()
// 		if err != nil {
// 			fmt.Println(err)
// 			os.Exit(1)
// 		}
//
// 		// Search config in home directory with name ".s3" (without extension).
// 		viper.AddConfigPath(home)
// 		viper.SetConfigName(".s3")
// 	}
//
// 	viper.AutomaticEnv() // read in environment variables that match
//
// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Println("Using config file:", viper.ConfigFileUsed())
// 	}
// }
