@echo off
chcp 65001 > nul
echo ==========================================
echo    Tortoise AI Agent 停止脚本
echo ==========================================
echo.

REM 停止 API 服务器
echo [1/2] 停止 API 服务器...
taskkill /f /im tortoise-api.exe >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq Tortoise-API*" >nul 2>&1

REM 停止前端
echo [2/2] 停止 Web UI...
taskkill /f /im node.exe >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq Tortoise-Web*" >nul 2>&1

echo.
echo ==========================================
echo    所有服务已停止
echo ==========================================
echo.
pause
