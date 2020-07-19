GOS3_VERSION=$(shell cat .version)

# These tests are for development
test:
	make -j5 test-go vet-go

test-go:
	go test s3/tests

vet-go:
	go vet

update-version-go:
	echo "package main\n\nconst gos3Version = \"$(GOS3_VERSION)\"" > version.go

platform-all: clean update-version-go test
	make -j9 \
		platform-windows \
		platform-darwin \
		platform-freebsd \
		platform-freebsd-arm64 \
		platform-linux \
		platform-linux-arm64 \
		platform-linux-ppc64le

platform-windows:
	GOOS=windows GOARCH=amd64 go build "-ldflags=-s -w" -o build/gos3-windows-64/gos3.exe

platform-unixlike:
	test -n "$(GOOS)" && test -n "$(GOARCH)" && test -n "$(BUILDDIR)"
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build "-ldflags=-s -w" -o "$(BUILDDIR)/gos3"

platform-darwin:
	make GOOS=darwin GOARCH=amd64 BUILDDIR=build/gos3-darwin-64 platform-unixlike

platform-freebsd:
	make GOOS=freebsd GOARCH=amd64 BUILDDIR=build/gos3-freebsd-64 platform-unixlike

platform-freebsd-arm64:
	make GOOS=freebsd GOARCH=arm64 BUILDDIR=build/gos3-freebsd-arm64 platform-unixlike

platform-linux:
	make GOOS=linux GOARCH=amd64 BUILDDIR=build/gos3-linux-64 platform-unixlike

platform-linux-arm64:
	make GOOS=linux GOARCH=arm64 BUILDDIR=build/gos3-linux-arm64 platform-unixlike

platform-linux-ppc64le:
	make GOOS=linux GOARCH=ppc64le BUILDDIR=build/gos3-linux-ppc64le platform-unixlike

publish-all: update-version-go test-all
	make -j8 \
		publish-windows \
		publish-darwin \
		publish-freebsd \
		publish-freebsd-arm64 \
		publish-linux \
		publish-linux-arm64 \
		publish-linux-ppc64le
	git commit -am "publish $(GOS3_VERSION)"
	git tag "v$(GOS3_VERSION)"
	git push origin master "v$(GOS3_VERSION)"

publish-windows: platform-windows
	test -n "$(OTP)" && cd build/gos3-windows-64 && npm publish --otp="$(OTP)"

publish-darwin: platform-darwin
	test -n "$(OTP)" && cd build/gos3-darwin-64 && npm publish --otp="$(OTP)"

publish-freebsd: platform-freebsd
	test -n "$(OTP)" && cd build/gos3-freebsd-64 && npm publish --otp="$(OTP)"

publish-freebsd-arm64: platform-freebsd-arm64
	test -n "$(OTP)" && cd build/gos3-freebsd-arm64 && npm publish --otp="$(OTP)"

publish-linux: platform-linux
	test -n "$(OTP)" && cd build/gos3-linux-64 && npm publish --otp="$(OTP)"

publish-linux-arm64: platform-linux-arm64
	test -n "$(OTP)" && cd build/gos3-linux-arm64 && npm publish --otp="$(OTP)"

publish-linux-ppc64le: platform-linux-ppc64le
	test -n "$(OTP)" && cd build/gos3-linux-ppc64le && npm publish --otp="$(OTP)"

clean:
	rm -rf build
	go clean -testcache ./tests/...
