ver ?= 0.0.0
bin = sq
sources = $(wildcard *.go) $(wildcard ../../cli/*.go) $(wildcard ../../builtin/*.go)

.PHONY: all
all: test $(bin)

.PHONY: build-for-release
build-for-release: test $(bin) # the dependency $(bin) is required to build the binary to run on the local env for e2e-test
	GOOS=linux   GOARCH=amd64   go build -o dist/$(ver)/linux-amd64/$(bin)
	GOOS=linux   GOARCH=arm64   go build -o dist/$(ver)/linux-arm64/$(bin)
	GOOS=linux   GOARCH=arm     go build -o dist/$(ver)/linux-arm/$(bin)
	GOOS=linux   GOARCH=riscv64 go build -o dist/$(ver)/linux-riscv64/$(bin)

.PHONY: package
package:
	cp ../../README.md ../../LICENSE dist/$(ver)/linux-amd64/
	cp ../../README.md ../../LICENSE dist/$(ver)/linux-arm64/
	cp ../../README.md ../../LICENSE dist/$(ver)/linux-arm/
	cp ../../README.md ../../LICENSE dist/$(ver)/linux-riscv64/

	mkdir -p dist/$(ver)/pkg

	tar -C dist/$(ver)/linux-amd64   -zcf dist/$(ver)/pkg/$(bin)_$(ver)_linux_amd64.tar.gz   $(bin) README.md LICENSE
	tar -C dist/$(ver)/linux-arm64   -zcf dist/$(ver)/pkg/$(bin)_$(ver)_linux_arm64.tar.gz   $(bin) README.md LICENSE
	tar -C dist/$(ver)/linux-arm     -zcf dist/$(ver)/pkg/$(bin)_$(ver)_linux_arm.tar.gz     $(bin) README.md LICENSE
	tar -C dist/$(ver)/linux-riscv64 -zcf dist/$(ver)/pkg/$(bin)_$(ver)_linux_riscv64.tar.gz $(bin) README.md LICENSE

.PHONY: release
release:
	ghr -prerelease -replace -owner bitbears-dev -repository $(bin) v$(ver) dist/$(ver)/pkg/

$(bin): $(sources)
	go build

.PHONY: test
test:
	go test

.PHONY: clean
clean:
	rm -f $(bin)
	rm -rf dist/
