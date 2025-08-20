.PHONY: help lint fmt test build build-wails build-cli build-server build-mcp clean all vscode-install vscode-build vscode-package vscode-clean

# Default target
all: lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  lint            - Run linter and check formatting (Go + VSCode extension)"
	@echo "  fmt             - Format code with gofmt"
	@echo "  test            - Run all tests (Go + VSCode extension)"
	@echo "  build           - Build all applications (Wails app, CLI, server, MCP)"
	@echo "  build-wails     - Build the Wails application only"
	@echo "  build-cli       - Build the CLI application only"
	@echo "  build-server    - Build the server application only"
	@echo "  build-mcp       - Build the MCP server application only"
	@echo "  clean           - Clean build artifacts"
	@echo "  all             - Run lint, test, and build (default)"
	@echo ""
	@echo "VSCode Extension targets:"
	@echo "  vscode-install  - Install VSCode extension dependencies"
	@echo "  vscode-build    - Compile VSCode extension TypeScript"
	@echo "  vscode-package  - Package VSCode extension as VSIX"
	@echo "  vscode-clean    - Clean VSCode extension build artifacts"

# Format code with gofmt
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -w .

# Lint and check formatting
.PHONY: lint
lint: fmt vscode-install
	@echo "Running Go linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running basic checks..."; \
		go vet ./...; \
		gofmt -l . | grep -v "^$$" && echo "Files not formatted properly" && exit 1 || echo "All files properly formatted"; \
	fi
	@echo "Running VSCode extension linter..."
	cd vscode-extension && npm run lint

# Run tests
.PHONY: test
test: build-cli vscode-build
	@echo "Running Go tests..."
	go test -v ./tests/...
	@echo "Running VSCode extension tests..."
	cd vscode-extension && npm test

# Build all applications
.PHONY: build
build: build-wails build-cli build-server build-mcp
	ls -tarl build/bin/
	@echo "All applications built successfully!"

# Build the Wails application
.PHONY: build-wails
build-wails:
	@echo "Building Wails application..."
	wails build
	@if [ "$(shell uname -s)" = "Darwin" ] && [ "$CI" != "true" ]; then \
		echo "Copying M2E.app to /Applications..."; \
		cp -R build/bin/M2E.app /Applications/; \
	fi

# Build the CLI application
.PHONY: build-cli
build-cli:
	@echo "Building CLI application..."
	go build -o build/bin/m2e ./cmd/m2e

# Build the server application
.PHONY: build-server
build-server:
	@echo "Building server application..."
	go build -o build/bin/m2e-server ./cmd/m2e-server

# Build the MCP server application
.PHONY: build-mcp
build-mcp:
	@echo "Building MCP server application..."
	go build -o build/bin/m2e-mcp ./cmd/m2e-mcp

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/bin/
	go clean ./...

# Install development dependencies
.PHONY: install-deps
install-deps:
	@echo "Installing development dependencies..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found, installing..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

# Install-app command that copies the built build/bin/M2E.app to /Applications (overrides existing app)
.PHONY: install-app
install-app:
	@echo "Installing M2E.app to /Applications..."
	@if [ -d "build/bin/M2E.app" ]; then \
		if [ "$(shell uname -s)" = "Darwin" ]; then \
			cp -R build/bin/M2E.app /Applications/; \
			echo "M2E.app installed successfully!"; \
		else \
			echo "This command is only supported on macOS."; \
		fi \
	else \
		echo "M2E.app not found in build/bin/ directory."; \
	fi

# Run MCP's inspector tool
.PHONY: inspect
inspect:
	@echo "Running MCP inspector tool..."
	DANGEROUSLY_OMIT_AUTH=true npx -y @modelcontextprotocol/inspector "

# VSCode Extension targets
.PHONY: vscode-install
vscode-install:
	@echo "Installing VSCode extension dependencies..."
	cd vscode-extension && npm ci

.PHONY: vscode-build
vscode-build: vscode-install
	@echo "Compiling VSCode extension TypeScript..."
	cd vscode-extension && npm run compile

.PHONY: vscode-package
vscode-package: vscode-build
	@echo "Packaging VSCode extension as VSIX..."
	cd vscode-extension && npm run package

.PHONY: vscode-clean
vscode-clean:
	@echo "Cleaning VSCode extension build artifacts..."
	cd vscode-extension && rm -rf out dist node_modules *.vsix
	cd vscode-extension && rm -rf resources/bin
