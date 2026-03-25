@echo off
REM ERT (Windows Emergency Response Tool) Build Script

setlocal enabledelayedexpansion

REM Project paths
set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

REM Build configuration
set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ert.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"

REM Get build time using PowerShell (wmic is deprecated on Windows 11)
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set "BUILD_TIME=%%a"

REM Set ldflags (no -trimpath as it's not supported on Windows linker)
set "LDFLAGS=-s -H windowsgui"
set "LDFLAGS=!LDFLAGS! -X main.Version=%VERSION%"
set "LDFLAGS=!LDFLAGS! -X main.BuildTime=%BUILD_TIME%"

REM Create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%"
)

echo.
echo ==============================================
echo   Windows Emergency Response Tool v%VERSION%
echo   Build Script
echo ==============================================
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

where node >nul 2>&1
if !ERRORLEVEL! equ 0 (
    for /f "tokens=*" %%i in ('node --version') do echo [INFO] Node.js version: %%i
) else (
    echo [WARN] Node.js is not installed, frontend build will be skipped
)

echo [SUCCESS] Environment check passed
echo.
goto :eof

:build_gui
echo [INFO] Building GUI version...

echo [INFO] Cleaning old build artifacts...
if exist "%GUI_OUTPUT%" del /q "%GUI_OUTPUT%"

echo [INFO] Compiling Go code...
go build -ldflags "%LDFLAGS%" -o "%GUI_OUTPUT%" ./cmd/gui/

if exist "%GUI_OUTPUT%" (
    for %%A in ("%GUI_OUTPUT%") do set "FILE_SIZE=%%~zA"
    set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
    echo [SUCCESS] GUI build succeeded: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] GUI build failed
    exit /b 1
)
echo.
goto :eof

:build_cli
echo [INFO] Building CLI version...

echo [INFO] Cleaning old build artifacts...
if exist "%CLI_OUTPUT%" del /q "%CLI_OUTPUT%"

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
goto :eof

:build_frontend
echo [INFO] Checking frontend build...

if exist "%PROJECT_DIR%app\package.json" (
    echo [INFO] Frontend directory exists, skipping npm install
) else (
    echo [WARN] Frontend directory not found, skipping frontend build
)
echo.
goto :eof

:clean
echo [INFO] Cleaning build artifacts...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
del /q "%PROJECT_DIR%ert.exe" 2>nul
del /q "%PROJECT_DIR%ert-cli.exe" 2>nul
echo [SUCCESS] Cleanup complete
goto :eof

:check
echo [INFO] Running code checks...
go vet ./...
echo [SUCCESS] Code check complete
goto :eof

:main
if "%~1"=="" goto :build_all
if /i "%~1"=="gui" goto :build_gui_only
if /i "%~1"=="cli" goto :build_cli_only
if /i "%~1"=="all" goto :build_all
if /i "%~1"=="clean" goto :clean
if /i "%~1"=="check" goto :check
if /i "%~1"=="help" goto :show_help

:show_help
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   gui     Build GUI version
echo   cli     Build CLI version
echo   all     Build all versions ^(default^)
echo   clean   Clean build artifacts
echo   check   Run code checks
echo   help    Show this help message
echo.
echo Examples:
echo   build.bat          Build all versions
echo   build.bat gui      Build GUI version only
echo   build.bat cli      Build CLI version only
echo   build.bat clean    Clean artifacts
echo.
goto :eof

:build_all
call :check_env
call :build_gui
call :build_cli
call :build_frontend
echo ==============================================
echo   Build Complete!
echo ==============================================
echo.
echo Build outputs:
echo   GUI: %GUI_OUTPUT%
echo   CLI: %CLI_OUTPUT%
echo.
goto :eof

:build_gui_only
call :check_env
call :build_gui
goto :eof

:build_cli_only
call :check_env
call :build_cli
goto :eof
