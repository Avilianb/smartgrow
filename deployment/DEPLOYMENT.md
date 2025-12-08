# ==========================================
# SmartGrow 部署说明文档
# ==========================================

## 📋 目录

1. [环境准备](#环境准备)
2. [首次部署](#首次部署)
3. [日常更新](#日常更新)
4. [常见问题](#常见问题)
5. [回滚操作](#回滚操作)

---

## 🔧 环境准备

### 服务器要求

- **系统**: Debian 13 / Ubuntu 20.04+
- **内存**: 512MB+
- **磁盘**: 1GB+
- **软件**: Go 1.21+, GCC, Git

### 安装必要软件

```bash
# 更新软件源
sudo apt update

# 安装 Go
sudo apt install -y golang-go

# 安装 GCC（CGO编译需要）
sudo apt install -y build-essential

# 安装 Git
sudo apt install -y git

# 验证安装
go version    # 应显示 go1.21 或更高
gcc --version
git --version
```

---

## 🚀 首次部署

### 步骤1: 克隆项目到服务器

```bash
# SSH 登录服务器
ssh root@your-server-ip

# 克隆项目
cd /root
git clone https://github.com/你的用户名/smartgrow.git smart-grow
cd smart-grow
```

### 步骤2: 配置文件

```bash
# 复制配置模板
cd backend/configs
cp config.example.yaml config.yaml

# 编辑配置文件
nano config.yaml

# 必须修改的配置：
# 1. security.jwt_secret - 生成新的JWT密钥
# 2. security.device_api_key - 生成新的设备API密钥
# 3. security.allowed_origins - 添加您的域名
# 4. weather.api_key - 和风天气API密钥
```

生成密钥的命令：

```bash
# 生成JWT密钥
openssl rand -base64 48

# 生成设备API密钥
openssl rand -hex 32
```

### 步骤3: 上传部署脚本

```bash
# 在本地执行（Windows）
cd C:\coding\deployment
scp deploy.sh root@your-server-ip:/root/smart-grow/

# 在服务器上添加执行权限
ssh root@your-server-ip "chmod +x /root/smart-grow/deploy.sh"
```

### 步骤4: 执行部署

```bash
# SSH到服务器
ssh root@your-server-ip

# 进入项目目录
cd /root/smart-grow

# 执行部署脚本
./deploy.sh

# 脚本会自动完成：
# ✓ 环境检查
# ✓ 备份（如有旧版本）
# ✓ 编译后端
# ✓ 配置systemd服务
# ✓ 启动服务
# ✓ 健康检查
```

### 步骤5: 配置Nginx反向代理（可选）

如果使用Nginx Proxy Manager：

1. 登录NPM管理界面
2. 添加代理主机：
   - **Domain Names**: `iot.netr0.com`
   - **Scheme**: `http`
   - **Forward Hostname/IP**: `localhost`
   - **Forward Port**: `8080`
3. 启用SSL（Let's Encrypt）
4. 保存

### 步骤6: 验证部署

```bash
# 检查服务状态
systemctl status smartgrow

# 测试API
curl http://localhost:8080/health

# 查看日志
tail -f /root/smart-grow/logs/server.log
```

访问 `http://your-server-ip:8080` 应该能看到登录界面。

---

## 🔄 日常更新

### 方式一：本地一键部署（推荐）

在本地开发机上：

```bash
# 1. 提交代码到GitHub
git add .
git commit -m "更新说明"
git push origin main

# 2. 使用部署脚本
cd deployment

# 部署全部
deploy.bat all

# 或只部署前端/后端
deploy.bat frontend
deploy.bat backend
```

### 方式二：服务器端拉取

在服务器上：

```bash
# SSH到服务器
ssh root@your-server-ip

# 进入项目目录
cd /root/smart-grow

# 执行更新
./deploy.sh

# 脚本会自动：
# ✓ 备份当前版本
# ✓ 从GitHub拉取最新代码
# ✓ 编译后端
# ✓ 重启服务
# ✓ 健康检查
```

### 更新流程说明

```
本地开发
   ↓
修改代码
   ↓
git commit + git push
   ↓
GitHub仓库更新
   ↓
服务器执行 ./deploy.sh
   ↓
自动拉取 → 构建 → 部署
```

---

## ❓ 常见问题

### 1. Git拉取失败

**问题**: `git pull` 提示认证失败

**解决**:

```bash
# 方法1: 使用SSH密钥（推荐）
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
cat ~/.ssh/id_rsa.pub
# 将公钥添加到 GitHub Settings → SSH Keys

# 方法2: 使用Personal Access Token
# 在GitHub生成token，然后：
git remote set-url origin https://YOUR_TOKEN@github.com/username/repo.git
```

### 2. 编译失败

**问题**: `go build` 失败

**解决**:

```bash
# 清理缓存
go clean -cache -modcache

# 重新下载依赖
cd backend
go mod download
go mod tidy

# 重新编译
CGO_ENABLED=1 go build -o ../server ./cmd/server
```

### 3. 端口被占用

**问题**: `address already in use`

**解决**:

```bash
# 查找占用进程
ss -tlnp | grep 8080

# 杀死进程
kill -9 <PID>

# 或直接重启服务
systemctl restart smartgrow
```

### 4. 数据库权限问题

**问题**: 无法写入数据库

**解决**:

```bash
# 创建目录
sudo mkdir -p /opt/irrigation/db

# 设置权限
sudo chown -R root:root /opt/irrigation
sudo chmod -R 755 /opt/irrigation
```

### 5. 前端无法连接后端

**问题**: API调用失败，CORS错误

**解决**:

检查后端配置 `backend/configs/config.yaml`:

```yaml
security:
  allowed_origins:
    - "https://your-actual-domain.com"  # 确保域名正确
```

然后重启服务：
```bash
systemctl restart smartgrow
```

---

## ⏮️ 回滚操作

### 自动回滚

部署脚本会自动备份，出现问题时可快速回滚：

```bash
# 在服务器上执行
cd /root/smart-grow
./deploy.sh rollback
```

### 手动回滚

如果自动回滚失败，可手动操作：

```bash
# 查看备份
ls -la /root/smart-grow/backups/

# 选择要回滚的备份（例如 backup-20251208-120000）
BACKUP=backup-20251208-120000

# 恢复可执行文件
cp backups/$BACKUP/server ./server
chmod +x server

# 恢复前端（如果有）
rm -rf frontend/dist
cp -r backups/$BACKUP/dist frontend/

# 重启服务
systemctl restart smartgrow

# 验证
systemctl status smartgrow
curl http://localhost:8080/health
```

### 回滚到指定Git版本

```bash
# 查看提交历史
git log --oneline

# 回滚到指定版本
git checkout <commit-hash>

# 重新部署
./deploy.sh

# 如需永久回退
git reset --hard <commit-hash>
git push -f origin main  # 谨慎使用
```

---

## 🔐 安全最佳实践

### 1. 修改默认密码

首次登录后立即修改管理员密码：

1. 登录系统
2. 进入"修改密码"页面
3. 设置强密码（至少12位，包含大小写字母、数字、特殊字符）

### 2. 保护配置文件

```bash
# 确保配置文件不被提交到Git
echo "backend/configs/config.yaml" >> .gitignore

# 设置文件权限
chmod 600 backend/configs/config.yaml
```

### 3. 启用HTTPS

使用Nginx Proxy Manager启用SSL证书（Let's Encrypt）

### 4. 定期备份数据库

```bash
# 创建备份脚本
cat > /root/backup_db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/root/db_backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
cp /opt/irrigation/db/irrigation.db $BACKUP_DIR/irrigation_$DATE.db
# 保留最近30天的备份
find $BACKUP_DIR -name "irrigation_*.db" -mtime +30 -delete
EOF

chmod +x /root/backup_db.sh

# 添加到crontab（每天凌晨2点备份）
(crontab -l 2>/dev/null; echo "0 2 * * * /root/backup_db.sh") | crontab -
```

---

## 📊 监控与维护

### 查看系统状态

```bash
# 服务状态
systemctl status smartgrow

# 实时日志
journalctl -u smartgrow -f

# 应用日志
tail -f /root/smart-grow/logs/server.log

# 资源使用
htop
df -h
free -h
```

### 性能优化

1. **启用日志轮转**:

```bash
cat > /etc/logrotate.d/smartgrow << EOF
/root/smart-grow/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
EOF
```

2. **数据库优化**:

定期执行 VACUUM 清理数据库：

```bash
sqlite3 /opt/irrigation/db/irrigation.db "VACUUM;"
```

---

## 📞 获取帮助

- 查看项目Wiki
- 提交Issue: https://github.com/你的用户名/smartgrow/issues
- 查看部署日志: `cat /root/smart-grow/logs/deploy.log`

---

**祝您部署顺利！** 🎉
