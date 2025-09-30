# Go Installation Script for Windows
# Run this in PowerShell as Administrator

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Go Installation Helper for Windows" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is already installed
$goExists = Get-Command go -ErrorAction SilentlyContinue
if ($goExists) {
    Write-Host "Go is already installed!" -ForegroundColor Green
    go version
    Write-Host ""
    Write-Host "If you want to reinstall, please uninstall the current version first." -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 0
}

Write-Host "Go is not currently installed." -ForegroundColor Yellow
Write-Host ""

# Configuration
$GO_VERSION = "1.23.4"
$GO_INSTALLER = "go$GO_VERSION.windows-amd64.msi"
$DOWNLOAD_URL = "https://go.dev/dl/$GO_INSTALLER"
$INSTALLER_PATH = "$env:TEMP\$GO_INSTALLER"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Step 1: Downloading Go $GO_VERSION" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Downloading from: $DOWNLOAD_URL" -ForegroundColor White
Write-Host "Saving to: $INSTALLER_PATH" -ForegroundColor White
Write-Host "Please wait..." -ForegroundColor Yellow
Write-Host ""

try {
    # Download the installer
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $INSTALLER_PATH -UseBasicParsing
    Write-Host "Download complete!" -ForegroundColor Green
} catch {
    Write-Host ""
    Write-Host "ERROR: Download failed!" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please download manually from: https://go.dev/dl/" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Step 2: Installing Go" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "The installer will now launch." -ForegroundColor White
Write-Host "Please follow the installation wizard:" -ForegroundColor White
Write-Host "  1. Accept the license agreement" -ForegroundColor White
Write-Host "  2. Use default installation path (C:\Program Files\Go)" -ForegroundColor White
Write-Host "  3. Click Install" -ForegroundColor White
Write-Host "  4. Click Finish when complete" -ForegroundColor White
Write-Host ""
Read-Host "Press Enter to start the installer"

try {
    # Run the installer
    Start-Process msiexec.exe -ArgumentList "/i `"$INSTALLER_PATH`" /passive /norestart" -Wait
    Write-Host ""
    Write-Host "Installation completed!" -ForegroundColor Green
} catch {
    Write-Host ""
    Write-Host "Installation encountered an issue." -ForegroundColor Red
    Write-Host "Please try running the installer manually from: $INSTALLER_PATH" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Step 3: Next Steps" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "IMPORTANT: You must close this window and open a NEW PowerShell window!" -ForegroundColor Yellow
Write-Host ""
Write-Host "Then verify the installation by running:" -ForegroundColor White
Write-Host "  go version" -ForegroundColor Cyan
Write-Host ""
Write-Host "If Go is not found, you may need to:" -ForegroundColor White
Write-Host "  1. Log out and log back in to Windows" -ForegroundColor White
Write-Host "  2. OR restart your computer" -ForegroundColor White
Write-Host ""
Write-Host "After Go is verified, build the unified-thinking server:" -ForegroundColor White
Write-Host "  cd C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking" -ForegroundColor Cyan
Write-Host "  go mod download" -ForegroundColor Cyan
Write-Host "  go build -o bin\unified-thinking.exe .\cmd\server" -ForegroundColor Cyan
Write-Host ""
Read-Host "Press Enter to exit"
