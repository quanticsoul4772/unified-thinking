@echo off
echo ============================================
echo Go Installation Helper for Windows
echo ============================================
echo.

REM Check if Go is already installed
where go >nul 2>&1
if %errorlevel% equ 0 (
    echo Go is already installed!
    go version
    echo.
    echo If you want to reinstall, please uninstall the current version first.
    pause
    exit /b 0
)

echo Go is not currently installed.
echo.
echo ============================================
echo Step 1: Downloading Go 1.23.4 for Windows
echo ============================================
echo.

set GO_VERSION=1.23.4
set GO_INSTALLER=go%GO_VERSION%.windows-amd64.msi
set DOWNLOAD_URL=https://go.dev/dl/%GO_INSTALLER%

echo Downloading from: %DOWNLOAD_URL%
echo Please wait...
echo.

REM Download using PowerShell
powershell -Command "& {Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TEMP%\%GO_INSTALLER%'}"

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Download failed!
    echo Please download manually from: https://go.dev/dl/
    pause
    exit /b 1
)

echo.
echo Download complete!
echo.
echo ============================================
echo Step 2: Running Go Installer
echo ============================================
echo.
echo The installer will now launch.
echo Please follow the installation wizard:
echo   1. Accept the license agreement
echo   2. Use default installation path (C:\Program Files\Go)
echo   3. Click Install
echo   4. Click Finish when complete
echo.
pause

REM Run the installer
msiexec /i "%TEMP%\%GO_INSTALLER%" /passive /norestart

if %errorlevel% neq 0 (
    echo.
    echo Installation encountered an issue.
    echo Please try running the installer manually from: %TEMP%\%GO_INSTALLER%
    pause
    exit /b 1
)

echo.
echo ============================================
echo Step 3: Verifying Installation
echo ============================================
echo.
echo Waiting for installation to complete...
timeout /t 10 /nobreak >nul

REM Refresh PATH by restarting script with updated environment
echo Please close this window and open a NEW command prompt, then run:
echo   go version
echo.
echo If Go is not found, you may need to:
echo   1. Log out and log back in (to refresh environment)
echo   2. OR restart your computer
echo.
echo After Go is verified, run:
echo   cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
echo   make install-deps
echo   make build
echo.

pause
