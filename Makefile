.PHONY: build test run docker-build docker-run clean help

# Binary name
BINARY_NAME=imgcheck

# Build directory
BUILD_DIR=bin

# Default target
.DEFAULT_GOAL := help

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/imgcheck
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application (requires FOLDER argument)
run: build
	@if [ -z "$(FOLDER)" ]; then \
		echo "Usage: make run FOLDER=/path/to/images"; \
		exit 1; \
	fi
	@echo "Running $(BINARY_NAME) on $(FOLDER)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) $(FOLDER)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .
	@echo "Docker image built: $(BINARY_NAME):latest"

# Run Docker container (requires FOLDER argument)
docker-run: docker-build
	@if [ -z "$(FOLDER)" ]; then \
		echo "Usage: make docker-run FOLDER=/path/to/images"; \
		exit 1; \
	fi
	@echo "Running $(BINARY_NAME) in Docker on $(FOLDER)..."
	@docker run --rm -v "$(FOLDER):/images" $(BINARY_NAME):latest /images

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run || echo "golangci-lint not installed, skipping..."

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  run          - Run the application (requires FOLDER=/path/to/images)"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container (requires FOLDER=/path/to/images)"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  help         - Show this help message"

