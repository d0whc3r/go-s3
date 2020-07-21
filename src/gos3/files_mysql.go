package gos3

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"

	"s3/src/config"
)

func createDumpFile() (*string, error) {
	conf := config.Config()
	host := conf.MysqlHost
	port := conf.MysqlPort
	user := conf.MysqlUser
	pass := conf.MysqlPassword
	dbname := conf.MysqlDatabase
	if user == "" || pass == "" || dbname == "" {
		return nil, errors.New("error in mysql-dump environment variables not defined: $MYSQL_USER, $MYSQL_PASSWORD, $MYSQL_DATABASE, $MYSQL_HOST, $MYSQL_PORT")
	}
	dir, err := ioutil.TempDir("", "s3mysql")
	if err != nil {
		return nil, err
	}
	mdfile := fmt.Sprintf("mysqldump-%s-%s", dbname, time.Now().Format("2006-01-02.150405"))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, dbname))
	if err != nil {
		return nil, err
	}
	dumper, err := mysqldump.Register(db, dir, mdfile)
	defer dumper.Close()
	if err != nil {
		return nil, err
	}
	result, err := dumper.Dump()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(result)
	defer file.Close()
	return &result, err
}

func (m S3Manager) UploadMysql(bucket string, folder string, options *UploadOptions) error {
	dump, err := createDumpFile()
	if err != nil {
		return err
	}
	return m.UploadFiles(bucket, []string{*dump}, folder, options)
}
