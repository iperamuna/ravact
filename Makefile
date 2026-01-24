.PHONY: all build build-linux build-darwin build-all test test-coverage clean install docker-test help

# Binary name
BINARY_NAME=ravact
VERSION?=0.1.3
BUILD_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

all: test build

## build: Build binary for current platform
build:
	@echo "Building for current platform..."
	@echo "Copying assets for embedding..."
	@rm -rf cmd/ravact/assets
	@cp -r assets cmd/ravact/
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/ravact
	@rm -rf cmd/ravact/assets

## build-linux: Build for Linux amd64
build-linux:
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	@echo "Copying assets for embedding..."
	@rm -rf cmd/ravact/assets
	@cp -r assets cmd/ravact/
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ravact
	@rm -rf cmd/ravact/assets
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

## build-linux-arm64: Build for Linux arm64
build-linux-arm64:
	@echo "Building for Linux arm64..."
	@mkdir -p $(BUILD_DIR)
	@echo "Copying assets for embedding..."
	@rm -rf cmd/ravact/assets
	@cp -r assets cmd/ravact/
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/ravact
	@rm -rf cmd/ravact/assets
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

## build-darwin: Build for macOS (for testing)
build-darwin:
	@echo "Building for macOS amd64..."
	@mkdir -p $(BUILD_DIR)
	@echo "Copying assets for embedding..."
	@rm -rf cmd/ravact/assets
	@cp -r assets cmd/ravact/
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ravact
	@rm -rf cmd/ravact/assets
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

## build-darwin-arm64: Build for macOS arm64 (Apple Silicon)
build-darwin-arm64:
	@echo "Building for macOS arm64..."
	@mkdir -p $(BUILD_DIR)
	@echo "Copying assets for embedding..."
	@rm -rf cmd/ravact/assets
	@cp -r assets cmd/ravact/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ravact
	@rm -rf cmd/ravact/assets
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

## build-all: Build for all target platforms
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64
	@echo "All builds complete!"
	@ls -lh $(BUILD_DIR)/

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-integration: Run integration tests (requires Linux or Docker)
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## install: Install binary to GOPATH/bin
install:
	@echo "Installing..."
	$(GOCMD) install $(LDFLAGS) ./cmd/ravact

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## docker-test: Run tests in Ubuntu 24.04 Docker container
docker-test:
	@echo "Running tests in Docker (Ubuntu 24.04)..."
	docker build -t ravact-test -f Dockerfile.test .
	docker run --rm ravact-test

## docker-shell: Open shell in Ubuntu 24.04 container for manual testing
docker-shell:
	@echo "Opening shell in Ubuntu 24.04 container..."
	docker run --rm -it -v $(PWD):/workspace -w /workspace ubuntu:24.04 /bin/bash

## help: Show this help message
help:
	@echo "Ravact - Makefile commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
