# Makefile for jta - JSON Translation Agent

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build settings
BINARY_NAME := jta
MAIN_PATH := ./cmd/jta
BUILD_DIR := dist

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Linker flags to inject version information
LDFLAGS := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)"

.PHONY: all build clean test coverage install uninstall help

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build and install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Uninstall from GOPATH/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "✓ Uninstalled"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	@echo "\nCoverage summary:"
	@$(GOCMD) tool cover -func=coverage.out | tail -1

# View coverage in browser
coverage-html: coverage
	@echo "Opening coverage report in browser..."
	@$(GOCMD) tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@echo "✓ Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✓ Dependencies updated"

# Run the application (dev mode)
run:
	@$(GOCMD) run $(MAIN_PATH)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Multi-platform build complete"
	@ls -lh $(BUILD_DIR)

# Show version information that will be embedded
version-info:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

# Help target
help:
	@echo "Jta - JSON Translation Agent - Makefile commands:"
	@echo ""
	@echo "  make build        - Build the application"
	@echo "  make install      - Build and install to GOPATH/bin"
	@echo "  make uninstall    - Remove from GOPATH/bin"
	@echo "  make test         - Run all tests"
	@echo "  make coverage     - Run tests with coverage report"
	@echo "  make coverage-html - View coverage report in browser"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Download and tidy dependencies"
	@echo "  make run          - Run in development mode"
	@echo "  make build-all    - Build for multiple platforms"
	@echo "  make version-info - Show version information"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Version info will be injected:"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Date:    $(DATE)"
