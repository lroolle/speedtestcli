BINARY_NAME = speedtest
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION_PKG = github.com/lroolle/speedtestcli/internal/version
LDFLAGS = -ldflags "-s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).GitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PKG).BuildDate=$(BUILD_DATE)"

.PHONY: build install test lint clean cross

build:
	go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) .

install: build
	@if [ -n "$(GOPATH)" ]; then \
		cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME); \
	else \
		cp $(BINARY_NAME) ~/go/bin/$(BINARY_NAME); \
	fi

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

cross:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME)-windows-amd64.exe .
