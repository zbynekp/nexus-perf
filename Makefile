.PHONY: help build run test clean lint fmt install deps

# Variables
BINARY_NAME=nexus-perf
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"
VERSION?=1.0.0

help:
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make install        - Install the application"
	@echo "  make deps           - Download dependencies"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run in Docker"

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux
	@echo "Build complete: $(BINARY_NAME)-linux"

build-macos:
	@echo "Building $(BINARY_NAME) for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)-macos
	@echo "Build complete: $(BINARY_NAME)-macos"

build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME).exe
	@echo "Build complete: $(BINARY_NAME).exe"

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(BINARY_NAME)-macos $(BINARY_NAME).exe
	rm -f coverage.out coverage.html
	@echo "Clean complete"

lint:
	@echo "Running linters..."
	$(GO) vet ./...
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not installed"

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed"

install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS)
	@echo "Installation complete"

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies downloaded"

docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

docker-run:
	docker run --rm -it \
		-e NEXUS_PERF_NEXUS_ENDPOINT=$(NEXUS_ENDPOINT) \
		-e NEXUS_PERF_USERNAME=$(NEXUS_USERNAME) \
		-e NEXUS_PERF_PASSWORD=$(NEXUS_PASSWORD) \
		-e NEXUS_PERF_REPOSITORY_NAME=$(NEXUS_REPO) \
		-e NEXUS_PERF_NUM_THREADS=$(THREADS) \
		-e NEXUS_PERF_NUM_FILES=$(NUM_FILES) \
		$(BINARY_NAME):latest test

all: clean deps build
	@echo "Build successful"

.PHONY: all build build-linux build-macos build-windows run test test-coverage clean lint fmt install deps docker-build docker-run
