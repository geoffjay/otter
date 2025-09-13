# Otter Development Makefile

.PHONY: build test clean install deps fmt vet lint run-example

# Build variables
BINARY_NAME=otter
BUILD_DIR=./bin
MAIN_FILE=./main.go

# Default target
all: deps fmt vet test build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run tests
test:
	go test -v ./...

# Build binary
build: deps fmt vet
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
build-all: deps fmt vet
	mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# macOS
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Install binary globally
install: build
	go install .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean
	rm -rf .otter/

# Run linting (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run

# Development workflow
dev: fmt vet test build

# Run example workflow
run-example: build
	@echo "Running otter init example..."
	cd /tmp && rm -rf otter-test && mkdir otter-test && cd otter-test && \
	$(CURDIR)/$(BUILD_DIR)/$(BINARY_NAME) init && \
	echo "Created test Otterfile with sample layer" && \
	echo "LAYER https://github.com/github/gitignore.git TARGET gitignore-templates" > Otterfile && \
	echo "Running otter build..." && \
	$(CURDIR)/$(BUILD_DIR)/$(BINARY_NAME) build || true

# Docker targets
docker-build:
	docker build -t otter:latest .

docker-run: docker-build
	docker run --rm -it otter:latest

docker-shell: docker-build
	docker run --rm -it --entrypoint /bin/sh otter:latest

# Release target (for GitHub Actions)
release: clean build-all
	mkdir -p $(BUILD_DIR)/release
	# Create archives
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip release/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

# Help
help:
	@echo "Available targets:"
	@echo "  all            - Run full development workflow (deps, fmt, vet, test, build)"
	@echo "  deps           - Install/update Go dependencies"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  test           - Run tests"
	@echo "  build          - Build binary"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  install        - Install binary globally"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run golangci-lint (requires golangci-lint installation)"
	@echo "  dev            - Quick development workflow"
	@echo "  run-example    - Build and run example workflow"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Build and run Docker container"
	@echo "  docker-shell   - Build and run Docker container with shell access"
	@echo "  release        - Create release archives"
	@echo "  help           - Show this help message"
