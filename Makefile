# Set the name of the binary
BINARY_NAME=trackit

# Go environment variables
GO=go
GOPATH=$(shell go env GOPATH)
GOOS=darwin
GOARCH=amd64

# Directories
SRC_DIR=.
BUILD_DIR=./build
DARWIN_DIR=$(BUILD_DIR)/darwin/amd64
WINDOWS_DIR=$(BUILD_DIR)/windows/amd64
UBANTU_X86_64_DIR=$(BUILD_DIR)/ubantu/x86_64
SCHEMA_SRC=./internal/db/schema.sql
SCHEMA_DEST=./cmd/schema.sql

# File paths
MAIN_FILE=$(SRC_DIR)/main.go
MAC_BINARY_PATH=$(DARWIN_DIR)/$(BINARY_NAME)
WIN_BINARY_PATH=$(WINDOWS_DIR)/$(BINARY_NAME).exe
UBANTU_X86_BINARY_PATH=$(UBANTU_X86_64_DIR)/$(BINARY_NAME)

# Default target
.PHONY: all
all: copy-schema build build-windows build-ubantu

# Copy schema.sql from internal/db to cmd/schema.sql
.PHONY: copy-schema
copy-schema:
	@echo "Copying schema.sql from internal/db to cmd..."
	cp $(SCHEMA_SRC) $(SCHEMA_DEST)

# Build the binary for macOS
.PHONY: build
build: fmt vet
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(DARWIN_DIR)
	$(GO) build -o $(MAC_BINARY_PATH) $(MAIN_FILE)

# Build the binary for Ubantu
.PHONY: build-ubantu
build-ubantu: fmt vet
	@echo "Building $(BINARY_NAME) for Ubantu x86_64"
	@mkdir -p $(UBANTU_X86_64_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(UBANTU_X86_BINARY_PATH) $(MAIN_FILE)

# Build the binary for Windows
.PHONY: build-windows
build-windows: fmt vet
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(WINDOWS_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(WIN_BINARY_PATH) $(MAIN_FILE)

# Format the code using gofmt
.PHONY: fmt
fmt:
	@echo "Running gofmt..."
	$(GO) fmt ./...

# Run static analysis (go vet)
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./...

# Clean build files
.PHONY: clean
clean:
	@echo "Cleaning build files..."
	rm -rf $(BUILD_DIR)

# Install the binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(MAC_BINARY_PATH) /usr/local/bin/$(BINARY_NAME)

# Lint code (if you have a linter, like golangci-lint)
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Run a quick local dev mode (if applicable)
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	$(MAC_BINARY_PATH) $(ARGS)

# Pass subcommands (e.g., make run init will call trackit init)
.PHONY: subcommand
subcommand:
	@echo "Passing subcommand to $(BINARY_NAME)..."
	$(MAC_BINARY_PATH) $(ARGS)
