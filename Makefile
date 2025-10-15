.PHONY: install watch build test clean

install:
	go install ./cmd/galaxy

watch:
	watchexec -w pkg -w cmd -w internal -e go -- go install ./cmd/galaxy

build:
	go build -o galaxy ./cmd/galaxy

test:
	go test ./...

clean:
	rm -f galaxy
	rm -rf tmp/
