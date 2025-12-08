@echo off
REM ==========================================
REM SmartGrow 一键部署脚本（本地使用）
REM 用途：从本地一键部署到服务器
REM ==========================================
chcp 65001 >nul
setlocal enabledelayedexpansion

REM ========== 配置区域 ==========
set SERVER=202.155.123.28
set SERVER_USER=root
set SERVER_DIR=/root/smart-grow
set SERVER_PORT=22

REM ========== 显示帮助 ==========
if "%1"=="" goto :show_help
if "%1"=="help" goto :show_help
if "%1"=="-h" goto :show_help
if "%1"=="--help" goto :show_help

goto :parse_command

:show_help
echo.
echo ==========================================
echo SmartGrow 一键部署工具
echo ==========================================
echo.
echo 用法:
echo   deploy.bat [命令]
echo.
echo 命令:
echo   all         - 部署前端和后端（默认）
echo   frontend    - 只部署前端
echo   backend     - 只部署后端
echo   status      - 查看服务器状态
echo   logs        - 查看服务器日志
echo   restart     - 重启服务器服务
echo   help        - 显示此帮助信息
echo.
echo 示例:
echo   deploy.bat all          部署全部
echo   deploy.bat frontend     只部署前端
echo   deploy.bat status       查看状态
echo.
goto :eof

REM ========== 解析命令 ==========
:parse_command
set COMMAND=%1

echo.
echo ==========================================
echo SmartGrow 部署脚本
echo ==========================================
echo 目标服务器: %SERVER%
echo 部署目录: %SERVER_DIR%
echo 命令: %COMMAND%
echo ==========================================
echo.

if "%COMMAND%"=="all" goto :deploy_all
if "%COMMAND%"=="frontend" goto :deploy_frontend
if "%COMMAND%"=="backend" goto :deploy_backend
if "%COMMAND%"=="status" goto :check_status
if "%COMMAND%"=="logs" goto :show_logs
if "%COMMAND%"=="restart" goto :restart_service

echo [错误] 未知命令: %COMMAND%
echo 使用 'deploy.bat help' 查看帮助
exit /b 1

REM ========== 部署全部 ==========
:deploy_all
echo [1/2] 部署前端...
call :deploy_frontend_internal
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 前端部署失败
    exit /b 1
)

echo.
echo [2/2] 部署后端...
call :deploy_backend_internal
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 后端部署失败
    exit /b 1
)

echo.
echo ==========================================
echo [成功] 全部部署完成！
echo ==========================================
goto :eof

REM ========== 部署前端 ==========
:deploy_frontend
call :deploy_frontend_internal
if %ERRORLEVEL% NEQ 0 exit /b 1
echo.
echo ==========================================
echo [成功] 前端部署完成！
echo ==========================================
goto :eof

:deploy_frontend_internal
echo [前端] 检查 Node.js 环境...
where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 Node.js
    exit /b 1
)

echo [前端] 安装依赖...
cd frontend
call npm install --silent
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 依赖安装失败
    cd ..
    exit /b 1
)

echo [前端] 构建生产版本...
call npm run build
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 构建失败
    cd ..
    exit /b 1
)
cd ..

echo [前端] 上传到服务器...
scp -r frontend/dist %SERVER_USER%@%SERVER%:%SERVER_DIR%/frontend/
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 上传失败
    exit /b 1
)

echo [前端] 重启服务...
ssh %SERVER_USER%@%SERVER% "systemctl restart smartgrow"
if %ERRORLEVEL% NEQ 0 (
    echo [警告] 重启失败，请手动重启
    exit /b 1
)

timeout /t 3 /nobreak >nul
echo [前端] 健康检查...
ssh %SERVER_USER%@%SERVER% "curl -s http://localhost:8080/health || echo 'Health check failed'"

exit /b 0

REM ========== 部署后端 ==========
:deploy_backend
call :deploy_backend_internal
if %ERRORLEVEL% NEQ 0 exit /b 1
echo.
echo ==========================================
echo [成功] 后端部署完成！
echo ==========================================
goto :eof

:deploy_backend_internal
echo [后端] 上传源码到服务器...
scp -r backend/cmd backend/internal backend/go.mod backend/go.sum %SERVER_USER%@%SERVER%:%SERVER_DIR%/backend/
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 上传失败
    exit /b 1
)

echo [后端] 在服务器上编译...
ssh %SERVER_USER%@%SERVER% "cd %SERVER_DIR%/backend && CGO_ENABLED=1 go build -o ../server -ldflags '-s -w' ./cmd/server"
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 编译失败
    exit /b 1
)

echo [后端] 重启服务...
ssh %SERVER_USER%@%SERVER% "systemctl restart smartgrow"
if %ERRORLEVEL% NEQ 0 (
    echo [警告] 重启失败，请手动重启
    exit /b 1
)

timeout /t 3 /nobreak >nul
echo [后端] 健康检查...
ssh %SERVER_USER%@%SERVER% "curl -s http://localhost:8080/health || echo 'Health check failed'"

exit /b 0

REM ========== 查看状态 ==========
:check_status
echo [状态] 查询服务器状态...
ssh %SERVER_USER%@%SERVER% "systemctl status smartgrow --no-pager"
goto :eof

REM ========== 查看日志 ==========
:show_logs
echo [日志] 最近50行日志...
ssh %SERVER_USER%@%SERVER% "tail -50 %SERVER_DIR%/logs/server.log"
goto :eof

REM ========== 重启服务 ==========
:restart_service
echo [重启] 重启服务器服务...
ssh %SERVER_USER%@%SERVER% "systemctl restart smartgrow && sleep 2 && systemctl status smartgrow --no-pager"
echo.
echo [重启] 健康检查...
ssh %SERVER_USER%@%SERVER% "curl -s http://localhost:8080/health"
goto :eof
