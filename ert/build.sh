#!/bin/bash
#
# ERT (Windows Emergency Response Tool) Build Script
# 一键打包脚本
#
# 使用方法:
#   ./build.sh          # 默认构建 GUI 版本
#   ./build.sh cli      # 构建 CLI 版本
#   ./build.sh clean    # 清理构建产物
#   ./build.sh check    # 代码检查
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目路径
PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

# 构建配置
OUTPUT_DIR="$PROJECT_DIR/bin"
GUI_OUTPUT="$OUTPUT_DIR/ert.exe"
CLI_OUTPUT="$OUTPUT_DIR/ert-cli.exe"
VERSION="13.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 打印函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示横幅
show_banner() {
    echo ""
    echo "=============================================="
    echo "  Windows 应急响应工具 (ERT) v${VERSION}"
    echo "  Build Script"
    echo "=============================================="
    echo ""
}

# 检查环境
check_environment() {
    print_info "检查构建环境..."
    
    # 检查 Go 版本
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go 1.21+"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+(\.\d+)?' | head -1)
    print_success "Go 版本: $GO_VERSION"
    
    # 检查 Node.js (可选)
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        print_success "Node.js 版本: $NODE_VERSION"
    else
        print_warn "Node.js 未安装，前端构建将被跳过"
    fi
    
    # 检查 pnpm (可选)
    if command -v pnpm &> /dev/null; then
        PNPM_VERSION=$(pnpm --version)
        print_success "pnpm 版本: $PNPM_VERSION"
    else
        print_warn "pnpm 未安装"
    fi
}

# 代码检查
check_code() {
    print_info "运行代码检查..."
    
    # 语法检查
    print_info "检查 Go 语法..."
    go vet ./... 2>/dev/null || true
    
    # 格式检查
    print_info "检查代码格式..."
    gofmt -l . 2>/dev/null || true
    
    print_success "代码检查完成"
}

# 构建 GUI 版本
build_gui() {
    print_info "开始构建 GUI 版本..."
    
    # 设置编译参数
    LDFLAGS="-s -w -H=windowsgui -trimpath"
    LDFLAGS="$LDFLAGS -X main.Version=${VERSION}"
    LDFLAGS="$LDFLAGS -X main.BuildTime=${BUILD_TIME}"
    
    # 编译
    print_info "编译 Go 代码..."
    GOOS=windows GOARCH=amd64 go build \
        -ldflags="$LDFLAGS" \
        -o "$GUI_OUTPUT" \
        ./cmd/gui/
    
    if [ -f "$GUI_OUTPUT" ]; then
        FILE_SIZE=$(stat -c%s "$GUI_OUTPUT" 2>/dev/null || stat -f%z "$GUI_OUTPUT" 2>/dev/null)
        FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1024 / 1024" | bc 2>/dev/null || echo "$((FILE_SIZE / 1024 / 1024)) MB")
        print_success "GUI 版本构建成功: $GUI_OUTPUT (${FILE_SIZE_MB})"
    else
        print_error "GUI 版本构建失败"
        exit 1
    fi
}

# 构建 CLI 版本
build_cli() {
    print_info "开始构建 CLI 版本..."
    
    # 设置编译参数
    LDFLAGS="-s -w -trimpath"
    LDFLAGS="$LDFLAGS -X main.Version=${VERSION}"
    LDFLAGS="$LDFLAGS -X main.BuildTime=${BUILD_TIME}"
    
    # 编译
    print_info "编译 CLI 代码..."
    GOOS=windows GOARCH=amd64 go build \
        -ldflags="$LDFLAGS" \
        -o "$CLI_OUTPUT" \
        ./cmd/cli/
    
    if [ -f "$CLI_OUTPUT" ]; then
        FILE_SIZE=$(stat -c%s "$CLI_OUTPUT" 2>/dev/null || stat -f%z "$CLI_OUTPUT" 2>/dev/null)
        FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1024 / 1024" | bc 2>/dev/null || echo "$((FILE_SIZE / 1024 / 1024)) MB")
        print_success "CLI 版本构建成功: $CLI_OUTPUT (${FILE_SIZE_MB})"
    else
        print_error "CLI 版本构建失败"
        exit 1
    fi
}

# 清理构建产物
clean() {
    print_info "清理构建产物..."
    
    rm -rf "$OUTPUT_DIR"
    rm -f "$PROJECT_DIR/ert.exe"
    rm -f "$PROJECT_DIR/ert-cli.exe"
    
    print_success "清理完成"
}

# 显示帮助
show_help() {
    show_banner
    echo "用法: ./build.sh [命令]"
    echo ""
    echo "命令:"
    echo "  gui     构建 GUI 版本 (默认)"
    echo "  cli     构建 CLI 版本"
    echo "  all     构建所有版本"
    echo "  check   运行代码检查"
    echo "  clean   清理构建产物"
    echo "  help    显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./build.sh          # 构建 GUI 版本"
    echo "  ./build.sh all      # 构建所有版本"
    echo "  ./build.sh cli      # 构建 CLI 版本"
    echo "  ./build.sh clean    # 清理"
    echo ""
}

# 主函数
main() {
    show_banner
    
    case "${1:-gui}" in
        gui)
            check_environment
            build_gui
            ;;
        cli)
            check_environment
            build_cli
            ;;
        all)
            check_environment
            build_gui
            build_cli
            ;;
        check)
            check_environment
            check_code
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
    
    echo ""
    print_success "构建流程完成!"
    echo ""
}

# 运行
main "$@"
