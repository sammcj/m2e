#!/bin/bash
set -e

# Build script for macOS (Apple Silicon optimized)
echo "Building 'Murican to English Converter for macOS (Apple Silicon)..."

# Check for Wails CLI
if ! command -v wails &>/dev/null; then
  echo "Wails CLI not found. Installing..."
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2
fi

# Set environment variables for optimized build
export CGO_ENABLED=1
export GOOS=darwin
export GOARCH=arm64
export MACOSX_DEPLOYMENT_TARGET=15.0

# Clean previous build
rm -rf build/bin

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd frontend && npm install && cd ..

# Build the application with optimizations
echo "Building application..."
wails build -platform darwin/arm64 -o "murican-to-english" -ldflags="-s -w" -trimpath -tags "production"

# Copy custom Info.plist to the app bundle
echo "Applying custom Info.plist with macOS service support..."
cp build/darwin/custom-Info.plist "build/bin/Murican to English.app/Contents/Info.plist"

# Create a DMG for distribution (requires create-dmg)
if command -v create-dmg &>/dev/null; then
  echo "Creating DMG..."
  create-dmg \
    --volname "Murican to English" \
    --volicon "appicon.png" \
    --window-pos 200 120 \
    --window-size 800 400 \
    --icon-size 100 \
    --icon "Murican to English.app" 200 190 \
    --hide-extension "Murican to English.app" \
    --app-drop-link 600 185 \
    "build/bin/murican-to-english-arm64.dmg" \
    "build/bin/Murican to English.app"
  echo "DMG created at build/bin/murican-to-english-arm64.dmg"
else
  echo "create-dmg not found. Skipping DMG creation."
  echo "To create a DMG, install create-dmg: brew install create-dmg"
fi

# Create a zip archive as an alternative
echo "Creating ZIP archive..."
cd build/bin
zip -r "murican-to-english-arm64.zip" "Murican to English.app"
cd ../..

echo "Build completed successfully!"
echo "Application available at:"
echo "  - build/bin/Murican to English.app"
echo "  - build/bin/murican-to-english-arm64.zip"
if command -v create-dmg &>/dev/null; then
  echo "  - build/bin/murican-to-english-arm64.dmg"
fi
