@echo off
REM ERT (Windows Emergency Response Tool) Build Script
REM Windows 一键打包脚本

setlocal enabledelayedexpansion

REM 颜色定义
set RED=[91m
set GREEN=[92m
set YELLOW=[93m
set BLUE=[94m
set NC=[0m

REM 项目路径
set PROJECT_DIR=%~dp0
cd /d "%PROJECT_DIR%"

REM 构建配置
set OUTPUT_DIR=%PROJECT_DIR%bin
set GUI_OUTPUT=%OUTPUT_DIR%\ert.exe
set CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe
set VERSION=13.0.0
REM 获取当前时间
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set BUILD_TIME=%dt:~0,4%-%dt:~4,2%-%dt:~6,2% %dt:~8,2%:%dt:~10,2%:%dt:~12,2%

set LDFLAGS=-s -w -H=windowsgui -trimpath
set LDFLAGS=%LDFLAGS% -X main.Version=%VERSION%
set LDFLAGS=%LDFLAGS% -X main.BuildTime=%BUILD_TIME%

REM 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo.
echo ==============================================
echo   Windows 应急响应工具 (ERT) v%VERSION%
echo   Build Script
echo ==============================================
echo.

:check_env
echo [INFO] 检查构建环境...

where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go 未安装，请先安装 Go 1.21+
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo [INFO] Go 版本: !GO_VERSION!

REM 检查 Node.js (可选)
where node >nul 2>&1
if %ERRORLEVEL% equ 0 (
    for /f "tokens=*" %%i in ('node --version') do echo [INFO] Node.js 版本: %%i
) else (
    echo [WARN] Node.js 未安装，前端构建将被跳过
)

echo [SUCCESS] 环境检查完成
echo.

:build_gui
echo [INFO] 开始构建 GUI 版本...

echo [INFO] 清理旧构建产物...
if exist "%GUI_OUTPUT%" del /q "%GUI_OUTPUT%"

echo [INFO] 编译 Go 代码...
cmd /c "set GOOS=windows && set GOARCH=amd64 && go build -ldflags=\"%LDFLAGS%\" -o \"%GUI_OUTPUT%\" ./cmd/gui/"

if exist "%GUI_OUTPUT%" (
    for %%A in ("%GUI_OUTPUT%") do set FILE_SIZE=%%~zA
    set /a FILE_SIZE_MB=%FILE_SIZE% / 1024 / 1024
    echo [SUCCESS] GUI 版本构建成功: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] GUI 版本构建失败
    exit /b 1
)
echo.

:build_cli
echo [INFO] 开始构建 CLI 版本...

echo [INFO] 清理旧构建产物...
if exist "%CLI_OUTPUT%" del /q "%CLI_OUTPUT%"

echo [INFO] 编译 CLI 代码...
cmd /c "set GOOS=windows && set GOARCH=amd64 && go build -ldflags=\"%LDFLAGS%\" -o \"%CLI_OUTPUT%\" ./cmd/cli/"

if exist "%CLI_OUTPUT%" (
    for %%A in ("%CLI_OUTPUT%") do set FILE_SIZE=%%~zA
    set /a FILE_SIZE_MB=%FILE_SIZE% / 1024 / 1024
    echo [SUCCESS] CLI 版本构建成功: %CLI_OUTPUT% (!FILE_SIZE_MB! MB)
) else (
    echo [ERROR] CLI 版本构建失败
    exit /b 1
)
echo.

:build_frontend
echo [INFO] 检查前端构建...

if exist "%PROJECT_DIR%app\package.json" (
    echo [INFO] 前端目录存在，跳过 npm 安装
    REM 如果需要构建前端，取消下面的注释
    REM cd /d "%PROJECT_DIR%app"
    REM call npm install
    REM call npm run build
    REM cd /d "%PROJECT_DIR%"
) else (
    echo [WARN] 前端目录不存在，跳过前端构建
)

echo.
echo ==============================================
echo   构建完成!
echo ==============================================
echo.
echo 构建产物:
echo   GUI: %GUI_OUTPUT%
echo   CLI: %CLI_OUTPUT%
echo.

goto :eof

:clean
echo [INFO] 清理构建产物...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
del /q "%PROJECT_DIR%ert.exe" 2>nul
del /q "%PROJECT_DIR%ert-cli.exe" 2>nul
echo [SUCCESS] 清理完成
goto :eof

:check
echo [INFO] 运行代码检查...
go vet ./...
echo [SUCCESS] 代码检查完成
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
echo 用法: build.bat [命令]
echo.
echo 命令:
echo   gui     构建 GUI 版本
echo   cli     构建 CLI 版本
echo   all     构建所有版本 ^(默认^)
echo   clean   清理构建产物
echo   check   运行代码检查
echo   help    显示此帮助信息
echo.
echo 示例:
echo   build.bat          构建所有版本
echo   build.bat gui     仅构建 GUI 版本
echo   build.bat cli     仅构建 CLI 版本
echo   build.bat clean   清理
echo.
goto :eof

:build_all
call :check_env
call :build_gui
call :build_cli
call :build_frontend
goto :eof

:build_gui_only
call :check_env
call :build_gui
goto :eof

:build_cli_only
call :check_env
call :build_cli
goto :eof
