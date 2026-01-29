@echo off
REM Start the Secure Multi-Room Chat with Web Interface
REM This script starts the complete system with all 5 security features

cd /d "%~dp0"

echo.
echo ============================================
echo   Secure Multi-Room Chat System
echo   Starting Web Interface + TCP Server
echo ============================================
echo.

if exist chatish.exe (
    echo [+] Using compiled binary
    chatish.exe web
) else (
    echo [+] Building from source...
    go build -o chatish.exe .
    if %errorlevel% equ 0 (
        echo [+] Build successful
        chatish.exe web
    ) else (
        echo [-] Build failed
        echo [!] Falling back to 'go run'
        go run . web
    )
)
