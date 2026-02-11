# Suggested Development Commands

## Build and Test Commands
```bash
make                # Default: lint, test, build all
make lint           # Run golangci-lint and formatting checks  
make test           # Run all tests in tests/ directory
make build          # Build all applications (GUI, CLI, server, MCP)
make clean          # Clean build artifacts
```

## Individual Build Commands
```bash
make build-wails    # GUI application (macOS app bundle)
make build-cli      # Command-line interface
make build-server   # HTTP API server  
make build-mcp      # MCP server for AI tools
make vscode-build   # VSCode extension
```

## Development Mode
```bash
wails dev           # Hot-reload development server for GUI
```

## Quality & Security
```bash
make test-coverage  # Generate coverage report
make security       # Run govulncheck for vulnerabilities
make install-deps   # Install development dependencies
```

## Running Applications
```bash
go run .                    # Run GUI application (Wails)
go run ./cmd/m2e           # Run CLI application
go run ./cmd/m2e-server    # Run HTTP API server
go run ./cmd/m2e-mcp       # Run MCP server
```

## Testing Individual Components
```bash
go test -v ./tests/converter_test.go      # Test core converter
go test -v ./tests/cli_test.go           # Test CLI interface
go test -v ./tests/unit_*_test.go        # Test unit conversion
```

## macOS Specific
```bash
make install-app    # Install built app to /Applications
```

## System Information
- Platform: macOS (Darwin)
- Go version: 1.23+ required
- Node.js: 22+ required for frontend
- Wails CLI: v2.10.1+ required