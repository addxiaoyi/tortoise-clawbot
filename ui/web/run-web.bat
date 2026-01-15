@echo off
chcp 65001 > nul
echo ==========================================
echo   Tortoise Web UI 启动器
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

echo [1/3] 检查依赖安装...
if not exist "node_modules" (
    echo 正在安装依赖...
    call npm install
)

echo.
echo [2/3] 启动开发服务器...
echo 访问地址: http://localhost:3000
echo.

npm run dev

pause
