@echo off
chcp 65001 > nul
echo ==========================================
echo   Tortoise API Server 启动器
echo ==========================================
echo.

cd /d "%~dp0"

REM 检查 Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go 1.21+
    pause
    exit /b 1
)

echo [1/4] 下载依赖...
go mod tidy

echo.
echo [2/4] 编译中...
go build -o tortoise-api.exe ./cmd/api

if %errorlevel% neq 0 (
    echo.
    echo [错误] 编译失败！
    pause
    exit /b 1
)

echo.
echo [3/4] 启动服务器...
echo.
echo 访问地址:
echo   HTTP API:  http://localhost:18792/api/v1
echo   WebSocket: ws://localhost:18792/ws
echo.
echo 按 Ctrl+C 停止服务器
echo.

.\tortoise-api.exe

pause
