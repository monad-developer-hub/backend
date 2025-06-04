.PHONY: help run build test clean deps fmt lint docker-build docker-run

# Default target
help:
	@echo "Available commands:"
	@echo "  run          - Run the application in development mode"
	@echo "  build        - Build the application binary"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Run linter"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"

# Run the application
run:
	@echo "Starting Monad DevHub API..."
	go run cmd/api/main.go

# Build the application
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Install linter
install-lint:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air for live reload
install-air:
	@echo "Installing air for live reload..."
	go install github.com/cosmtrek/air@latest

# Run with live reload
dev:
	@echo "Starting with live reload..."
	air

# Database commands
db-create:
	@echo "Creating database..."
	createdb monad_devhub

db-drop:
	@echo "Dropping database..."
	dropdb monad_devhub

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t monad-devhub-api .

docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up --build 