.PHONY: all test test-go test-python build build-all godoc godoc-html clean

all: test build

## Run all tests (Go & Python)
test: test-go test-python

## Run Go unit tests and examples
test-go:
	go vet ./...
	go test -v ./...

## Run Python unit tests
test-python: build
	PYTHONPATH=python python3 -m unittest tests/test_python_loader.py

## Build Go motion-parser binary
build:
	mkdir -p bin
	go build -v -o bin/motion-parser ./cmd/motion-parser

## Cross-compile Go binaries for all platforms
build-all:
	mkdir -p release_dist
	GOOS=linux GOARCH=amd64 go build -o release_dist/motion-parser-linux-amd64 ./cmd/motion-parser
	GOOS=linux GOARCH=arm64 go build -o release_dist/motion-parser-linux-arm64 ./cmd/motion-parser
	GOOS=darwin GOARCH=amd64 go build -o release_dist/motion-parser-darwin-amd64 ./cmd/motion-parser
	GOOS=darwin GOARCH=arm64 go build -o release_dist/motion-parser-darwin-arm64 ./cmd/motion-parser
	GOOS=windows GOARCH=amd64 go build -o release_dist/motion-parser-windows-amd64.exe ./cmd/motion-parser

## Run interactive local Godoc server (http://localhost:6060/pkg/github.com/Mireu-Lab/Load-motion-animation/)
godoc:
	./scripts/run_godoc.sh

## Generate static Godoc HTML documentation for GitLab Pages / offline viewing
godoc-html:
	./scripts/generate_godoc.sh public

## Clean built artifacts and documentation output
clean:
	rm -rf bin release_dist public python/dist python/build python/*.egg-info
