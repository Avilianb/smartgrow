# 🌱 SmartGrow - 智能灌溉系统

<div align="center">

**基于IoT的智能农业灌溉管理系统**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18.2-61DAFB?logo=react)](https://reactjs.org/)
[![Platform](https://img.shields.io/badge/Platform-ESP32--S3-E7352C?logo=espressif)](https://www.espressif.com/)

[English](README_EN.md) | 简体中文

</div>

---

## 🚀 快速开始（一键安装）

### 自动安装（推荐）

适用于 Debian 13 / Ubuntu 20.04+ 系统，自动完成所有配置：

```bash
# 1. 下载安装脚本
wget https://raw.githubusercontent.com/Avilianb/smartgrow/main/install.sh

# 2. 添加执行权限
chmod +x install.sh

# 3. 运行安装（需要root权限）
sudo bash install.sh

# 国内服务器推荐：启用国内镜像加速
# 运行时会自动询问是否使用国内镜像，或者手动指定：
# USE_CHINA_MIRROR=true sudo bash install.sh
```

**安装脚本会自动完成：**
- ✅ 安装Node.js 20.x LTS和Go 1.21
- ✅ 克隆项目代码
- ✅ 安装前后端依赖
- ✅ 构建前后端代码
- ✅ 配置数据库
- ✅ 创建systemd服务
- ✅ 启动SmartGrow服务

**国内服务器优化：**
- 🚀 自动询问是否启用国内镜像加速
- 🌐 Go下载：使用阿里云镜像 mirrors.aliyun.com
- 📦 npm安装：使用阿里云镜像 npmmirror.com
- 🔧 Go模块：配置阿里云代理 mirrors.aliyun.com/goproxy
- 💾 GitHub：使用 ghproxy.com 加速克隆

**安装完成后：**
- 访问地址：`http://你的服务器IP:8080`
- 默认账户：`admin` / `admin123`
- 服务管理：`systemctl {start|stop|restart|status} smartgrow`

### 手动安装

如果自动安装失败或需要自定义配置，请参考下方的[详细部署文档](#-部署指南)。

---

## 📖 项目简介

SmartGrow 是一个完整的物联网智能灌溉解决方案，通过ESP32设备采集环境数据，结合天气预报和动态规划算法，实现自动化、智能化的农作物灌溉管理。

### ✨ 核心特性

- 🌡️ **实时监控** - 温度、湿度、土壤湿度、降雨量实时采集
- 🤖 **智能规划** - 基于动态规划算法优化15天灌溉计划
- ☁️ **天气集成** - 集成和风天气API获取精准气象数据
- 📱 **远程控制** - Web端远程控制水泵、遮阳装置
- 📊 **数据可视化** - 历史数据图表展示和分析
- 👥 **多用户管理** - 支持管理员和普通用户角色
- 🔐 **安全可靠** - JWT认证、密码加密、审计日志

---

## 🏗️ 技术架构

### 系统架构图

```
┌─────────────────┐
│   ESP32-S3      │  ← 传感器数据采集
│   (Arduino)     │  ← 水泵/遮阳控制
└────────┬────────┘
         │ HTTP API
         ▼
┌─────────────────┐
│   Go Backend    │  ← RESTful API
│   (Gin)         │  ← 业务逻辑
│   SQLite        │  ← 数据存储
└────────┬────────┘
         │ JSON
         ▼
┌─────────────────┐
│  React Frontend │  ← Web界面
│  (Vite+TS)      │  ← 数据可视化
└─────────────────┘
```

### 技术栈

**固件层 (Firmware)**
- ESP32-S3 微控制器
- Arduino Framework
- PlatformIO 构建系统
- DHT11/22 温湿度传感器
- 电容式土壤湿度传感器

**后端 (Backend)**
- Go 1.21+
- Gin Web Framework
- SQLite 数据库
- JWT 认证
- 和风天气 API
- 动态规划算法

**前端 (Frontend)**
- React 18.2 + TypeScript
- Vite 构建工具
- Tailwind CSS 4.x
- React Router 6
- Recharts 图表库
- Leaflet 地图

---

## 🚀 快速开始

### 前置要求

- **开发环境**
  - Node.js 16+ (前端开发)
  - Go 1.21+ (后端开发)
  - Git

- **服务器环境**
  - Debian 13 / Ubuntu 20.04+
  - Go 1.21+
  - GCC (CGO 编译需要)
  - 512MB+ 内存

### 本地开发

#### 1. 克隆项目

```bash
git clone https://github.com/Avilianb/smartgrow.git
cd smartgrow
```

#### 2. 启动后端

```bash
cd backend

# 安装依赖
go mod download

# 创建配置文件
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml，修改必要的配置

# 运行
go run cmd/server/main.go
```

后端将在 `http://localhost:8080` 启动

#### 3. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动

---

## 📦 部署到服务器

### 方式一：使用自动化脚本（推荐）

项目提供了完整的自动化部署脚本：

#### 从本地部署

```bash
# 在项目根目录执行
cd deployment

# 部署全部（前端+后端）
deploy.bat all

# 只部署前端
deploy.bat frontend

# 只部署后端
deploy.bat backend

# 查看服务状态
deploy.bat status
```

#### 在服务器上部署

```bash
# SSH 到服务器
ssh root@your-server

# 克隆项目（首次）
git clone https://github.com/你的用户名/smartgrow.git /root/smart-grow
cd /root/smart-grow

# 上传部署脚本
scp deployment/deploy.sh root@your-server:/root/smart-grow/

# 添加执行权限
chmod +x deploy.sh

# 执行部署
./deploy.sh

# 后续更新只需执行
./deploy.sh
```

### 方式二：手动部署

详细步骤请查看 [部署文档](deployment/DEPLOYMENT.md)

---

## 🔧 配置说明

### 后端配置

编辑 `backend/configs/config.yaml`：

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release  # debug/release

database:
  path: /opt/irrigation/db/irrigation.db

security:
  jwt_secret: "你的JWT密钥（至少32字符）"
  jwt_expire_hours: 24
  allowed_origins:
    - "https://your-domain.com"
  device_api_key: "你的设备API密钥"

weather:
  api_host: "devapi.qweather.com"
  api_key: "你的和风天气API密钥"
```

**重要提示：** 生产环境务必修改：
- `jwt_secret` - 随机生成的强密钥
- `device_api_key` - 设备认证密钥
- `allowed_origins` - 允许的前端域名

### 前端配置

前端使用环境变量，在 `.env.local` 中配置：

```env
VITE_API_BASE_URL=http://your-server:8080/api
```

---

## 📱 使用指南

### 默认登录信息

**管理员账户**
- 用户名: `admin`
- 密码: `admin123`

⚠️ **首次登录后请立即修改密码！**

### 功能模块

#### 仪表盘 (Dashboard)
- 实时设备状态查看
- 传感器数据监控
- 今日灌溉计划
- 历史数据图表

#### 位置管理 (Location Manager)
- 设备地理位置设置
- 地图标注
- 天气预报更新

#### 系统日志 (System Logs)
- 设备操作日志
- 系统事件记录
- 日志筛选和搜索

#### 用户管理 (仅管理员)
- 创建普通用户
- 分配设备
- 用户权限管理

---

## 🔌 ESP32 固件配置

在 `firmware/src/main.cpp` 中配置：

```cpp
// WiFi 配置
const char* WIFI_SSID = "你的WiFi名称";
const char* WIFI_PASSWORD = "你的WiFi密码";

// 服务器配置
const char* SERVER_URL = "http://your-server:8080";
const char* DEVICE_ID = "esp32s3-1";
const char* DEVICE_API_KEY = "你的设备API密钥";  // 与后端配置一致
```

烧录固件：

```bash
cd firmware
pio run --target upload
```

---

## 📊 API 文档

### 认证接口

#### 管理员登录
```http
POST /api/auth/admin/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

#### 普通用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "password"
}
```

### 设备接口

#### 上传传感器数据
```http
POST /api/device/data
Content-Type: application/json
Authorization: DeviceAPIKey <your-api-key>

{
  "device_id": "esp32s3-1",
  "timestamp": "2025-12-08T10:00:00Z",
  "temperature_c": 25.5,
  "humidity_pct": 60.0,
  "soil_raw": 2000,
  "rain_analog": 1023,
  "rain_digital": 1,
  "pump_state": "off",
  "shade_state": "closed"
}
```

更多API详情请查看 [API文档](docs/API.md)

---

## 🛠️ 开发指南

### 项目结构

```
smartgrow/
├── backend/                 # 后端服务
│   ├── cmd/
│   │   └── server/         # 主程序入口
│   ├── internal/
│   │   ├── api/            # API定义
│   │   ├── config/         # 配置管理
│   │   ├── database/       # 数据库
│   │   ├── handler/        # 请求处理
│   │   ├── middleware/     # 中间件
│   │   ├── models/         # 数据模型
│   │   ├── planner/        # 灌溉算法
│   │   ├── repository/     # 数据仓储
│   │   ├── service/        # 业务逻辑
│   │   └── weather/        # 天气API
│   └── configs/            # 配置文件
│
├── frontend/               # 前端应用
│   ├── src/
│   │   ├── components/    # 可复用组件
│   │   ├── contexts/      # React Context
│   │   ├── pages/         # 页面组件
│   │   ├── services/      # API服务
│   │   └── types/         # TypeScript类型
│   └── public/            # 静态资源
│
├── firmware/              # ESP32固件
│   ├── src/               # 源代码
│   ├── include/           # 头文件
│   └── platformio.ini     # PlatformIO配置
│
└── deployment/            # 部署脚本
    ├── deploy.bat         # Windows本地部署
    ├── deploy.sh          # 服务器端部署
    └── scripts/           # 辅助脚本
```

### 添加新功能

1. **后端添加API**
   - 在 `internal/models/` 定义数据模型
   - 在 `internal/repository/` 添加数据库操作
   - 在 `internal/service/` 实现业务逻辑
   - 在 `internal/handler/` 添加HTTP处理器
   - 在 `handler.go` 注册路由

2. **前端添加页面**
   - 在 `src/pages/` 创建页面组件
   - 在 `src/services/api.ts` 添加API调用
   - 在 `App.tsx` 添加路由
   - 在 `Sidebar.tsx` 添加菜单

---

## 🧪 测试

### 后端测试

```bash
cd backend
go test ./...
```

### 前端测试

```bash
cd frontend
npm test
```

---

## 📝 常见问题

### 1. 部署后无法访问？

检查防火墙是否开放8080端口：
```bash
sudo ufw allow 8080
```

### 2. 数据库初始化失败？

确保数据库目录存在且有写权限：
```bash
sudo mkdir -p /opt/irrigation/db
sudo chmod 755 /opt/irrigation/db
```

### 3. ESP32连接失败？

- 检查WiFi配置是否正确
- 确认服务器地址可访问
- 验证设备API密钥是否匹配

### 4. 前端API调用失败？

检查CORS配置，确保前端域名在 `allowed_origins` 中

更多问题请查看 [FAQ](docs/FAQ.md)

---

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Go Web框架
- [React](https://reactjs.org/) - 前端框架
- [和风天气](https://www.qweather.com/) - 天气数据API
- [ESP32](https://www.espressif.com/) - 物联网芯片

---

## 📞 联系方式

- 项目主页: https://github.com/你的用户名/smartgrow
- 问题反馈: https://github.com/你的用户名/smartgrow/issues
- 邮箱: your-email@example.com

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给一个Star！**

Made with ❤️ by [您的名字]

</div>
