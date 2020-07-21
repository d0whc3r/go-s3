GOS3_VERSION=$(shell cat .version)

# These tests are for development
test:
	make -j5 test-go vet-go

test-go:
	go test -cover -coverprofile=coverage.out s3/src/...

vet-go:
	go vet

update-version-go:
	echo "package main\n\nconst gos3Version = \"$(GOS3_VERSION)\"" > version.go

build:
	make platform-linux

build-all: clean
	make -j9 \
		platform-windows \
		platform-darwin \
		platform-freebsd \
		platform-freebsd-arm64 \
		platform-linux \
		platform-linux-arm64 \
		platform-linux-arm \
		platform-linux-ppc64le

platform-windows:
	GOOS=windows GOARCH=amd64 go build "-ldflags=-s -w" -o "build/gos3_v${GOS3_VERSION}_windows-amd64.exe"

platform-unixlike:
	test -n "$(GOOS)" && test -n "$(GOARCH)" && test -n "$(BUILDPATH)"
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build "-ldflags=-s -w" -o "$(BUILDPATH)"

platform-darwin:
	make GOOS=darwin GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_darwin-amd64" platform-unixlike

platform-freebsd:
	make GOOS=freebsd GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_freebsd-amd64" platform-unixlike

platform-freebsd-arm64:
	make GOOS=freebsd GOARCH=arm64 BUILDPATH="build/gos3_v${GOS3_VERSION}_freebsd-arm64" platform-unixlike

platform-linux:
	make GOOS=linux GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-amd64" platform-unixlike

platform-linux-arm64:
	make GOOS=linux GOARCH=arm64 BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-arm64" platform-unixlike

platform-linux-arm:
	make GOOS=linux GOARCH=arm BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-arm" platform-unixlike

platform-linux-ppc64le:
	make GOOS=linux GOARCH=ppc64le BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-ppc64le" platform-unixlike

clean:
	rm -rf build
	go clean -testcache ./tests/...
