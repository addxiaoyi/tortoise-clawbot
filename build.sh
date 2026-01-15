#!/bin/bash

# Tortoise 构建脚本
# 构建所有平台的二进制文件

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# ========== 构建 Go 服务器 ==========
build_server() {
    log_info "构建 Go 服务器..."
    
    cd "$SCRIPT_DIR/server"
    
    # Linux
    log_info "构建 Linux amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/tortoise-server-linux-amd64" .
    
    # Linux ARM64
    log_info "构建 Linux arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$BUILD_DIR/tortoise-server-linux-arm64" .
    
    # macOS
    log_info "构建 macOS amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/tortoise-server-darwin-amd64" .
    
    log_info "构建 macOS arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/tortoise-server-darwin-arm64" .
    
    # Windows
    log_info "构建 Windows amd64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/tortoise-server.exe" .
    
    # 重命名为通用名称
    cp "$BUILD_DIR/tortoise-server" "$BUILD_DIR/tortoise-server-linux-amd64" 2>/dev/null || true
    
    cd "$SCRIPT_DIR"
    log_info "Go 服务器构建完成!"
}

# ========== 构建 Flutter ==========
build_flutter() {
    if ! command -v flutter &> /dev/null; then
        log_warn "Flutter 未安装，跳过 Flutter 构建"
        return
    fi
    
    log_info "构建 Flutter 应用..."
    
    cd "$SCRIPT_DIR/flutter"
    
    # Web
    log_info "构建 Web..."
    flutter build web --release -o "$BUILD_DIR/web"
    
    # Linux
    log_info "构建 Linux..."
    flutter build linux --release -o "$BUILD_DIR/linux"
    
    # macOS
    log_info "构建 macOS..."
    flutter build macos --release -o "$BUILD_DIR/macos"
    
    # Windows
    log_info "构建 Windows..."
    flutter build windows --release -o "$BUILD_DIR/windows"
    
    cd "$SCRIPT_DIR"
    log_info "Flutter 构建完成!"
}

# ========== 创建发布包 ==========
create_packages() {
    log_info "创建发布包..."
    
    cd "$BUILD_DIR"
    
    # Linux 包
    if [ -f "tortoise-server-linux-amd64" ]; then
        tar -czvf tortoise-server-linux-amd64.tar.gz tortoise-server-linux-amd64
        log_info "创建: tortoise-server-linux-amd64.tar.gz"
    fi
    
    # Windows 包
    if [ -f "tortoise-server.exe" ]; then
        powershell -Command "Compress-Archive -Path tortoise-server.exe -DestinationPath tortoise-server-windows-amd64.zip -Force"
        log_info "创建: tortoise-server-windows-amd64.zip"
    fi
    
    # Web 包
    if [ -d "web" ]; then
        tar -czvf tortoise-web.tar.gz web
        log_info "创建: tortoise-web.tar.gz"
    fi
    
    cd "$SCRIPT_DIR"
}

# ========== 显示构建结果 ==========
show_results() {
    log_info "构建结果:"
    echo ""
    ls -lh "$BUILD_DIR"
    echo ""
}

# ========== 主逻辑 ==========
case "${1:-all}" in
    server)
        build_server
        ;;
    flutter|app)
        build_flutter
        ;;
    packages)
        create_packages
        ;;
    all)
        build_server
        build_flutter
        create_packages
        show_results
        ;;
    clean)
        log_info "清理构建目录..."
        rm -rf "$BUILD_DIR"
        mkdir -p "$BUILD_DIR"
        log_info "清理完成"
        ;;
    help|--help|-h)
        echo "Tortoise 构建脚本"
        echo ""
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  server   只构建 Go 服务器"
        echo "  flutter  只构建 Flutter 应用"
        echo "  packages 创建发布包"
        echo "  all      构建所有 (默认)"
        echo "  clean    清理构建目录"
        echo "  help     显示帮助"
        ;;
    *)
        log_error "未知选项: $1"
        echo "使用 $0 help 查看帮助"
        exit 1
        ;;
esac
