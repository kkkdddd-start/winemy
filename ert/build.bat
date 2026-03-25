@echo off
REM ERT (Windows Emergency Response Tool) Build Script
REM This script builds both GUI (Wails) and CLI versions

setlocal enabledelayedexpansion

REM Project paths
set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

REM Build configuration
set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ERT.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"

REM Get build time using PowerShell
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set "BUILD_TIME=%%a"

REM Set ldflags
set "LDFLAGS=-s -H windowsgui"
set "LDFLAGS=!LDFLAGS! -X main.Version=%VERSION%"
set "LDFLAGS=!LDFLAGS! -X main.BuildTime=%BUILD_TIME%"

REM Create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%"
)

echo.
echo ===============================================
echo   Windows Emergency Response Tool v%VERSION%
echo   Build Script
echo ===============================================
echo.

:check_env
echo [INFO] Checking build environment...

where go >nul 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21+
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set "GO_VERSION=%%i"
echo [INFO] Go version: !GO_VERSION!

where wails >nul 2>&1
if !ERRORLEVEL! equ 0 (
    for /f "tokens=*" %%i in ('wails version') do set "WAILS_VERSION=%%i"
    echo [INFO] Wails version: !WAILS_VERSION!
) else (
    echo [WARN] Wails is not installed, GUI build will be skipped
    echo [INFO] Install Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest
)

where node >nul 2>&1
if !ERRORLEVEL! equ 0 (
    for /f "tokens=*" %%i in ('node --version') do echo [INFO] Node.js version: %%i
) else (
    echo [WARN] Node.js is not installed
)

echo [SUCCESS] Environment check passed
echo.

:build_gui
echo [INFO] Building GUI version (Wails)...

REM Check if wails is available
where wails >nul 2>&1
if !ERRORLEVEL! neq 0 (
    echo [WARN] Wails not found, skipping GUI build
    echo [INFO] GUI build requires Wails to be installed
    echo [INFO] Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    goto :build_cli
)

REM Clean old build artifacts
if exist "%GUI_OUTPUT%" (
    echo [INFO] Cleaning old GUI build...
    del /q "%GUI_OUTPUT%" 2>nul
)

REM Build frontend first
echo [INFO] Building frontend...
cd /d "%PROJECT_DIR%app"
if exist "package.json" (
    call npm install 2>nul
    if exist "node_modules" (
        echo [INFO] Frontend dependencies installed
    )
    call npm run build
    if exist "dist" (
        echo [SUCCESS] Frontend build succeeded
    ) else (
        echo [WARN] Frontend build may have issues
    )
) else (
    echo [WARN] Frontend directory not found
)
cd /d "%PROJECT_DIR%"

REM Build Wails application
echo [INFO] Compiling Wails application...
wails build -platform windows/amd64 -outputname "%GUI_OUTPUT%"

if exist "%GUI_OUTPUT%" (
    for %%A in ("%GUI_OUTPUT%") do set "FILE_SIZE=%%~zA"
    set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
    echo [SUCCESS] GUI build succeeded: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [WARN] GUI build did not produce output file
    echo [INFO] Try running: wails build -dev for development build
)
echo.

:build_cli
echo [INFO] Building CLI version...

REM Clean old build artifacts
if exist "%CLI_OUTPUT%" (
    echo [INFO] Cleaning old CLI build...
    del /q "%CLI_OUTPUT%" 2>nul
)

REM Build CLI version
echo [INFO] Compiling CLI code...
go build -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/

if exist "%CLI_OUTPUT%" (
    for %%A in ("%CLI_OUTPUT%") do set "FILE_SIZE=%%~zA"
    set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
    echo [SUCCESS] CLI build succeeded: %CLI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] CLI build failed
    exit /b 1
)
echo.
goto :show_result

:show_result
echo ==============================================
echo   Build Complete!
echo ==============================================
echo.
echo Build outputs:
if exist "%GUI_OUTPUT%" (
    echo   GUI: %GUI_OUTPUT%
) else (
    echo   GUI: (not built, Wails required)
)
if exist "%CLI_OUTPUT%" (
    echo   CLI: %CLI_OUTPUT%
) else (
    echo   CLI: (build failed)
)
echo.
echo For GUI development: wails dev
echo For GUI build: wails build
echo.

:clean
echo [INFO] Cleaning build artifacts...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
del /q "%PROJECT_DIR%app\dist" 2>nul
del /q "%PROJECT_DIR%app\node_modules" 2>nul
del /q "%PROJECT_DIR%app\.wails" 2>nul
echo [SUCCESS] Cleanup complete
goto :eof

:check
echo [INFO] Running code checks...
go vet ./...
echo [SUCCESS] Code check complete
goto :eof

:help
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   (none)   Build all versions (GUI + CLI)
echo   gui      Build GUI version only
echo   cli      Build CLI version only
echo   clean    Clean build artifacts
echo   check     Run code checks
echo   help      Show this help message
echo.
echo Examples:
echo   build.bat          Build all versions
echo   build.bat gui      Build GUI version only
echo   build.bat cli      Build CLI version only
echo   build.bat clean    Clean artifacts
echo.
goto :eof

:main
if "%~1"=="" goto :build_gui
if /i "%~1"=="gui" goto :build_gui
if /i "%~1"=="cli" goto :build_cli
if /i "%~1"=="all" goto :build_gui
if /i "%~1"=="clean" goto :clean
if /i "%~1"=="check" goto :check
if /i "%~1"=="help" goto :help
goto :help
