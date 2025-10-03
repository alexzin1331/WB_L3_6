.PHONY: test test-coverage test-verbose test-race test-unit test-integration benchmark clean deps test-deps

# Default Make target
.DEFAULT_GOAL := test

# Test targets
test: test-unit test-integration

# Run unit tests (no external dependencies)
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/storage/storage_test.go -run=".*"

# Run integration tests with testcontainers
test-integration:
	@echo "Running integration tests with testcontainers..."
	go test -v ./internal/storage/ -run="TestStorage"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./internal/storage/

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/storage/
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run verbose tests with detailed output
test-verbose:
	@echo "Running verbose tests..."
	go test -v ./internal/storage/

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./internal/storage/

# Clean test artifacts
clean:
	go clean -testcache
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install test dependencies
test-deps:
	go install github.com/testcontainers/testcontainers-go
	go install github.com/stretchr/testify

# Build the application
build:
	go build -o bin/app ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Docker targets
docker-build:
	docker build -t l3_6 .

docker-run:
	docker run -p 8080:8080 l3_6

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  test           - Run both unit and integration tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests with testcontainers"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  benchmark      - Run benchmarks"
	@echo "  clean          - Clean test artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  test-deps      - Install test dependencies"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help message"
