# VSCode Setup Guide for POE2 Chaos Crafter

## Prerequisites

1. **Install Go**
   - Download from: https://go.dev/dl/
   - Install and add to PATH
   - Verify: `go version`

2. **Install Tesseract OCR**
   
   **Windows:**
   - Download: https://github.com/UB-Mannheim/tesseract/wiki
   - Install to: `C:\Program Files\Tesseract-OCR`
   - Add to PATH: `C:\Program Files\Tesseract-OCR`
   - Verify: `tesseract --version`

   **Linux (Ubuntu/Debian):**
   ```bash
   sudo apt-get update
   sudo apt-get install tesseract-ocr libtesseract-dev
   ```

   **macOS:**
   ```bash
   brew install tesseract
   ```

3. **Install VSCode**
   - Download from: https://code.visualstudio.com/
   - Install the **Go extension** (by Go Team at Google)

4. **Install Build Tools**
   
   **Windows:**
   - Install MinGW-w64 or TDM-GCC
   - Or install Visual Studio Build Tools

   **Linux:**
   ```bash
   sudo apt-get install gcc libx11-dev libxtst-dev libpng-dev
   ```

   **macOS:**
   ```bash
   xcode-select --install
   ```

## Quick Setup

### Automatic Setup (Recommended)

**Windows:**
```cmd
setup_vscode.bat
```

**Linux/macOS:**
```bash
chmod +x setup_vscode.sh
./setup_vscode.sh
```

This will:
- ✓ Check dependencies
- ✓ Initialize Go module
- ✓ Install Go packages
- ✓ Create VSCode config files
- ✓ Build the program

### Manual Setup

If you prefer manual setup:

1. **Initialize Go module:**
   ```bash
   go mod init poe2-crafter
   ```

2. **Install dependencies:**
   ```bash
   go get github.com/go-vgo/robotgo
   go get github.com/otiai10/gosseract/v2
   ```

3. **Copy VSCode config files:**
   - Copy `.vscode/tasks.json`
   - Copy `.vscode/launch.json`
   - Copy `.vscode/settings.json`

4. **Build:**
   ```bash
   go build poe2_crafter.go
   ```

## Using VSCode

### Method 1: Build and Run Tasks

1. **Build the program:**
   - Press `Ctrl+Shift+B` (or `Cmd+Shift+B` on Mac)
   - Or: `Terminal` → `Run Build Task`
   - This compiles `poe2_crafter.go` → `poe2_crafter.exe`

2. **Run the program:**
   - Press `Ctrl+Shift+P` (or `Cmd+Shift+P`)
   - Type: `Tasks: Run Task`
   - Select: `Run POE2 Crafter`

3. **Build and Run together:**
   - `Ctrl+Shift+P` → `Tasks: Run Task`
   - Select: `Build and Run`

### Method 2: Debug Mode (F5)

1. **Set breakpoints** (optional)
   - Click left of line numbers to add breakpoints

2. **Start debugging:**
   - Press `F5`
   - Or: `Run` → `Start Debugging`

3. **Debug controls:**
   - Continue (F5)
   - Step Over (F10)
   - Step Into (F11)
   - Stop (Shift+F5)

### Method 3: Integrated Terminal

1. **Open terminal in VSCode:**
   - Press `` Ctrl+` `` (backtick)
   - Or: `Terminal` → `New Terminal`

2. **Build:**
   ```bash
   go build poe2_crafter.go
   ```

3. **Run:**
   ```bash
   # Windows
   .\poe2_crafter.exe

   # Linux/Mac
   ./poe2_crafter
   ```

## VSCode Keyboard Shortcuts

| Action | Windows/Linux | macOS |
|--------|---------------|-------|
| Build | `Ctrl+Shift+B` | `Cmd+Shift+B` |
| Debug | `F5` | `F5` |
| Run Task | `Ctrl+Shift+P` | `Cmd+Shift+P` |
| Terminal | `` Ctrl+` `` | `` Cmd+` `` |
| Save All | `Ctrl+K S` | `Cmd+K S` |

## VSCode Tasks Available

All available in: `Terminal` → `Run Task`

1. **Build POE2 Crafter** - Compile the program
2. **Run POE2 Crafter** - Run the compiled program
3. **Build and Run** - Build then run automatically
4. **Install Dependencies** - Re-install Go packages
5. **Clean Build** - Clean build artifacts

## Folder Structure

```
poe2-crafter/
├── poe2_crafter.go          # Main source code
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums (auto-generated)
├── poe2_crafter.exe         # Compiled program (Windows)
├── poe2_crafter             # Compiled program (Linux/Mac)
├── .vscode/
│   ├── tasks.json           # Build/run tasks
│   ├── launch.json          # Debug configuration
│   └── settings.json        # Go settings
├── setup_vscode.bat         # Windows setup script
├── setup_vscode.sh          # Linux/Mac setup script
├── debug_*.png              # Debug screenshots (if enabled)
└── success_*.png            # Success screenshots
```

## Customizing Build

### Change Output Name

Edit `.vscode/tasks.json`:
```json
"args": [
    "build",
    "-o",
    "my_custom_name${executableExtension}",  // Change this
    "${workspaceFolder}/poe2_crafter.go"
]
```

### Add Build Flags

For smaller binary:
```json
"args": [
    "build",
    "-ldflags", "-s -w",  // Strip debug info
    "-o", "poe2_crafter${executableExtension}",
    "${workspaceFolder}/poe2_crafter.go"
]
```

For static binary (Linux):
```json
"args": [
    "build",
    "-ldflags", "-extldflags -static",
    "-o", "poe2_crafter",
    "${workspaceFolder}/poe2_crafter.go"
]
```

## Troubleshooting

### "go: command not found"

**Problem:** Go not in PATH

**Solution:**
- Windows: Add Go to PATH via System Environment Variables
- Linux/Mac: Add to `.bashrc` or `.zshrc`:
  ```bash
  export PATH=$PATH:/usr/local/go/bin
  ```

### "tesseract: command not found"

**Problem:** Tesseract not installed or not in PATH

**Solution:**
- Install Tesseract (see Prerequisites)
- Add to PATH
- Restart VSCode

### Build fails with "gcc: command not found"

**Problem:** C compiler not installed (needed for robotgo)

**Solution:**
- Windows: Install MinGW-w64 or TDM-GCC
- Linux: `sudo apt-get install gcc`
- macOS: `xcode-select --install`

### "cannot find package" error

**Problem:** Dependencies not installed

**Solution:**
```bash
go mod download
go get github.com/go-vgo/robotgo
go get github.com/otiai10/gosseract/v2
```

### VSCode Go extension not working

**Problem:** Extension not installed or Go tools missing

**Solution:**
1. Install Go extension from VSCode marketplace
2. Press `Ctrl+Shift+P` → `Go: Install/Update Tools`
3. Select all tools → OK

### Debug mode not working

**Problem:** Delve debugger not installed

**Solution:**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

## Tips for Development

1. **Auto-format on save:** Already configured in `settings.json`

2. **Auto-import packages:** Type the package name, VSCode will suggest import

3. **View function signatures:** Hover over function names

4. **Go to definition:** `F12` or `Ctrl+Click`

5. **Find references:** `Shift+F12`

6. **Rename symbol:** `F2`

7. **Code snippets:** Type `for`, `if`, `func`, etc. and press Tab

## Running from Terminal

After building, you can run from any terminal:

**Windows:**
```cmd
cd path\to\poe2-crafter
poe2_crafter.exe
```

**Linux/macOS:**
```bash
cd path/to/poe2-crafter
./poe2_crafter
```

## Next Steps

Once setup is complete:

1. ✓ Open folder in VSCode
2. ✓ Press `Ctrl+Shift+B` to build
3. ✓ Press `F5` to run
4. ✓ Follow the interactive setup wizard
5. ✓ Start crafting!

## Support

If you encounter issues:

1. Check this README
2. Verify all prerequisites are installed
3. Run setup script again
4. Check VSCode Output panel for errors
5. Check Go extension status in VSCode

Happy coding! 🚀
