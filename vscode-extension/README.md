# M2E - 'Murican to English Converter

Convert 'Murican to standard international English and freedom units to standard metric units directly within Visual Studio Code.

![M2E Extension](https://img.shields.io/badge/M2E-VSCode%20Extension-blue)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

## Overview

M2E is a VSCode extension that provides seamless conversion of American English spellings to British English, with optional imperial-to-metric unit conversion. The extension integrates with the existing M2E converter infrastructure through a local HTTP API server, providing accurate text conversion without leaving your IDE.

M2E provides a CLI, MCP and GUI as well as the VSCode extension, see the the project on Github for more information and documentation: https://github.com/sammcj/m2e

### Key Features

- **Smart Text Conversion**: Convert American spellings to standard English (color → colour, organize → organise)
- **Unit Conversion**: Transform imperial measurements to metric (5 feet → 1.52m, 70°F → 21°C)
- **Code-Aware Processing**: Preserve syntax while converting comments and strings in code files
- **Real-Time Diagnostics**: Highlight American spellings as you type
- **Quick Fixes**: One-click conversion suggestions via VSCode's built-in Quick Fix system
- **Preview Mode**: Review changes using VSCode's diff viewer before applying
- **Multiple Conversion Modes**: Selection, entire file, or comments-only conversion
- **Automatic Server Management**: Zero-configuration setup with bundled binaries

## Supported Platforms

- **macOS**: Intel and Apple Silicon (darwin-arm64)
- **Linux**: x64 and ARM64 architectures
- **Windows**: Not supported (Unix-like systems only)

## Installation

### From VSCode Marketplace

1. Open VSCode
2. Go to Extensions (Ctrl+Shift+X / Cmd+Shift+X)
3. Search for "M2E - American to British English Converter"
4. Click Install

You can also browse to the [VSCoode Marketplace online](https://marketplace.visualstudio.com/items?itemName=SamMcLeod.m2e-vscode).

### From VSIX File

1. Download the latest `.vsix` file from [GitHub Releases](https://github.com/sammcj/m2e/releases)
2. Open VSCode Command Palette (Ctrl+Shift+P / Cmd+Shift+P)
3. Run "Extensions: Install from VSIX..."
4. Select the downloaded `.vsix` file

### Manual Installation

```bash
# Clone the repository
git clone https://github.com/sammcj/m2e.git
cd m2e/vscode-extension

# Install dependencies
npm install

# Build the extension
npm run compile

# Package the extension
npm install -g @vscode/vsce
vsce package

# Install the generated .vsix file
code --install-extension m2e-vscode-*.vsix
```

## Quick Start

1. **Install the extension** (see installation instructions above)
2. **Open any text file** or code file in VSCode
3. **Select text** you want to convert
4. **Use keyboard shortcut** `Cmd+Shift+M` (macOS) or `Ctrl+Shift+M` (Linux)
5. **View the conversion** - converted text replaces the selection

The extension automatically starts its local server when needed. No additional configuration required.

## Features

### Text Conversion Commands

#### Convert Selection
Convert currently selected text from American to British English.

- **Command**: `M2E: Convert Selection`
- **Keyboard Shortcut**: `Cmd+Shift+M` (macOS) / `Ctrl+Shift+M` (Linux)
- **Context Menu**: Right-click on selected text
- **When**: Available when text is selected

#### Convert File
Convert the entire active file from American to British English.

- **Command**: `M2E: Convert File`
- **Keyboard Shortcut**: `Cmd+Alt+M` (macOS) / `Ctrl+Alt+M` (Linux)
- **Context Menu**: Right-click on file in Explorer
- **When**: Always available

#### Convert Comments Only
Convert only comments and strings in code files, preserving syntax.

- **Command**: `M2E: Convert Comments Only`
- **When**: Available for programming language files

#### Convert and Preview
Preview changes using VSCode's built-in diff viewer before applying.

- **Command**: `M2E: Convert and Preview`
- **Keyboard Shortcut**: `Cmd+Shift+Alt+M` (macOS) / `Ctrl+Shift+Alt+M` (Linux)
- **When**: Always available

### Diagnostic Features

#### Real-Time Spelling Detection
The extension continuously scans open documents and highlights American spellings with informational diagnostics.

- **Severity Levels**: Information (default), Warning, or Error
- **File Types**: All supported text and code files
- **Exclusions**: Automatically excludes `node_modules`, `.git`, `dist`, `build` directories

#### Quick Fix Actions
Click on highlighted American spellings to see available Quick Fix actions:

- **Convert Word**: Replace individual American spelling with British equivalent
- **Convert All in File**: Convert all American spellings in the current file
- **Ignore Word**: Add word to local ignore list
- **Ignore in Workspace**: Add word to workspace-specific ignore list

### Server Management

#### Automatic Server Lifecycle
The extension manages a local M2E server process automatically:

- **Auto-Start**: Server starts when first needed
- **Port Discovery**: Automatically finds available port (default: 18181)
- **Health Monitoring**: Continuous health checks ensure server availability
- **Auto-Restart**: Intelligent restart on configuration changes
- **Clean Shutdown**: Graceful server termination when VSCode closes

### Unit Conversion

Convert imperial measurements to metric equivalents:

- **Temperature**: 70°F → 21°C, 32°F → 0°C
- **Distance**: 5 feet → 1.52m, 2 miles → 3.22km
- **Weight**: 10 pounds → 4.54kg, 1 ounce → 28.35g
- **Volume**: 1 gallon → 3.79L, 1 pint → 0.47L
- **Area**: 1 acre → 0.40 hectares, 1 sq ft → 0.09 sq m

## Configuration

### Extension Settings

Access settings via `File > Preferences > Settings` and search for "M2E":

#### Diagnostics Configuration

```json
{
  "m2e.enableDiagnostics": true,
  "m2e.diagnosticSeverity": "Information"
}
```

- **Enable Diagnostics**: Show real-time American spelling detection
- **Diagnostic Severity**: Set severity level (Information, Warning, Error)

#### Conversion Configuration

```json
{
  "m2e.enableUnitConversion": true,
  "m2e.codeAwareConversion": true,
  "m2e.preserveCodeSyntax": true
}
```

- **Enable Unit Conversion**: Include imperial-to-metric conversion
- **Code Aware Conversion**: Use intelligent code-aware processing
- **Preserve Code Syntax**: Maintain syntax highlighting and formatting

#### Server Configuration

```json
{
  "m2e.serverPort": 18181,
  "m2e.customServerPath": "",
  "m2e.debugLogging": false
}
```

- **Server Port**: Preferred port for local server (auto-increments if busy)
- **Custom Server Path**: Path to custom m2e-server binary (optional)
- **Debug Logging**: Enable detailed logging for development

#### File Processing Configuration

```json
{
  "m2e.excludePatterns": [
    "**/node_modules/**",
    "**/.git/**",
    "**/dist/**",
    "**/build/**"
  ]
}
```

- **Exclude Patterns**: Additional glob patterns to exclude from processing

### Workspace Configuration

Create `.vscode/settings.json` in your workspace for project-specific settings:

```json
{
  "m2e.enableDiagnostics": false,
  "m2e.enableUnitConversion": false,
  "m2e.excludePatterns": [
    "**/docs/**",
    "**/examples/**"
  ]
}
```

### User Dictionary

Create custom word mappings in `~/.config/m2e/user_dictionary.json`:

```json
{
  "specialization": "specialisation",
  "customization": "customisation",
  "optimization": "optimisation"
}
```

### Unit Conversion Configuration

Customise unit conversion settings in `~/.config/m2e/unit_config.json`:

```json
{
  "temperature": {
    "enabled": true,
    "precision": 1
  },
  "distance": {
    "enabled": true,
    "precision": 2
  },
  "weight": {
    "enabled": true,
    "precision": 2
  }
}
```

## Supported File Types

The extension processes these file types:

### Text Files
- Plain text (`.txt`)
- Markdown (`.md`, `.markdown`)
- Documentation files (`.rst`, `.adoc`)

### Programming Languages
- JavaScript/TypeScript (`.js`, `.ts`, `.jsx`, `.tsx`)
- Python (`.py`)
- Go (`.go`)
- Java (`.java`)
- C/C++ (`.c`, `.cpp`, `.h`, `.hpp`)
- C# (`.cs`)
- PHP (`.php`)
- Ruby (`.rb`)
- Rust (`.rs`)

### Configuration Files
- JSON (`.json`)
- YAML (`.yml`, `.yaml`)
- XML (`.xml`)
- INI (`.ini`)

## Command Reference

### Available Commands

| Command | Description | Keyboard Shortcut |
|---------|-------------|-------------------|
| `M2E: Convert Selection` | Convert selected text | `Cmd+Shift+M` / `Ctrl+Shift+M` |
| `M2E: Convert File` | Convert entire file | `Cmd+Alt+M` / `Ctrl+Alt+M` |
| `M2E: Convert Comments Only` | Convert only comments in code | - |
| `M2E: Convert and Preview` | Preview changes before applying | `Cmd+Shift+Alt+M` / `Ctrl+Shift+Alt+M` |
| `M2E: Restart Server` | Restart the local server | - |
| `M2E: Ignore Word` | Add word to ignore list | - |
| `M2E: Manage Ignore List` | Edit workspace ignore list | - |
| `M2E: Refresh Diagnostics` | Refresh diagnostic highlighting | - |

### Context Menus

#### Editor Context Menu
Right-click on selected text:
- "Convert to British English" - Convert selection

#### Explorer Context Menu
Right-click on files in Explorer:
- "Convert File to British English" - Convert entire file

## Usage Examples

### Basic Text Conversion

**Input:**
```text
The color of the aluminum organization center was recognized for its specialized optimization.
```

**Output:**
```text
The colour of the aluminium organisation centre was recognised for its specialised optimisation.
```

### Unit Conversion

**Input:**
```text
The server room is 70°F and measures 10 feet by 12 feet.
The cooling system handles 5 gallons per minute at 30 PSI.
```

**Output:**
```text
The server room is 21°C and measures 3.05m by 3.66m.
The cooling system handles 18.93L per minute at 206.84 kPa.
```

### Code-Aware Conversion

**Input (JavaScript):**
```javascript
// This function optimizes the color analyzer
function colorAnalyzer(colorValue) {
    // Analyze the color for optimization
    return optimizeColor(colorValue);
}
```

**Output (JavaScript):**
```javascript
// This function optimises the colour analyser
function colorAnalyzer(colorValue) {
    // Analyse the colour for optimisation
    return optimizeColor(colorValue);
}
```

Note: Function names and variables are preserved while comments are converted.

## Troubleshooting

### Common Issues

#### Extension Not Activating
**Symptoms**: Commands not available, no server status
**Solutions**:
1. Check if you're on a supported platform (macOS or Linux)
2. Reload VSCode window (`Developer: Reload Window`)
3. Check Extension Host output for errors
4. Verify extension is enabled in Extensions panel

#### Server Not Starting
**Symptoms**: "Failed to start M2E server" error
**Solutions**:
1. Check if port 18181 is available: `lsof -i :18181`
2. Verify binary permissions: `chmod +x ~/.vscode/extensions/*/resources/bin/*/m2e-server`
3. Try custom server path in settings
4. Check M2E output channel for detailed error messages

#### Commands Not Working
**Symptoms**: Commands run but no conversion happens
**Solutions**:
1. Ensure text is selected for "Convert Selection"
2. Check server status in status bar
3. Verify file is supported type
4. Check exclude patterns in settings
5. Restart server using `M2E: Restart Server`

#### Slow Performance
**Symptoms**: Conversion takes long time
**Solutions**:
1. Check file size - large files (>500KB) take longer
2. Disable diagnostics for large workspaces
3. Add exclude patterns for unnecessary directories
4. Close unused documents to reduce diagnostic load

#### Diagnostics Not Showing
**Symptoms**: American spellings not highlighted
**Solutions**:
1. Enable diagnostics in settings (`m2e.enableDiagnostics`: true)
2. Check if file type is supported
3. Verify file is not in excluded patterns
4. Run `M2E: Refresh Diagnostics` command
5. Check diagnostic severity setting

### Debug Information

#### Enable Debug Logging
Set `m2e.debugLogging` to `true` for detailed logs in the M2E output channel.

#### Check Server Status
Run this command to check server health:
```bash
curl http://localhost:18181/api/v1/health
```

#### Manual Server Start
Start server manually for debugging:
```bash
# Find the binary location
find ~/.vscode/extensions -name "m2e-server" -type f

# Run manually
API_PORT=18181 /path/to/m2e-server
```

#### Reset Configuration
Remove configuration files to reset to defaults:
```bash
rm -rf ~/.config/m2e/
```

### Getting Help

1. **Check Output Channel**: View `M2E` output channel for detailed logs
2. **Restart Server**: Use `M2E: Restart Server` command
3. **Reload Window**: Use `Developer: Reload Window` command
4. **Report Issues**: Create issue at [GitHub Issues](https://github.com/sammcj/m2e/issues)

When reporting issues, include:
- Operating system and architecture
- VSCode version
- Extension version
- Error messages from M2E output channel
- Steps to reproduce the issue

## Development

### Prerequisites

- Node.js 20+ and npm
- Go 1.24+ (for building m2e-server)
- VSCode or VSCode Insiders
- Git

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/sammcj/m2e.git
cd m2e

# Build server component
make build-server

# Setup extension
cd vscode-extension
npm install
npm run compile
```

### Development Workflow

```bash
# Watch mode for TypeScript compilation
npm run watch

# Open extension in development mode
code .
# Press F5 to launch Extension Development Host

# Run tests
npm test

# Lint code
npm run lint

# Package extension
vsce package
```

### Testing

```bash
# Run all tests
npm test

# Run specific test suites
npm run test:unit
npm run test:integration

# Run with coverage
npm run test:coverage
```

### Project Structure

```
vscode-extension/
├── src/                    # TypeScript source code
│   ├── extension.ts        # Main extension entry point
│   ├── commands/           # Command implementations
│   ├── providers/          # VSCode providers (diagnostics, code actions)
│   ├── services/           # Core services (server, client)
│   └── utils/              # Utility functions
├── test/                   # Test files
├── resources/              # Icons and assets
├── out/                    # Compiled JavaScript
├── package.json            # Extension manifest
└── README.md               # This file
```

### Contributing

Areas for contribution:

- **Language Support**: Add support for additional programming languages
- **Unit Conversions**: Expand unit conversion types and accuracy
- **Performance**: Optimise processing for large files
- **User Experience**: Improve diagnostics and Quick Fix suggestions
- **Testing**: Expand test coverage and add edge cases

## License

MIT License - see [LICENSE](../LICENSE) file for details.

## Changelog

See the github release notes: https://github.com/sammcj/m2e/releases

## Related Projects

- **M2E Core**: [Main M2E project](https://github.com/sammcj/m2e) with CLI, GUI, and server
- **M2E MCP**: Model Context Protocol server for AI integration
- **M2E API**: HTTP API server for integration with other tools

---

**Note**: This extension requires macOS or Linux. Windows is not supported as the project focuses on Unix-like systems.
