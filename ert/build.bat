@echo off

REM ============================================================================
REM ERT Build Script
REM ============================================================================

set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ERT.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"

echo.
echo ================================================
echo   ERT Build Script v%VERSION%
echo ================================================
echo.

REM ============================================================================
REM Check Arguments
REM ============================================================================
if "%~1"=="" goto :build_all
if /i "%~1"=="gui" goto :build_gui
if /i "%~1"=="cli" goto :build_cli
if /i "%~1"=="dev" goto :build_dev
if /i "%~1"=="clean" goto :clean
if /i "%~1"=="check" goto :check
goto :show_help

REM ============================================================================
REM Build All
REM ============================================================================
:build_all
echo [STEP] Checking environment...

where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)
for /f "tokens=*" %%i in ('go version 2^<^&1') do echo [INFO] %%i

where wails >nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo [INFO] Wails is installed
    set "HAS_WAILS=1"
) else (
    echo [WARN] Wails not found, GUI build will be skipped
    set "HAS_WAILS=0"
)

echo [STEP] Setting up China mirrors...
go env -w GOPROXY=https://goproxy.cn,direct >nul 2>&1
go env -w GOSUMDB=off >nul 2>&1

if exist "%PROJECT_DIR%app\package.json" (
    cd /d "%PROJECT_DIR%app"
    (
        echo registry=https://registry.npmmirror.com
        echo sass_binary_site=https://npmmirror.com/mirrors/node-sass
    ) > .npmrc
    cd /d "%PROJECT_DIR%"
)

echo.
echo [STEP] Building GUI...
if "%HAS_WAILS%"=="1" goto :build_gui
echo [SKIP] GUI build skipped (Wails not installed)
echo.

:build_cli
echo [STEP] Building CLI...
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd\'"') do set "BUILD_DATE=%%a"

go build -ldflags "-s -H windowsgui -X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE%" -o "%CLI_OUTPUT%" ./cmd/cli/
if %ERRORLEVEL% equ 0 (
    echo [SUCCESS] CLI: %CLI_OUTPUT%
) else (
    echo [ERROR] CLI build failed
    exit /b 1
)

echo.
echo =============================================
echo   Build Complete
echo =============================================
echo.
if exist "%GUI_OUTPUT%" (
    echo [OK] GUI: %GUI_OUTPUT%
) else (
    echo [--] GUI: not built
)
if exist "%CLI_OUTPUT%" (
    echo [OK] CLI: %CLI_OUTPUT%
)
echo.
exit /b 0

REM ============================================================================
REM Build GUI Only
REM ============================================================================
:build_gui
where wails >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Wails is not installed
    echo [INFO] Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    exit /b 1
)

echo [INFO] Wails found, starting GUI build...
echo [INFO] Wails will automatically build the frontend

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo [STEP] Building Wails application...
wails build -platform windows/amd64 -outputname "%GUI_OUTPUT%"
if %ERRORLEVEL% equ 0 (
    if exist "%GUI_OUTPUT%" (
        echo [SUCCESS] GUI: %GUI_OUTPUT%
    ) else (
        echo [ERROR] GUI output not found
        exit /b 1
    )
) else (
    echo [ERROR] Wails build failed
    exit /b 1
)

exit /b 0

REM ============================================================================
REM Build CLI Only
REM ============================================================================
:build_cli
echo [STEP] Building CLI...
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format \'yyyy-MM-dd\'"') do set "BUILD_DATE=%%a"

go build -ldflags "-s -H windowsgui -X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE%" -o "%CLI_OUTPUT%" ./cmd/cli/
if %ERRORLEVEL% equ 0 (
    echo [SUCCESS] CLI: %CLI_OUTPUT%
) else (
    echo [ERROR] CLI build failed
    exit /b 1
)
exit /b 0

REM ============================================================================
REM Development Mode
REM ============================================================================
:build_dev
where wails >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Wails is not installed
    exit /b 1
)

cd /d "%PROJECT_DIR%"
echo [INFO] Starting Wails dev server...
echo [INFO] Press Ctrl+C to stop
wails dev
exit /b 0

REM ============================================================================
REM Clean
REM ============================================================================
:clean
echo [STEP] Cleaning...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
if exist "%PROJECT_DIR%app\dist" rmdir /s /q "%PROJECT_DIR%app\dist"
if exist "%PROJECT_DIR%app\node_modules" rmdir /s /q "%PROJECT_DIR%app\node_modules"
if exist "%PROJECT_DIR%app\.wails" rmdir /s /q "%PROJECT_DIR%app\.wails"
if exist "%PROJECT_DIR%app\.npmrc" del /q "%PROJECT_DIR%app\.npmrc"
echo [OK] Done
exit /b 0

REM ============================================================================
REM Code Check
REM ============================================================================
:check
echo [STEP] Running code checks...
go vet ./...
if %ERRORLEVEL% equ 0 (
    echo [OK] go vet passed
) else (
    echo [WARN] go vet found issues
)
go build -v ./...
if %ERRORLEVEL% equ 0 (
    echo [OK] Go build check passed
) else (
    echo [WARN] Go build check failed
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
echo   (none)  - Build all (GUI + CLI)
echo   gui     - Build GUI only
echo   cli     - Build CLI only
echo   dev     - Development mode
echo   clean   - Clean artifacts
echo   check   - Code checks
echo   help    - Show this help
echo.
echo Prerequisites:
echo   - Go 1.21+
echo   - Node.js 18+ (for GUI)
echo   - Wails v2 (for GUI)
echo.
echo Manual GUI Build:
echo   1. cd app
echo   2. npm install
echo   3. npm run build
echo   4. cd ..
echo   5. wails build
echo.
exit /b 0
