@echo off
chcp 65001 > nul
echo ==========================================
echo    Tortoise AI Agent 一键启动器
echo ==========================================
echo.

cd /d "%~dp0"

REM 检查 Node.js
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Node.js，请先安装 Node.js 18+
    pause
    exit /b 1
)

REM 检查 Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go 1.21+
    pause
    exit /b 1
)

echo [1/5] 准备启动前后端服务...
echo.

REM 创建日志目录
if not exist "logs" mkdir logs

REM 启动后端 API 服务
echo [2/5] 启动 API 服务器 (端口 18792)...
cd server
start "Tortoise-API" cmd /c "go run ./cmd/api > ..\logs\api.log 2>&1"
cd ..

REM 等待后端启动
echo 等待后端启动...
timeout /t 3 /nobreak > nul

REM 启动前端 Web UI
echo [3/5] 启动 Web UI (端口 3000)...
cd ui\web
start "Tortoise-Web" cmd /c "npm run dev > ..\..\logs\web.log 2>&1"
cd ..\..

echo.
echo ==========================================
echo    启动完成！
echo ==========================================
echo.
echo 服务地址:
echo   - Web UI:    http://localhost:3000
echo   - API Server: http://localhost:18792
echo.
echo 日志文件:
echo   - API 日志: logs\api.log
echo   - Web 日志: logs\web.log
echo.
echo 提示: 关闭此窗口不会停止服务
echo       如需停止，请关闭对应的命令窗口
echo.
pause
