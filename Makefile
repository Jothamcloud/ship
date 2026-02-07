.PHONY: build clean install test lint

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/jothamarinze/shipfe/pkg/version.Version=$(VERSION) \
	-X github.com/jothamarinze/shipfe/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/jothamarinze/shipfe/pkg/version.BuildDate=$(BUILD_DATE)"

# Output directory
DIST_DIR := dist

# Default target
all: build

# Build for current platform
build:
	go build $(LDFLAGS) -o $(DIST_DIR)/shipfe .

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/shipfe-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/shipfe-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/shipfe-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/shipfe-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/shipfe-windows-amd64.exe .

# Install locally
install:
	go install $(LDFLAGS) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Verify dependencies
verify:
	go mod verify
