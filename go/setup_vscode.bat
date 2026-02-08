@echo off
echo ================================================
echo   POE2 Chaos Crafter - VSCode Setup Script
echo ================================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [X] Go is not installed!
    echo     Please install Go from: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo [OK] Go is installed: %GO_VERSION%
echo.

REM Check if Tesseract is installed
where tesseract >nul 2>nul
if %errorlevel% neq 0 (
    echo [!] Tesseract OCR is not installed!
    echo.
    echo Installation instructions:
    echo   Download from: https://github.com/UB-Mannheim/tesseract/wiki
    echo   Install to: C:\Program Files\Tesseract-OCR
    echo   Add to PATH
    echo.
    set /p CONTINUE="Continue anyway? (y/n): "
    if /i not "%CONTINUE%"=="y" exit /b 1
) else (
    for /f "tokens=*" %%i in ('tesseract --version 2^>^&1 ^| findstr /C:"tesseract"') do set TESS_VERSION=%%i
    echo [OK] Tesseract is installed: %TESS_VERSION%
)

echo.
echo Step 1: Initializing Go module...
go mod init poe2-crafter 2>nul
if %errorlevel% equ 0 (
    echo   [OK] Go module initialized
) else (
    echo   [OK] Module already exists
)

echo.
echo Step 2: Installing Go dependencies...
echo   Installing robotgo...
go get github.com/go-vgo/robotgo

echo   Installing gosseract...
go get github.com/otiai10/gosseract/v2

echo.
echo Step 3: Creating .vscode directory...
if not exist ".vscode" mkdir .vscode

echo.
echo Step 4: VSCode configuration files ready...
echo   - tasks.json (build and run tasks)
echo   - launch.json (debug configuration)
echo   - settings.json (Go settings)

echo.
echo Step 5: Building the program...
go build -o poe2_crafter.exe poe2_crafter.go

if %errorlevel% equ 0 (
    echo   [OK] Build successful!
) else (
    echo   [X] Build failed!
    pause
    exit /b 1
)

echo.
echo ================================================
echo   [OK] Setup Complete!
echo ================================================
echo.
echo VSCode is now configured with:
echo   - Build task (Ctrl+Shift+B)
echo   - Run task (Terminal menu ^> Run Task)
echo   - Debug configuration (F5)
echo.
echo Quick start in VSCode:
echo   1. Open this folder in VSCode
echo   2. Press Ctrl+Shift+B to build
echo   3. Press F5 to run with debugging
echo   OR
echo   4. Terminal ^> Run Task ^> 'Build and Run'
echo.
echo To run from command prompt:
echo   poe2_crafter.exe
echo.
pause
