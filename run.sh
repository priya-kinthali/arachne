#!/bin/bash

# Go Web Scraper Runner Script
# Automatically finds and runs Go source files (excluding test files)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go from https://golang.org/dl/"
    exit 1
fi

# Find all Go source files (excluding test files)
GO_FILES=$(find . -name "*.go" -not -name "*_test.go" | sort)

if [ -z "$GO_FILES" ]; then
    print_error "No Go source files found in current directory"
    exit 1
fi

print_info "Found Go files:"
echo "$GO_FILES" | sed 's/^/  /'

print_info "Running Go Web Scraper..."
echo ""

# Run the scraper with all arguments passed to this script
go run $GO_FILES "$@"

print_success "Scraper completed!" 