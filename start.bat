@echo off
chcp 65001 >nul
title Tortoise - AI Agent Framework

echo.
echo  ╔═══════════════════════════════════════════════╗
echo  ║       🐢 Tortoise AI Agent Framework         ║
echo  ╚═══════════════════════════════════════════════╝
echo.

:: 检查 Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] Go 未安装，无法启动后端服务
    set "HAS_GO=0"
) else (
    echo [OK] Go 已安装
    set "HAS_GO=1"
)

:: 检查 Flutter
where flutter >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] Flutter 未安装，无法启动前端
    set "HAS_FLUTTER=0"
) else (
    echo [OK] Flutter 已安装
    set "HAS_FLUTTER=1"
)

echo.
echo 请选择操作:
echo.
echo   [1] 启动全部服务 (后端 + 前端)
echo   [2] 只启动后端服务
echo   [3] 只启动前端 (Flutter)
echo   [4] 启动 Docker
echo   [5] 构建项目
echo   [6] 查看帮助
echo   [0] 退出
echo.

set /p choice="请输入选项 (0-6): "

if "%choice%"=="1" goto :all
if "%choice%"=="2" goto :backend
if "%choice%"=="3" goto :frontend
if "%choice%"=="4" goto :docker
if "%choice%"=="5" goto :build
if "%choice%"=="6" goto :help
if "%choice%"=="0" goto :exit

:all
    echo.
    echo [INFO] 启动全部服务...
    
    :: 启动后端
    if "%HAS_GO%"=="1" (
        start "Tortoise Backend" cmd /k "cd /d %~dp0server && go run main.go"
        echo [OK] 后端服务启动中 (端口 8080)
    )
    
    :: 等待后端启动
    timeout /t 3 /nobreak >nul
    
    :: 启动前端
    if "%HAS_FLUTTER%"=="1" (
        start cmd /k "cd /d %~dp0flutter && flutter run"
        echo [OK] 前端启动中...
    )
    
    goto :end

:backend
    if "%HAS_GO%"=="0" (
        echo [ERROR] Go 未安装，无法启动后端
        goto :end
    )
    
    echo.
    echo [INFO] 启动后端服务...
    cd /d %~dp0server
    go run main.go
    
    goto :end

:frontend
    if "%HAS_FLUTTER%"=="0" (
        echo [ERROR] Flutter 未安装，无法启动前端
        goto :end
    )
    
    echo.
    echo [INFO] 启动前端...
    cd /d %~dp0flutter
    flutter run
    
    goto :end

:docker
    echo.
    echo [INFO] 启动 Docker 服务...
    docker-compose up -d
    docker-compose logs -f
    
    goto :end

:build
    echo.
    echo [INFO] 构建项目...
    call build.bat
    goto :end

:help
    echo.
    echo ════════════════════════════════════════════
    echo                    帮助信息
    echo ════════════════════════════════════════════
    echo.
    echo 前置要求:
    echo   - Go 1.22+ (用于后端)
    echo   - Flutter SDK (用于前端)
    echo   - Docker (可选)
    echo.
    echo 环境变量配置:
    echo   复制 server/.env.example 为 server/.env
    echo   并填入你的 API Keys
    echo.
    echo 详细文档请查看:
    echo   - docs/deployment.md (部署指南)
    echo   - docs/getting-started.md (快速开始)
    echo.
    echo ════════════════════════════════════════════
    goto :end

:exit
    exit /b 0

:end
    echo.
    pause
