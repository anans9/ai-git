# AI-Git CLI Makefile

# Variables
BINARY_NAME=ai-git
MAIN_PATH=./main.go
BUILD_DIR=build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.PHONY: all
all: clean build

# Help target
.PHONY: help
help:
	@echo "AI-Git CLI Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build           Build the binary"
	@echo "  test            Run tests"
	@echo "  clean           Clean build artifacts"
	@echo "  install         Install binary globally"
	@echo "  release         Create release build"
	@echo "  format          Format Go code"

# Build targets
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Code quality targets
.PHONY: format
format:
	@echo "Formatting Go code..."
	gofmt -w .
	go mod tidy

# Installation targets
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) globally..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Release target
.PHONY: release
release: clean build
	@echo "Creating release build..."
	@echo "Version: $(VERSION)"
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning..."
	go clean
	@rm -rf $(BUILD_DIR)

# Phony targets
.PHONY: all help build test format install release clean
