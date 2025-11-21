.PHONY: build run test clean docker-build docker-run help lint lint-fix lint-install dev temporal-start temporal-stop build-client trigger-workflow db db-down

# Variables
BINARY_NAME=worker
CLIENT_BINARY_NAME=client
CMD_PATH=./cmd/worker
CLIENT_CMD_PATH=./cmd/client
GOLANGCI_LINT_VERSION=v1.55.2

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/$(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Built successfully: bin/$(BINARY_NAME)"

# Build the client application
build-client:
	@echo "Building client application..."
	@go build -o bin/$(CLIENT_BINARY_NAME) $(CLIENT_CMD_PATH)
	@echo "✅ Built successfully: bin/$(CLIENT_BINARY_NAME)"

# Run the application (starts Temporal dev server and worker)
# Note: Start the database separately with 'make db' before running this
run: build
	@echo "Starting Temporal dev server and worker..."
	@echo "Note: Make sure PostgreSQL is running (use 'make db' in another terminal)"
	@echo "Temporal Service will be available on localhost:7233"
	@echo "Temporal Web UI will be available at http://localhost:8233"
	@trap 'pkill -f "temporal server start-dev" || true' EXIT INT TERM; \
	temporal server start-dev > /tmp/temporal.log 2>&1 & \
	TEMPORAL_PID=$$!; \
	echo "Temporal dev server started (PID: $$TEMPORAL_PID)"; \
	echo "Waiting for Temporal server to be ready..."; \
	sleep 3; \
	echo "Starting worker..."; \
	./bin/$(BINARY_NAME); \
	EXIT_CODE=$$?; \
	echo "Worker stopped, shutting down Temporal server..."; \
	kill $$TEMPORAL_PID 2>/dev/null || true; \
	exit $$EXIT_CODE

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
lint: lint-install deps
	@echo "Linting code..."
	@go build ./... 2>/dev/null || true  # Build to generate export data
	@golangci-lint run 2>&1 | tee /tmp/golangci_output.txt; \
	LINT_EXIT=$$?; \
	if [ $$LINT_EXIT -ne 0 ]; then \
		if grep -q "unsupported version: 2" /tmp/golangci_output.txt && ! grep -qE "^[a-zA-Z].*\.go:[0-9]+:[0-9]+:" /tmp/golangci_output.txt; then \
			echo "⚠️  Note: Export data version issue with external packages (Go 1.24+ compatibility, not a code issue)"; \
			exit 0; \
		fi; \
	fi; \
	exit $$LINT_EXIT

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

# Start Temporal dev server
temporal-start:
	@echo "Starting Temporal dev server..."
	@echo "Temporal Service will be available on localhost:7233"
	@echo "Temporal Web UI will be available at http://localhost:8233"
	@echo "Press CTRL+C to stop the server"
	@temporal server start-dev

# Start Temporal dev server with persistent database
temporal-start-persist:
	@echo "Starting Temporal dev server with persistent database..."
	@echo "Temporal Service will be available on localhost:7233"
	@echo "Temporal Web UI will be available at http://localhost:8233"
	@echo "Database will be saved to temporal.db"
	@echo "Press CTRL+C to stop the server"
	@temporal server start-dev --db-filename temporal.db

# Trigger a workflow execution (builds and runs the client)
trigger-workflow: build-client
	@echo "Triggering AudioProcessingWorkflow..."
	@echo "Usage: make trigger-workflow [FILE_PATH=data/sine440.wav]"
	@if [ -z "$(FILE_PATH)" ]; then \
		./bin/$(CLIENT_BINARY_NAME); \
	else \
		./bin/$(CLIENT_BINARY_NAME) $(FILE_PATH); \
	fi

# Start PostgreSQL database (tears down on Ctrl+C)
db:
	@echo "Starting PostgreSQL database..."
	@echo "Database will be available on localhost:5432"
	@echo "Press CTRL+C to stop and tear down the database"
	@trap 'echo ""; echo "Tearing down database..."; docker-compose down postgres 2>/dev/null || true; exit 0' EXIT INT TERM; \
	docker-compose up postgres

# Stop and remove PostgreSQL database (including volume/data)
db-down:
	@echo "Stopping and removing PostgreSQL database (this will delete all data)..."
	@docker-compose down postgres -v
	@echo "✅ Database stopped and removed (volume deleted)"

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application binary"
	@echo "  run                - Start Temporal dev server and worker"
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
	@echo "  temporal-start     - Start Temporal dev server (in-memory)"
	@echo "  temporal-start-persist - Start Temporal dev server (persistent DB)"
	@echo "  build-client       - Build the workflow client binary"
	@echo "  trigger-workflow   - Trigger AudioProcessingWorkflow (default: data/sine440.wav)"
	@echo "  db                 - Start PostgreSQL database (tears down on Ctrl+C)"
	@echo "  db-down            - Stop and remove PostgreSQL database (including volume/data)"

