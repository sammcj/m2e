# Changelog

All notable changes to the M2E VSCode extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Additional language support for more programming languages
- Enhanced unit conversion accuracy and types
- Performance optimisations for large workspaces
- Configurable diagnostic update frequency
- Telemetry for usage analytics (opt-in)

## [0.1.0] - 2024-08-19

### Added
- **Initial release** of M2E VSCode extension
- **Text conversion commands**:
  - Convert Selection: Transform selected American text to British English
  - Convert File: Convert entire file contents
  - Convert Comments Only: Code-aware conversion preserving syntax
  - Convert and Preview: Preview changes using VSCode's diff viewer
- **Real-time diagnostics**: Highlight American spellings as you type
- **Quick Fix actions**: One-click conversion suggestions
- **Automatic server management**: Zero-configuration setup with bundled binaries
- **Unit conversion support**: Imperial to metric conversions (temperature, distance, weight, volume)
- **Multi-platform support**: macOS (Intel/Apple Silicon) and Linux (x64/ARM64)
- **Code-aware processing**: Intelligent handling of programming language files
- **Comprehensive configuration**: 12+ settings for customising behaviour
- **Status bar integration**: Real-time server status monitoring
- **Context menus**: Right-click conversion options in editor and explorer
- **Keyboard shortcuts**: Quick access to conversion commands
- **File type support**: 15+ programming languages plus text/markdown files
- **Workspace integration**: Project-specific settings and ignore lists
- **User dictionary support**: Custom word mappings via configuration files
- **Exclude patterns**: Configurable file/directory exclusion for processing
- **Error handling**: Graceful degradation and comprehensive error reporting
- **Development tools**: Hot-reload development environment and testing framework

### Features by Category

#### Core Conversion Engine
- British English spelling conversion (200+ word mappings)
- Imperial to metric unit conversion (temperature, distance, weight, volume, area)
- Context-aware word processing (avoiding false positives)
- Code syntax preservation during conversion
- Large file handling with progress indicators

#### VSCode Integration
- Native VSCode diagnostic provider
- Code action provider for Quick Fix suggestions
- Status bar integration with server health monitoring
- Command palette integration with 8 commands
- Context menu integration for editor and explorer
- Keyboard shortcut support (Cmd/Ctrl+Shift+M, etc.)
- Settings UI integration with 12+ configuration options

#### Development Experience
- TypeScript implementation with strict type checking
- Comprehensive error handling and user feedback
- Detailed logging with configurable verbosity
- Hot-reload development environment
- Automated testing framework
- ESLint integration for code quality

#### Platform Support
- **macOS**: Intel and Apple Silicon architectures
- **Linux**: x64 and ARM64 architectures  
- **Bundled binaries**: Pre-compiled m2e-server for all supported platforms
- **Auto-detection**: Automatic platform/architecture detection
- **Custom paths**: Support for user-provided server binaries

#### File Processing
- **Text files**: Plain text, Markdown, reStructuredText
- **Programming languages**: JavaScript, TypeScript, Python, Go, Java, C/C++, C#, PHP, Ruby, Rust
- **Configuration files**: JSON, YAML, XML, INI
- **Smart exclusions**: Automatic exclusion of node_modules, .git, dist, build directories
- **Size limits**: Performance optimisation for large files (100KB+ warnings)

#### Configuration System
- **User settings**: Global configuration via VSCode settings
- **Workspace settings**: Project-specific configuration  
- **User dictionary**: Custom word mappings in ~/.config/m2e/
- **Unit conversion**: Configurable precision and enabled conversions
- **Exclude patterns**: Glob pattern support for file exclusion
- **Debug logging**: Detailed logging for development and troubleshooting

### Technical Implementation

#### Server Management
- **Automatic lifecycle**: Start/stop server as needed
- **Port discovery**: Auto-increment from preferred port (18181)
- **Health monitoring**: Continuous server health checks
- **Process cleanup**: Graceful shutdown on VSCode exit
- **Error recovery**: Automatic restart on server failures

#### Performance Optimisations
- **Lazy activation**: Extension activates only when needed
- **Debounced diagnostics**: Efficient real-time highlighting
- **Streaming processing**: Handle large files without blocking UI
- **Memory management**: Efficient resource usage and cleanup
- **Caching**: Intelligent caching of conversion results

#### Quality Assurance
- **Error boundaries**: Comprehensive error handling at all levels
- **User feedback**: Clear error messages and status communication
- **Logging**: Detailed logging for debugging and support
- **Validation**: Input validation and sanitisation
- **Testing**: Unit and integration test coverage

### Documentation
- **Comprehensive README**: Installation, configuration, usage, and troubleshooting
- **API documentation**: TypeScript interfaces and JSDoc comments
- **Configuration guide**: Detailed explanation of all settings
- **Development setup**: Instructions for extension development
- **Troubleshooting guide**: Common issues and solutions

### Known Limitations
- **Windows not supported**: Extension designed for Unix-like systems only
- **Large file performance**: Files >500KB may experience slower processing
- **Language coverage**: Limited to 15 programming languages currently
- **Unit conversion scope**: Basic imperial-to-metric conversions only
- **Offline operation**: Requires local server for full functionality

---

## Release Notes Format

### Version Numbering
- **Major** (X.0.0): Breaking changes, major feature additions
- **Minor** (0.X.0): New features, improvements, non-breaking changes  
- **Patch** (0.0.X): Bug fixes, documentation updates, small improvements

### Change Categories
- **Added**: New features and functionality
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Features removed in this version
- **Fixed**: Bug fixes and corrections
- **Security**: Security vulnerability fixes

### Migration Guides
When breaking changes are introduced, migration guides will be provided to help users update their configurations and workflows.

### Support Policy
- **Current version**: Full support with bug fixes and security updates
- **Previous major version**: Security updates only
- **Older versions**: No longer supported

---

*For the complete version history and detailed release information, see the [GitHub Releases](https://github.com/sammcj/m2e/releases) page.*