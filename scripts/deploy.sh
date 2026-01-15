#!/bin/bash
set -e

# ============ Tortoise AI Framework 部署脚本 ============

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local missing=()
    
    if ! command_exists docker; then
        missing+=("docker")
    fi
    
    if ! command_exists docker-compose; then
        missing+=("docker-compose")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "缺少必要的依赖: ${missing[*]}"
        echo "请安装以下软件后重试:"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 加载环境变量
load_env() {
    if [ -f .env ]; then
        log_info "加载环境变量..."
        export $(grep -v '^#' .env | xargs)
        log_success "环境变量已加载"
    else
        log_warning ".env 文件不存在，将使用默认值"
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要目录..."
    
    mkdir -p data
    mkdir -p docker
    mkdir -p logs
    
    log_success "目录创建完成"
}

# 构建 Docker 镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    docker-compose build
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    docker-compose up -d
    
    log_success "服务启动完成"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    
    docker-compose down
    
    log_success "服务已停止"
}

# 查看日志
view_logs() {
    docker-compose logs -f
}

# 查看状态
view_status() {
    docker-compose ps
}

# 清理
cleanup() {
    log_warning "清理 Docker 资源..."
    
    docker-compose down -v --rmi all
    
    log_success "清理完成"
}

# 重新部署
redeploy() {
    log_info "重新部署..."
    stop_services
    cleanup
    create_directories
    build_images
    start_services
}

# 备份数据
backup() {
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    
    log_info "创建备份: $backup_dir"
    
    mkdir -p "$backup_dir"
    
    # 备份数据
    if [ -d "data" ]; then
        cp -r data "$backup_dir/"
    fi
    
    # 备份环境变量
    if [ -f ".env" ]; then
        cp .env "$backup_dir/"
    fi
    
    log_success "备份完成: $backup_dir"
}

# 恢复数据
restore() {
    local backup_dir=$1
    
    if [ -z "$backup_dir" ]; then
        log_error "请指定备份目录"
        echo "用法: $0 restore <backup_dir>"
        exit 1
    fi
    
    if [ ! -d "$backup_dir" ]; then
        log_error "备份目录不存在: $backup_dir"
        exit 1
    fi
    
    log_info "恢复数据: $backup_dir"
    
    # 停止服务
    stop_services
    
    # 恢复数据
    if [ -d "$backup_dir/data" ]; then
        rm -rf data
        cp -r "$backup_dir/data" .
    fi
    
    if [ -f "$backup_dir/.env" ]; then
        cp "$backup_dir/.env" .
    fi
    
    # 重新启动
    start_services
    
    log_success "数据恢复完成"
}

# 初始化 (首次运行)
init() {
    log_info "初始化 Tortoise AI Framework..."
    
    check_dependencies
    create_directories
    
    # 创建 .env 文件
    if [ ! -f ".env" ]; then
        log_info "创建 .env 文件..."
        cp .env.example .env
        log_warning "请编辑 .env 文件填入你的 API Key"
    fi
    
    # 构建并启动
    build_images
    start_services
    
    log_success "初始化完成!"
    echo ""
    echo "访问以下地址:"
    echo "  - 前端面板: http://localhost:3000"
    echo "  - API 文档: http://localhost:18792/docs"
    echo "  - Prometheus: http://localhost:9090 (可选)"
    echo ""
}

# 显示帮助
show_help() {
    echo "Tortoise AI Framework 部署脚本"
    echo ""
    echo "用法: $0 <command>"
    echo ""
    echo "可用命令:"
    echo "  init       - 初始化并启动 (首次运行)"
    echo "  start      - 启动服务"
    echo "  stop       - 停止服务"
    echo "  restart    - 重启服务"
    echo "  logs       - 查看日志"
    echo "  status     - 查看状态"
    echo "  build      - 构建镜像"
    echo "  redeploy   - 重新部署"
    echo "  backup     - 备份数据"
    echo "  restore    - 恢复数据"
    echo "  cleanup    - 清理资源"
    echo "  help       - 显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 init        # 首次运行"
    echo "  $0 start       # 启动服务"
    echo "  $0 logs        # 查看日志"
    echo "  $0 backup      # 备份"
}

# 主程序
main() {
    local command=${1:-help}
    
    case $command in
        init)
            init
            ;;
        start)
            check_dependencies
            load_env
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            check_dependencies
            load_env
            stop_services
            start_services
            ;;
        logs)
            view_logs
            ;;
        status)
            view_status
            ;;
        build)
            check_dependencies
            load_env
            build_images
            ;;
        redeploy)
            check_dependencies
            load_env
            redeploy
            ;;
        backup)
            backup
            ;;
        restore)
            restore $2
            ;;
        cleanup)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
