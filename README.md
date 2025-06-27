# M2E - 'Murican to English Converter

A lightweight application for converting text from American to International English spellings.

## Features

- Converts pasted text from American to International English
- Fast and responsive and minimalist interface
- Native desktop application for macOS
- Also gets rid of those pesky "smart" quotes and em-dashes that break everything
- CLI support for file conversion
- Clipboard conversion
- MCP (~~Murican Conversion Protocol~~ Model Context Protocol) server for use with AI agents and agentic coding tools
- Code-aware conversion that preserves code syntax while converting comments (BETA)
- macOS Services integration

GUI
![smaller cars please](screenshots/app-screenshot.png)

MCP Server Use
![MCP Server](screenshots/mcp-screenshot.png)

---

- [M2E - 'Murican to English Converter](#m2e---murican-to-english-converter)
  - [Features](#features)
  - [Installation](#installation)
    - [GUI](#gui)
    - [CLI](#cli)
    - [MCP Server](#mcp-server)
  - [How It Works](#how-it-works)
    - [Adding New Words](#adding-new-words)
    - [macOS Services Integration](#macos-services-integration)
  - [Technology Stack](#technology-stack)
  - [Development](#development)
    - [Prerequisites](#prerequisites)
    - [Development](#development-1)
    - [CLI Usage](#cli-usage)
    - [Clipboard Usage](#clipboard-usage)
    - [MCP Server Usage](#mcp-server-usage)
    - [API Usage](#api-usage)
    - [Development Mode](#development-mode)
    - [Building](#building)
  - [Project Structure](#project-structure)
  - [License](#license)

Simply paste your freedom text in the left and on the right you'll find the international English spelling.

Note: There is a known issue with typing directly in the freedom text box, I'll get to it eventually but for now pasting works fine.

## Installation

### GUI

- Visit [releases](https://github.com/sammcj/m2e/releases) and download the latest zip

### CLI

```bash
go install github.com/sammcj/m2e/cmd/m2e-cli@HEAD
```

### MCP Server

```bash
go install github.com/sammcj/m2e/cmd/m2e-mcp@HEAD
```

---

## How It Works

The application uses JSON dictionaries to map between American and English spellings. The conversion logic is implemented in Go, which provides fast and efficient text processing. The frontend is built with React, providing a modern and responsive user interface.

### Adding New Words

There are two ways to add new words to the dictionary:

1. **User Dictionary (Recommended)**: The application automatically creates a user dictionary at `$HOME/.config/m2e/american_spellings.json` when first run. You can edit this file to add your own custom word mappings. This file will be merged with the built-in dictionary, and user entries will override built-in ones.

   Example user dictionary:
   ```json
   {
     "customize": "customise",
     "myword": "myword-british",
     "color": "colour-custom"
   }
   ```

2. **Built-in Dictionary**: For permanent additions to the application, edit the [american_spellings.json](pkg/converter/data/american_spellings.json) file. This requires rebuilding the application as the dictionary is embedded at build time.

The user dictionary provides several advantages:
- No need to rebuild the application
- Survives application updates
- Can override built-in dictionary entries
- Robust error handling - invalid JSON will show a warning but won't break the application
- Automatically created with an example entry on first run

### macOS Services Integration

The application integrates with macOS Services, allowing you to convert text from any application.

1. **Convert selected text**: Select text in any application, right-click, and choose "Convert to British English" from the Services menu.
2. **Convert files**: Right-click on a file in Finder, and select "Convert File to British English" from the Services menu to convert the file's contents.

This feature makes it easy to convert text without having to open the application directly. After installation, you may need to log out and log back in for the services to be registered with macOS.

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

### Development

1. Clone the repository

```bash
git clone https://github.com/sammcj/m2e.git
cd m2e
```

2. Install dependencies

```bash
go mod tidy
```

3. Build

The project includes a Makefile for common development tasks:

```bash
make          # Run lint, test, and build (default)
make lint     # Run linter and format check
make test     # Run all tests
make build    # Build the application
make clean    # Clean build artifacts
```

### CLI Usage

The application can be run from the command line to convert files or piped text.

**Build the CLI from source:**
```bash
make build-cli
```

**Convert a file:**

If installed via `go install`:
```bash
m2e-cli -input yourfile.txt -output converted.txt
```

If built from source:
```bash
./build/bin/m2e-cli -input yourfile.txt -output converted.txt
```

**Convert piped text:**

If installed via `go install`:
```bash
echo "I love color and flavor." | m2e-cli
```

If built from source:
```bash
echo "I love color and flavor." | ./build/bin/m2e-cli
```

### Clipboard Usage

The CLI can also be used to convert the contents of the clipboard directly.

**Convert clipboard text:**

Set the `M2E_CLIPBOARD` environment variable to `1` or `true` and call the CLI:

```bash
M2E_CLIPBOARD=1 m2e-cli
```

The converted text will be copied back to the clipboard.

---

### MCP Server Usage

The application can be run as an MCP (Model Context Protocol) server to provide conversion functionality to AI agents and tools.

**Build the MCP server from source:**
```bash
make build-mcp
```

**Run the MCP server in Streamable HTTP mode:**

If installed via `go install`:
```bash
m2e-mcp
```

If built from source:
```bash
./build/bin/m2e-mcp
```
The server will serve up on /mcp and start on port 8081 by default. You can change this by setting the `MCP_PORT` environment variable.

MCP client configuration:

```json
{
  "mcpServers": {
    "m2e": {
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

If installed via `go install`:
```bash
MCP_TRANSPORT=stdio m2e-mcp
```

If built from source:
```bash
MCP_TRANSPORT=stdio ./build/bin/m2e-mcp
```

**Available Tools:**
- `convert_text`: Converts American English text to British English
  - Parameters: `text` (string, required) - The text to convert
- `convert_file`: Converts a file from American English to British English and saves it back
  - Parameters: `file_path` (string, required) - The fully qualified path to the file to convert
  - Uses intelligent processing: for plain text files (.txt, .md, etc.), converts all text but preserves code within markdown blocks. For code/config files (.go, .js, .py, etc.), only converts comments to preserve functionality.

**Available Resources:**
- `dictionary://american-to-british`: Access to the American-to-British dictionary mapping

**Example MCP client usage:**

Convert text:
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

Convert a file:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "convert_file",
    "arguments": {
      "file_path": "/path/to/your/file.txt"
    }
  },
  "id": 2
}
```

---

### API Usage

The application can be run as an API server to provide conversion functionality programmatically.

**Install the API server via go install:**
```bash
go install github.com/sammcj/m2e/cmd/m2e-server@HEAD
```

**Build the server from source:**
```bash
make build-server
```

**Run the server:**

If installed via `go install`:
```bash
m2e-server
```

If built from source:
```bash
./build/bin/m2e-server
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

---

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

---

## Project Structure

```
m2e/
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
│   ├── chroma_test.go    # Syntax highlighting tests
│   └── mcp_convert_file_test.go # MCP convert_file tool tests
├── Makefile              # Development automation
├── main.go               # Main application entry
├── app.go                # Application setup and binding to frontend
└── wails.json            # Wails configuration
```
## License

- Copyright © 2025 Sam McLeod
- [MIT](LICENSE)
