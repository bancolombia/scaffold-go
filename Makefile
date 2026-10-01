.PHONY: build test test-unit test-integration test-properties test-fast clean install run help docs-start docs-build

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -X github.com/bancolombia/scaffold-go/internal/commands.Version=$(VERSION) \
           -X github.com/bancolombia/scaffold-go/internal/commands.BuildDate=$(BUILD_DATE) \
           -X github.com/bancolombia/scaffold-go/internal/commands.GitCommit=$(GIT_COMMIT)

# Build the binary
build:
	@echo "Building scaffold-go..."
	@go build -ldflags "$(LDFLAGS)" -o bin/scaffold-go cmd/main.go
	@echo "✅ Build complete: ./bin/scaffold-go"

# Build the binary in debug mode (with optimizations disabled)
build-debug:
	@echo "Building scaffold-go (with optimizations disabled)..."
	@go build -gcflags="all=-N -l" -ldflags "$(LDFLAGS)" -o bin/scaffold-go cmd/main.go
	@echo "✅ Build complete: ./bin/scaffold-go"

# Run all tests (slow - includes property-based tests with 100 iterations)
test:
	@echo "Running all tests..."
	@go test ./... -v

# Run only unit tests (fast - excludes property-based tests)
test-unit:
	@echo "Running unit tests (excluding property-based tests)..."
	@go test ./... -v -short

# Run only integration tests
test-integration:
	@echo "Running integration tests..."
	@go test ./tests/integration/... -v

# Run only property-based tests with reduced iterations (faster)
test-properties:
	@echo "Running property-based tests with reduced iterations..."
	@GOPTER_MIN_SUCCESS=20 go test ./... -v -run "Property"

# Fast test suite for development (unit tests + quick property tests)
test-fast:
	@echo "Running fast test suite..."
	@go test ./... -v -short -timeout 30s

# Run tests in parallel with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -v -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f bin/scaffold-go coverage.out coverage.html
	@echo "✅ Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing scaffold-go..."
	@go install -ldflags "$(LDFLAGS)" cmd/main.go
	@echo "✅ Installed to GOPATH/bin"

# Run the CLI
run:
	@go run cmd/main.go

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Preview Docusaurus docs (requires Node.js 18+)
docs-start:
	@npm --prefix docs install
	@npm --prefix docs start

# Production Docusaurus build into website/build
docs-build:
	@npm --prefix docs install
	@npm --prefix docs run build

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  test             - Run all tests (slow)"
	@echo "  test-unit        - Run only unit tests (fast)"
	@echo "  test-integration - Run only integration tests"
	@echo "  test-properties  - Run property-based tests with reduced iterations"
	@echo "  test-fast        - Run fast test suite for development"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  clean            - Remove build artifacts"
	@echo "  install          - Install to GOPATH/bin"
	@echo "  run              - Run the CLI"
	@echo "  tidy             - Tidy dependencies"
	@echo "  fmt              - Format code"
	@echo "  docs-start       - Preview documentation site (Docusaurus)"
	@echo "  docs-build       - Build documentation site"
	@echo "  help             - Show this help"

# Default target
.DEFAULT_GOAL := build
