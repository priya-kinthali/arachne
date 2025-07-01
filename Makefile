# Go Web Scraper Makefile
# Demonstrates build automation and development workflow

.PHONY: help build test test-verbose benchmark clean run run-debug lint format check-deps install

# Get all Go source files (excluding test files)
GO_FILES := $(shell find . -name "*.go" -not -name "*_test.go" -not -name "circuit_breaker_test.go")

# Default target
help:
	@echo "🚀 Go Web Scraper - Available Commands:"
	@echo ""
	@echo "📦 Build & Run:"
	@echo "  build        - Build the scraper binary"
	@echo "  run          - Run the scraper with default settings"
	@echo "  run-debug    - Run with debug logging enabled"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test         - Run unit tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  benchmark    - Run performance benchmarks"
	@echo ""
	@echo "🔧 Development:"
	@echo "  lint         - Run code linting (requires golangci-lint)"
	@echo "  format       - Format code with gofmt"
	@echo "  check-deps   - Check and tidy dependencies"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean        - Remove build artifacts and output files"
	@echo "  install      - Install dependencies"
	@echo ""
	@echo "📊 Examples:"
	@echo "  make run concurrent=5 timeout=5s"
	@echo "  make run-debug log-level=debug"

# Build the scraper
build:
	@echo "🔨 Building Go Web Scraper..."
	go build -o scraper $(GO_FILES)
	@echo "✅ Build complete: ./scraper"

# Run the scraper with default settings
run:
	@echo "🚀 Running Go Web Scraper..."
	go run $(GO_FILES)

# Run with debug logging
run-debug:
	@echo "🔍 Running with debug logging..."
	go run $(GO_FILES) -log-level=debug

# Run unit tests
test:
	@echo "🧪 Running unit tests..."
	go test -v ./...

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	go test -v -count=1 ./...

# Run performance benchmarks
benchmark:
	@echo "⚡ Running performance benchmarks..."
	go test -bench=. -benchmem ./...

# Run code linting (requires golangci-lint)
lint:
	@echo "🔍 Running code linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
format:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatting complete"

# Check and tidy dependencies
check-deps:
	@echo "📦 Checking dependencies..."
	go mod tidy
	go mod verify
	@echo "✅ Dependencies verified"

# Clean build artifacts and output files
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f scraper
	rm -f scraping_results.json
	rm -f scraping_metrics.json
	@echo "✅ Clean complete"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	go mod download
	@echo "✅ Dependencies installed"

# Development setup
setup: install
	@echo "🔧 Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✅ Development environment ready"

# Run with custom parameters (example usage: make run-custom concurrent=5 timeout=5s)
run-custom:
	@echo "🚀 Running with custom parameters..."
	go run $(GO_FILES) -concurrent=$(concurrent) -timeout=$(timeout) -log-level=$(log-level)

# Performance test with different concurrency levels
perf-test:
	@echo "📊 Performance testing with different concurrency levels..."
	@echo "Testing with 1 concurrent request..."
	go run $(GO_FILES) -concurrent=1 -metrics=true
	@echo ""
	@echo "Testing with 3 concurrent requests..."
	go run $(GO_FILES) -concurrent=3 -metrics=true
	@echo ""
	@echo "Testing with 5 concurrent requests..."
	go run $(GO_FILES) -concurrent=5 -metrics=true

# Show help for command-line flags
help-flags:
	@echo "📋 Available command-line flags:"
	@go run $(GO_FILES) -h 2>/dev/null || echo "Run 'go run $(GO_FILES) -h' to see available flags"

# Create release build
release: clean
	@echo "🏷️  Creating release build..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o scraper-linux-amd64 $(GO_FILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o scraper-darwin-amd64 $(GO_FILES)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o scraper-windows-amd64.exe $(GO_FILES)
	@echo "✅ Release builds created:"
	@ls -la scraper-*

# Show project statistics
stats:
	@echo "📊 Project Statistics:"
	@echo "Lines of code:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo ""
	@echo "Go files:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo ""
	@echo "Test coverage:"
	@go test -cover ./... 2>/dev/null || echo "Run tests to see coverage" 