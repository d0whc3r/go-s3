
# :floppy_disk: Go s3

Utility to use s3 storage using golang. It could upload content to s3 storage to use as backup service.

# :eyes: Project status

[![Actions Status](https://github.com/d0whc3r/node-s3/workflows/Nodejs/badge.svg)](https://github.com/d0whc3r/node-s3/actions)
[![pipeline status](https://gitlab.com/d0whc3r/node-s3/badges/master/pipeline.svg)](https://github.com/d0whc3r/node-s3)

[![codecov](https://codecov.io/gh/d0whc3r/node-s3/branch/master/graph/badge.svg)](https://codecov.io/gh/d0whc3r/node-s3)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/0ff5030a11234255a13694d619b5bcf6)](https://www.codacy.com/manual/d0whc3r/node-s3?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=d0whc3r/node-s3&amp;utm_campaign=Badge_Grade)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=d0whc3r_node-s3&metric=alert_status)](https://sonarcloud.io/dashboard?id=d0whc3r_node-s3)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=d0whc3r_node-s3&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=d0whc3r_node-s3)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=d0whc3r_node-s3&metric=security_rating)](https://sonarcloud.io/dashboard?id=d0whc3r_node-s3)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=d0whc3r_node-s3&metric=bugs)](https://sonarcloud.io/dashboard?id=d0whc3r_node-s3)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=d0whc3r_node-s3&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=d0whc3r_node-s3)

![dependencies](https://img.shields.io/david/d0whc3r/node-s3.svg)
[![npm version](https://img.shields.io/npm/v/@d0whc3r%2Fnode-s3.svg)](https://www.npmjs.com/package/@d0whc3r/node-s3)

[![](https://img.shields.io/docker/cloud/build/d0whc3r/s3.svg)](https://hub.docker.com/r/d0whc3r/s3)
[![](https://images.microbadger.com/badges/version/d0whc3r/s3.svg)](https://hub.docker.com/r/d0whc3r/s3)
[![](https://images.microbadger.com/badges/image/d0whc3r/s3.svg)](https://hub.docker.com/r/d0whc3r/s3)

## :key: Create keys

`Access key` and `Secre key` should be defined in environment variables check [example.env](./example.env) for more info about environment variables

## :boat: Docker usage

You could use cli app in docker

### :rowboat: Build docker image

```bash
docker build -t s3 .
```

## :beginner: Environment variables

- `ENDPOINT`: Endpoint to connect, could be any s3 compatible endpoint (it could be defined in commandline with `--endpoint` or `-e`, example: http://s3.eu-central-1.amazonaws.com:9000)
- `ACCESS_KEY`: Access key to use (required)
- `SECRET_KEY`: Secret key to use (required)
- `BUCKET`: Bucket name to connect (it could be created using `-c` or it could be defined in commandline with `--bucket`)
- `MAX_RETRIES`: Maximum retry connections when fail (optional, example: 3)
- `FORCE_PATH_STYLE`: Force path style (optional, example: true)
- `SSL_ENABLED`: Enable ssl connection to endpoint (optional, example: false)

## :checkered_flag: Cli help output

### Docker usage

Using docker image from [hub.docker.com](https://hub.docker.com/r/d0whc3r/gos3)

```bash
docker run --rm d0whc3r/gos3 --help
```

### Cli usage

TODO

```
Help for go-s3

  Usage of go-s3 in command line. 

Options

  -e, --endpoint url                             Destination url (can be defined by $ENDPOINT env variable)                    
  --bucket bucket                                Destination bucket (can be defined by $BUCKET env variable)                   
  -l, --list                                     List all files                                                                
  -b, --backup file*                             Backup files                                                                  
  -z, --zip zipname.zip                          Zip backup files                                                              
  -r, --replace                                  Replace files if already exists when backup upload                            
  -c, --create                                   Create destination upload bucket                                              
  -f, --folder foldername                        Folder name to upload file/s                                                  
  -d, --delete foldername=duration OR duration   Clean files older than duration in foldername                                 
  -m, --mysql                                    Mysql backup using environment variables to connect mysql server              
                                                 ($MYSQL_USER, $MYSQL_PASSWORD, $MYSQL_DATABASE, $MYSQL_HOST, $MYSQL_PORT)     
  -h, --help                                     Print this usage guide.                                                       

Examples

  1. List files in "sample" bucket.                                                                             $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -l                                         
  2. Backup multiple files to "backupFolder" folder.                                                            $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/index.ts -b images/logo.png -f      
                                                                                                                backupFolder                                                                                                
  3. Backup files using wildcard to "backup" folder.                                                            $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -f backup             
  4. Backup files using wildcard and zip into "zipped" folder, bucket will be created if it doesn't exists.     $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z -f zipped.zip -c   
  5. Backup files using wildcard and zip using "allfiles.zip" as filename into "zipped" folder, bucket will     $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -b src/* -b images/* -z allfiles.zip -f    
  be created if it doesn't exists and zipfile will be replaced if it exists                                     zipped -c -r                                                                                                
  6. Delete files in "uploads" folder older than 2days and files in "monthly" folder older than 1month          $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -d uploads=2d -d monthly=1M                
  7. Delete files in "uploads" folder older than 1minute                                                        $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f uploads -d 1m                           
  8. Generate mysql dump file zip it and upload to "mysql-backup" folder                                        $ go-s3 -e http://s3.eu-central-1.amazonaws.com --bucket sample -f mysql-backup -m -z                      
```

## Alternatives

The same interface/api using nodejs: [node-s3](https://github.com/d0whc3r/node-s3)
