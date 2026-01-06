.PHONY: build test test-mock test-cgo clean examples lint fmt

# Default target
all: test

# Build verification (just compile)
build:
	CGO_ENABLED=1 go build ./...

# Build without CGO (mock mode)
build-mock:
	CGO_ENABLED=0 go build ./...

# Test with mock (no zenoh-c required)
test-mock:
	CGO_ENABLED=0 go test -v ./...

# Test with CGO (requires zenoh-c installed)
test-cgo:
	CGO_ENABLED=1 go test -v ./...

# Default test uses mock
test: test-mock

# Build examples (CGO)
examples:
	CGO_ENABLED=1 go build -o bin/pub ./examples/pub
	CGO_ENABLED=1 go build -o bin/sub ./examples/sub

# Install zenoh-c (Ubuntu/Debian ARM64)
install-zenohc-arm64:
	curl -L https://github.com/eclipse-zenoh/zenoh-c/releases/download/1.0.0/zenoh-c-1.0.0-aarch64-unknown-linux-gnu.deb -o /tmp/zenoh-c.deb
	sudo dpkg -i /tmp/zenoh-c.deb

# Install zenoh-c (Ubuntu/Debian x86_64)
install-zenohc-x64:
	curl -L https://github.com/eclipse-zenoh/zenoh-c/releases/download/1.0.0/zenoh-c-1.0.0-x86_64-unknown-linux-gnu.deb -o /tmp/zenoh-c.deb
	sudo dpkg -i /tmp/zenoh-c.deb

# Install zenoh-c (macOS)
install-zenohc-mac:
	brew tap eclipse-zenoh/homebrew-zenoh
	brew install zenoh-c

# Clean
clean:
	rm -rf bin/
	go clean

# Lint
lint:
	golangci-lint run

# Format
fmt:
	go fmt ./...

# Help
help:
	@echo "zenoh-go Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build           Build with CGO"
	@echo "  build-mock      Build without CGO (mock mode)"
	@echo "  test-mock       Test without CGO (default)"
	@echo "  test-cgo        Test with CGO (requires zenoh-c)"
	@echo "  examples        Build examples"
	@echo "  install-zenohc-arm64  Install zenoh-c on ARM64 Linux"
	@echo "  install-zenohc-x64    Install zenoh-c on x86_64 Linux"
	@echo "  install-zenohc-mac    Install zenoh-c on macOS"
	@echo "  clean           Remove build artifacts"
	@echo "  lint            Run linter"
	@echo "  fmt             Format code"








