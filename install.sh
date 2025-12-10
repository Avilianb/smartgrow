#!/bin/bash

##############################################################################
# SmartGrow 智能灌溉系统 - 一键安装脚本
#
# 适用于: Debian 13 / Ubuntu 20.04+
# 用途: 自动安装和配置SmartGrow系统的所有组件
##############################################################################

set -e  # 遇到错误立即退出

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

# 显示欢迎信息
show_banner() {
    echo "======================================"
    echo "  SmartGrow 智能灌溉系统安装程序"
    echo "======================================"
    echo ""
}

# 检查是否以root运行
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        log_info "使用方法: sudo bash install.sh"
        exit 1
    fi
}

# 检查系统
check_system() {
    log_info "检查系统环境..."

    # 检查系统类型
    if [ -f /etc/debian_version ]; then
        log_success "系统检查通过: Debian/Ubuntu"
    else
        log_warning "此脚本为Debian/Ubuntu优化，其他系统可能需要调整"
    fi

    # 检查架构
    ARCH=$(uname -m)
    log_info "系统架构: $ARCH"
}

# 安装系统依赖
install_dependencies() {
    log_info "更新系统包列表..."
    apt-get update -qq

    log_info "安装基础工具..."
    apt-get install -y -qq curl wget git build-essential

    log_success "基础工具安装完成"
}

# 安装Node.js
install_nodejs() {
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node -v)
        log_success "Node.js已安装: $NODE_VERSION"
    else
        log_info "安装Node.js 20.x LTS..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        log_success "Node.js安装完成: $(node -v)"
    fi

    # 检查npm
    if command -v npm &> /dev/null; then
        log_success "npm已安装: $(npm -v)"
    fi
}

# 安装Go
install_golang() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        log_success "Go已安装: $GO_VERSION"
    else
        log_info "安装Go 1.21..."
        GO_VERSION="1.21.5"
        GO_FILE="go${GO_VERSION}.linux-amd64.tar.gz"

        # 显示下载进度，设置超时为10分钟
        log_info "下载 $GO_FILE (约140MB，请耐心等待)..."
        if ! wget --progress=bar:force --timeout=600 https://go.dev/dl/$GO_FILE 2>&1 | \
            grep --line-buffered "%" | \
            sed -u -e "s,\.,,g" | \
            awk '{printf("\r[下载进度] %s", $0)}'; then
            log_error "下载失败，请检查网络连接"
            exit 1
        fi
        echo ""  # 换行

        log_info "解压Go安装包..."
        rm -rf /usr/local/go
        tar -C /usr/local -xzf $GO_FILE
        rm $GO_FILE

        # 添加到PATH
        if ! grep -q "/usr/local/go/bin" /etc/profile; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        fi
        export PATH=$PATH:/usr/local/go/bin

        log_success "Go安装完成: $(go version)"
    fi
}

# 克隆项目
clone_project() {
    PROJECT_DIR="/root/smart-grow"

    if [ -d "$PROJECT_DIR" ]; then
        log_warning "项目目录已存在: $PROJECT_DIR"
        read -p "是否删除并重新克隆? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$PROJECT_DIR"
        else
            log_info "使用现有项目目录"
            return
        fi
    fi

    log_info "克隆项目到 $PROJECT_DIR..."
    git clone https://github.com/Avilianb/smartgrow.git "$PROJECT_DIR"
    cd "$PROJECT_DIR"
    log_success "项目克隆完成"
}

# 安装前端依赖
install_frontend() {
    log_info "安装前端依赖..."
    cd "$PROJECT_DIR/frontend"
    npm install --silent
    log_success "前端依赖安装完成"

    log_info "构建前端..."
    npm run build
    log_success "前端构建完成"
}

# 安装后端依赖
install_backend() {
    log_info "安装后端依赖..."
    cd "$PROJECT_DIR/backend"
    go mod download
    log_success "后端依赖下载完成"

    log_info "编译后端..."
    go build -o server cmd/server/main.go
    log_success "后端编译完成"
}

# 配置数据库
setup_database() {
    log_info "初始化数据库..."
    cd "$PROJECT_DIR/backend"

    # 创建数据目录
    mkdir -p data

    # 数据库会在首次启动时自动创建
    log_success "数据库配置完成"
}

# 配置文件
setup_config() {
    log_info "配置应用..."

    cd "$PROJECT_DIR/backend/configs"

    # 如果配置文件不存在，从示例复制
    if [ ! -f config.yaml ]; then
        if [ -f config.example.yaml ]; then
            cp config.example.yaml config.yaml
            log_info "已从config.example.yaml创建配置文件"
        fi
    fi

    # 生成新的密钥（如果需要）
    JWT_SECRET=$(openssl rand -hex 32)
    DEVICE_API_KEY=$(openssl rand -hex 32)

    log_info "配置文件位置: $PROJECT_DIR/backend/configs/config.yaml"
    log_warning "请根据需要修改配置文件"
    log_info "JWT密钥: $JWT_SECRET"
    log_info "设备API密钥: $DEVICE_API_KEY"
}

# 创建systemd服务
setup_systemd() {
    log_info "创建systemd服务..."

    cat > /etc/systemd/system/smartgrow.service << EOF
[Unit]
Description=SmartGrow Irrigation System
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_DIR/backend
ExecStart=$PROJECT_DIR/backend/server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    systemctl daemon-reload

    # 启用服务
    systemctl enable smartgrow

    log_success "Systemd服务创建完成"
}

# 启动服务
start_service() {
    log_info "启动SmartGrow服务..."
    systemctl start smartgrow

    sleep 2

    if systemctl is-active --quiet smartgrow; then
        log_success "服务启动成功！"
        systemctl status smartgrow --no-pager
    else
        log_error "服务启动失败"
        systemctl status smartgrow --no-pager
        exit 1
    fi
}

# 显示完成信息
show_completion() {
    echo ""
    echo "======================================"
    log_success "SmartGrow安装完成！"
    echo "======================================"
    echo ""
    log_info "服务状态: $(systemctl is-active smartgrow)"
    log_info "项目目录: $PROJECT_DIR"
    log_info "访问地址: http://$(hostname -I | awk '{print $1}'):8080"
    echo ""
    log_info "默认管理员账户:"
    log_info "  用户名: admin"
    log_info "  密码: admin123"
    echo ""
    log_info "常用命令:"
    echo "  启动服务: systemctl start smartgrow"
    echo "  停止服务: systemctl stop smartgrow"
    echo "  重启服务: systemctl restart smartgrow"
    echo "  查看状态: systemctl status smartgrow"
    echo "  查看日志: journalctl -u smartgrow -f"
    echo ""
    log_warning "请记得修改默认密码！"
    echo "======================================"
}

# 主函数
main() {
    show_banner
    check_root
    check_system

    log_info "开始安装SmartGrow系统..."
    echo ""

    install_dependencies
    install_nodejs
    install_golang
    clone_project
    install_frontend
    install_backend
    setup_database
    setup_config
    setup_systemd
    start_service

    show_completion
}

# 运行主函数
main
