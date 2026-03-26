@echo off
setlocal enabledelayedexpansion

REM ============================================================================
REM ERT (Windows Emergency Response Tool) Build Script
REM 
REM Features:
REM   1. China mirror acceleration (npm, Go modules)
REM   2. Detailed error collection and logging
REM   3. One-click build GUI + CLI versions
REM
REM Usage:
REM   build.bat          - Build all versions
REM   build.bat gui      - Build GUI only
REM   build.bat cli      - Build CLI only
REM   build.bat dev      - Development mode
REM   build.bat clean    - Clean build artifacts
REM   build.bat check    - Code check
REM   build.bat help     - Show help
REM ============================================================================

set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

REM ============================================================================
REM Configuration
REM ============================================================================
set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ERT.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"
set "BUILD_LOG=%PROJECT_DIR%build.log"
set "ERROR_LOG=%PROJECT_DIR%build_errors.log"

REM ============================================================================
REM Initialize Log
REM ============================================================================
:init_log
echo ========================================= > "%BUILD_LOG%"
echo ERT Build Log - %date% %time% >> "%BUILD_LOG%"
echo ========================================= >> "%BUILD_LOG%"
echo. >> "%BUILD_LOG%"
exit /b 0

REM ============================================================================
REM Logging Functions
REM ============================================================================
:log_info
echo [INFO] %~1
echo [INFO] %~1 >> "%BUILD_LOG%"
exit /b 0

:log_success
echo [SUCCESS] %~1
echo [SUCCESS] %~1 >> "%BUILD_LOG%"
exit /b 0

:log_warn
echo [WARN] %~1
echo [WARN] %~1 >> "%BUILD_LOG%"
exit /b 0

:log_error
echo [ERROR] %~1
echo [ERROR] %~1 >> "%BUILD_LOG%"
echo [ERROR] %~1 >> "%ERROR_LOG%"
exit /b 1

:log_step
echo.
echo [STEP] %~1
echo. >> "%BUILD_LOG%"
echo [STEP] %~1 >> "%BUILD_LOG%"
exit /b 0

REM ============================================================================
REM Main Entry
REM ============================================================================
:main
call :init_log

echo.
echo ================================================
echo   Windows Emergency Response Tool (ERT) v%VERSION%
echo   Build Script
echo ================================================
echo.

call :log_step "Starting build process"

REM Check arguments
if "%~1"=="" goto :build_all
if /i "%~1"=="gui" goto :build_gui_only
if /i "%~1"=="cli" goto :build_cli_only
if /i "%~1"=="all" goto :build_all
if /i "%~1"=="dev" goto :build_dev
if /i "%~1"=="clean" goto :clean
if /i "%~1"=="check" goto :check
if /i "%~1"=="help" goto :show_help
goto :show_help

REM ============================================================================
REM Environment Check
REM ============================================================================
:check_env
call :log_info "Checking build environment..."

REM Check Go
where go >nul 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_error "Go is not installed. Please install Go 1.21+"
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set "GO_VERSION=%%i"
call :log_info "Go version: !GO_VERSION!"

REM Check Node.js
where node >nul 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_warn "Node.js is not installed, some features may be limited"
) else (
    for /f "tokens=*" %%i in ('node -v') do set "NODE_VERSION=%%i"
    call :log_info "Node.js version: !NODE_VERSION!"
)

REM Check Wails
where wails >nul 2>&1
if !ERRORLEVEL! equ 0 (
    for /f "tokens=*" %%i in ('wails version 2^|findstr /i version') do set "WAILS_VERSION=%%i"
    call :log_info "Wails is installed"
    set "WAILS_INSTALLED=1"
) else (
    call :log_warn "Wails is not installed, GUI build will be skipped"
    call :log_info "Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    set "WAILS_INSTALLED=0"
)

REM Check frontend directory
if not exist "%PROJECT_DIR%app\package.json" (
    call :log_warn "app\package.json not found, frontend build may fail"
)

exit /b 0

REM ============================================================================
REM Setup China Mirrors
REM ============================================================================
:setup_mirrors
call :log_info "Setting up China mirrors..."

REM Setup Go mirror
call :log_info "Configuring Go module mirror..."
go env -w GOPROXY=https://goproxy.cn,direct >> "%BUILD_LOG%" 2>&1
go env -w GOSUMDB=off >> "%BUILD_LOG%" 2>&1

REM Setup npm mirror
call :log_info "Configuring npm mirror..."
if exist "%PROJECT_DIR%app\package.json" (
    cd /d "%PROJECT_DIR%app"
    
    (
        echo registry=https://registry.npmmirror.com
        echo sass_binary_site=https://npmmirror.com/mirrors/node-sass
        echo phantomjs_cdnurl=https://npmmirror.com/mirrors/phantomjs
        echo electron_mirror=https://npmmirror.com/mirrors/electron
        echo electron_builder_binary_mirror=https://npmmirror.com/mirrors/electron-builder-binpackages
    ) > .npmrc
    
    cd /d "%PROJECT_DIR%"
    call :log_info "npm mirror configured"
)

exit /b 0

REM ============================================================================
REM Build GUI Version
REM ============================================================================
:build_gui
call :log_step "Building GUI version..."

REM Check Wails
if "!WAILS_INSTALLED!"=="0" (
    call :log_warn "Skipping GUI build (Wails not installed)"
    exit /b 0
)

REM Create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%" >> "%BUILD_LOG%" 2>&1
    call :log_info "Created output directory: !OUTPUT_DIR!"
)

REM Clean old build
call :log_info "Cleaning old build files..."
if exist "%GUI_OUTPUT%" (
    del /q "%GUI_OUTPUT%" >> "%BUILD_LOG%" 2>&1
)
if exist "%PROJECT_DIR%app\dist" (
    rmdir /s /q "%PROJECT_DIR%app\dist" >> "%BUILD_LOG%" 2>&1
)

REM Install frontend dependencies
call :log_info "Installing frontend dependencies..."
cd /d "%PROJECT_DIR%app"
if exist "package.json" (
    call npm install >> "%BUILD_LOG%" 2>&1
    if !ERRORLEVEL! neq 0 (
        call :log_error "npm install failed. Check %ERROR_LOG% for details"
        call :collect_error "npm install failed" "Frontend dependency installation failed"
        cd /d "%PROJECT_DIR%"
        exit /b 1
    )
    call :log_success "Frontend dependencies installed"
)

REM Build frontend
call :log_info "Building frontend..."
call npm run build >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_error "Frontend build failed. Check %ERROR_LOG% for details"
    call :collect_error "npm run build failed" "Frontend build failed"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

if not exist "dist" (
    call :log_error "Frontend dist directory not generated"
    call :collect_error "dist not found" "Frontend build output not found"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)
call :log_success "Frontend build succeeded"

cd /d "%PROJECT_DIR%"

REM Build Wails application
call :log_info "Compiling Wails application..."

REM Get build time
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd HH:mm:ss\'"') do set "BUILD_TIME=%%a"

REM Set build flags
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM Execute Wails build
wails build -platform windows/amd64 -outputname "%GUI_OUTPUT%" >> "%BUILD_LOG%" 2>&1
set "BUILD_RESULT=!ERRORLEVEL!"

if !BUILD_RESULT! neq 0 (
    call :log_error "Wails build failed (error code: !BUILD_RESULT!)"
    call :collect_error "wails build failed with code !BUILD_RESULT!" "Wails GUI build failed"
    call :dump_wails_errors
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

if not exist "%GUI_OUTPUT%" (
    call :log_error "GUI output file not generated"
    call :collect_error "GUI output not found" "GUI build artifact not found"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

for %%A in ("%GUI_OUTPUT%") do set "FILE_SIZE=%%~zA"
set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
call :log_success "GUI build succeeded: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)"

cd /d "%PROJECT_DIR%"
exit /b 0

REM ============================================================================
REM Collect Wails Error Details
REM ============================================================================
:dump_wails_errors
call :log_info "Collecting build error information..."

echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Wails Diagnostic Info - %date% %time% >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"

wails doctor >> "%ERROR_LOG%" 2>&1

echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Environment >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
go env >> "%ERROR_LOG%" 2>&1

exit /b 0

REM ============================================================================
REM Collect Generic Errors
REM ============================================================================
:collect_error
echo. >> "%ERROR_LOG%"
echo [%time%] %~1 >> "%ERROR_LOG%"
echo Details: %~2 >> "%ERROR_LOG%"
echo. >> "%ERROR_LOG%"
exit /b 0

REM ============================================================================
REM Build CLI Version
REM ============================================================================
:build_cli
call :log_step "Building CLI version..."

REM Create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%" >> "%BUILD_LOG%" 2>&1
)

REM Clean old build
if exist "%CLI_OUTPUT%" (
    del /q "%CLI_OUTPUT%" >> "%BUILD_LOG%" 2>&1
)

REM Get build time
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd HH:mm:ss\'"') do set "BUILD_TIME=%%a"

REM Set build flags
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM Execute build
call :log_info "Compiling CLI code..."
go build -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/ >> "%BUILD_LOG%" 2>&1
set "BUILD_RESULT=!ERRORLEVEL!"

if !BUILD_RESULT! neq 0 (
    call :log_error "CLI build failed (error code: !BUILD_RESULT!)"
    call :collect_error "go build cli failed with code !BUILD_RESULT!" "CLI build failed"
    call :dump_go_errors
    exit /b 1
)

if not exist "%CLI_OUTPUT%" (
    call :log_error "CLI output file not generated"
    call :collect_error "CLI output not found" "CLI build artifact not found"
    exit /b 1
)

for %%A in ("%CLI_OUTPUT%") do set "FILE_SIZE=%%~zA"
set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
call :log_success "CLI build succeeded: %CLI_OUTPUT% (!FILE_SIZE_MB! MB)"

exit /b 0

REM ============================================================================
REM Collect Go Build Error Details
REM ============================================================================
:dump_go_errors
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Build Errors - %date% %time% >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"

go build -v -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/ >> "%ERROR_LOG%" 2>&1

echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Vet Results >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
go vet ./... >> "%ERROR_LOG%" 2>&1

exit /b 0

REM ============================================================================
REM Development Mode
REM ============================================================================
:build_dev
call :log_step "Starting development mode..."
call :check_env

if "!WAILS_INSTALLED!"=="0" (
    call :log_error "Wails is not installed, cannot start development mode"
    exit /b 1
)

call :setup_mirrors

cd /d "%PROJECT_DIR%"
call :log_info "Starting Wails development server..."
call :log_info "Press Ctrl+C to stop development server"

wails dev
exit /b 0

REM ============================================================================
REM Build All Versions
REM ============================================================================
:build_all
call :check_env
call :setup_mirrors
call :build_gui
set "GUI_RESULT=!ERRORLEVEL!"
call :build_cli
set "CLI_RESULT=!ERRORLEVEL!"

echo.
echo ==============================================
echo   Build Complete
echo ==============================================
echo.
echo Build outputs:
if exist "%GUI_OUTPUT%" (
    echo   [SUCCESS] GUI: %GUI_OUTPUT%
) else if "!WAILS_INSTALLED!"=="1" (
    echo   [FAILED]  GUI: Build failed, check error log
) else (
    echo   [SKIPPED] GUI: Wails not installed
)
if exist "%CLI_OUTPUT%" (
    echo   [SUCCESS] CLI: %CLI_OUTPUT%
) else (
    echo   [FAILED]  CLI: Build failed, check error log
)
echo.
echo Log files:
echo   Build log: %BUILD_LOG%
echo   Error log: %ERROR_LOG%
echo.

if !GUI_RESULT! neq 0 (
    if !CLI_RESULT! neq 0 (
        exit /b 1
    )
)

REM Clean empty logs
if exist "%ERROR_LOG%" (
    for %%A in ("%ERROR_LOG%") do if %%~zA equ 0 del "%ERROR_LOG%"
)

exit /b 0

REM ============================================================================
REM Build GUI Only
REM ============================================================================
:build_gui_only
call :check_env
call :setup_mirrors
call :build_gui

echo.
echo ==============================================
echo   GUI Build Complete
echo ==============================================
if exist "%GUI_OUTPUT%" (
    echo   [SUCCESS] GUI: %GUI_OUTPUT%
    exit /b 0
) else (
    echo   [FAILED]  GUI: Build failed
    echo   Check: %ERROR_LOG%
    exit /b 1
)

REM ============================================================================
REM Build CLI Only
REM ============================================================================
:build_cli_only
call :check_env
call :setup_mirrors
call :build_cli

echo.
echo ==============================================
echo   CLI Build Complete
echo ==============================================
if exist "%CLI_OUTPUT%" (
    echo   [SUCCESS] CLI: %CLI_OUTPUT%
    exit /b 0
) else (
    echo   [FAILED]  CLI: Build failed
    echo   Check: %ERROR_LOG%
    exit /b 1
)

REM ============================================================================
REM Clean
REM ============================================================================
:clean
call :log_info "Cleaning build artifacts..."

if exist "%OUTPUT_DIR%" (
    rmdir /s /q "%OUTPUT_DIR%"
    call :log_info "Deleted: !OUTPUT_DIR!"
)

if exist "%PROJECT_DIR%app\dist" (
    rmdir /s /q "%PROJECT_DIR%app\dist"
    call :log_info "Deleted: app\dist"
)

if exist "%PROJECT_DIR%app\node_modules" (
    rmdir /s /q "%PROJECT_DIR%app\node_modules"
    call :log_info "Deleted: app\node_modules"
)

if exist "%PROJECT_DIR%app\.wails" (
    rmdir /s /q "%PROJECT_DIR%app\.wails"
    call :log_info "Deleted: app\.wails"
)

if exist "%PROJECT_DIR%app\.npmrc" (
    del /q "%PROJECT_DIR%app\.npmrc"
    call :log_info "Deleted: app\.npmrc"
)

if exist "%BUILD_LOG%" del /q "%BUILD_LOG%"
if exist "%ERROR_LOG%" del /q "%ERROR_LOG%"

call :log_success "Cleanup complete"
exit /b 0

REM ============================================================================
REM Code Check
REM ============================================================================
:check
call :log_step "Running code checks..."

call :log_info "Running go vet..."
go vet ./... >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_warn "go vet found issues, check log"
) else (
    call :log_success "go vet passed"
)

call :log_info "Checking Go syntax..."
go build -v ./... >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_warn "Go build check found issues"
) else (
    call :log_success "Go build check passed"
)

call :log_success "Code check complete"
exit /b 0

REM ============================================================================
REM Help
REM ============================================================================
:show_help
echo.
echo ================================================
echo   Windows Emergency Response Tool (ERT) Build
echo ================================================
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   (none)  - Build all versions (GUI + CLI)
echo   gui     - Build GUI version only
echo   cli     - Build CLI version only
echo   dev     - Start development mode (requires Wails)
echo   clean   - Clean all build artifacts
echo   check   - Run code checks
echo   help    - Show this help message
echo.
echo China Acceleration:
echo   - Go modules: goproxy.cn
echo   - npm: npmmirror.com
echo.
echo Log Files:
echo   - Build log: %PROJECT_DIR%build.log
echo   - Error log: %PROJECT_DIR%build_errors.log
echo.
echo Examples:
echo   build.bat      - One-click build
echo   build.bat gui  - Build GUI only
echo   build.bat dev  - Development mode
echo   build.bat clean - Clean artifacts
echo.
echo Prerequisites:
echo   1. Install Go 1.21+
echo   2. Install Node.js 18+
echo   3. Install Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest
echo.
exit /b 0
