.PHONY: help  build  clean lint fmt  deps docker-build docker-run all

PKG         := github.com/company/nexus-perf
BINARY_NAME := nexus-perf
GO          := go
VERSION     ?= 0.1.0

# docker-run defaults (override on the command line)
NEXUS_ENDPOINT ?=
NEXUS_USERNAME ?=
NEXUS_PASSWORD ?=
NEXUS_REPO     ?=
THREADS        ?= 4
NUM_FILES      ?= 10

help:
	@echo "Available targets:"
	@echo "  make build           - Cross-compile for Linux amd64"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make lint            - Run go vet and golangci-lint"
	@echo "  make fmt             - Format code with gofmt and goimports"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Run in Docker"
	@echo "                         Required: NEXUS_ENDPOINT NEXUS_USERNAME NEXUS_PASSWORD NEXUS_REPO"
	@echo "                         Optional: THREADS (default 4) NUM_FILES (default 10)"


build:
	@echo "Building $(BINARY_NAME) for Linux amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -v \
		-ldflags "-X $(PKG)/cmd.appVersion=$(VERSION) -X $(PKG)/cmd.buildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-o $(BINARY_NAME)-linux-amd64-$(VERSION) .
	@echo "Build complete: $(BINARY_NAME)-linux-amd64-$(VERSION)"

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -f $(BINARY_NAME)-* $(BINARY_NAME) $(BINARY_NAME).exe
	rm -f coverage.out coverage.html
	rm -f *.log
	@echo "Clean complete"

lint:
	@echo "Running linters..."
	$(GO) vet ./...
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed, skipping"


deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies updated"

docker-build:
	@echo "Building Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

docker-run:
	docker run --rm -it \
		-e NEXUS_PERF_NEXUS_ENDPOINT="$(NEXUS_ENDPOINT)" \
		-e NEXUS_PERF_USERNAME="$(NEXUS_USERNAME)" \
		-e NEXUS_PERF_PASSWORD="$(NEXUS_PASSWORD)" \
		-e NEXUS_PERF_REPOSITORY_NAME="$(NEXUS_REPO)" \
		-e NEXUS_PERF_NUM_THREADS="$(THREADS)" \
		-e NEXUS_PERF_NUM_FILES="$(NUM_FILES)" \
		$(BINARY_NAME):latest test

all: clean deps build
	@echo "Build successful"
