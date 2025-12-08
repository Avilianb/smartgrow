# SmartGrow 项目测试改进流程

本文档详细说明了从本地开发到服务器部署、再到GitHub版本管理的完整工作流程。

## 📋 目录

- [开发环境准备](#开发环境准备)
- [本地开发流程](#本地开发流程)
- [测试流程](#测试流程)
- [服务器部署流程](#服务器部署流程)
- [GitHub版本管理](#github版本管理)
- [完整工作流示例](#完整工作流示例)
- [常见问题](#常见问题)

---

## 开发环境准备

### 本地环境要求

- **Node.js**: 18+ (前端开发)
- **Go**: 1.21+ (后端开发)
- **Git**: 版本管理
- **SSH**: 服务器连接（需配置密钥认证）

### 服务器配置

- **地址**: 202.155.123.28
- **用户**: root
- **项目路径**: `/root/smart-grow/`
- **服务名称**: `smartgrow.service`
- **访问地址**: http://iot.netr0.com 或 http://202.155.123.28:8080

### SSH密钥配置（首次）

```bash
# 生成SSH密钥（如果还没有）
ssh-keygen -t rsa -b 4096

# 上传公钥到服务器
type %USERPROFILE%\.ssh\id_rsa.pub | ssh root@202.155.123.28 "cat >> ~/.ssh/authorized_keys"
```

---

## 本地开发流程

### 1. 获取最新代码

```bash
# 拉取最新代码
git pull origin main

# 查看当前分支和状态
git status
```

### 2. 前端开发

#### 启动开发服务器

```bash
cd frontend
npm install          # 首次或package.json变更后
npm run dev          # 启动开发服务器 (http://localhost:5173)
```

#### 开发规范

- **组件路径**: `frontend/src/components/`
- **页面路径**: `frontend/src/pages/`
- **API服务**: `frontend/src/services/api.ts`
- **类型定义**: `frontend/src/types/index.ts`

#### 本地测试

```bash
# 运行TypeScript类型检查
npm run build

# 检查代码格式
npm run lint
```

### 3. 后端开发

#### 启动后端服务器

```bash
cd backend
go mod download      # 首次或依赖变更后
go run cmd/server/main.go
```

后端将在 `http://localhost:8080` 启动

#### 开发规范

- **配置文件**: `backend/configs/config.yaml`
- **API处理器**: `backend/internal/handlers/`
- **业务逻辑**: `backend/internal/services/`
- **数据模型**: `backend/internal/models/`

#### 本地测试

```bash
# 运行测试
go test ./...

# 构建检查
go build -o server cmd/server/main.go
```

### 4. ESP32固件开发

```bash
cd firmware

# 使用PlatformIO编译
pio run

# 上传到设备
pio run -t upload
```

配置文件：`firmware/include/config.h`

---

## 测试流程

### 前端测试清单

- [ ] 页面加载无错误（检查浏览器Console）
- [ ] 所有功能按钮正常工作
- [ ] 表单验证正常
- [ ] API调用成功（检查Network标签）
- [ ] 响应式布局正常（测试不同屏幕尺寸）
- [ ] 地图显示正常（如适用）

### 后端测试清单

- [ ] API端点返回正确状态码
- [ ] 数据库操作正常
- [ ] JWT认证正常工作
- [ ] 错误处理正确
- [ ] CORS配置正确

### 集成测试

```bash
# 1. 启动后端
cd backend
go run cmd/server/main.go

# 2. 在另一个终端启动前端
cd frontend
npm run dev

# 3. 在浏览器访问 http://localhost:5173 进行测试
```

---

## 服务器部署流程

### 方式一：使用自动化脚本（推荐）

#### Windows本地部署脚本

```bash
# 部署前端
deployment\deploy.bat frontend

# 部署后端
deployment\deploy.bat backend

# 部署全部
deployment\deploy.bat all

# 查看服务状态
deployment\deploy.bat status

# 查看日志
deployment\deploy.bat logs

# 重启服务
deployment\deploy.bat restart
```

#### Linux服务器脚本

```bash
# SSH到服务器后执行
cd /root/smart-grow
./deployment/deploy.sh
```

### 方式二：手动部署

#### 部署前端

```bash
# 1. 构建前端
cd frontend
npm run build

# 2. 上传到服务器
scp -r dist/* root@202.155.123.28:/root/smart-grow/frontend/dist/

# 3. 重启服务
ssh root@202.155.123.28 "systemctl restart smartgrow"
```

#### 部署后端

```bash
# 1. 上传代码
cd backend
scp -r * root@202.155.123.28:/root/smart-grow/backend/

# 2. 在服务器上编译
ssh root@202.155.123.28 "cd /root/smart-grow/backend && go build -o server cmd/server/main.go"

# 3. 重启服务
ssh root@202.155.123.28 "systemctl restart smartgrow"
```

### 验证部署

```bash
# 检查服务状态
ssh root@202.155.123.28 "systemctl status smartgrow"

# 查看日志
ssh root@202.155.123.28 "journalctl -u smartgrow -f"

# 测试API端点
curl http://202.155.123.28:8080/health
```

### 服务器端常用命令

```bash
# 启动服务
systemctl start smartgrow

# 停止服务
systemctl stop smartgrow

# 重启服务
systemctl restart smartgrow

# 查看状态
systemctl status smartgrow

# 查看实时日志
journalctl -u smartgrow -f

# 查看最近100行日志
journalctl -u smartgrow -n 100

# 重新加载systemd配置
systemctl daemon-reload
```

---

## GitHub版本管理

### 提交规范（Semantic Commit Messages）

使用语义化提交信息，格式：`<type>: <description>`

#### 提交类型

- `feat:` 新功能
- `fix:` Bug修复
- `refactor:` 重构代码（不改变功能）
- `docs:` 文档更新
- `style:` 代码格式调整
- `test:` 添加或修改测试
- `chore:` 构建工具或辅助工具的变动
- `perf:` 性能优化

#### 提交示例

```bash
# Bug修复示例
git commit -m "fix: 修复位置管理页面的持久化问题"

# 新功能示例
git commit -m "feat: 添加用户权限管理功能"

# 重构示例
git commit -m "refactor: 优化首次使用体验和错误处理"
```

### 完整Git工作流

#### 1. 查看项目状态

```bash
# 查看工作区状态
git status

# 查看未暂存的修改
git diff

# 查看已暂存的修改
git diff --cached

# 查看最近提交
git log --oneline -5
```

#### 2. 提交代码

```bash
# 添加修改的文件
git add frontend/src/pages/LocationManager.tsx

# 或添加所有修改
git add .

# 提交（使用标准格式）
git commit -m "$(cat <<'EOF'
fix: 修复位置管理页面的持久化问题

问题修复：
1. 每次打开页面显示上次保存的位置
2. 首次使用提示只在第一次显示

技术实现：
- 使用localStorage存储用户保存的经纬度坐标
- 使用localStorage记录用户是否已保存过位置

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

#### 3. 推送到GitHub

```bash
# 推送到主分支
git push origin main

# 强制推送（谨慎使用）
git push -f origin main
```

#### 4. 拉取最新代码

```bash
# 拉取并合并
git pull origin main

# 拉取但不合并
git fetch origin
```

### 分支管理（可选）

```bash
# 创建新分支
git checkout -b feature/new-feature

# 切换分支
git checkout main

# 合并分支
git merge feature/new-feature

# 删除分支
git branch -d feature/new-feature
```

### 撤销操作

```bash
# 撤销工作区修改
git restore frontend/src/pages/LocationManager.tsx

# 撤销暂存区
git restore --staged frontend/src/pages/LocationManager.tsx

# 修改最后一次提交
git commit --amend

# 回退到上一次提交（危险操作）
git reset --hard HEAD^
```

---

## 完整工作流示例

### 场景一：修复前端Bug

```bash
# 1. 确保代码最新
git pull origin main

# 2. 本地开发和测试
cd frontend
npm run dev
# [进行修改和测试]

# 3. 构建检查
npm run build

# 4. 提交到Git
git add frontend/src/pages/LocationManager.tsx
git commit -m "fix: 修复地图显示问题"
git push origin main

# 5. 部署到服务器
npm run build
scp -r dist/* root@202.155.123.28:/root/smart-grow/frontend/dist/
ssh root@202.155.123.28 "systemctl restart smartgrow"

# 6. 验证部署
# 访问 http://iot.netr0.com 测试
```

### 场景二：添加后端功能

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 本地开发
cd backend
# [编写代码]

# 3. 本地测试
go test ./...
go run cmd/server/main.go
# [测试API]

# 4. 提交代码
git add backend/
git commit -m "feat: 添加数据导出API"
git push origin main

# 5. 部署到服务器
scp -r * root@202.155.123.28:/root/smart-grow/backend/
ssh root@202.155.123.28 "cd /root/smart-grow/backend && go build -o server cmd/server/main.go && systemctl restart smartgrow"

# 6. 验证
ssh root@202.155.123.28 "systemctl status smartgrow"
```

### 场景三：更新ESP32固件

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 修改固件代码
cd firmware
# [修改 src/main.cpp 或 include/config.h]

# 3. 本地编译测试
pio run

# 4. 提交代码
git add firmware/
git commit -m "feat: 更新ESP32固件适配生产服务器"
git push origin main

# 5. 烧录到设备（如果有硬件）
pio run -t upload
```

---

## 常见问题

### Q1: 部署后页面显示旧内容？

**原因**: 浏览器缓存

**解决**:
- 硬刷新: `Ctrl+F5` (Windows) 或 `Cmd+Shift+R` (Mac)
- 清除浏览器缓存
- 在浏览器开发者工具中禁用缓存

### Q2: 服务启动失败？

```bash
# 查看详细错误日志
ssh root@202.155.123.28 "journalctl -u smartgrow -n 50"

# 常见原因：
# - 端口被占用：使用 lsof -i :8080 查看
# - 配置文件错误：检查 backend/configs/config.yaml
# - 数据库文件权限问题
```

### Q3: Git推送被拒绝？

```bash
# 拉取远程最新代码
git pull origin main --rebase

# 解决冲突后
git push origin main
```

### Q4: SSH连接超时？

```bash
# 检查服务器是否在线
ping 202.155.123.28

# 检查SSH服务
ssh -v root@202.155.123.28

# 如果密钥问题，重新上传
type %USERPROFILE%\.ssh\id_rsa.pub | ssh root@202.155.123.28 "cat >> ~/.ssh/authorized_keys"
```

### Q5: 前端API调用失败？

**检查清单**:
1. 后端服务是否启动: `systemctl status smartgrow`
2. CORS配置是否正确: 查看 `backend/configs/config.yaml` 的 `allowed_origins`
3. API地址是否正确: 前端 `api.ts` 中的 `BASE_URL`
4. 网络请求状态: 浏览器开发者工具 Network 标签

### Q6: 数据库更改后如何迁移？

```bash
# 1. 备份现有数据库
ssh root@202.155.123.28 "cp /root/smart-grow/backend/data/smartgrow.db /root/smart-grow/backend/data/smartgrow.db.backup"

# 2. 上传新的schema.sql
scp backend/configs/schema.sql root@202.155.123.28:/root/smart-grow/backend/configs/

# 3. 重新初始化数据库（会清空数据）
ssh root@202.155.123.28 "cd /root/smart-grow/backend && rm data/smartgrow.db && systemctl restart smartgrow"
```

---

## 快速命令参考

### 本地开发

```bash
# 前端
cd frontend && npm run dev

# 后端
cd backend && go run cmd/server/main.go

# 固件
cd firmware && pio run
```

### 部署

```bash
# 快速部署前端
cd frontend && npm run build && scp -r dist/* root@202.155.123.28:/root/smart-grow/frontend/dist/ && ssh root@202.155.123.28 "systemctl restart smartgrow"

# 快速部署后端
cd backend && scp -r * root@202.155.123.28:/root/smart-grow/backend/ && ssh root@202.155.123.28 "cd /root/smart-grow/backend && go build -o server cmd/server/main.go && systemctl restart smartgrow"
```

### Git操作

```bash
# 快速提交
git add . && git commit -m "feat: 新功能描述" && git push

# 查看状态
git status && git log --oneline -5
```

### 服务器管理

```bash
# 查看状态和日志
ssh root@202.155.123.28 "systemctl status smartgrow && journalctl -u smartgrow -n 20"

# 重启服务
ssh root@202.155.123.28 "systemctl restart smartgrow"
```

---

## 最佳实践

### 开发规范

1. **小步提交**: 每完成一个小功能就提交，不要攒太多修改
2. **清晰的提交信息**: 使用语义化提交，说明"为什么"而不只是"改了什么"
3. **代码审查**: 重要修改前先本地全面测试
4. **配置分离**: 敏感信息使用环境变量，不提交到Git

### 测试规范

1. **本地先测**: 所有改动先在本地测试通过
2. **多场景测试**: 考虑边界情况和错误情况
3. **浏览器兼容**: 测试主流浏览器（Chrome, Firefox, Safari）
4. **移动端适配**: 测试响应式布局

### 部署规范

1. **备份优先**: 重要修改前先备份数据库
2. **分步部署**: 先部署后端，再部署前端
3. **验证部署**: 部署后立即访问网站验证
4. **监控日志**: 部署后观察日志是否有异常

### Git规范

1. **经常拉取**: 开始工作前先 `git pull`
2. **不要强推**: 避免使用 `git push -f`，除非明确知道后果
3. **分支管理**: 大功能使用分支开发，小修改直接在main
4. **忽略文件**: 确保 `.gitignore` 正确配置，不提交敏感文件

---

## 相关文档

- [项目部署文档](DEPLOYMENT.md)
- [GitHub上传指南](GITHUB_UPLOAD.md)
- [ESP32固件文档](../firmware/README.md)
- [项目README](../README.md)

---

**文档版本**: v1.0
**最后更新**: 2025-12-08
**维护者**: SmartGrow Team
