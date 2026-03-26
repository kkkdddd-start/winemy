@echo off
setlocal enabledelayedexpansion

REM ============================================================================
REM ERT (Windows Emergency Response Tool) Build Script
REM ============================================================================

set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ERT.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"
set "BUILD_LOG=%PROJECT_DIR%build.log"
set "ERROR_LOG=%PROJECT_DIR%build_errors.log"

REM ============================================================================
REM Main Entry
REM ============================================================================
call :init_log

echo.
echo ================================================
echo   Windows Emergency Response Tool (ERT) v%VERSION%
echo   Build Script
echo ================================================
echo.

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
REM Initialize Log
REM ============================================================================
:init_log
echo ========================================= > "%BUILD_LOG%"
echo ERT Build Log - %date% %time% >> "%BUILD_LOG%"
echo ========================================= >> "%BUILD_LOG%"
exit /b 0

REM ============================================================================
REM Check Environment
REM ============================================================================
:check_env
echo [INFO] Checking build environment... >> "%BUILD_LOG%"
echo [INFO] Checking build environment...

where go >nul 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] Go is not installed >> "%BUILD_LOG%"
    echo [ERROR] Go is not installed
    exit /b 1
)
for /f "tokens=*" %%i in ('go version 2^<^&1') do set "GO_VERSION=%%i"
echo [INFO] Go version: !GO_VERSION! >> "%BUILD_LOG%"
echo [INFO] Go version: !GO_VERSION!

where wails >nul 2>&1
if !ERRORLEVEL! equ 0 (
    echo [INFO] Wails is installed >> "%BUILD_LOG%"
    echo [INFO] Wails is installed
    set "WAILS_INSTALLED=1"
) else (
    echo [WARN] Wails is not installed, GUI build will be skipped >> "%BUILD_LOG%"
    echo [WARN] Wails is not installed, GUI build will be skipped
    set "WAILS_INSTALLED=0"
)

exit /b 0

REM ============================================================================
REM Setup China Mirrors
REM ============================================================================
:setup_mirrors
echo [INFO] Setting up China mirrors... >> "%BUILD_LOG%"
echo [INFO] Setting up China mirrors...

go env -w GOPROXY=https://goproxy.cn,direct >> "%BUILD_LOG%" 2>&1
go env -w GOSUMDB=off >> "%BUILD_LOG%" 2>&1
echo [INFO] Go mirror configured >> "%BUILD_LOG%"

if exist "%PROJECT_DIR%app\package.json" (
    cd /d "%PROJECT_DIR%app"
    (
        echo registry=https://registry.npmmirror.com
        echo sass_binary_site=https://npmmirror.com/mirrors/node-sass
        echo electron_mirror=https://npmmirror.com/mirrors/electron
    ) > .npmrc
    cd /d "%PROJECT_DIR%"
    echo [INFO] npm mirror configured >> "%BUILD_LOG%"
)

exit /b 0

REM ============================================================================
REM Build GUI
REM ============================================================================
:build_gui
echo [STEP] Building GUI version... >> "%BUILD_LOG%"
echo [STEP] Building GUI version...

if "!WAILS_INSTALLED!"=="0" (
    echo [SKIP] GUI build skipped (Wails not installed) >> "%BUILD_LOG%"
    echo [SKIP] GUI build skipped (Wails not installed)
    exit /b 0
)

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

cd /d "%PROJECT_DIR%app"
echo [INFO] Installing frontend dependencies... >> "%BUILD_LOG%"
echo [INFO] Installing frontend dependencies...

call npm install >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] npm install failed >> "%BUILD_LOG%"
    echo [ERROR] npm install failed, check logs
    exit /b 1
)
echo [INFO] Frontend dependencies installed >> "%BUILD_LOG%"

call npm run build >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] Frontend build failed >> "%BUILD_LOG%"
    echo [ERROR] Frontend build failed, check logs
    exit /b 1
)
echo [INFO] Frontend built >> "%BUILD_LOG%"

cd /d "%PROJECT_DIR%"

echo [INFO] Building Wails application... >> "%BUILD_LOG%"
echo [INFO] Building Wails application...

for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd HH:mm:ss\'"') do set "BUILD_TIME=%%a"
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

wails build -platform windows/amd64 -outputname "%GUI_OUTPUT%" >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] Wails build failed >> "%BUILD_LOG%"
    echo [ERROR] Wails build failed
    exit /b 1
)

if exist "%GUI_OUTPUT%" (
    for %%A in ("%GUI_OUTPUT%") do set "FILE_SIZE=%%~zA"
    set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
    echo [SUCCESS] GUI: %GUI_OUTPUT% (!FILE_SIZE_MB! MB) >> "%BUILD_LOG%"
    echo [SUCCESS] GUI: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] GUI output not found >> "%BUILD_LOG%"
    echo [ERROR] GUI output not found
    exit /b 1
)

exit /b 0

REM ============================================================================
REM Build CLI
REM ============================================================================
:build_cli
echo [STEP] Building CLI version... >> "%BUILD_LOG%"
echo [STEP] Building CLI version...

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd HH:mm:ss\'"') do set "BUILD_TIME=%%a"
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

echo [INFO] Compiling CLI... >> "%BUILD_LOG%"
echo [INFO] Compiling CLI...

go build -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/ >> "%BUILD_LOG%" 2>&1
if !ERRORLEVEL! neq 0 (
    echo [ERROR] CLI build failed >> "%BUILD_LOG%"
    echo [ERROR] CLI build failed
    exit /b 1
)

if exist "%CLI_OUTPUT%" (
    for %%A in ("%CLI_OUTPUT%") do set "FILE_SIZE=%%~zA"
    set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
    echo [SUCCESS] CLI: %CLI_OUTPUT% (!FILE_SIZE_MB! MB) >> "%BUILD_LOG%"
    echo [SUCCESS] CLI: %CLI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] CLI output not found >> "%BUILD_LOG%"
    echo [ERROR] CLI output not found
    exit /b 1
)

exit /b 0

REM ============================================================================
REM Build All
REM ============================================================================
:build_all
call :check_env
if !ERRORLEVEL! neq 0 (
    echo [ERROR] Environment check failed >> "%BUILD_LOG%"
    exit /b 1
)

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
echo [RESULT] GUI: 
if exist "%GUI_OUTPUT%" (
    echo   SUCCESS: %GUI_OUTPUT%
) else (
    echo   SKIPPED or FAILED
)
echo [RESULT] CLI:
if exist "%CLI_OUTPUT%" (
    echo   SUCCESS: %CLI_OUTPUT%
) else (
    echo   FAILED
)
echo.
echo [LOG] Check %BUILD_LOG% for details
echo.

exit /b 0

REM ============================================================================
REM Build GUI Only
REM ============================================================================
:build_gui_only
call :check_env
if !ERRORLEVEL! neq 0 exit /b 1
call :setup_mirrors
call :build_gui
exit /b !ERRORLEVEL!

REM ============================================================================
REM Build CLI Only
REM ============================================================================
:build_cli_only
call :check_env
if !ERRORLEVEL! neq 0 exit /b 1
call :setup_mirrors
call :build_cli
exit /b !ERRORLEVEL!

REM ============================================================================
REM Development Mode
REM ============================================================================
:build_dev
call :check_env
if !ERRORLEVEL! neq 0 exit /b 1
call :setup_mirrors
cd /d "%PROJECT_DIR%"
wails dev
exit /b 0

REM ============================================================================
REM Clean
REM ============================================================================
:clean
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
if exist "%PROJECT_DIR%app\dist" rmdir /s /q "%PROJECT_DIR%app\dist"
if exist "%PROJECT_DIR%app\node_modules" rmdir /s /q "%PROJECT_DIR%app\node_modules"
if exist "%PROJECT_DIR%app\.wails" rmdir /s /q "%PROJECT_DIR%app\.wails"
if exist "%PROJECT_DIR%app\.npmrc" del /q "%PROJECT_DIR%app\.npmrc"
if exist "%BUILD_LOG%" del /q "%BUILD_LOG%"
if exist "%ERROR_LOG%" del /q "%ERROR_LOG%"
echo [SUCCESS] Cleanup complete
exit /b 0

REM ============================================================================
REM Code Check
REM ============================================================================
:check
go vet ./... 2>&1
if !ERRORLEVEL! equ 0 (
    echo [SUCCESS] go vet passed
) else (
    echo [WARN] go vet found issues
)
exit /b 0

REM ============================================================================
REM Help
REM ============================================================================
:show_help
echo.
echo ================================================
echo   ERT Build Script
echo ================================================
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   (none)  - Build all versions
echo   gui     - Build GUI only
echo   cli     - Build CLI only
echo   dev     - Development mode
echo   clean   - Clean artifacts
echo   check   - Code checks
echo   help    - Show help
echo.
echo Prerequisites:
echo   - Go 1.21+
echo   - Node.js 18+ (for GUI)
echo   - Wails (for GUI): go install github.com/wailsapp/wails/v2/cmd/wails@latest
echo.
exit /b 0
