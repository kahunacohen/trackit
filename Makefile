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
SCHEMA_SRC=./internal/db/schema.sql
SCHEMA_DEST=./cmd/schema.sql

# File paths
MAIN_FILE=$(SRC_DIR)/main.go
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Default target
.PHONY: all
all: copy-schema build

# Copy schema.sql from internal/db to cmd/schema.sql
.PHONY: copy-schema
copy-schema:
	@echo "Copying schema.sql from internal/db to cmd..."
	cp $(SCHEMA_SRC) $(SCHEMA_DEST)

# Build the binary for macOS
.PHONY: build
build: fmt vet
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BINARY_PATH) $(MAIN_FILE)

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
	cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)

# Lint code (if you have a linter, like golangci-lint)
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Run a quick local dev mode (if applicable)
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)
