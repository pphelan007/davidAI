.PHONY: build run test clean docker-build docker-run help lint lint-fix lint-install dev

# Variables
BINARY_NAME=worker
CMD_PATH=./cmd/worker
GOLANGCI_LINT_VERSION=v1.55.2

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/$(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Built successfully: bin/$(BINARY_NAME)"

# Run the application (starts the server)
run: build
	@echo "Starting worker server..."
	@./bin/$(BINARY_NAME)

# Check if code compiles, passes lint, and builds binary (development check)
dev: lint
	@echo "Building binary..."
	@go build -o bin/$(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Code compiles and binary built successfully: bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Install golangci-lint
lint-install:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		INSTALLED_VERSION=$$(golangci-lint version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo ""); \
		EXPECTED_VERSION=$$(echo "$(GOLANGCI_LINT_VERSION)" | sed 's/^v//'); \
		if [ -n "$$INSTALLED_VERSION" ] && [ "$$INSTALLED_VERSION" = "$$EXPECTED_VERSION" ]; then \
			echo "golangci-lint $(GOLANGCI_LINT_VERSION) is already installed"; \
		else \
			echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) (found $$INSTALLED_VERSION)..."; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
		fi; \
	else \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

# Lint code
lint: lint-install
	@echo "Linting code..."
	@golangci-lint run

# Lint code with auto-fix
lint-fix: lint-install
	@echo "Linting code with auto-fix..."
	@golangci-lint run --fix

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest -f Dockerfile .

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker-compose up

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application binary"
	@echo "  run                - Start the worker server"
	@echo "  dev                - Lint code and build binary (development check)"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  clean              - Clean build artifacts"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  lint-fix           - Lint code with auto-fix"
	@echo "  lint-install       - Install golangci-lint"
	@echo "  deps               - Download dependencies"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run with docker-compose"

