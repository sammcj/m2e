# M2E Project Onboarding

## Project Purpose
M2E (Murican to English Converter) is a text conversion application that converts American English to standard international/British English spellings with optional imperial-to-metric unit conversion.

## Tech Stack
- **Backend**: Go 1.23+ with Wails v2 framework
- **Frontend**: React 18+ with Vite build system
- **UI Libraries**: Charm libraries (glamour for markdown rendering, lipgloss for styling) - used in CLI report mode
- **Framework**: Wails v2.10+ for native desktop application

## Project Structure
```
m2e/
├── main.go                 # Main GUI application entry point (Wails)
├── app.go                  # Application binding and methods
├── cmd/                    # Different executable entry points
│   ├── m2e/               # CLI application
│   ├── m2e-server/        # HTTP API server
│   └── m2e-mcp/           # MCP server for AI tools
├── pkg/converter/         # Core conversion logic
├── pkg/report/           # Report generation (uses glamour)
├── frontend/             # React SPA
└── tests/               # Comprehensive test suite
```

## Main Entry Points
1. **GUI App**: `go run .` or `make build-wails` - Creates native desktop app
2. **CLI Tool**: `go run ./cmd/m2e` or `make build-cli` - Command-line interface
3. **API Server**: `go run ./cmd/m2e-server` or `make build-server` - HTTP API
4. **MCP Server**: `go run ./cmd/m2e-mcp` or `make build-mcp` - MCP protocol server

## Development Commands
- `make` - Default: lint, test, build all
- `make lint` - Run linter and format check (golangci-lint)
- `make test` - Run all tests
- `make build` - Build all applications
- `wails dev` - Development mode with hot reload

## Code Style & Conventions
- Go best practices with embedded resources
- British English spellings throughout
- Minimal comments on complex logic only
- Multi-interface design with shared core converter
- Code-aware processing for different file types

## Testing
All tests in `tests/` directory:
- Unit tests for converter functions
- Integration tests for interfaces
- Performance tests for unit conversion
- Real-world example testing

## Important Notes
- This is NOT a TUI (Terminal User Interface) application
- Main entry point creates a Wails GUI application
- Uses Charm libraries only for CLI report formatting, not for TUI
- No interactive menu system or edit modes in this codebase