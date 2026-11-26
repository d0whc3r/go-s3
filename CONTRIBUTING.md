# Contributing to go-s3

Thank you for considering contributing to **go-s3**! This guide covers everything you need to set up your local development environment, run tests, and submit pull requests.

## Prerequisites

- **Go** `1.27` or newer (matching `go.mod`)
- **Docker** and **Docker Compose** (optional, only required for integration tests against MinIO and MySQL)
- **golangci-lint** (optional, recommended for local linting)

## Project Structure

```text
.
├── cmd/                # Cobra CLI commands and flag wiring
├── internal/
│   ├── config/         # Environment & CLI configuration, AWS SDK v2 S3 client factory
│   ├── gos3/           # Core S3 operations, zip archiving, duration parser, MySQL dump
│   └── version/        # Build-time version metadata
├── tests/              # Integration tests (build tag: integration) and sample fixtures
├── Dockerfile          # Multi-stage container build (distroless non-root runtime)
├── docker-compose.yml  # Local MinIO, MySQL, PostgreSQL, and MongoDB stack for integration testing
└── Makefile            # Common build, test, and lint targets
```

## Development Workflow

### 1. Clone and install dependencies

```bash
git clone https://github.com/d0whc3r/go-s3.git
cd go-s3
go mod download
```

### 2. Run unit tests and static checks

Unit tests use an in-memory S3 mock and require no external services:

```bash
make vet
make test
```

If you have `golangci-lint` installed:

```bash
make lint
```

### 3. Run integration tests (MinIO + MySQL)

1. Copy the test environment template:

   ```bash
   cp test-example.env test.env
   ```

2. Start the local MinIO and MySQL containers:

   ```bash
   make docker-up
   ```

3. Run the integration test suite:

   ```bash
   make test-integration
   ```

4. Stop the containers when finished:

   ```bash
   make docker-down
   ```

### 4. Build locally

Build a binary for your current platform:

```bash
make build
./build/gos3 --help
```

Build the Docker image locally:

```bash
docker build -t gos3:dev --build-arg VERSION=dev .
```

Cross-compile release binaries for all supported platforms:

```bash
make build-all
```

## Coding Guidelines

- Follow standard Go idioms (`gofmt`, `goimports`, `go vet`).
- Return wrapped errors (`fmt.Errorf("context: %w", err)`) instead of calling `log.Fatal` or `os.Exit` inside library packages.
- Propagate `context.Context` through all network and database operations.
- Never log secrets (`SECRET_KEY`, `MYSQL_PASSWORD`, etc.), even when `--debug` is enabled.
- Add unit tests for every bug fix and new feature.

## Commit Messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

- `feat: add support for ...`
- `fix: handle ...`
- `docs: update ...`
- `refactor: simplify ...`
- `test: add unit tests for ...`
- `chore(deps): update ...`

## Pull Request Process

1. Fork the repository and create a feature branch from `master`.
2. Make your changes with clear, focused commits.
3. Ensure `make vet` and `make test` pass without race conditions.
4. Open a Pull Request describing the motivation, changes, and how you tested them.

## Code of Conduct

All contributors are expected to follow our [Code of Conduct](./CODE_OF_CONDUCT.md).
