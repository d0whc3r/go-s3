GOS3_VERSION ?= $(shell cat .version)
LDFLAGS := -s -w -X github.com/d0whc3r/go-s3/internal/version.Version=$(GOS3_VERSION)

.PHONY: help build build-all test test-integration vet lint clean docker-up docker-down

help:
	@echo "Available targets:"
	@echo "  build             Build gos3 binary for the current OS/arch"
	@echo "  build-all         Cross-compile release binaries for all supported platforms"
	@echo "  test              Run unit tests with race detector and coverage"
	@echo "  test-integration  Run integration tests (requires MinIO and MySQL via docker-up)"
	@echo "  vet               Run go vet"
	@echo "  lint              Run golangci-lint"
	@echo "  docker-up         Start local MinIO and MySQL services"
	@echo "  docker-down       Stop local MinIO and MySQL services"
	@echo "  clean             Remove build artifacts and test cache"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

test:
	mkdir -p coverage
	go test -race -coverprofile=coverage/coverage.out ./...

test-integration:
	go test -race -tags=integration -v ./tests/...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o build/gos3 .

build-all: clean
	$(MAKE) -j8 \
		build-linux-amd64 \
		build-linux-arm64 \
		build-linux-arm \
		build-linux-ppc64le \
		build-darwin-amd64 \
		build-darwin-arm64 \
		build-freebsd-amd64 \
		build-freebsd-arm64 \
		build-windows-amd64

build-unixlike:
	@test -n "$(GOOS)" && test -n "$(GOARCH)" && test -n "$(BUILDPATH)"
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -trimpath -ldflags="$(LDFLAGS)" -o "$(BUILDPATH)" .

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o "build/gos3_v$(GOS3_VERSION)_windows-amd64.exe" .

build-darwin-amd64:
	$(MAKE) GOOS=darwin GOARCH=amd64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_darwin-amd64" build-unixlike

build-darwin-arm64:
	$(MAKE) GOOS=darwin GOARCH=arm64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_darwin-arm64" build-unixlike

build-freebsd-amd64:
	$(MAKE) GOOS=freebsd GOARCH=amd64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_freebsd-amd64" build-unixlike

build-freebsd-arm64:
	$(MAKE) GOOS=freebsd GOARCH=arm64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_freebsd-arm64" build-unixlike

build-linux-amd64:
	$(MAKE) GOOS=linux GOARCH=amd64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_linux-amd64" build-unixlike

build-linux-arm64:
	$(MAKE) GOOS=linux GOARCH=arm64 BUILDPATH="build/gos3_v$(GOS3_VERSION)_linux-arm64" build-unixlike

build-linux-arm:
	$(MAKE) GOOS=linux GOARCH=arm BUILDPATH="build/gos3_v$(GOS3_VERSION)_linux-arm" build-unixlike

build-linux-ppc64le:
	$(MAKE) GOOS=linux GOARCH=ppc64le BUILDPATH="build/gos3_v$(GOS3_VERSION)_linux-ppc64le" build-unixlike

clean:
	rm -rf build
	go clean -testcache
