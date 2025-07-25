# M2E - 'Murican to English Converter

A lightweight application for converting text from American to International English spellings.

## Features

- Converts pasted text from American to International English
- Freedom Unit to standard metric unit conversion**: Automatically converts imperial units (feet, pounds, °F, etc.) to standard metric equivalents
- Fast and responsive and minimalist interface
- Native desktop application for macOS
- Also gets rid of those pesky "smart" quotes and em-dashes that break everything
- CLI support for file conversion
- Clipboard conversion
- MCP (~~Murican Conversion Protocol~~ Model Context Protocol) server for use with AI agents and agentic coding tools
- Code-aware conversion that preserves code syntax while converting comments (BETA)
- Configurable unit conversion with user preferences
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
  - [Freedom Unit Conversion](#freedom-unit-conversion)
    - [Supported Unit Types](#supported-unit-types)
    - [Examples](#examples)
    - [Configuration](#configuration)
    - [Interface Integration](#interface-integration)
    - [Troubleshooting](#troubleshooting)
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

## Freedom Unit Conversion

M2E includes intelligent imperial-to-metric unit conversion that works alongside spelling conversion. The feature is designed to be code-aware, converting units only in appropriate contexts while preserving code functionality.

### Supported Unit Types

- **Length**: feet, inches, yards, miles → metres, centimetres, kilometres
- **Mass**: pounds, ounces, tons → kilograms, grams, tonnes  
- **Volume**: gallons, quarts, pints, fluid ounces → litres, millilitres
- **Temperature**: Fahrenheit → Celsius
- **Area**: square feet, acres → square metres, hectares

### Examples

**Basic conversions:**
```
"The room is 12 feet wide" → "The room is 3.7 metres wide"
"Temperature was 75°F" → "Temperature was 24°C"
"The package weighs 5 pounds" → "The package weighs 2.3 kg"
"I drove 10 miles to work" → "I drove 16 km to work"
```

**Code-aware processing:**
```go
// The buffer should be 1024 bytes in size (no conversion - bytes not imperial)
// Set the width to 100 inches for display → Set the width to 254 cm for display
const ROOM_WIDTH_FEET = 12  // ← No conversion in variable names or values
```

**Smart detection (avoids idioms):**
```
"I'm miles away from home" → (no conversion - idiomatic usage)
"They moved inch by inch" → (no conversion - idiomatic usage)
"The room is 6 feet tall" → "The room is 1.8 metres tall" (converts measurements)
```

### Configuration

Unit conversion can be customized through a configuration file at `$HOME/.config/m2e/unit_config.json`.

**Create example configuration:**

The configuration file is automatically created with default values when first needed. You can manually create and edit `$HOME/.config/m2e/unit_config.json` to customize unit conversion behavior.

**Configuration options:**

```json
{
  "enabled": true,
  "enabledUnitTypes": ["length", "mass", "volume", "temperature", "area"],
  "precision": {
    "length": 1,
    "mass": 1,
    "volume": 1,
    "temperature": 0,
    "area": 1
  },
  "customMappings": {
    "customize": "customise"
  },
  "excludePatterns": [
    "miles?\\s+(?:away|apart|from\\s+home|ahead)",
    "inch\\s+by\\s+inch"
  ],
  "preferences": {
    "preferWholeNumbers": true,
    "maxDecimalPlaces": 2,
    "temperatureFormat": "°C",
    "useSpaceBetweenValueAndUnit": true,
    "roundingThreshold": 0.1
  },
  "detection": {
    "minConfidence": 0.5,
    "maxNumberDistance": 3,
    "detectCompoundUnits": true,
    "detectWrittenNumbers": true
  }
}
```

**Key configuration options:**

- `enabled`: Enable/disable all unit conversion
- `enabledUnitTypes`: Array of unit types to convert
- `precision`: Decimal places for each unit type
- `customMappings`: Custom unit mappings (American → British)
- `excludePatterns`: Regex patterns to exclude from conversion
- `preferences.preferWholeNumbers`: Round to whole numbers when close (e.g., 2.98 → 3)
- `preferences.temperatureFormat`: Use "°C" or "degrees Celsius"
- `detection.minConfidence`: Minimum confidence (0.0-1.0) to convert a detected unit
- `detection.maxNumberDistance`: Maximum words between number and unit

### Interface Integration

Unit conversion is available across all interfaces:

**GUI**: Toggle unit conversion in the application interface

**CLI**: Use the `-units` flag
```bash
m2e-cli -units "The room is 12 feet wide"
echo "Temperature was 75°F" | m2e-cli -units
```

**MCP Server**: Add `convert_units` parameter
```json
{
  "name": "convert_text",
  "arguments": {
    "text": "The room is 12 feet wide",
    "convert_units": "true"
  }
}
```

**API Server**: Include `convert_units` in request body
```json
{
  "text": "The room is 12 feet wide",
  "convert_units": true
}
```

### Troubleshooting

For detailed troubleshooting help, see the [Unit Conversion Troubleshooting Guide](docs/UNIT_CONVERSION_GUIDE.md).

**Quick fixes:**

1. **Units not converting**: Check if unit conversion is enabled and the unit type is in `enabledUnitTypes`
2. **Unexpected conversions**: Add exclusion patterns for specific phrases in `excludePatterns`
3. **Wrong precision**: Adjust precision settings for each unit type
4. **Configuration errors**: Validate JSON syntax and check error messages

**Reset to defaults:**
```bash
# Remove user configuration to use defaults
rm ~/.config/m2e/unit_config.json
```

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
m2e-cli -input yourfile.txt -output converted.txt -units  # with unit conversion
```

If built from source:
```bash
./build/bin/m2e-cli -input yourfile.txt -output converted.txt
./build/bin/m2e-cli -input yourfile.txt -output converted.txt -units  # with unit conversion
```

**Convert piped text:**

If installed via `go install`:
```bash
echo "I love color and flavor." | m2e-cli
echo "The room is 12 feet wide." | m2e-cli -units  # with unit conversion
```

If built from source:
```bash
echo "I love color and flavor." | ./build/bin/m2e-cli
echo "The room is 12 feet wide." | ./build/bin/m2e-cli -units  # with unit conversion
```

**CLI Options:**
- `-input`: Input file to convert (reads from stdin if not specified)
- `-output`: Output file to write to (writes to stdout if not specified)  
- `-units`: Freedom Unit Conversion (default: false)
- `-no-smart-quotes`: Disable smart quote normalisation (default: false)
- `-h`, `-help`: Show help message

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
- `convert_text`: Converts American English text to British English with optional unit conversion
  - Parameters: 
    - `text` (string, required) - The text to convert
    - `convert_units` (string, optional) - Freedom Unit Conversion ("true"/"false", default: "false")
    - `normalise_smart_quotes` (string, optional) - Normalise smart quotes to regular quotes ("true"/"false", default: "true")
- `convert_file`: Converts a file from American English to British English and saves it back
  - Parameters: 
    - `file_path` (string, required) - The fully qualified path to the file to convert
    - `convert_units` (string, optional) - Freedom Unit Conversion ("true"/"false", default: "false")
    - `normalise_smart_quotes` (string, optional) - Normalise smart quotes to regular quotes ("true"/"false", default: "true")
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
      "text": "I love color and flavor. The room is 12 feet wide.",
      "convert_units": "true"
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
      "file_path": "/path/to/your/file.txt",
      "convert_units": "true",
      "normalise_smart_quotes": "true"
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

  Converts text from American to British English with optional unit conversion.

  **Request Body:**
  ```json
  {
    "text": "I love color and flavor. The room is 12 feet wide.",
    "convert_units": true,
    "normalise_smart_quotes": true
  }
  ```

  **Parameters:**
  - `text` (string, required): The text to convert
  - `convert_units` (boolean, optional): Freedom Unit Conversion (default: false)
  - `normalise_smart_quotes` (boolean, optional): Normalise smart quotes to regular quotes (default: true)

  **Response:**
  ```json
  {
    "text": "I love colour and flavour. The room is 3.7 metres wide."
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
