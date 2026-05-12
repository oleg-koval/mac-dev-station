.PHONY: build test lint fmt cover tidy run clean

BINARY := bin/mac-dev-station
PKG := ./...

build:
	go build -o $(BINARY) ./cmd/mac-dev-station

run:
	go run ./cmd/mac-dev-station

test:
	go test -race -count=1 $(PKG)

cover:
	go test -race -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run $(PKG)

fmt:
	go fmt $(PKG)
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out coverage.html
