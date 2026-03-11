# Default binary name
APP_NAME := satellite-visualizer
BUILD_DIR := bin

.PHONY: all build run clean test lint

all: clean build test

# Build the binary into the bin/ directory
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/satellite-visualizer

# Run the binary
run:
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Test the code
test:
	@echo "Testing..."
	@go test ./...

# Lint the code
lint:
	@echo "Linting..."
	@go vet ./...
