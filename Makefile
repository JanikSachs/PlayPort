.PHONY: build run test clean install dev help

# Binary name
BINARY_NAME=playport
MAIN_PATH=./cmd/playport

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete!"

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run without building
dev:
	@echo "Running in development mode..."
	$(GORUN) $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed!"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Run all checks (fmt, lint, test)
check: fmt test
	@echo "All checks passed!"

# Display help
help:
	@echo "PlayPort - Makefile commands:"
	@echo ""
	@echo "  make build         - Build the application"
	@echo "  make run           - Build and run the application"
	@echo "  make dev           - Run without building (faster for development)"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Lint code (requires golangci-lint)"
	@echo "  make check         - Run fmt and test"
	@echo "  make help          - Display this help message"
	@echo ""

# Default target
.DEFAULT_GOAL := help
