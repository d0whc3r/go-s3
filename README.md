# gos3

A lightweight command-line utility for backing up files, directories, and MySQL databases to any S3-compatible storage (AWS S3, MinIO, Cloudflare R2, DigitalOcean Spaces, Backblaze B2, etc.), with built-in zip compression and automated retention cleanup.

## Features

- **File & folder backups** with wildcard/glob support and optional zip compression.
- **MySQL database backups** dumped directly and uploaded (raw `.sql` or `.zip`).
- **Retention cleanup** to delete objects older than a given age (`30m`, `2d`, `1w`, `1M`, `1y`, or compound durations like `1M15d`).
- **S3-compatible** — works with AWS S3, MinIO, Cloudflare R2, DigitalOcean Spaces, and other S3 APIs.
- **Single static binary & minimal Docker image** — easy to run in cron jobs, servers, or CI pipelines.

## Installation

### Download prebuilt binary

Download the latest binary for your OS and architecture from the [GitHub Releases](https://github.com/d0whc3r/go-s3/releases) page, make it executable, and place it on your `PATH`:

```bash
chmod +x gos3
sudo mv gos3 /usr/local/bin/gos3
```

### Using Docker

```bash
docker run --rm d0whc3r/gos3 --help
```

### Using Go

```bash
go install github.com/d0whc3r/go-s3@latest
```

## Configuration

`gos3` can be configured via environment variables, command-line flags, or both. Command-line flags take precedence over environment variables. See [example.env](./example.env) for a ready-to-use template.

### S3 Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `ENDPOINT` | Yes (or `-e`) | — | S3-compatible endpoint URL (e.g. `https://s3.eu-central-1.amazonaws.com` or `http://localhost:9000`) |
| `BUCKET` | Yes (or `--bucket`) | — | Target bucket name |
| `ACCESS_KEY` | Yes\* | — | S3 access key ID |
| `SECRET_KEY` | Yes\* | — | S3 secret access key |
| `REGION` | No | `us-east-1` | S3 region (or `--region`) |
| `MAX_RETRIES` | No | `3` | Maximum retry attempts on transient S3 errors |
| `FORCE_PATH_STYLE` | No | `true` | Use path-style S3 URLs (`true` for MinIO / self-hosted S3) |
| `SSL_ENABLED` | No | `false` | Default to `https://` when `ENDPOINT` omits a URL scheme |
| `DEBUG` | No | `false` | Enable verbose debug logs to stderr (or `--debug`) |

\* If `ACCESS_KEY` and `SECRET_KEY` are omitted, `gos3` falls back to the standard AWS credential chain (shared credentials file, IAM role, etc.).

### MySQL Environment Variables (for `-m` / `--mysql`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MYSQL_HOST` | No | `localhost` | MySQL server hostname |
| `MYSQL_PORT` | No | `3306` | MySQL server port |
| `MYSQL_USER` | Yes | — | MySQL username |
| `MYSQL_PASSWORD` | Yes | — | MySQL password |
| `MYSQL_DATABASE` | Yes | — | Database name to dump |

## Usage

```text
gos3 [flags]
```

### Flags

| Flag | Short | Description |
| --- | --- | --- |
| `--endpoint <url>` | `-e` | S3-compatible endpoint URL (overrides `$ENDPOINT`) |
| `--bucket <name>` | | Destination bucket name (overrides `$BUCKET`) |
| `--region <region>` | | S3 region (overrides `$REGION`) |
| `--list` | `-l` | List objects in the bucket (optionally filtered by `--folder`) |
| `--backup <path>` | `-b` | File, directory, or glob pattern to upload (repeatable) |
| `--folder <prefix>` | `-f` | Destination folder prefix inside the bucket |
| `--zip` | `-z` | Compress backup files into a timestamped `.zip` archive |
| `--zipname <name>` | `-n` | Custom `.zip` archive filename (implies `--zip`) |
| `--replace` | `-r` | Overwrite existing objects in S3 |
| `--create` | `-c` | Create the destination bucket if it does not exist |
| `--delete <spec>` | `-d` | Delete files older than a duration (e.g. `7d`, `1M`, or `folder=7d`; repeatable) |
| `--mysql` | `-m` | Dump the configured MySQL database and upload it |
| `--debug` | | Print detailed debug logs to stderr |
| `--version` | `-v` | Print the installed version and exit |
| `--help` | `-h` | Show help message |

### Supported Duration Formats (`-d` / `--delete`)

| Unit | Aliases | Meaning |
| --- | --- | --- |
| Seconds | `s`, `sec`, `second`, `seconds` | Seconds |
| Minutes | `m`, `min`, `minute`, `minutes` | Minutes |
| Hours | `h`, `hr`, `hour`, `hours` | Hours |
| Days | `d`, `day`, `days` | Calendar days |
| Weeks | `w`, `week`, `weeks` | Weeks (7 days) |
| Months | `M`, `mo`, `month`, `months` | Calendar months |
| Years | `y`, `yr`, `year`, `years` | Calendar years |

> **Note:** Lowercase `m` means **minutes** and uppercase `M` means **months**. You can also combine units, such as `2d12h` or `1M15d`.

## Examples

### List files in a bucket

```bash
gos3 -e https://s3.eu-central-1.amazonaws.com --bucket my-backups -l
```

List files inside a specific folder:

```bash
gos3 --bucket my-backups -f daily -l
```

### Back up files and directories

Upload specific files to a folder:

```bash
gos3 --bucket my-backups -b /etc/nginx/nginx.conf -b /var/log/app.log -f server-config
```

Upload files matching a wildcard (quote globs so `gos3` expands them):

```bash
gos3 --bucket my-backups -b "logs/*.log" -f logs
```

### Back up as a zip archive

Create a timestamped zip archive and create the bucket if it does not exist:

```bash
gos3 --bucket my-backups -b ./data -z -f archives -c
```

Create a zip with a fixed filename, overwriting the previous archive:

```bash
gos3 --bucket my-backups -b ./data -n latest.zip -f archives -c -r
```

### Back up a MySQL database

Dump a MySQL database, compress it as a `.zip`, and upload to `mysql-backups/`:

```bash
export MYSQL_HOST=localhost MYSQL_PORT=3306 MYSQL_USER=backup MYSQL_PASSWORD=secret MYSQL_DATABASE=app
gos3 --bucket my-backups -f mysql-backups -m -z
```

### Clean up old backups (retention policy)

Delete files older than 2 days in `daily/` and older than 1 month in `monthly/`:

```bash
gos3 --bucket my-backups -d daily=2d -d monthly=1M
```

Delete files older than 30 minutes in `temp/`:

```bash
gos3 --bucket my-backups -f temp -d 30m
```

### Troubleshoot with debug output

```bash
gos3 --bucket my-backups -b ./report.pdf -f reports --debug
```

### Run with Docker

Mount the files you want to back up and pass your environment file:

```bash
docker run --rm \
  --env-file .env \
  -v "$(pwd)/data:/data:ro" \
  d0whc3r/gos3 -b /data -f daily -z
```

## Alternatives

Looking for a Node.js implementation with a similar interface? Check out [node-s3](https://github.com/d0whc3r/node-s3).

## Community & Support

- Found a bug or want to request a feature? Open an [issue](https://github.com/d0whc3r/go-s3/issues).
- Want to contribute? Read [CONTRIBUTING.md](./CONTRIBUTING.md) and our [Code of Conduct](./CODE_OF_CONDUCT.md).
- Need to report a security vulnerability? See [SECURITY.md](./SECURITY.md).

## License

Licensed under the [Apache License, Version 2.0](./LICENSE).
