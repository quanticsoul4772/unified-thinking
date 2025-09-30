@echo off
REM Build script for Unified Thinking Server with automatic SDK fetching
REM Run this after Go is installed

echo ============================================
echo Building Unified Thinking Server
echo ============================================
echo.

REM Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH!
    echo.
    echo Please:
    echo 1. Install Go from https://go.dev/dl/
    echo 2. Close this window and open a NEW command prompt
    echo 3. Run this script again
    echo.
    pause
    exit /b 1
)

echo Go is installed:
go version
echo.

echo Navigating to project directory...
cd /d C:\Development\Projects\MCP\project-root\mcp-servers\unified-thinking
echo.

echo ============================================
echo Step 1: Initializing module and fetching SDK...
echo ============================================
echo.

REM Initialize fresh go.mod without version specification
echo Creating go.mod...
echo module unified-thinking> go.mod
echo.>> go.mod
echo go 1.23>> go.mod
echo.>> go.mod

REM Fetch the latest version of the SDK
echo Fetching latest MCP SDK...
go get github.com/modelcontextprotocol/go-sdk/mcp@latest
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to fetch MCP SDK
    echo.
    echo The SDK might not have any tagged releases yet.
    echo Trying to fetch from main branch...
    go get github.com/modelcontextprotocol/go-sdk/mcp@main
    if %errorlevel% neq 0 (
        echo ERROR: Failed to fetch SDK from main branch
        pause
        exit /b 1
    )
)

echo.
echo Running go mod tidy...
go mod tidy
if %errorlevel% neq 0 (
    echo ERROR: go mod tidy failed
    pause
    exit /b 1
)

echo Dependencies configured successfully!
echo.

echo ============================================
echo Step 2: Building server...
echo ============================================
echo.

REM Create bin directory if it doesn't exist
if not exist bin mkdir bin

go build -o bin\unified-thinking.exe .\cmd\server
if %errorlevel% neq 0 (
    echo ERROR: Build failed!
    echo.
    echo Please check for errors above.
    pause
    exit /b 1
)

echo.
echo ============================================
echo Build Complete!
echo ============================================
echo.

REM Verify the executable was created
if exist bin\unified-thinking.exe (
    echo Executable created: bin\unified-thinking.exe
    echo.
    
    REM Show file size
    for %%A in (bin\unified-thinking.exe) do echo File size: %%~zA bytes
    echo.
    
    echo ============================================
    echo Next Steps:
    echo ============================================
    echo.
    echo 1. Update your Claude Desktop config:
    echo    File: %%APPDATA%%\Claude\claude_desktop_config.json
    echo.
    echo 2. Add this configuration:
    echo {
    echo   "mcpServers": {
    echo     "unified-thinking": {
    echo       "command": "C:\\Development\\Projects\\MCP\\project-root\\mcp-servers\\unified-thinking\\bin\\unified-thinking.exe",
    echo       "transport": "stdio",
    echo       "env": {
    echo         "DEBUG": "true"
    echo       }
    echo     }
    echo   }
    echo }
    echo.
    echo 3. Restart Claude Desktop completely
    echo.
    echo 4. Test with prompts like:
    echo    - "Think step by step about..."
    echo    - "Explore multiple branches of..."
    echo    - "What's a creative solution to..."
    echo.
) else (
    echo ERROR: Executable not found after build!
    echo Something went wrong during the build process.
    echo.
)

pause
