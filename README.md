# gos3

A lightweight command-line utility for backing up files, directories, and databases (**MySQL/MariaDB**, **PostgreSQL**, **SQLite**, and **MongoDB**) to any S3-compatible storage (AWS S3, MinIO, Cloudflare R2, DigitalOcean Spaces, Backblaze B2, etc.), with built-in zip compression and automated retention cleanup.

## Features

- **File & folder backups** with wildcard/glob support and optional zip compression.
- **Multi-database backups (zero external binaries required)**:
  - **MySQL / MariaDB** — SQL dump
  - **PostgreSQL** — pure-Go schema + data SQL dump
  - **SQLite** — consistent live snapshot via `VACUUM INTO`
  - **MongoDB** — Extended JSON dump of all collections
- **Retention cleanup** to delete objects older than a given age (`30m`, `2d`, `1w`, `1M`, `1y`, or compound durations like `1M15d`).
- **S3-compatible** — works with AWS S3, MinIO, Cloudflare R2, DigitalOcean Spaces, and other S3 APIs.
- **Single static binary & minimal Docker image** — easy to run in cron jobs, servers, or CI pipelines.

## Installation

### Quick install (Linux, macOS, FreeBSD)

```bash
curl -fsSL https://raw.githubusercontent.com/d0whc3r/go-s3/master/install.sh | sh
```

The installer detects your OS and architecture, verifies the SHA-256 checksum, and installs `gos3` to `/usr/local/bin`.

To install a specific version or use a custom directory:

```bash
curl -fsSL https://raw.githubusercontent.com/d0whc3r/go-s3/master/install.sh | VERSION=v2.0.0 INSTALL_DIR="$HOME/.local/bin" sh
```

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

### Updating

To update an existing installation to the latest release (with SHA-256 verification):

```bash
gos3 --update
```

*(If installed in `/usr/local/bin`, run `sudo gos3 --update`.)*

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

### Database Environment Variables

#### MySQL / MariaDB (`--mysql` / `-m` or `--db mysql` / `--db mariadb`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MYSQL_HOST` | No | `localhost` | MySQL / MariaDB hostname |
| `MYSQL_PORT` | No | `3306` | MySQL / MariaDB port |
| `MYSQL_USER` | Yes | — | Database username |
| `MYSQL_PASSWORD` | Yes | — | Database password |
| `MYSQL_DATABASE` | Yes | — | Database name to dump |

#### PostgreSQL (`--postgres` or `--db postgres`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `POSTGRES_URL` | No\* | — | Full connection URI (e.g. `postgres://user:pass@host:5432/db?sslmode=require`) |
| `POSTGRES_HOST` | No | `localhost` | PostgreSQL hostname |
| `POSTGRES_PORT` | No | `5432` | PostgreSQL port |
| `POSTGRES_USER` | Yes\* | — | Database username |
| `POSTGRES_PASSWORD` | No | — | Database password |
| `POSTGRES_DATABASE` | Yes\* | — | Database name to dump |
| `POSTGRES_SSLMODE` | No | `prefer` | SSL mode (`disable`, `prefer`, `require`, `verify-full`) |
| `POSTGRES_SCHEMA` | No | `public` | Schema to dump |

\* Provide either `POSTGRES_URL` or `POSTGRES_USER` + `POSTGRES_DATABASE`.

#### SQLite (`--sqlite <path>` or `--db sqlite`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `SQLITE_PATH` | Yes (or `--sqlite`) | — | Path to the local `.db` / `.sqlite` file |

#### MongoDB (`--mongodb` or `--db mongodb`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MONGO_URI` | No | — | Full MongoDB URI (e.g. `mongodb://user:pass@host:27017` or `mongodb+srv://...`) |
| `MONGO_HOST` | No | `localhost` | MongoDB hostname (used when `MONGO_URI` is empty) |
| `MONGO_PORT` | No | `27017` | MongoDB port |
| `MONGO_USER` | No | — | MongoDB username |
| `MONGO_PASSWORD` | No | — | MongoDB password |
| `MONGO_DATABASE` | Yes | — | Database name to dump |

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
| `--db <engine>` | | Database engine to dump: `mysql`, `mariadb`, `postgres`, `sqlite`, `mongodb` (repeatable) |
| `--mysql` | `-m` | Shorthand for `--db mysql` |
| `--postgres` | | Shorthand for `--db postgres` |
| `--sqlite <path>` | | Shorthand for `--db sqlite` with database file path |
| `--mongodb` | | Shorthand for `--db mongodb` |
| `--update` | | Self-update `gos3` to the latest GitHub Release |
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

### Back up databases

**MySQL / MariaDB:**

```bash
export MYSQL_HOST=localhost MYSQL_PORT=3306 MYSQL_USER=backup MYSQL_PASSWORD=secret MYSQL_DATABASE=app
gos3 --bucket my-backups -f mysql-backups --mysql -z
```

**PostgreSQL:**

```bash
export POSTGRES_URL="postgres://backup:secret@localhost:5432/app?sslmode=require"
gos3 --bucket my-backups -f pg-backups --postgres -z
```

**SQLite (live consistent snapshot):**

```bash
gos3 --bucket my-backups -f sqlite-backups --sqlite /var/lib/app/data.db -z
```

**MongoDB (Extended JSON):**

```bash
export MONGO_URI="mongodb://backup:secret@localhost:27017" MONGO_DATABASE=app
gos3 --bucket my-backups -f mongo-backups --mongodb -z
```

**Multiple databases in one run:**

```bash
gos3 --bucket my-backups -f db-backups --db postgres --db mongodb -z
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
