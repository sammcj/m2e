.PHONY: help lint fmt test build build-wails build-cli build-server build-mcp clean all

# Default target
all: lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  lint        - Run linter and check formatting"
	@echo "  fmt         - Format code with gofmt"
	@echo "  test        - Run all tests"
	@echo "  build       - Build all applications (Wails app, CLI, server, MCP)"
	@echo "  build-wails - Build the Wails application only"
	@echo "  build-cli   - Build the CLI application only"
	@echo "  build-server- Build the server application only"
	@echo "  build-mcp   - Build the MCP server application only"
	@echo "  clean       - Clean build artifacts"
	@echo "  all         - Run lint, test, and build (default)"

# Format code with gofmt
fmt:
	@echo "Formatting code..."
	gofmt -w .

# Lint and check formatting
lint: fmt
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running basic checks..."; \
		go vet ./...; \
		gofmt -l . | grep -v "^$$" && echo "Files not formatted properly" && exit 1 || echo "All files properly formatted"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	go test -v ./tests/...

# Build all applications
build: build-wails build-cli build-server build-mcp
	ls -tarl build/bin/
	@echo "All applications built successfully!"

# Build the Wails application
build-wails:
	@echo "Building Wails application..."
	wails build
	@if [ "$(shell uname -s)" = "Darwin" ] && [ "$CI" != "true" ]; then \
		echo "Copying M2E.app to /Applications..."; \
		cp -R build/bin/M2E.app /Applications/; \
	fi


# Build the CLI application
build-cli:
	@echo "Building CLI application..."
	go build -o build/bin/m2e ./cmd/m2e-cli

# Build the server application
build-server:
	@echo "Building server application..."
	go build -o build/bin/m2e-server ./cmd/m2e-server

# Build the MCP server application
build-mcp:
	@echo "Building MCP server application..."
	go build -o build/bin/m2e-mcp ./cmd/m2e-mcp

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/bin/
	go clean ./...

# Install development dependencies
install-deps:
	@echo "Installing development dependencies..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found, installing..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

# Run all quality checks
quality: lint test security
	@echo "All quality checks passed!"
