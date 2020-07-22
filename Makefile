GOS3_VERSION=$(shell cat .version)

docker:
	docker-compose up -d

docker-down:
	docker-compose down

test:
	make -j2 test-go vet-go

# go test -cover -coverprofile=coverage.out s3/...
# ginkgo -p -cover -outputdir=coverage ./...
test-go:
	go test -cover -coverprofile=coverage/coverage.coverprofile ./...

vet-go:
	go vet

update-version-go:
	echo "package main\n\nconst gos3Version = \"$(GOS3_VERSION)\"" > version.go

build:
	make build-linux

build-all: clean
	make -j9 \
		build-windows \
		build-darwin \
		build-freebsd \
		build-freebsd-arm64 \
		build-linux \
		build-linux-arm64 \
		build-linux-arm \
		build-linux-ppc64le

build-windows:
	GOOS=windows GOARCH=amd64 go build "-ldflags=-s -w" -o "build/gos3_v${GOS3_VERSION}_windows-amd64.exe"

build-unixlike:
	test -n "$(GOOS)" && test -n "$(GOARCH)" && test -n "$(BUILDPATH)"
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build "-ldflags=-s -w" -o "$(BUILDPATH)"

build-darwin:
	make GOOS=darwin GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_darwin-amd64" build-unixlike

build-freebsd:
	make GOOS=freebsd GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_freebsd-amd64" build-unixlike

build-freebsd-arm64:
	make GOOS=freebsd GOARCH=arm64 BUILDPATH="build/gos3_v${GOS3_VERSION}_freebsd-arm64" build-unixlike

build-linux:
	make GOOS=linux GOARCH=amd64 BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-amd64" build-unixlike

build-linux-arm64:
	make GOOS=linux GOARCH=arm64 BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-arm64" build-unixlike

build-linux-arm:
	make GOOS=linux GOARCH=arm BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-arm" build-unixlike

build-linux-ppc64le:
	make GOOS=linux GOARCH=ppc64le BUILDPATH="build/gos3_v${GOS3_VERSION}_linux-ppc64le" build-unixlike

clean:
	rm -rf build
	go clean -testcache ./...
