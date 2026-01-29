@echo off
REM Start the Secure Chat Server
REM Clients can connect via CLI or Web

cd /d "%~dp0"

echo.
echo ============================================
echo   Secure Chat Server
echo   Listening on 127.0.0.1:5000
echo ============================================
echo.

if exist chatish.exe (
    echo [+] Using compiled binary
    chatish.exe server
) else (
    echo [+] Building from source...
    go build -o chatish.exe .
    if %errorlevel% equ 0 (
        echo [+] Build successful
        chatish.exe server
    ) else (
        echo [-] Build failed
        echo [!] Falling back to 'go run'
        go run . server
    )
)
