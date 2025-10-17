.PHONY: install watch build test clean

VERSION := $(shell cat VERSION)
LDFLAGS := -X github.com/cameron-webmatter/galaxy/pkg/cli.Version=$(VERSION)

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/galaxy

watch:
	watchexec -w pkg -w cmd -w internal -e go -- ./scripts/build.sh

build:
	go build -ldflags "$(LDFLAGS)" -o galaxy ./cmd/galaxy

test:
	go test ./...

clean:
	rm -f galaxy
	rm -rf tmp/
