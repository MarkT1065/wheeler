.PHONY: build test clean run help

BINARY_NAME=wheeler
BUILD_DIR=./bin

# Detect Go installation
GO_BIN := $(shell which go 2>/dev/null)
ifndef GO_BIN
    $(error Go is not installed or not in PATH)
endif

# Get the actual GOROOT and GOPATH from the go binary
DETECTED_GOROOT := $(shell $(GO_BIN) env GOROOT)
DETECTED_GOPATH := $(shell $(GO_BIN) env GOPATH)

# Override any incorrect environment settings
export GOROOT := $(DETECTED_GOROOT)
export GOPATH := $(DETECTED_GOPATH)

build:
	@echo "Building $(BINARY_NAME)..."
	@echo "Using Go: $(shell go version)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "Running all tests..."
	@CGO_ENABLED=1 go test -v ./...
	@echo ""
	@echo "All tests complete!"

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f test_*.db
	@rm -f ./data/integration_test.db*
	@echo "Clean complete"

help:
	@echo "Wheeler - Available Commands:"
	@echo ""
	@echo "  make test      - Run all tests"
	@echo "  make build     - Build the application"
	@echo "  make run       - Build and run"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make help      - Show this help"
	@echo ""
	@echo "Requirements:"
	@echo "  - Go 1.19+ with CGO support"
	@echo "  - SQLite3 development libraries"
