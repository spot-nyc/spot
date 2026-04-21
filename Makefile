.PHONY: all test lint build snapshot clean test-install

all: test lint build

test:
	go test ./... -race -coverprofile=coverage.out

lint:
	golangci-lint run ./...

build:
	go build -o dist/spot ./cmd/spot

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf dist/ coverage.out

test-install:
	sh scripts/test-install.sh
