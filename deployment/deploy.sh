#!/bin/bash
# ==========================================
# SmartGrow 服务器端部署脚本
# 用途：在服务器上从GitHub拉取最新代码并部署
# ==========================================

set -e  # 遇到错误立即退出

# ========== 配置区域 ==========
PROJECT_DIR="/root/smart-grow"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"
LOG_FILE="$PROJECT_DIR/logs/deploy.log"
BACKUP_DIR="$PROJECT_DIR/backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ========== 日志函数 ==========
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1" | tee -a "$LOG_FILE"
}

# ========== 显示帮助 ==========
show_help() {
    echo ""
    echo "=========================================="
    echo "SmartGrow 服务器端部署工具"
    echo "=========================================="
    echo ""
    echo "用法:"
    echo "  ./deploy.sh [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  pull        - 从GitHub拉取最新代码"
    echo "  build       - 构建前后端"
    echo "  deploy      - 拉取代码 + 构建 + 重启（默认）"
    echo "  rollback    - 回滚到上一个版本"
    echo "  status      - 查看服务状态"
    echo "  logs        - 查看运行日志"
    echo "  help        - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh              完整部署流程"
    echo "  ./deploy.sh pull         只拉取代码"
    echo "  ./deploy.sh status       查看状态"
    echo ""
}

# ========== 检查环境 ==========
check_environment() {
    log "检查运行环境..."

    # 检查是否在正确的目录
    if [ ! -d "$PROJECT_DIR" ]; then
        log_error "项目目录不存在: $PROJECT_DIR"
        exit 1
    fi

    # 检查Git
    if ! command -v git &> /dev/null; then
        log_error "未安装 Git，请先安装: sudo apt install git"
        exit 1
    fi

    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "未安装 Go，请先安装: sudo apt install golang-go"
        exit 1
    fi

    # 创建必要目录
    mkdir -p "$BACKUP_DIR"
    mkdir -p "$(dirname "$LOG_FILE")"

    log_success "环境检查通过"
}

# ========== 备份当前版本 ==========
backup_current() {
    log "备份当前版本..."

    BACKUP_NAME="backup-$(date +'%Y%m%d-%H%M%S')"
    BACKUP_PATH="$BACKUP_DIR/$BACKUP_NAME"

    mkdir -p "$BACKUP_PATH"

    # 备份可执行文件
    if [ -f "$PROJECT_DIR/server" ]; then
        cp "$PROJECT_DIR/server" "$BACKUP_PATH/"
        log_success "已备份可执行文件"
    fi

    # 备份前端
    if [ -d "$FRONTEND_DIR/dist" ]; then
        cp -r "$FRONTEND_DIR/dist" "$BACKUP_PATH/"
        log_success "已备份前端文件"
    fi

    # 保留最近5个备份
    cd "$BACKUP_DIR"
    ls -t | tail -n +6 | xargs -r rm -rf

    echo "$BACKUP_NAME" > "$PROJECT_DIR/.last_backup"
    log_success "备份完成: $BACKUP_NAME"
}

# ========== 从GitHub拉取代码 ==========
pull_code() {
    log "从 GitHub 拉取最新代码..."

    cd "$PROJECT_DIR"

    # 检查是否是Git仓库
    if [ ! -d ".git" ]; then
        log_error "当前目录不是Git仓库，请先克隆项目"
        log "执行: git clone <你的GitHub仓库地址> $PROJECT_DIR"
        exit 1
    fi

    # 暂存本地修改（如果有）
    if ! git diff-index --quiet HEAD --; then
        log_warning "检测到本地修改，正在暂存..."
        git stash
    fi

    # 拉取最新代码
    git pull origin main || git pull origin master

    if [ $? -eq 0 ]; then
        log_success "代码拉取成功"
        git log -1 --oneline | tee -a "$LOG_FILE"
    else
        log_error "代码拉取失败"
        exit 1
    fi
}

# ========== 构建后端 ==========
build_backend() {
    log "构建后端..."

    cd "$BACKEND_DIR"

    # 下载依赖
    log "下载 Go 依赖..."
    go mod download

    # 编译
    log "编译后端服务..."
    CGO_ENABLED=1 go build -o "$PROJECT_DIR/server" -ldflags "-s -w" ./cmd/server

    if [ $? -eq 0 ]; then
        chmod +x "$PROJECT_DIR/server"
        log_success "后端编译成功 ($(du -h $PROJECT_DIR/server | cut -f1))"
    else
        log_error "后端编译失败"
        exit 1
    fi
}

# ========== 重启服务 ==========
restart_service() {
    log "重启 SmartGrow 服务..."

    systemctl restart smartgrow

    if [ $? -eq 0 ]; then
        sleep 3
        log_success "服务重启成功"

        # 健康检查
        log "执行健康检查..."
        if curl -s http://localhost:8080/health | grep -q "ok"; then
            log_success "健康检查通过 ✓"
        else
            log_warning "健康检查失败，请查看日志"
        fi
    else
        log_error "服务重启失败"
        exit 1
    fi
}

# ========== 回滚 ==========
rollback() {
    log "执行回滚操作..."

    if [ ! -f "$PROJECT_DIR/.last_backup" ]; then
        log_error "未找到备份信息"
        exit 1
    fi

    LAST_BACKUP=$(cat "$PROJECT_DIR/.last_backup")
    BACKUP_PATH="$BACKUP_DIR/$LAST_BACKUP"

    if [ ! -d "$BACKUP_PATH" ]; then
        log_error "备份不存在: $BACKUP_PATH"
        exit 1
    fi

    log "回滚到: $LAST_BACKUP"

    # 恢复文件
    if [ -f "$BACKUP_PATH/server" ]; then
        cp "$BACKUP_PATH/server" "$PROJECT_DIR/"
        chmod +x "$PROJECT_DIR/server"
        log_success "已恢复可执行文件"
    fi

    if [ -d "$BACKUP_PATH/dist" ]; then
        rm -rf "$FRONTEND_DIR/dist"
        cp -r "$BACKUP_PATH/dist" "$FRONTEND_DIR/"
        log_success "已恢复前端文件"
    fi

    # 重启服务
    restart_service

    log_success "回滚完成"
}

# ========== 查看状态 ==========
show_status() {
    echo ""
    echo "=========================================="
    echo "SmartGrow 服务状态"
    echo "=========================================="
    systemctl status smartgrow --no-pager
    echo ""
    echo "端口监听:"
    ss -tlnp | grep 8080 || echo "端口 8080 未监听"
    echo ""
}

# ========== 查看日志 ==========
show_logs() {
    echo "=========================================="
    echo "最近 50 行运行日志"
    echo "=========================================="
    tail -50 "$PROJECT_DIR/logs/server.log"
}

# ========== 完整部署流程 ==========
full_deploy() {
    echo ""
    echo "=========================================="
    echo "SmartGrow 完整部署流程"
    echo "=========================================="
    echo ""

    check_environment
    backup_current
    pull_code
    build_backend
    restart_service

    echo ""
    echo "=========================================="
    log_success "部署完成！"
    echo "=========================================="
    echo ""
    echo "访问地址: http://202.155.123.28:8080"
    echo "域名访问: https://iot.netr0.com"
    echo ""
}

# ========== 主函数 ==========
main() {
    COMMAND=${1:-deploy}

    case $COMMAND in
        pull)
            check_environment
            pull_code
            ;;
        build)
            check_environment
            build_backend
            ;;
        deploy)
            full_deploy
            ;;
        rollback)
            rollback
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
