# 📤 GitHub 上传指南

## 快速上传到 GitHub

### 步骤 1: 在 GitHub 创建仓库

1. 登录 [GitHub](https://github.com)
2. 点击右上角 `+` → `New repository`
3. 填写仓库信息：
   - **Repository name**: `smartgrow` 或你喜欢的名称
   - **Description**: SmartGrow 智能灌溉系统
   - **Public/Private**: 根据需求选择
   - ⚠️ **不要**勾选 "Add a README file"
   - ⚠️ **不要**勾选 "Add .gitignore"
   - ⚠️ **不要**选择 License（我们已经准备好了）
4. 点击 `Create repository`

### 步骤 2: 本地初始化并上传

在项目目录 `C:\coding` 打开命令行（已经初始化了Git），然后执行：

```bash
# 如果还没初始化，执行这个（已执行可跳过）
git init

# 添加所有文件到暂存区
git add .

# 提交
git commit -m "Initial commit: SmartGrow智能灌溉系统"

# 添加远程仓库（替换成你的GitHub用户名和仓库名）
git remote add origin https://github.com/你的用户名/smartgrow.git

# 推送到GitHub
git push -u origin main
```

⚠️ **注意替换**: 将 `你的用户名/smartgrow` 替换为你实际的GitHub用户名和仓库名

### 步骤 3: 验证上传

访问你的GitHub仓库页面，应该能看到所有文件已上传成功。

---

## 🔐 如果遇到认证问题

### 方式一：使用 Personal Access Token（推荐）

GitHub不再支持密码认证，需要使用Token：

1. **生成Token**:
   - 访问 https://github.com/settings/tokens
   - 点击 `Generate new token` → `Generate new token (classic)`
   - Note: `SmartGrow Deploy`
   - Expiration: 选择过期时间（建议选择长期）
   - Select scopes: 勾选 `repo`（所有权限）
   - 点击 `Generate token`
   - ⚠️ **复制Token并保存**（只显示一次！）

2. **使用Token**:

```bash
# 方式A: 在push时输入
# 用户名: 你的GitHub用户名
# 密码: 刚才复制的Token（不是GitHub密码）

# 方式B: 配置远程仓库URL包含Token
git remote set-url origin https://YOUR_TOKEN@github.com/你的用户名/smartgrow.git
git push -u origin main
```

### 方式二：使用 SSH Key

```bash
# 1. 生成SSH密钥（如果没有）
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
# 一路回车即可

# 2. 查看公钥
cat ~/.ssh/id_rsa.pub
# 复制输出内容

# 3. 添加到GitHub
# 访问 https://github.com/settings/ssh/new
# Title: My Computer
# Key: 粘贴刚才复制的公钥
# 点击 Add SSH key

# 4. 修改远程仓库URL
git remote set-url origin git@github.com:你的用户名/smartgrow.git

# 5. 推送
git push -u origin main
```

---

## 📝 日常提交流程

以后修改代码后，上传到GitHub的步骤：

```bash
# 1. 查看修改的文件
git status

# 2. 添加所有修改
git add .

# 3. 提交（写清楚修改内容）
git commit -m "添加新功能：XXX"

# 4. 推送到GitHub
git push

# 完成！代码已上传到GitHub
```

---

## 🌿 分支管理（可选）

如果想用分支开发：

```bash
# 创建开发分支
git checkout -b dev

# 在dev分支开发
# ... 修改代码 ...
git add .
git commit -m "开发新功能"
git push origin dev

# 合并到主分支
git checkout main
git merge dev
git push
```

---

## 🚨 常见问题

### 1. `fatal: not a git repository`

```bash
# 确保在项目根目录
cd C:\coding
git init
```

### 2. `remote origin already exists`

```bash
# 删除现有的remote并重新添加
git remote remove origin
git remote add origin https://github.com/你的用户名/smartgrow.git
```

### 3. `failed to push some refs`

```bash
# 先拉取远程代码
git pull origin main --allow-unrelated-histories

# 解决冲突（如果有）后再推送
git push -u origin main
```

### 4. `.gitignore` 没生效

```bash
# 清除Git缓存
git rm -r --cached .
git add .
git commit -m "Update .gitignore"
git push
```

### 5. 误提交了敏感文件

```bash
# 从Git历史中删除（谨慎操作！）
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch backend/configs/config.yaml" \
  --prune-empty --tag-name-filter cat -- --all

# 强制推送
git push origin --force --all
```

---

## ✅ 验证清单

上传到GitHub前，请确认：

- [ ] 敏感信息已从代码中移除（密钥、密码等）
- [ ] `.gitignore` 已正确配置
- [ ] `README.md` 中的GitHub链接已更新为你的实际链接
- [ ] `backend/configs/config.yaml` 没有被提交（被.gitignore排除）
- [ ] 前端 `node_modules/` 没有被提交
- [ ] 后端编译产物 `server` 没有被提交

---

## 🎉 成功上传后

### 更新服务器配置

在服务器上配置Git访问：

```bash
# SSH到服务器
ssh root@202.155.123.28

# 配置Git用户信息
git config --global user.name "Your Name"
git config --global user.email "your-email@example.com"

# 克隆你的仓库
cd /root
rm -rf smart-grow  # 删除旧的
git clone https://github.com/你的用户名/smartgrow.git smart-grow

# 或配置SSH Key（推荐）
ssh-keygen -t rsa -b 4096 -C "server@example.com"
cat ~/.ssh/id_rsa.pub
# 将公钥添加到 GitHub → Settings → SSH Keys

# 使用SSH方式克隆
git clone git@github.com:你的用户名/smartgrow.git smart-grow
```

### 开始使用自动化部署

现在你可以：

1. **本地修改代码** → `git push`
2. **服务器更新** → `./deploy.sh`
3. **自动部署完成** ✨

---

**恭喜！现在您的项目已托管在GitHub上了！** 🎊
