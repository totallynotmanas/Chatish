@echo off
REM Start the Secure Chat Client
REM Connects to server on 127.0.0.1:5000

cd /d "%~dp0"

echo.
echo ============================================
echo   Secure Chat Client
echo   Connecting to 127.0.0.1:5000
echo ============================================
echo.
echo Test Credentials:
echo   alice / password123 (Admin)
echo   bob / secure456 (Member)
echo   charlie / guest789 (Guest)
echo.

if exist chatish.exe (
    echo [+] Using compiled binary
    chatish.exe client
) else (
    echo [+] Building from source...
    go build -o chatish.exe .
    if %errorlevel% equ 0 (
        echo [+] Build successful
        chatish.exe client
    ) else (
        echo [-] Build failed
        echo [!] Falling back to 'go run'
        go run . client
    )
)
