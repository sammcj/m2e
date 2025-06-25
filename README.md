# Freedom Translator - 'Murican to English Converter

A modern, lightweight application for converting text from American to International English spellings.

## Features

- Converts pasted text from American to International English
- Fast and responsive and minimalist interface
- Native desktop application for macOS
- Also gets rid of those pesky "smart" quotes and em-dashes that break everything
- CLI support for file conversion
- MCP (~~Muricon Conversion~~ Model Context Protocol) server for use with AI agents and tools
- Code-aware conversion that preserves code syntax while converting comments (BETA)
- macOS Services integration (WIP)

---

- [Freedom Translator - 'Murican to English Converter](#freedom-translator---murican-to-english-converter)
  - [Features](#features)
    - [CLI Usage](#cli-usage)
    - [API Usage](#api-usage)
    - [MCP Server Usage](#mcp-server-usage)
  - [Installation](#installation)
  - [Technology Stack](#technology-stack)
  - [Development](#development)
    - [Prerequisites](#prerequisites)
    - [Setup](#setup)
    - [Development Commands](#development-commands)
    - [Development Mode](#development-mode)
    - [Building](#building)
  - [Project Structure](#project-structure)
  - [How It Works](#how-it-works)
    - [macOS Services Integration](#macos-services-integration)
  - [License](#license)

### CLI Usage

The application can be run from the command line to convert files or piped text.

**Build the CLI:**
```bash
make build-cli
```

**Convert a file:**
```bash
./build/bin/murican-to-english-cli -input yourfile.txt -output converted.txt
```

**Convert piped text:**
```bash
echo "I love color and flavor." | ./build/bin/murican-to-english-cli
```

### API Usage

The application can be run as an API server to provide conversion functionality programmatically.

**Build the server:**
```bash
make build-server
```

**Run the server:**
```bash
./build/bin/murican-to-english-server
```
The server will start on port 8080 by default. You can change this by setting the `API_PORT` environment variable.

**Endpoints:**

- `POST /api/v1/convert`

  Converts text from American to British English.

  **Request Body:**
  ```json
  {
    "text": "I love color and flavor."
  }
  ```

  **Response:**
  ```json
  {
    "text": "I love colour and flavour."
  }
  ```

- `GET /api/v1/health`

  Returns a 200 OK status if the server is running.

### MCP Server Usage

The application can be run as an MCP (Model Context Protocol) server to provide conversion functionality to AI agents and tools.

**Build the MCP server:**
```bash
make build-mcp
```

**Run the MCP server in Streamable HTTP mode:**
```bash
./build/bin/murican-to-english-mcp
```
The server will serve up on /mcp and start on port 8081 by default. You can change this by setting the `MCP_PORT` environment variable.

MCP client configuration:

```json
{
  "mcpServers": {
    "murican-to-english": {
      "timeout": 60,
      "type": "streamableHttp",
      "url": "http://localhost:8081/mcp",
      "autoApprove": [
        "convert_text"
      ]
    }
  }
}
```

**Run the MCP server in STDIO mode:**
```bash
MCP_TRANSPORT=stdio ./build/bin/murican-to-english-mcp
```

**Available Tools:**
- `convert_text`: Converts American English text to British English
  - Parameters: `text` (string, required) - The text to convert

**Available Resources:**
- `dictionary://american-to-british`: Access to the American-to-British dictionary mapping

**Example MCP client usage:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "convert_text",
    "arguments": {
      "text": "I love color and flavor."
    }
  },
  "id": 1
}
```

![smaller cars please](screenshots/app-screenshot.png)

Simply paste your freedom text in the left and on the right you'll find the international English spelling.

Note: There is a known issue with typing directly in the freedom text box, I'll get to it eventually but for now pasting works fine.

## Installation

- Visit [releases](https://github.com/sammcj/murican-to-english/releases) and download the latest zip

---

## Technology Stack

- **Backend**: Go >= 1.24
- **Frontend**: React >= 18
- **Framework**: Wails v2.10
- **Build System**: Go modules and npm

## Development

### Prerequisites

- Go 1.24 or later
- Node.js 22 or later
- Wails CLI v2.10.1 or later

### Setup

1. Install the Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1
```

2. Clone the repository:

```bash
git clone https://github.com/sammcj/murican-to-english.git
cd murican-to-english
```

3. Install dependencies:

```bash
wails dev
```

### Development Commands

The project includes a Makefile for common development tasks:

```bash
make          # Run lint, test, and build (default)
make lint     # Run linter and format check
make test     # Run all tests
make build    # Build the application
make clean    # Clean build artifacts
```

### Development Mode

To run the application in development mode:

```bash
wails dev
```

This will start the application with hot reloading for both frontend and backend changes.

### Building

To build the application for production:

```bash
make build
# or
./build-macos.sh
```

This will create a native application optimised for Apple Silicon in the `build/bin` directory.

For more detailed build instructions, including GitHub Actions automated builds and universal binary creation, see [BUILD.md](BUILD.md).

## Project Structure

```
murican-to-english/
├── build/                # Build artifacts
├── frontend/             # Frontend code using React
│   ├── src/
│   │   ├── App.jsx       # Main application component
│   │   └── App.css       # Application styles
│   ├── index.html
│   └── package.json
├── pkg/
│   └── converter/        # Go package for conversion logic
│       ├── converter.go  # Main conversion functionality
│       ├── codeaware.go  # Code-aware conversion with syntax highlighting
│       └── data/         # JSON dictionaries
├── tests/                # Test files
│   ├── converter_test.go # Basic conversion tests
│   ├── codeaware_test.go # Code-aware functionality tests
│   └── chroma_test.go    # Syntax highlighting tests
├── Makefile              # Development automation
├── main.go               # Main application entry
├── app.go                # Application setup and binding to frontend
└── wails.json            # Wails configuration
```

## How It Works

The application uses JSON dictionaries to map between American and English spellings. The conversion logic is implemented in Go, which provides fast and efficient text processing. The frontend is built with React, providing a modern and responsive user interface.

### macOS Services Integration

The application integrates with macOS Services, allowing you to:

1. **Convert selected text**: Select text in any application, right-click, and choose "Convert to British English" from the Services menu.
2. **Convert files**: Right-click on a file in Finder, and select "Convert File to British English" from the Services menu to convert the file's contents.

This feature makes it easy to convert text without having to open the application directly. After installation, you may need to log out and log back in for the services to be registered with macOS.

## License

- Copyright © 2025 Sam McLeod
- [MIT](LICENSE)
