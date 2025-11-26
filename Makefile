# opgl-gateway Makefile

.PHONY: all build run test clean docker-build docker-run lint vet help

# Variables
APP_NAME := opgl-gateway
GO := go
DOCKER := docker
PORT := 8080

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	$(GO) build -o $(APP_NAME) main.go

# Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	$(GO) run main.go

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -f coverage.out
	rm -f coverage.html

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t $(APP_NAME):latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	$(DOCKER) run -p $(PORT):$(PORT) --env-file .env $(APP_NAME):latest

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	$(DOCKER) stop $(APP_NAME) || true

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Build the application (default)"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy dependencies"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-stop   - Stop Docker container"
	@echo "  help          - Show this help message"
