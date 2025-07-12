# FAISS Go Bindings Makefile

.PHONY: all build test clean example install check deps format

# Default target
all: build

# Build the project
build:
	@echo "Building FAISS Go bindings..."
	go build -v ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Build example
example:
	@echo "Building example..."
	@if [ -f "example/main.go" ]; then \
		cd example && go build -v -o example_faiss main.go; \
		echo "Example built successfully. Run with: ./example/example_faiss"; \
	else \
		echo "Example not found"; \
	fi

# Run example
run-example: example
	@echo "Running example..."
	@cd example && ./example_faiss

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean ./...
	@if [ -f "example/example_faiss" ]; then rm example/example_faiss; fi
	@if [ -f "example_index.faiss" ]; then rm example_index.faiss; fi
	@if [ -f "test_index.faiss" ]; then rm test_index.faiss; fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...

# Check for common issues
check:
	@echo "Running checks..."
	go vet ./...
	go mod verify

# Install the package
install:
	@echo "Installing package..."
	go install ./...

# Show help
help:
	@echo "FAISS Go Bindings Makefile"
	@echo "=========================="
	@echo "Available targets:"
	@echo "  all         - Build the project (default)"
	@echo "  build       - Build the project"
	@echo "  test        - Run tests"
	@echo "  example     - Build example"
	@echo "  run-example - Build and run example"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  format      - Format code"
	@echo "  check       - Run code checks"
	@echo "  install     - Install the package"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Requirements:"
	@echo "  - Go 1.21 or later"
	@echo "  - FAISS C library with C API support"
	@echo "  - libfaiss_c.so in library path"
	@echo ""
	@echo "To install FAISS C library:"
	@echo "  git clone https://github.com/facebookresearch/faiss.git"
	@echo "  cd faiss"
	@echo "  cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON ."
	@echo "  make -C build"
	@echo "  sudo make -C build install"
	@echo "  sudo cp build/c_api/libfaiss_c.so /usr/lib" 