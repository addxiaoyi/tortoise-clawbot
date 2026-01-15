@echo off
chcp 65001 > nul
echo ==========================================
echo   Tortoise API Server 构建脚本
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

echo [1/4] 清理旧构建...
if exist "cmd\api\tortoise-api.exe" del "cmd\api\tortoise-api.exe"
if exist "dist" rd /s /q "dist"

echo.
echo [2/4] 下载依赖...
go mod tidy

echo.
echo [3/4] 构建 API 服务器...
go build -o cmd\api\tortoise-api.exe ./cmd/api

if %errorlevel% neq 0 (
    echo.
    echo [错误] 构建失败！
    pause
    exit /b 1
)

echo.
echo [4/4] 创建输出目录...
mkdir dist 2>nul
copy "cmd\api\tortoise-api.exe" "dist\" >nul

echo.
echo ==========================================
echo   构建完成！
echo ==========================================
echo.
echo 可执行文件位置:
echo   %cd%\cmd\api\tortoise-api.exe
echo   %cd%\dist\tortoise-api.exe
echo.
echo 启动命令:
echo   tortoise-api.exe
echo.
echo 默认监听端口: 18792
echo.
pause
