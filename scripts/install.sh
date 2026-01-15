#!/bin/bash
# Tortoise 安装脚本

set -e

echo "=========================================="
echo "  Tortoise AI Agent Framework Installer"
echo "=========================================="
echo ""

# 检查系统
check_system() {
    case "$(uname -s)" in
        Linux*)     SYSTEM="linux";;
        Darwin*)    SYSTEM="macos";;
        CYGWIN*)    SYSTEM="windows";;
        MINGW*)     SYSTEM="windows";;
        *)          SYSTEM="unknown";;
    esac
    echo "检测到系统: $SYSTEM"
}

# 检查依赖
check_deps() {
    echo ""
    echo "检查依赖..."

    # Rust
    if command -v cargo &> /dev/null; then
        RUST_VERSION=$(cargo --version | awk '{print $2}')
        echo "  ✓ Rust $RUST_VERSION"
    else
        echo "  ✗ Rust 未安装"
        echo "  请运行: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        exit 1
    fi

    # Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        echo "  ✓ Go $GO_VERSION"
    else
        echo "  ✗ Go 未安装 (可选，用于插件开发)"
    fi

    # Flutter
    if command -v flutter &> /dev/null; then
        FLUTTER_VERSION=$(flutter --version | head -1 | awk '{print $2}')
        echo "  ✓ Flutter $FLUTTER_VERSION"
    else
        echo "  ✗ Flutter 未安装 (可选，用于 UI 开发)"
    fi
}

# 安装 Rust 依赖
install_rust_deps() {
    echo ""
    echo "安装 Rust 依赖..."
    cd core
    cargo fetch
    cargo build --release
    cd ..
    echo "  ✓ Rust 核心构建完成"
}

# 创建配置
create_config() {
    echo ""
    echo "创建配置目录..."

    CONFIG_DIR="$HOME/.tortoise"
    mkdir -p "$CONFIG_DIR"

    if [ ! -f "$CONFIG_DIR/config.toml" ]; then
        cp config/config.toml.example "$CONFIG_DIR/config.toml"
        echo "  ✓ 配置文件已创建: $CONFIG_DIR/config.toml"
    else
        echo "  - 配置文件已存在"
    fi

    # 创建数据目录
    mkdir -p "$CONFIG_DIR/memory"
    mkdir -p "$CONFIG_DIR/sessions"
    mkdir -p "$CONFIG_DIR/plugins"
    mkdir -p "$CONFIG_DIR/logs"
    echo "  ✓ 数据目录已创建"
}

# 安装二进制
install_binary() {
    echo ""
    echo "安装二进制文件..."

    BIN_DIR="$HOME/.local/bin"
    mkdir -p "$BIN_DIR"

    if [ -f "target/release/tortoise" ]; then
        cp target/release/tortoise "$BIN_DIR/"
        echo "  ✓ Tortoise 已安装到 $BIN_DIR"
    else
        echo "  ! 需要先构建: make build"
    fi
}

# 主函数
main() {
    check_system
    check_deps

    read -p "是否安装 Rust 核心? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        install_rust_deps
    fi

    create_config

    read -p "是否安装二进制到 ~/.local/bin? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        install_binary
    fi

    echo ""
    echo "=========================================="
    echo "  安装完成!"
    echo "=========================================="
    echo ""
    echo "快速开始:"
    echo "  tortoise start    # 启动服务"
    echo "  tortoise chat     # 开始聊天"
    echo "  tortoise status    # 查看状态"
    echo ""
}

main "$@"
