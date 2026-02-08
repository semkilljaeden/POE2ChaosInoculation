#!/bin/bash

echo "================================================"
echo "  POE2 Chaos Crafter - VSCode Setup Script"
echo "================================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    echo "   Please install Go from: https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo ""

# Check if Tesseract is installed
if ! command -v tesseract &> /dev/null; then
    echo "⚠ Tesseract OCR is not installed!"
    echo ""
    echo "Installation instructions:"
    echo "  Ubuntu/Debian: sudo apt-get install tesseract-ocr libtesseract-dev"
    echo "  macOS: brew install tesseract"
    echo "  Windows: Download from https://github.com/UB-Mannheim/tesseract/wiki"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✓ Tesseract is installed: $(tesseract --version | head -n1)"
fi

echo ""
echo "Step 1: Initializing Go module..."
go mod init poe2-crafter 2>/dev/null || echo "  (Module already exists)"

echo ""
echo "Step 2: Installing Go dependencies..."
echo "  → Installing robotgo..."
go get github.com/go-vgo/robotgo

echo "  → Installing gosseract..."
go get github.com/otiai10/gosseract/v2

echo ""
echo "Step 3: Creating .vscode directory..."
mkdir -p .vscode

echo ""
echo "Step 4: Copying VSCode configuration files..."
# Files should already be in .vscode from previous creation

echo ""
echo "Step 5: Building the program..."
go build -o poe2_crafter poe2_crafter.go

echo ""
echo "================================================"
echo "  ✓ Setup Complete!"
echo "================================================"
echo ""
echo "VSCode is now configured with:"
echo "  • Build task (Ctrl+Shift+B)"
echo "  • Run task (from Tasks menu)"
echo "  • Debug configuration (F5)"
echo ""
echo "Quick start in VSCode:"
echo "  1. Open this folder in VSCode"
echo "  2. Press Ctrl+Shift+B to build"
echo "  3. Press F5 to run with debugging"
echo "  OR"
echo "  4. Terminal → Run Task → 'Build and Run'"
echo ""
echo "To run from terminal:"
echo "  ./poe2_crafter"
echo ""
