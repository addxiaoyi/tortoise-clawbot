@echo off
chcp 65001 > nul
echo ==========================================
echo    Tortoise Core 核心引擎启动器
echo ==========================================
echo.

cd /d "%~dp0"

REM 检查 Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go 1.22+
    pause
    exit /b 1
)

echo [1/4] 检查环境...
go version

echo.
echo [2/4] 下载依赖...
go mod tidy

echo.
echo [3/4] 编译中...
go build -o tortoise-core.exe ./cmd/tortoise

if %errorlevel% neq 0 (
    echo.
    echo [错误] 编译失败！
    pause
    exit /b 1
)

echo.
echo [4/4] 启动核心引擎...
echo.
echo ═══════════════════════════════════════════════════
echo.
echo    启动 Tortoise Core 高性能 AI Agent 框架
echo.
echo    功能:
echo      - Runtime Engine     ^(^高性能运行时^)
echo      - Memory System      ^(^三层记忆系统^)
echo      - AI Engine         ^(^多模型路由^)
echo      - Plugin Host        ^(^插件系统^)
echo      - Channel Manager    ^(^消息渠道^)
echo      - MCP Protocol       ^(^协议支持^)
echo      - Discovery          ^(^设备发现^)
echo      - WebSocket          ^(^实时通信^)
echo      - Gateway           ^(^网关服务^)
echo.
echo ═══════════════════════════════════════════════════
echo.

.\tortoise-core.exe

pause
