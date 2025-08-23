.PHONY: build test clean run install-deps

BINARY_NAME=wheeler
BUILD_DIR=./bin

build: install-deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@export GOROOT=/home/mturansk/go/go1.24.4.linux-amd64/go && \
	 export PATH=$$GOROOT/bin:$$PATH && \
	 export GOPATH=/home/mturansk/go && \
	 CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test: install-deps
	@echo "Running tests..."
	@export GOROOT=/home/mturansk/go/go1.24.4.linux-amd64/go && \
	 export PATH=$$GOROOT/bin:$$PATH && \
	 export GOPATH=/home/mturansk/go && \
	 go test -v ./...

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

install-deps:
	@echo "Installing system dependencies..."
	@which pkg-config >/dev/null || which pkgconf >/dev/null || (echo "pkg-config is required" && exit 1)
	@echo "Checking for GTK+2..."
	@pkg-config --exists gtk+-2.0 2>/dev/null || pkgconf --exists gtk+-2.0 2>/dev/null || (echo "GTK+2 development headers are required. Install with: sudo dnf install gtk2-devel" && exit 1)
	@echo "Dependencies satisfied!"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f wheeler.db

help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  run         - Build and run the application"
	@echo "  clean       - Clean build artifacts"
	@echo "  install-deps- Check system dependencies"
	@echo "  help        - Show this help"