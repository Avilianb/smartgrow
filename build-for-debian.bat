@echo off
REM 完整打包脚本 - 包含前端构建和后端源码，用于 Debian 13 服务器部署
chcp 65001 >nul
echo ======================================
echo SmartGrow 完整项目打包脚本
echo 目标平台: Debian 13 (Linux amd64)
echo ======================================
echo.

REM 检查必要工具
echo [1/5] 检查必要工具...
where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 Node.js，请先安装 Node.js
    pause
    exit /b 1
)

where tar >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 tar 命令（需要 Windows 10 1803+）
    pause
    exit /b 1
)

echo [✓] 所有必要工具已就绪
echo.

REM 设置输出文件名
set OUTPUT_FILE=smartgrow-v2.1-debian13.tar.gz
set TEMP_DIR=smartgrow-deploy-temp

echo [2/5] 构建前端...
cd /d "%~dp0frontend"
if not exist "node_modules" (
    echo 正在安装前端依赖...
    call npm install
    if %ERRORLEVEL% NEQ 0 (
        echo [错误] 前端依赖安装失败
        cd /d "%~dp0"
        pause
        exit /b 1
    )
)

echo 正在构建前端生产版本...
call npm run build
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 前端构建失败
    cd /d "%~dp0"
    pause
    exit /b 1
)
echo [✓] 前端构建完成
echo.

cd /d "%~dp0"

echo [3/5] 准备后端源码...
echo 注意: 后端将在服务器上编译，因为需要CGO支持
echo.

echo [4/5] 准备部署文件...
REM 删除旧的临时目录
if exist %TEMP_DIR% rmdir /s /q %TEMP_DIR%
mkdir %TEMP_DIR%

REM 复制后端源码
echo 复制后端源码...
mkdir %TEMP_DIR%\backend
xcopy /E /I /Q backend\cmd %TEMP_DIR%\backend\cmd
xcopy /E /I /Q backend\internal %TEMP_DIR%\backend\internal
xcopy /E /I /Q backend\configs %TEMP_DIR%\backend\configs
copy backend\go.mod %TEMP_DIR%\backend\
copy backend\go.sum %TEMP_DIR%\backend\

REM 复制前端构建产物
echo 复制前端构建产物...
mkdir %TEMP_DIR%\frontend
xcopy /E /I /Q frontend\dist %TEMP_DIR%\frontend\dist

REM 创建必要的目录
echo 创建运行时目录...
mkdir %TEMP_DIR%\data
mkdir %TEMP_DIR%\logs
echo. > %TEMP_DIR%\data\.gitkeep
echo. > %TEMP_DIR%\logs\.gitkeep

REM 创建编译脚本
echo 创建编译脚本...
(
echo #!/bin/bash
echo # SmartGrow 编译脚本
echo.
echo cd "$(dirname "$0"^)/backend"
echo.
echo echo "正在编译 SmartGrow 服务..."
echo echo "这可能需要几分钟..."
echo.
echo # 检查Go环境
echo if ! command -v go ^&^> /dev/null; then
echo     echo "错误: 未找到 Go，请先安装 Go 1.21+"
echo     echo "安装命令: sudo apt install golang-go"
echo     exit 1
echo fi
echo.
echo # 检查gcc（CGO需要）
echo if ! command -v gcc ^&^> /dev/null; then
echo     echo "错误: 未找到 GCC，请先安装"
echo     echo "安装命令: sudo apt install build-essential"
echo     exit 1
echo fi
echo.
echo # 编译
echo CGO_ENABLED=1 go build -o ../server -ldflags "-s -w" ./cmd/server
echo.
echo if [ $? -eq 0 ]; then
echo     echo "✓ 编译成功！"
echo     echo "可执行文件: server"
echo else
echo     echo "✗ 编译失败"
echo     exit 1
echo fi
) > %TEMP_DIR%\build.sh

REM 创建安装脚本
(
echo #!/bin/bash
echo # SmartGrow 安装脚本
echo.
echo echo "======================================"
echo echo "SmartGrow v2.1 安装向导"
echo echo "======================================"
echo echo ""
echo.
echo # 检查依赖
echo echo "[1/3] 检查系统依赖..."
echo.
echo if ! command -v go ^&^> /dev/null; then
echo     echo "未找到 Go，正在安装..."
echo     sudo apt update
echo     sudo apt install -y golang-go
echo fi
echo.
echo if ! command -v gcc ^&^> /dev/null; then
echo     echo "未找到 GCC，正在安装..."
echo     sudo apt install -y build-essential
echo fi
echo.
echo echo "✓ 依赖检查完成"
echo echo ""
echo.
echo # 编译服务
echo echo "[2/3] 编译服务..."
echo chmod +x build.sh
echo ./build.sh
echo.
echo if [ ! -f "server" ]; then
echo     echo "错误: 编译失败"
echo     exit 1
echo fi
echo.
echo echo "✓ 服务编译完成"
echo echo ""
echo.
echo # 设置权限
echo echo "[3/3] 设置权限..."
echo chmod +x server start.sh start-daemon.sh stop.sh
echo echo "✓ 权限设置完成"
echo echo ""
echo.
echo echo "======================================"
echo echo "安装完成！"
echo echo "======================================"
echo echo ""
echo echo "启动服务:"
echo echo "  ./start.sh          # 前台运行"
echo echo "  ./start-daemon.sh   # 后台运行"
echo echo ""
echo echo "访问地址: http://服务器IP:8080"
echo echo "默认账号: admin / admin123"
echo echo ""
) > %TEMP_DIR%\install.sh

REM 创建启动脚本
(
echo #!/bin/bash
echo # SmartGrow 启动脚本
echo.
echo cd "$(dirname "$0"^)"
echo.
echo # 检查是否已编译
echo if [ ! -f "server" ]; then
echo     echo "错误: 未找到可执行文件，请先运行 ./install.sh"
echo     exit 1
echo fi
echo.
echo # 设置环境变量
echo export GIN_MODE=release
echo export PORT=8080
echo.
echo # 启动服务
echo echo "正在启动 SmartGrow 服务..."
echo echo "访问地址: http://localhost:8080"
echo echo "按 Ctrl+C 停止服务"
echo echo ""
echo ./server
) > %TEMP_DIR%\start.sh

REM 创建后台启动脚本
(
echo #!/bin/bash
echo # SmartGrow 后台启动脚本
echo.
echo cd "$(dirname "$0"^)"
echo.
echo # 检查是否已编译
echo if [ ! -f "server" ]; then
echo     echo "错误: 未找到可执行文件，请先运行 ./install.sh"
echo     exit 1
echo fi
echo.
echo # 检查是否已在运行
echo if [ -f "smartgrow.pid" ]; then
echo     PID=$(cat smartgrow.pid^)
echo     if ps -p $PID ^> /dev/null 2^>^&1; then
echo         echo "SmartGrow 已经在运行 (PID: $PID^)"
echo         exit 1
echo     fi
echo fi
echo.
echo # 设置环境变量
echo export GIN_MODE=release
echo export PORT=8080
echo.
echo # 启动服务
echo nohup ./server ^> logs/server.log 2^>^&1 ^& echo $! ^> smartgrow.pid
echo echo "SmartGrow 已在后台启动"
echo echo "PID: $(cat smartgrow.pid^)"
echo echo "日志: logs/server.log"
echo echo "访问地址: http://localhost:8080"
) > %TEMP_DIR%\start-daemon.sh

REM 创建停止脚本
(
echo #!/bin/bash
echo # SmartGrow 停止脚本
echo.
echo cd "$(dirname "$0"^)"
echo.
echo if [ ! -f "smartgrow.pid" ]; then
echo     echo "未找到 PID 文件，服务可能未运行"
echo     exit 1
echo fi
echo.
echo PID=$(cat smartgrow.pid^)
echo if ps -p $PID ^> /dev/null 2^>^&1; then
echo     kill $PID
echo     echo "正在停止 SmartGrow (PID: $PID^)..."
echo     sleep 2
echo     if ps -p $PID ^> /dev/null 2^>^&1; then
echo         kill -9 $PID
echo         echo "强制停止 SmartGrow"
echo     fi
echo     rm -f smartgrow.pid
echo     echo "SmartGrow 已停止"
echo else
echo     echo "服务未运行 (PID: $PID^)"
echo     rm -f smartgrow.pid
echo fi
) > %TEMP_DIR%\stop.sh

REM 创建systemd服务文件
(
echo [Unit]
echo Description=SmartGrow Irrigation System
echo After=network.target
echo.
echo [Service]
echo Type=simple
echo User=smartgrow
echo WorkingDirectory=/opt/smartgrow
echo ExecStart=/opt/smartgrow/server
echo Restart=on-failure
echo RestartSec=10
echo Environment="GIN_MODE=release"
echo Environment="PORT=8080"
echo.
echo StandardOutput=append:/opt/smartgrow/logs/server.log
echo StandardError=append:/opt/smartgrow/logs/server.log
echo.
echo [Install]
echo WantedBy=multi-user.target
) > %TEMP_DIR%\smartgrow.service

REM 创建README
(
echo # SmartGrow v2.1 - Debian 13 部署指南
echo.
echo ## 快速开始
echo.
echo ### 1. 解压文件
echo ```bash
echo tar -xzf smartgrow-v2.1-debian13.tar.gz
echo cd smartgrow-deploy-temp
echo ```
echo.
echo ### 2. 安装并编译
echo ```bash
echo chmod +x install.sh
echo ./install.sh
echo ```
echo.
echo 此脚本会:
echo - 安装必要的依赖 ^(Go, GCC^)
echo - 编译后端服务
echo - 设置正确的文件权限
echo.
echo ### 3. 启动服务
echo.
echo **前台运行 ^(测试^):**
echo ```bash
echo ./start.sh
echo ```
echo.
echo **后台运行:**
echo ```bash
echo ./start-daemon.sh
echo ```
echo.
echo **停止服务:**
echo ```bash
echo ./stop.sh
echo ```
echo.
echo ### 4. 访问系统
echo.
echo 打开浏览器: `http://服务器IP:8080`
echo.
echo 默认账号:
echo - 用户名: `admin`
echo - 密码: `admin123`
echo.
echo ^> **重要: 首次登录后请立即修改密码！**
echo.
echo ## 系统要求
echo.
echo - Debian 13 ^(或其他 Linux 发行版^)
echo - Go 1.21+ ^(安装脚本会自动安装^)
echo - GCC 编译器 ^(安装脚本会自动安装^)
echo - 512MB+ 内存
echo - 100MB+ 磁盘空间
echo.
echo ## 手动编译 ^(可选^)
echo.
echo 如果自动安装失败，可以手动编译:
echo.
echo ```bash
echo # 安装依赖
echo sudo apt update
echo sudo apt install -y golang-go build-essential
echo.
echo # 编译
echo cd backend
echo CGO_ENABLED=1 go build -o ../server ./cmd/server
echo cd ..
echo.
echo # 设置权限
echo chmod +x server start.sh start-daemon.sh stop.sh
echo ```
echo.
echo ## 使用 systemd ^(生产环境推荐^)
echo.
echo ```bash
echo # 复制到系统目录
echo sudo mkdir -p /opt/smartgrow
echo sudo cp -r * /opt/smartgrow/
echo sudo chown -R $USER:$USER /opt/smartgrow
echo.
echo # 安装服务
echo sudo cp /opt/smartgrow/smartgrow.service /etc/systemd/system/
echo sudo systemctl daemon-reload
echo sudo systemctl enable smartgrow
echo sudo systemctl start smartgrow
echo.
echo # 查看状态
echo sudo systemctl status smartgrow
echo sudo journalctl -u smartgrow -f
echo ```
echo.
echo ## 配置
echo.
echo 编辑 `backend/configs/config.yaml`:
echo - 修改端口 ^(默认 8080^)
echo - 修改 JWT 密钥 ^(生产环境必须^)
echo - 配置数据库路径
echo.
echo ## 故障排查
echo.
echo ### 查看日志
echo ```bash
echo tail -f logs/server.log
echo ```
echo.
echo ### 检查端口
echo ```bash
echo sudo netstat -tlnp ^| grep 8080
echo ```
echo.
echo ### 检查进程
echo ```bash
echo ps aux ^| grep server
echo ```
echo.
echo ## 目录结构
echo.
echo ```
echo smartgrow/
echo ├── backend/             # 后端源码
echo │   ├── cmd/
echo │   ├── internal/
echo │   ├── configs/
echo │   ├── go.mod
echo │   └── go.sum
echo ├── frontend/            # 前端构建产物
echo │   └── dist/
echo ├── data/                # 数据库文件
echo ├── logs/                # 日志文件
echo ├── install.sh           # 安装脚本
echo ├── build.sh             # 编译脚本
echo ├── start.sh             # 启动脚本
echo ├── start-daemon.sh      # 后台启动
echo ├── stop.sh              # 停止脚本
echo ├── smartgrow.service    # systemd 服务
echo └── README.md            # 本文档
echo ```
echo.
echo ## 更新部署
echo.
echo 1. 停止服务: `./stop.sh`
echo 2. 备份数据: `cp -r data data.backup`
echo 3. 解压新版本覆盖文件
echo 4. 重新编译: `./install.sh`
echo 5. 启动服务: `./start-daemon.sh`
echo.
echo ## Nginx 反向代理 ^(可选^)
echo.
echo ```nginx
echo server {
echo     listen 80;
echo     server_name your-domain.com;
echo.
echo     location / {
echo         proxy_pass http://localhost:8080;
echo         proxy_set_header Host $host;
echo         proxy_set_header X-Real-IP $remote_addr;
echo     }
echo }
echo ```
echo.
echo ## 技术支持
echo.
echo 如遇问题请查看日志文件或联系技术支持。
) > %TEMP_DIR%\README.md

echo [✓] 部署文件准备完成
echo.

echo [5/5] 打包文件...
tar -czf %OUTPUT_FILE% -C %TEMP_DIR% .
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 打包失败
    pause
    exit /b 1
)

echo [✓] 打包完成
echo.

echo 清理临时文件...
rmdir /s /q %TEMP_DIR%
echo [✓] 清理完成
echo.

echo ======================================
echo 打包成功！
echo ======================================
echo.
echo 输出文件: %OUTPUT_FILE%
dir /-c %OUTPUT_FILE% | find "%OUTPUT_FILE%"
echo.
echo ======================================
echo 部署步骤:
echo ======================================
echo.
echo 1. 上传到服务器:
echo    scp %OUTPUT_FILE% user@server:/home/user/
echo.
echo 2. 解压:
echo    tar -xzf smartgrow-v2.1-debian13.tar.gz
echo    cd smartgrow-deploy-temp
echo.
echo 3. 安装:
echo    chmod +x install.sh
echo    ./install.sh
echo.
echo 4. 启动:
echo    ./start-daemon.sh
echo.
echo 详细说明请查看压缩包内的 README.md
echo.
pause
