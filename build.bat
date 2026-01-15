@echo off
setlocal enabledelayedexpansion

:: Tortoise Windows 构建脚本
:: 使用方法: build.bat [server|flutter|all|clean]

set "SCRIPT_DIR=%~dp0"
set "BUILD_DIR=%SCRIPT_DIR%build"

:: 创建构建目录
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

echo [INFO] Tortoise 构建脚本
echo.

:: 解析参数
set "COMMAND=%~1"
if "%COMMAND%"=="" set "COMMAND=all"

goto :%COMMAND%

:server
    call :build_server
    goto :end

:flutter
    call :build_flutter
    goto :end

:all
    call :build_server
    call :build_flutter
    call :create_packages
    goto :end

:clean
    echo [INFO] 清理构建目录...
    if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
    mkdir "%BUILD_DIR%"
    echo [INFO] 清理完成
    goto :end

:help
    echo Tortoise 构建脚本
    echo.
    echo 用法: build.bat [选项]
    echo.
    echo 选项:
    echo   server   只构建 Go 服务器
    echo   flutter  只构建 Flutter 应用
    echo   all      构建所有 ^(默认^)
    echo   clean    清理构建目录
    echo   help     显示帮助
    goto :end

:: ========== 构建 Go 服务器 ==========
:build_server
    echo [INFO] 构建 Go 服务器...

    if not exist "%SCRIPT_DIR%server" (
        echo [WARN] server 目录不存在，跳过
        goto :eof
    )

    cd /d "%SCRIPT_DIR%server"

    :: 检查 Go
    where go >nul 2>&1
    if errorlevel 1 (
        echo [WARN] Go 未安装，跳过服务器构建
        cd /d "%SCRIPT_DIR%"
        goto :eof
    )

    :: 构建 Windows 版本
    echo [INFO] 构建 Windows amd64...
    go build -o "%BUILD_DIR%\tortoise-server.exe" .

    if errorlevel 1 (
        echo [ERROR] 构建失败
        cd /d "%SCRIPT_DIR%"
        goto :eof
    )

    echo [INFO] Go 服务器构建完成!
    cd /d "%SCRIPT_DIR%"
    goto :eof

:: ========== 构建 Flutter ==========
:build_flutter
    echo [INFO] 构建 Flutter 应用...

    if not exist "%SCRIPT_DIR%flutter" (
        echo [WARN] flutter 目录不存在，跳过
        goto :eof
    )

    :: 检查 Flutter
    where flutter >nul 2>&1
    if errorlevel 1 (
        echo [WARN] Flutter 未安装，跳过
        goto :eof
    )

    cd /d "%SCRIPT_DIR%flutter"

    :: Web
    echo [INFO] 构建 Web...
    flutter build web --release -o "%BUILD_DIR%\web"

    :: Windows
    echo [INFO] 构建 Windows...
    flutter build windows --release -o "%BUILD_DIR%\windows"

    echo [INFO] Flutter 构建完成!
    cd /d "%SCRIPT_DIR%"
    goto :eof

:: ========== 创建发布包 ==========
:create_packages
    echo [INFO] 创建发布包...

    cd /d "%BUILD_DIR%"

    :: Windows ZIP
    if exist "tortoise-server.exe" (
        powershell -Command "Compress-Archive -Path 'tortoise-server.exe' -DestinationPath 'tortoise-server-windows-amd64.zip' -Force"
        echo [INFO] 创建: tortoise-server-windows-amd64.zip
    )

    :: Web ZIP
    if exist "web" (
        powershell -Command "Compress-Archive -Path 'web' -DestinationPath 'tortoise-web.zip' -Force"
        echo [INFO] 创建: tortoise-web.zip
    )

    cd /d "%SCRIPT_DIR%"
    goto :eof

:: ========== 显示结果 ==========
:show_results
    echo.
    echo [INFO] 构建结果:
    echo.
    dir /b "%BUILD_DIR%"
    echo.
    goto :eof

:end
    if "%COMMAND%"=="all" call :show_results
    echo.
    echo [INFO] 完成
    endlocal
