# Build Instructions for 'Murican to English Converter

This document provides instructions for building the 'Murican to English Converter application for macOS, with a focus on Apple Silicon optimisation.

## Local Build

### Prerequisites

- macOS 11.0 or later
- Go 1.23 or later
- Node.js 22 or later
- Wails CLI v2.10.1 or later

### Building for macOS (Apple Silicon)

1. Make sure you have the Wails CLI installed:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1
```

2. Run the build script:

```bash
./build-macos.sh
```

This script will:
- Install dependencies if needed
- Build the application optimised for Apple Silicon (arm64)
- Create a ZIP archive of the application
- Optionally create a DMG if you have `create-dmg` installed

The built application will be available at:
- `build/bin/murican-to-english.app` (Application bundle)
- `build/bin/murican-to-english-arm64.zip` (ZIP archive)
- `build/bin/murican-to-english-arm64.dmg` (DMG installer, if `create-dmg` is installed)

### Building for macOS (Universal Binary)

To build a universal binary that works on both Intel and Apple Silicon Macs:

```bash
wails build -platform darwin/universal -o "murican-to-english" -ldflags="-s -w" -trimpath -tags "production"
```

## GitHub Actions Automated Build

The repository includes a GitHub Actions workflow that automatically builds the application for both Apple Silicon and as a Universal Binary.

### Workflow Triggers

The workflow is triggered on:
- Pushes to `main` or `master` branches
- Any tag starting with `v` (e.g., `v1.0.0`)
- Pull requests to `main` or `master` branches

### Build Artifacts

For each build, the workflow:
1. Builds the application
2. Creates a ZIP archive
3. Uploads the ZIP as a GitHub Actions artifact
4. For tagged releases, creates a GitHub Release with the ZIP files attached

### Creating a Release

To create a new release:

1. Tag your commit with a version number:

```bash
git tag v1.0.0
git push origin v1.0.0
```

2. The GitHub Actions workflow will automatically:
   - Build the application
   - Create a GitHub Release
   - Attach the built applications to the release

### Downloading Builds

- For the latest build from the main branch, download the artifacts from the latest GitHub Actions workflow run.
- For stable releases, download from the GitHub Releases page.

## Build Optimisations

The build process includes several optimisations:

- `-ldflags="-s -w"`: Reduces binary size by stripping debug information
- `-trimpath`: Removes file system paths from the binary for reproducible builds
- `-tags "production"`: Enables production-specific code paths
- `MACOSX_DEPLOYMENT_TARGET=11.0`: Sets the minimum macOS version required
- Apple Silicon specific optimisations for arm64 architecture

## Troubleshooting

If you encounter build issues:

1. Ensure you have the correct versions of Go and Node.js installed
2. Update the Wails CLI to the latest version
3. Clear the build cache: `wails build --clean`
4. Check the frontend dependencies: `cd frontend && npm install`
