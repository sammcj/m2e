# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

M2E is a text conversion application that converts American English to standard international / British English spellings with a tongue-and-cheek approach to it's functionality naming, with optional imperial-to-metric unit conversion. Built with Go backend and React frontend using Wails v2 framework.

## Core Architecture

### Main Components
- **Backend**: Go application (`app.go`, `main.go`) using Wails v2 for native desktop app
- **Frontend**: React SPA in `frontend/` directory with Vite build system
- **Converter Package**: Core logic in `pkg/converter/` handles text conversion and unit processing
- **Multiple Interface Modes**: GUI (Wails), CLI with integrated report mode, HTTP API server, and MCP server

### Key Files
- `app.go`: Application binding and frontend interface methods
- `pkg/converter/converter.go`: Core conversion logic with embedded dictionaries
- `pkg/converter/unit_*.go`: Unit conversion system (detector, patterns, config)
- `pkg/converter/codeaware.go`: Code-aware conversion preserving syntax
- `cmd/`: Different executable entry points (CLI with report mode, server, MCP)
- `pkg/report/`: Report generation and analysis functionality

## Development Commands

### Build and Test
```bash
make                # Default: lint, test, build all
make lint           # Run golangci-lint and formatting checks
make test           # Run all tests in tests/ directory
make build          # Build all applications (GUI, CLI, server, MCP)
```

#### Build VSCode Extension

```bash
make vscode-build   # Build VSCode extension for M2E - This is already triggered by `make build`
```

#### Individual Builds
```bash
make build-wails    # GUI application (macOS app bundle)
make build-cli      # Command-line interface with integrated report mode
make build-server   # HTTP API server
make build-mcp      # MCP server for AI tools
make vscode-build   # VSCode extension
```

### Lint Github Actions
```bash
actionlint
```

### Development Mode
```bash
wails dev           # Hot-reload development server
```

### Quality Checks
```bash
make test-coverage  # Generate coverage report
make security       # Run govulncheck for vulnerabilities
```

## Testing Strategy

All tests are in `tests/` directory with comprehensive coverage:
- Unit tests for converter functions
- Integration tests for different interface modes
- Performance tests for unit conversion
- Real-world example testing

Run single test file:
```bash
go test -v ./tests/converter_test.go
```

## Key Architectural Patterns

### Embedded Resources
- Dictionaries embedded at build time using `//go:embed`
- User dictionaries loaded from `~/.config/m2e/` at runtime
- Frontend assets embedded in Go binary

### Multi-Interface Design
- Single core converter package used by all interfaces
- Consistent API across CLI, GUI, HTTP, and MCP modes
- Code-aware processing for different file types

### Configuration System
- User dictionaries override built-in mappings
- Unit conversion preferences in `~/.config/m2e/unit_config.json`
- Robust error handling for invalid configurations

## Development Notes

### Go Version
Requires Go 1.24+ (specified in go.mod with toolchain go1.24.1)

### Frontend Dependencies
- React 18.3+ with Vite 6.2+ build system
- Install with `npm install` in frontend/ directory
- Build with `npm run build`

### Wails Integration
- Version 2.10.1 specified in wails.json
- Uses embedded assets and native window controls
- macOS-specific features like Services integration

### File Structure Conventions
- `pkg/` for reusable Go packages
- `cmd/` for executable entry points
- `tests/` for all test files
- `frontend/` for React application
- Embedded data in `pkg/converter/data/`

## Note

- Always run `make lint && make test` before stating you have completed the task.
