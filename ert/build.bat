@echo off
setlocal enabledelayedexpansion

REM ============================================================================
REM ERT (Windows Emergency Response Tool) 一键构建脚本
REM 
REM 功能：
REM   1. 国内镜像加速（npm、Go模块）
REM   2. 详细错误收集与日志记录
REM   3. 一键构建 GUI + CLI 版本
REM
REM 使用方法：
REM   build.bat          - 构建所有版本
REM   build.bat gui      - 仅构建 GUI 版本
REM   build.bat cli      - 仅构建 CLI 版本
REM   build.bat dev      - 开发模式启动
REM   build.bat clean    - 清理构建产物
REM   build.bat check    - 代码检查
REM   build.bat help     - 显示帮助
REM ============================================================================

set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

REM ============================================================================
REM 配置
REM ============================================================================
set "OUTPUT_DIR=%PROJECT_DIR%bin"
set "GUI_OUTPUT=%OUTPUT_DIR%\ERT.exe"
set "CLI_OUTPUT=%OUTPUT_DIR%\ert-cli.exe"
set "VERSION=13.0.0"
set "BUILD_LOG=%PROJECT_DIR%build.log"
set "ERROR_LOG=%PROJECT_DIR%build_errors.log"
set "TIME_STAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%"
set "TIME_STAMP=%TIME_STAMP: =0%"

REM ============================================================================
REM 初始化日志
REM ============================================================================
:init_log
echo ========================================= > "%BUILD_LOG%"
echo ERT Build Log - %date% %time% >> "%BUILD_LOG%"
echo ========================================= >> "%BUILD_LOG%"
echo. >> "%BUILD_LOG%"

REM ============================================================================
REM 输出函数
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
REM 主入口
REM ============================================================================
:main
call :init_log

echo.
echo ================================================
echo   Windows 应急响应工具 (ERT) v%VERSION%
echo   一键构建脚本
echo ================================================
echo.

call :log_step "开始构建流程"

REM 检查参数
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
REM 环境检查
REM ============================================================================
:check_env
call :log_info "检查构建环境..."

REM 检查 Go
where go >nul 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_error "Go 未安装，请先安装 Go 1.21+"
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set "GO_VERSION=%%i"
call :log_info "Go 版本: !GO_VERSION!"

REM 检查 Node.js
where node >nul 2>&1
if !ERRORLEVEL! neq 0 (
    call :log_warn "Node.js 未安装，部分功能可能受限"
) else (
    for /f "tokens=*" %%i in ('node -v') do set "NODE_VERSION=%%i"
    call :log_info "Node.js 版本: !NODE_VERSION!"
)

REM 检查 Wails
where wails >nul 2>&1
if !ERRORLEVEL! equ 0 (
    for /f "tokens=*" %%i in ('wails version') do set "WAILS_VERSION=%%i"
    call :log_info "Wails 版本: !WAILS_VERSION!"
    set "WAILS_INSTALLED=1"
) else (
    call :log_warn "Wails 未安装，将跳过 GUI 构建"
    call :log_info "安装命令: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    set "WAILS_INSTALLED=0"
)

REM 检查前端目录
if not exist "%PROJECT_DIR%app\package.json" (
    call :log_warn "app\package.json 不存在，前端构建可能失败"
)

exit /b 0

REM ============================================================================
REM 配置国内镜像
REM ============================================================================
:setup_mirrors
call :log_info "配置国内镜像加速..."

REM 配置 Go 镜像
call :log_info "配置 Go 模块镜像..."
go env -w GOPROXY=https://goproxy.cn,direct 2>>"%BUILD_LOG%"
go env -w GOPRIVATE= 2>>"%BUILD_LOG%"
go env -w GOSUMDB=off 2>>"%BUILD_LOG%"

REM 配置 npm 镜像
call :log_info "配置 npm 镜像..."
if exist "%PROJECT_DIR%app\package.json" (
    cd /d "%PROJECT_DIR%app"
    
    REM 尝试创建或修改 .npmrc
    (
        echo registry=https://registry.npmmirror.com
        echo sass_binary_site=https://npmmirror.com/mirrors/node-sass
        echo phantomjs_cdnurl=https://npmmirror.com/mirrors/phantomjs
        echo electron_mirror=https://npmmirror.com/mirrors/electron
        echo electron_builder_binaurant_mirror=https://npmmirror.com/mirrors/electron-builder-binpackages
    ) > .npmrc
    
    cd /d "%PROJECT_DIR%"
    call :log_info "npm 镜像配置完成"
)

exit /b 0

REM ============================================================================
REM 构建 GUI 版本
REM ============================================================================
:build_gui
call :log_step "开始构建 GUI 版本..."

REM 检查 Wails
if "!WAILS_INSTALLED!"=="0" (
    call :log_warn "跳过 GUI 构建（Wails 未安装）"
    exit /b 0
)

REM 创建输出目录
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%" 2>>"%BUILD_LOG%"
    call :log_info "创建输出目录: !OUTPUT_DIR!"
)

REM 清理旧构建
call :log_info "清理旧构建文件..."
if exist "%GUI_OUTPUT%" (
    del /q "%GUI_OUTPUT%" 2>>"%BUILD_LOG%"
)
if exist "%PROJECT_DIR%app\dist" (
    rmdir /s /q "%PROJECT_DIR%app\dist" 2>>"%BUILD_LOG%"
)

REM 安装前端依赖
call :log_info "安装前端依赖..."
cd /d "%PROJECT_DIR%app"
if exist "package.json" (
    call npm install 2>>"%BUILD_LOG%"
    if !ERRORLEVEL! neq 0 (
        call :log_error "npm install 失败，请查看 %ERROR_LOG%"
        call :collect_error "npm install failed" "GUI前端依赖安装失败"
        cd /d "%PROJECT_DIR%"
        exit /b 1
    )
    call :log_success "前端依赖安装成功"
)

REM 构建前端
call :log_info "构建前端..."
call npm run build 2>>"%BUILD_LOG%"
if !ERRORLEVEL! neq 0 (
    call :log_error "前端构建失败，请查看 %ERROR_LOG%"
    call :collect_error "npm run build failed" "前端构建失败"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

if not exist "dist" (
    call :log_error "前端 dist 目录未生成"
    call :collect_error "dist not found" "前端构建产物不存在"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)
call :log_success "前端构建成功"

cd /d "%PROJECT_DIR%"

REM 构建 Wails 应用
call :log_info "开始编译 Wails 应用..."

REM 获取构建时间
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set "BUILD_TIME=%%a"

REM 设置编译参数
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM 执行 Wails 构建
wails build -platform windows/amd64 -outputname "%GUI_OUTPUT%" 2>>"%BUILD_LOG%"
set "BUILD_RESULT=!ERRORLEVEL!"

if !BUILD_RESULT! neq 0 (
    call :log_error "Wails 构建失败（错误码: !BUILD_RESULT!）"
    call :collect_error "wails build failed with code !BUILD_RESULT!" "Wails GUI 构建失败"
    
    REM 尝试收集更多错误信息
    call :dump_wails_errors
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

if not exist "%GUI_OUTPUT%" (
    call :log_error "GUI 输出文件未生成"
    call :collect_error "GUI output not found" "GUI 构建产物不存在"
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

for %%A in ("%GUI_OUTPUT%") do set "FILE_SIZE=%%~zA"
set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
call :log_success "GUI 构建成功: %GUI_OUTPUT% (!FILE_SIZE_MB! MB)"

cd /d "%PROJECT_DIR%"
exit /b 0

REM ============================================================================
REM 收集 Wails 错误详情
REM ============================================================================
:dump_wails_errors
call :log_info "收集构建错误信息..."

REM 收集 wails 诊断信息
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Wails Diagnostic Info - %date% %time% >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"

wails doctor 2>&1 >> "%ERROR_LOG%" 2>&1

REM 收集 Go 环境信息
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Environment >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
go env >> "%ERROR_LOG%" 2>&1

REM 收集最近修改的文件列表
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Recent Modified Files >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
dir /t:w /o:d /a:-d /s /b "%PROJECT_DIR%\*.go" 2>nul | head /n 20 >> "%ERROR_LOG%"

exit /b 0

REM ============================================================================
REM 收集通用错误
REM ============================================================================
:collect_error
echo. >> "%ERROR_LOG%"
echo [%time%] %~1 >> "%ERROR_LOG%"
echo Details: %~2 >> "%ERROR_LOG%"
echo Stack: >> "%ERROR_LOG%"
dir /b "%PROJECT_DIR%internal\modules" 2>nul >> "%ERROR_LOG%"
exit /b 0

REM ============================================================================
REM 构建 CLI 版本
REM ============================================================================
:build_cli
call :log_step "开始构建 CLI 版本..."

REM 创建输出目录
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%" 2>>"%BUILD_LOG%"
)

REM 清理旧构建
if exist "%CLI_OUTPUT%" (
    del /q "%CLI_OUTPUT%" 2>>"%BUILD_LOG%"
)

REM 获取构建时间
for /f "tokens=2 delims==" %%a in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set "BUILD_TIME=%%a"

REM 设置编译参数
set "LDFLAGS=-s -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM 执行构建
call :log_info "编译 CLI 代码..."
go build -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/ 2>>"%BUILD_LOG%"
set "BUILD_RESULT=!ERRORLEVEL!"

if !BUILD_RESULT! neq 0 (
    call :log_error "CLI 构建失败（错误码: !BUILD_RESULT!）"
    call :collect_error "go build cli failed with code !BUILD_RESULT!" "CLI 构建失败"
    
    REM 收集 Go 错误详情
    call :dump_go_errors
    exit /b 1
)

if not exist "%CLI_OUTPUT%" (
    call :log_error "CLI 输出文件未生成"
    call :collect_error "CLI output not found" "CLI 构建产物不存在"
    exit /b 1
)

for %%A in ("%CLI_OUTPUT%") do set "FILE_SIZE=%%~zA"
set /a FILE_SIZE_MB=FILE_SIZE / 1024 / 1024
call :log_success "CLI 构建成功: %CLI_OUTPUT% (!FILE_SIZE_MB! MB)"

exit /b 0

REM ============================================================================
REM 收集 Go 编译错误详情
REM ============================================================================
:dump_go_errors
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Build Errors - %date% %time% >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"

go build -v -ldflags "%LDFLAGS%" -o "%CLI_OUTPUT%" ./cmd/cli/ 2>&1 | tail /n 50 >> "%ERROR_LOG%"

REM 收集 go vet 结果
echo. >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
echo Go Vet Results >> "%ERROR_LOG%"
echo ========================================= >> "%ERROR_LOG%"
go vet ./... 2>&1 >> "%ERROR_LOG%"

exit /b 0

REM ============================================================================
REM 开发模式
REM ============================================================================
:build_dev
call :log_step "启动开发模式..."
call :check_env

if "!WAILS_INSTALLED!"=="0" (
    call :log_error "Wails 未安装，无法启动开发模式"
    exit /b 1
)

call :setup_mirrors

cd /d "%PROJECT_DIR%"
call :log_info "启动 Wails 开发服务器..."
call :log_info "按 Ctrl+C 停止开发服务器"

wails dev
exit /b 0

REM ============================================================================
REM 构建所有版本
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
echo   构建完成
echo ==============================================
echo.
echo 构建结果:
if exist "%GUI_OUTPUT%" (
    echo   [SUCCESS] GUI: %GUI_OUTPUT%
) else if "!WAILS_INSTALLED!"=="1" (
    echo   [FAILED]  GUI: 构建失败，请查看错误日志
) else (
    echo   [SKIPPED] GUI: Wails 未安装
)
if exist "%CLI_OUTPUT%" (
    echo   [SUCCESS] CLI: %CLI_OUTPUT%
) else (
    echo   [FAILED]  CLI: 构建失败，请查看错误日志
)
echo.
echo 日志文件:
echo   构建日志: %BUILD_LOG%
echo   错误日志: %ERROR_LOG%
echo.

if !GUI_RESULT! neq 0 (
    if !CLI_RESULT! neq 0 (
        exit /b 1
    )
)

REM 清理空日志
if exist "%ERROR_LOG%" (
    for %%A in ("%ERROR_LOG%") do if %%~zA equ 0 del "%ERROR_LOG%"
)

exit /b 0

REM ============================================================================
REM 仅构建 GUI
REM ============================================================================
:build_gui_only
call :check_env
call :setup_mirrors
call :build_gui

echo.
echo ==============================================
echo   GUI 构建完成
echo ==============================================
if exist "%GUI_OUTPUT%" (
    echo   [SUCCESS] GUI: %GUI_OUTPUT%
    exit /b 0
) else (
    echo   [FAILED]  GUI: 构建失败
    echo   请查看: %ERROR_LOG%
    exit /b 1
)

REM ============================================================================
REM 仅构建 CLI
REM ============================================================================
:build_cli_only
call :check_env
call :setup_mirrors
call :build_cli

echo.
echo ==============================================
echo   CLI 构建完成
echo ==============================================
if exist "%CLI_OUTPUT%" (
    echo   [SUCCESS] CLI: %CLI_OUTPUT%
    exit /b 0
) else (
    echo   [FAILED]  CLI: 构建失败
    echo   请查看: %ERROR_LOG%
    exit /b 1
)

REM ============================================================================
REM 清理
REM ============================================================================
:clean
call :log_info "清理构建产物..."

if exist "%OUTPUT_DIR%" (
    rmdir /s /q "%OUTPUT_DIR%"
    call :log_info "已删除: !OUTPUT_DIR!"
)

if exist "%PROJECT_DIR%app\dist" (
    rmdir /s /q "%PROJECT_DIR%app\dist"
    call :log_info "已删除: app\dist"
)

if exist "%PROJECT_DIR%app\node_modules" (
    rmdir /s /q "%PROJECT_DIR%app\node_modules"
    call :log_info "已删除: app\node_modules"
)

if exist "%PROJECT_DIR%app\.wails" (
    rmdir /s /q "%PROJECT_DIR%app\.wails"
    call :log_info "已删除: app\.wails"
)

if exist "%PROJECT_DIR%app\.npmrc" (
    del /q "%PROJECT_DIR%app\.npmrc"
    call :log_info "已删除: app\.npmrc"
)

if exist "%BUILD_LOG%" del /q "%BUILD_LOG%"
if exist "%ERROR_LOG%" del /q "%ERROR_LOG%"

call :log_success "清理完成"
exit /b 0

REM ============================================================================
REM 代码检查
REM ============================================================================
:check
call :log_step "运行代码检查..."

call :log_info "运行 go vet..."
go vet ./... 2>>"%BUILD_LOG%"
if !ERRORLEVEL! neq 0 (
    call :log_warn "go vet 发现问题，请查看日志"
) else (
    call :log_success "go vet 检查通过"
)

call :log_info "检查 Go 语法..."
go build -v ./... 2>>"%BUILD_LOG%"
if !ERRORLEVEL! neq 0 (
    call :log_warn "Go 编译检查发现问题"
) else (
    call :log_success "Go 编译检查通过"
)

call :log_success "代码检查完成"
exit /b 0

REM ============================================================================
REM 帮助信息
REM ============================================================================
:show_help
echo.
echo ================================================
echo   Windows 应急响应工具 (ERT) 构建脚本
echo ================================================
echo.
echo 使用方法: build.bat [命令]
echo.
echo 命令:
echo   (无参数)  - 构建所有版本（GUI + CLI）
echo   gui       - 仅构建 GUI 版本
echo   cli       - 仅构建 CLI 版本
echo   dev       - 启动开发模式（需要 Wails）
echo   clean     - 清理所有构建产物
echo   check     - 运行代码检查
echo   help      - 显示帮助信息
echo.
echo 国内加速:
echo   - 自动配置 Go 模块镜像 (goproxy.cn)
echo   - 自动配置 npm 镜像 (npmmirror.com)
echo.
echo 日志文件:
echo   - 构建日志: %PROJECT_DIR%build.log
echo   - 错误日志: %PROJECT_DIR%build_errors.log
echo.
echo 示例:
echo   build.bat       - 一键构建
echo   build.bat gui   - 仅构建 GUI
echo   build.bat dev   - 开发模式
echo   build.bat clean - 清理
echo.
echo 首次构建前请确保:
echo   1. 安装 Go 1.21+
echo   2. 安装 Node.js 18+
echo   3. 安装 Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest
echo.
exit /b 0
