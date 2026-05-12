.PHONY: build test lint fmt cover tidy run clean install release

BINARY := bin/mac-dev-station
PKG := ./...
VERSION ?= v0.0.0-dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/mac-dev-station

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

install: build
	cp $(BINARY) /opt/homebrew/bin/mac-dev-station

release:
	@test -n "$(VERSION)" || (echo "VERSION=vX.Y.Z required" && exit 1)
	git tag -a $(VERSION) -m "$(VERSION)"
	git push origin $(VERSION)
	@echo "→ GoReleaser workflow will publish via GitHub Actions"
