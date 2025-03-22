# Freedom Translator - 'Murican to English Converter

A modern, lightweight application for converting text from American to English spellings.

## Features

- Convert text from American to English (and in reverse)
- Clean, modern user interface
- Fast and responsive
- Native desktop application for macOS

## Screenshots

![Application Screenshot](screenshots/app-screenshot.png)

## Technology Stack

- **Backend**: Go
- **Frontend**: React
- **Framework**: Wails v2.10
- **Build System**: Go modules and npm

## Development

### Prerequisites

- Go 1.23 or later
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

### Development Mode

To run the application in development mode:

```bash
wails dev
```

This will start the application with hot reloading for both frontend and backend changes.

### Building

To build the application for production:

```bash
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
│       └── data/         # JSON dictionaries
├── main.go               # Main application entry
├── app.go                # Application setup and binding to frontend
└── wails.json            # Wails configuration
```

## How It Works

The application uses JSON dictionaries to map between American and English spellings. The conversion logic is implemented in Go, which provides fast and efficient text processing. The frontend is built with React, providing a modern and responsive user interface.

## License

- Copyright © 2025 Sam McLeod
- [MIT](LICENSE)
