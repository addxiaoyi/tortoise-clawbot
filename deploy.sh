#!/bin/bash

# Tortoise 部署脚本 (Linux/macOS 无容器)
# 使用方法: ./deploy.sh [start|stop|restart|status]

set -e

APP_NAME="tortoise"
APP_DIR="/opt/tortoise"
DATA_DIR="/var/lib/tortoise"
LOG_DIR="/var/log/tortoise"
CONFIG_DIR="/etc/tortoise"

# 二进制文件名
BINARY_NAME="tortoise-server"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 权限运行此脚本"
        exit 1
    fi
}

# 创建用户
create_user() {
    if ! id "$APP_NAME" &>/dev/null; then
        log_info "创建用户 $APP_NAME..."
        useradd -r -s /bin/false -d "$APP_DIR" "$APP_NAME"
    fi
}

# 创建目录
create_dirs() {
    log_info "创建目录..."
    mkdir -p "$APP_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$CONFIG_DIR"
}

# 下载或构建二进制
setup_binary() {
    log_info "设置二进制文件..."
    
    if [ -f "./$BINARY_NAME" ]; then
        log_info "使用本地二进制文件..."
        cp "./$BINARY_NAME" "$APP_DIR/$BINARY_NAME"
    else
        log_warn "请先构建或下载 $BINARY_NAME 到当前目录"
        exit 1
    fi
    
    chmod +x "$APP_DIR/$BINARY_NAME"
}

# 配置环境变量
setup_env() {
    log_info "配置环境变量..."
    
    cat > "$CONFIG_DIR/env" << 'EOF'
# Tortoise 环境变量配置
# OpenAI
OPENAI_API_KEY=your-openai-api-key

# Anthropic
ANTHROPIC_API_KEY=your-anthropic-api-key

# Telegram
TELEGRAM_BOT_TOKEN=your-telegram-bot-token

# Discord
DISCORD_BOT_TOKEN=your-discord-bot-token

# 应用密钥
APP_SECRET_KEY=change-this-secret-key-in-production
EOF

    chmod 600 "$CONFIG_DIR/env"
}

# 创建 Systemd 服务
create_systemd_service() {
    log_info "创建 Systemd 服务..."
    
    cat > "/etc/systemd/system/$APP_NAME.service" << EOF
[Unit]
Description=Tortoise AI Agent Server
After=network.target

[Service]
Type=simple
User=$APP_NAME
Group=$APP_NAME
WorkingDirectory=$APP_DIR
EnvironmentFile=$CONFIG_DIR/env
ExecStart=$APP_DIR/$BINARY_NAME
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/stdout.log
StandardError=append:$LOG_DIR/stderr.log

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

[Install]
WantedBy=multi-user.target
EOF

    log_info "重新加载 systemd..."
    systemctl daemon-reload
}

# 创建启动脚本
create_startup_script() {
    log_info "创建启动脚本..."
    
    cat > "$APP_DIR/start.sh" << 'EOF'
#!/bin/bash
cd /opt/tortoise
export $(cat /etc/tortoise/env | xargs)
./tortoise-server
EOF

    chmod +x "$APP_DIR/start.sh"
}

# 安装
install() {
    check_root
    create_user
    create_dirs
    setup_binary
    setup_env
    create_systemd_service
    create_startup_script
    
    # 设置权限
    chown -R "$APP_NAME:$APP_NAME" "$APP_DIR" "$DATA_DIR" "$LOG_DIR"
    
    log_info "安装完成！"
    echo ""
    echo "请编辑配置文件: $CONFIG_DIR/env"
    echo ""
    echo "启动服务: systemctl start tortoise"
    echo "开机自启: systemctl enable tortoise"
    echo "查看日志: journalctl -u tortoise -f"
}

# 启动
start() {
    log_info "启动 Tortoise..."
    systemctl start "$APP_NAME"
    
    if systemctl is-active --quiet "$APP_NAME"; then
        log_info "Tortoise 已启动"
    else
        log_error "启动失败，请检查日志"
        exit 1
    fi
}

# 停止
stop() {
    log_info "停止 Tortoise..."
    systemctl stop "$APP_NAME"
    log_info "Tortoise 已停止"
}

# 重启
restart() {
    log_info "重启 Tortoise..."
    systemctl restart "$APP_NAME"
    log_info "Tortoise 已重启"
}

# 状态
status() {
    systemctl status "$APP_NAME" --no-pager
}

# 卸载
uninstall() {
    check_root
    log_warn "即将卸载 Tortoise..."
    systemctl stop "$APP_NAME" 2>/dev/null || true
    systemctl disable "$APP_NAME" 2>/dev/null || true
    rm -f "/etc/systemd/system/$APP_NAME.service"
    rm -rf "$APP_DIR" "$DATA_DIR" "$LOG_DIR" "$CONFIG_DIR"
    userdel "$APP_NAME" 2>/dev/null || true
    log_info "卸载完成"
}

# 显示帮助
show_help() {
    echo "Tortoise 部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  install    安装 Tortoise"
    echo "  start     启动服务"
    echo "  stop      停止服务"
    echo "  restart   重启服务"
    echo "  status    查看状态"
    echo "  uninstall 卸载"
    echo "  help      显示帮助"
    echo ""
}

# 主逻辑
case "${1:-help}" in
    install)
        install
        ;;
    start)
        check_root
        start
        ;;
    stop)
        check_root
        stop
        ;;
    restart)
        check_root
        restart
        ;;
    status)
        status
        ;;
    uninstall)
        uninstall
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "未知命令: $1"
        show_help
        exit 1
        ;;
esac
