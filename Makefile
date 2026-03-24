.PHONY: build test lint clean run help

# Default target
all: lint test build

# Build the application
build:
	@echo "Building..."
	@go build -o bin/crypto-bot ./cmd/main.go

# Run all tests with race detection and coverage
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage report..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Note: golangci-lint not installed. Run 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# Run the application
run:
	@go run ./cmd/main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Format code with gofmt"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Run the application"
	@echo "  help          - Show this help message"
