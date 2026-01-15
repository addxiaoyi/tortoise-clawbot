# Tortoise 部署脚本 (Windows PowerShell)
# 支持无容器部署和 Docker 部署

param(
    [Parameter(Position=0)]
    [ValidateSet("install", "start", "stop", "restart", "status", "uninstall", "help", "docker-install", "docker-start", "docker-stop", "docker-restart")]
    [string]$Command = "help",
    
    [string]$InstallPath = "$env:ProgramFiles\Tortoise",
    [string]$DataPath = "$env:LOCALAPPDATA\Tortoise\Data",
    [string]$ConfigPath = "$env:ProgramData\Tortoise\Config",
    [string]$LogPath = "$env:LOCALAPPDATA\Tortoise\Logs"
)

$ErrorActionPreference = "Stop"
$APP_NAME = "Tortoise"
$BINARY_NAME = "tortoise-server.exe"

# 颜色输出
function Write-Info { param($Message) Write-Host "[INFO] $Message" -ForegroundColor Green }
function Write-Warn { param($Message) Write-Host "[WARN] $Message" -ForegroundColor Yellow }
function Write-Err { param($Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }
function Write-Success { param($Message) Write-Host "[OK] $Message" -ForegroundColor Cyan }

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 创建目录
function New-Directories {
    $dirs = @($InstallPath, $DataPath, $ConfigPath, $LogPath)
    foreach ($dir in $dirs) {
        if (-not (Test-Path $dir)) {
            Write-Info "创建目录: $dir"
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
}

# 下载或复制二进制文件
function Install-Binary {
    Write-Info "安装二进制文件..."
    
    # 检查当前目录下是否有二进制文件
    $localBinary = Join-Path $PSScriptRoot $BINARY_NAME
    
    if (Test-Path $localBinary) {
        Write-Info "使用本地二进制文件: $localBinary"
        Copy-Item $localBinary -Destination (Join-Path $InstallPath $BINARY_NAME) -Force
    }
    else {
        Write-Warn "未找到本地二进制文件"
        Write-Info "请下载 $BINARY_NAME 并放置在当前目录下"
        Write-Info "下载地址: https://github.com/tortoise-ai/tortoise/releases"
        return $false
    }
    
    return $true
}

# 创建配置文件
function New-ConfigFile {
    Write-Info "创建配置文件..."
    
    $envFile = Join-Path $ConfigPath "env"
    
    if (-not (Test-Path $envFile)) {
        @"
# Tortoise 环境变量配置
# ================================

# AI API Keys
OPENAI_API_KEY=your-openai-api-key
ANTHROPIC_API_KEY=your-anthropic-api-key
GOOGLE_API_KEY=your-google-api-key

# 渠道配置
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
DISCORD_BOT_TOKEN=your-discord-bot-token
SLACK_BOT_TOKEN=your-slack-bot-token
SLACK_SIGNING_SECRET=your-slack-signing-secret

# WhatsApp 配置 (WhatsApp Business API)
WHATSAPP_API_URL=https://graph.facebook.com/v18.0
WHATSAPP_PHONE_NUMBER_ID=your-phone-number-id
WHATSAPP_ACCESS_TOKEN=your-whatsapp-access-token

# Signal 配置
SIGNAL_SERVER_HOST=your-signal-server
SIGNAL_SERVER_PORT=8080
SIGNAL_ACCOUNT_ID=your-account-id
SIGNAL_PASSWORD=your-signal-password

# 应用配置
APP_SECRET_KEY=change-this-secret-key-in-production
APP_PORT=8080
APP_HOST=0.0.0.0

# 数据库配置
DATABASE_PATH=$DataPath\data.db

# 日志配置
LOG_LEVEL=info
LOG_PATH=$LogPath

# TLS/SSL 配置 (可选)
# ENABLE_TLS=true
# TLS_CERT=path/to/cert.pem
# TLS_KEY=path/to/key.pem

# ================================
# 其他配置项请参考官方文档
"@ | Out-File -FilePath $envFile -Encoding UTF8
        Write-Success "配置文件已创建: $envFile"
    }
    else {
        Write-Info "配置文件已存在，跳过创建"
    }
}

# 创建 Windows 服务
function New-WindowsService {
    Write-Info "创建 Windows 服务..."
    
    $serviceName = $APP_NAME
    $serviceDisplayName = "Tortoise AI Agent Server"
    $serviceDescription = "Tortoise - 下一代 AI 代理框架"
    
    # 检查服务是否已存在
    $existingService = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    
    if ($existingService) {
        Write-Warn "服务 '$serviceName' 已存在"
        return
    }
    
    # 创建 nssm 配置脚本
    $nssmConfig = @"
# nssm 安装参数
set service-name $serviceName
set service-display-name "$serviceDisplayName"
set service-description "$serviceDescription"
set application "$InstallPath\$BINARY_NAME"
set app-parameters "--config $ConfigPath\env"
set app-environment EXTRA `"PATH=$env:PATH`"
set app-directory "$InstallPath"
set AppStdout "$LogPath\stdout.log"
set AppStderr "$LogPath\stderr.log"
set AppRotateFiles 1
set AppRotateBytes 10485760
set AppRotateSeconds 86400
set AppStopMethodConsole 6000
set AppRestartMethod 1 6000
set StartMode auto
"@
    
    # 检查是否安装了 nssm
    $nssmPath = $null
    $possiblePaths = @(
        "C:\Program Files\nssm\nssm.exe",
        "C:\Program Files (x86)\nssm\nssm.exe",
        "$env:ProgramFiles\nssm\nssm.exe"
    )
    
    foreach ($path in $possiblePaths) {
        if (Test-Path $path) {
            $nssmPath = $path
            break
        }
    }
    
    if (-not $nssmPath) {
        Write-Warn "未找到 nssm，请先安装 nssm"
        Write-Info "下载: https://nssm.cc/download"
        Write-Info "安装后运行: nssm install $serviceName"
        return
    }
    
    Write-Info "使用 nssm: $nssmPath"
    
    # 使用 sc 命令创建服务 (基础版本)
    $scPath = "sc.exe"
    
    # 创建服务
    & $scPath create $serviceName binPath= `"$InstallPath\$BINARY_NAME --config `"$ConfigPath\env`""` DisplayName= `"$serviceDisplayName`" start= auto
    DisplayError= 1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "服务 '$serviceName' 已创建"
        
        # 设置描述
        & $scPath description $serviceName `"$serviceDescription`"
        
        # 设置重启策略
        & $scPath failure $serviceName reset= 86400 actions= restart/60000/restart/120000/restart/300000
    }
    else {
        Write-Err "服务创建失败"
    }
}

# 安装 (无容器)
function Install-Tortoise {
    if (-not (Test-Administrator)) {
        Write-Err "请以管理员权限运行此脚本"
        exit 1
    }
    
    Write-Info "开始安装 Tortoise..."
    
    New-Directories
    
    if (-not (Install-Binary)) {
        return
    }
    
    New-ConfigFile
    
    # 尝试创建 Windows 服务
    New-WindowsService
    
    Write-Success "安装完成!"
    Write-Host ""
    Write-Host "配置说明:"
    Write-Host "  1. 编辑配置文件: $ConfigPath\env"
    Write-Host "  2. 配置你的 API Keys"
    Write-Host ""
    Write-Host "启动服务:"
    Write-Host "  - 自动: 重新启动计算机或使用 'sc start $APP_NAME'"
    Write-Host "  - 手动: .\deploy.ps1 start"
    Write-Host ""
    Write-Host "查看日志:"
    Write-Host "  - 事件查看器 -> Windows 日志 -> 应用程序"
    Write-Host "  - $LogPath"
}

# 启动服务
function Start-Service {
    if (-not (Test-Administrator)) {
        Write-Err "请以管理员权限运行此脚本"
        exit 1
    }
    
    Write-Info "启动 $APP_NAME..."
    
    $service = Get-Service -Name $APP_NAME -ErrorAction SilentlyContinue
    
    if ($service) {
        Start-Service -Name $APP_NAME
        Write-Success "服务已启动"
    }
    else {
        # 直接运行
        Write-Info "服务未安装，直接启动进程..."
        
        $envFile = Join-Path $ConfigPath "env"
        if (Test-Path $envFile) {
            # 加载环境变量
            Get-Content $envFile | ForEach-Object {
                if ($_ -match '^([^=]+)=(.*)$') {
                    [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
                }
            }
        }
        
        $binaryPath = Join-Path $InstallPath $BINARY_NAME
        if (Test-Path $binaryPath) {
            Start-Process -FilePath $binaryPath -WorkingDirectory $InstallPath -PassThru
            Write-Success "进程已启动"
        }
        else {
            Write-Err "未找到二进制文件: $binaryPath"
        }
    }
}

# 停止服务
function Stop-Service {
    if (-not (Test-Administrator)) {
        Write-Err "请以管理员权限运行此脚本"
        exit 1
    }
    
    Write-Info "停止 $APP_NAME..."
    
    $service = Get-Service -Name $APP_NAME -ErrorAction SilentlyContinue
    
    if ($service) {
        Stop-Service -Name $APP_NAME -Force
        Write-Success "服务已停止"
    }
    else {
        # 终止进程
        $processes = Get-Process -Name "tortoise-server" -ErrorAction SilentlyContinue
        if ($processes) {
            $processes | Stop-Process -Force
            Write-Success "进程已终止"
        }
        else {
            Write-Warn "服务未运行"
        }
    }
}

# 重启服务
function Restart-Service {
    if (-not (Test-Administrator)) {
        Write-Err "请以管理员权限运行此脚本"
        exit 1
    }
    
    Write-Info "重启 $APP_NAME..."
    
    $service = Get-Service -Name $APP_NAME -ErrorAction SilentlyContinue
    
    if ($service) {
        Restart-Service -Name $APP_NAME
        Write-Success "服务已重启"
    }
    else {
        Stop-Service
        Start-Sleep -Seconds 2
        Start-Service
    }
}

# 查看状态
function Get-Status {
    Write-Host ""
    Write-Host "=== $APP_NAME 状态 ===" -ForegroundColor Cyan
    Write-Host ""
    
    # 服务状态
    $service = Get-Service -Name $APP_NAME -ErrorAction SilentlyContinue
    
    if ($service) {
        Write-Host "服务状态: $($service.Status)"
        Write-Host "启动类型: $($service.StartType)"
    }
    else {
        Write-Host "服务状态: 未安装" -ForegroundColor Yellow
    }
    
    Write-Host ""
    
    # 进程状态
    $processes = Get-Process -Name "tortoise-server" -ErrorAction SilentlyContinue
    
    if ($processes) {
        Write-Host "运行中的进程:"
        foreach ($proc in $processes) {
            Write-Host "  PID: $($proc.Id), CPU: $($proc.CPU)s, 内存: $([math]::Round($proc.WorkingSet64/1MB, 2)) MB"
        }
    }
    else {
        Write-Host "进程状态: 未运行"
    }
    
    Write-Host ""
    
    # 配置文件
    $envFile = Join-Path $ConfigPath "env"
    if (Test-Path $envFile) {
        Write-Host "配置文件: $envFile [存在]"
    }
    else {
        Write-Host "配置文件: $envFile [不存在]" -ForegroundColor Yellow
    }
    
    # 数据目录
    if (Test-Path $DataPath) {
        Write-Host "数据目录: $DataPath [存在]"
    }
    else {
        Write-Host "数据目录: $DataPath [不存在]" -ForegroundColor Yellow
    }
    
    Write-Host ""
    
    # 健康检查
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            Write-Host "健康检查: OK" -ForegroundColor Green
            $content = $response.Content | ConvertFrom-Json
            if ($content.PSObject.Properties.Name -contains "version") {
                Write-Host "版本: $($content.version)"
            }
        }
        else {
            Write-Host "健康检查: HTTP $($response.StatusCode)" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "健康检查: 无法连接 (服务可能未启动)" -ForegroundColor Yellow
    }
    
    Write-Host ""
}

# 卸载
function Uninstall-Tortoise {
    if (-not (Test-Administrator)) {
        Write-Err "请以管理员权限运行此脚本"
        exit 1
    }
    
    Write-Warn "即将卸载 Tortoise..."
    
    # 停止服务
    Stop-Service -ErrorAction SilentlyContinue
    
    # 删除服务
    $service = Get-Service -Name $APP_NAME -ErrorAction SilentlyContinue
    if ($service) {
        & sc.exe delete $APP_NAME
        Write-Info "服务已删除"
    }
    
    # 删除文件和目录
    if (Test-Path $InstallPath) {
        Remove-Item -Path $InstallPath -Recurse -Force
        Write-Info "安装目录已删除"
    }
    
    Write-Warn "注意: 数据和配置文件保留在以下位置:"
    Write-Host "  - $DataPath"
    Write-Host "  - $ConfigPath"
    Write-Host "  - $LogPath"
    
    Write-Success "卸载完成"
}

# Docker 部署
function Docker-Install {
    Write-Info "安装 Docker 容器版本..."
    
    # 检查 Docker 是否安装
    try {
        & docker --version | Out-Null
    }
    catch {
        Write-Err "Docker 未安装"
        Write-Info "请先安装 Docker Desktop: https://docker.com/products/docker-desktop"
        return
    }
    
    Write-Info "Docker 已安装"
}

function Docker-Start {
    Write-Info "启动 Docker 容器..."
    
    # 检查 docker-compose.yml
    $composeFile = Join-Path $PSScriptRoot "docker-compose.yaml"
    
    if (-not (Test-Path $composeFile)) {
        Write-Err "未找到 docker-compose.yaml"
        return
    }
    
    # 启动容器
    docker-compose -f $composeFile up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "容器已启动"
        Write-Info "查看日志: docker-compose logs -f"
        Write-Info "停止容器: .\deploy.ps1 docker-stop"
    }
    else {
        Write-Err "容器启动失败"
    }
}

function Docker-Stop {
    Write-Info "停止 Docker 容器..."
    
    $composeFile = Join-Path $PSScriptRoot "docker-compose.yaml"
    
    if (Test-Path $composeFile) {
        docker-compose -f $composeFile down
        Write-Success "容器已停止"
    }
}

function Docker-Restart {
    Docker-Stop
    Start-Sleep -Seconds 2
    Docker-Start
}

# 显示帮助
function Show-Help {
    Write-Host @"

Tortoise 部署脚本
==================

用法: .\deploy.ps1 [命令]

命令:
  install        安装 Tortoise (无容器)
  start          启动服务
  stop           停止服务
  restart        重启服务
  status         查看状态
  uninstall      卸载
  
  docker-install 安装 Docker 版本
  docker-start   启动 Docker 容器
  docker-stop    停止 Docker 容器
  docker-restart 重启 Docker 容器
  
  help           显示帮助

示例:
  # 无容器安装
  .\deploy.ps1 install
  .\deploy.ps1 start
  .\deploy.ps1 status
  
  # Docker 安装
  .\deploy.ps1 docker-start
  .\deploy.ps1 docker-stop

要求:
  - 安装需要管理员权限
  - 无容器部署需要 Windows Server 2016+ 或 Windows 10+
  - Docker 部署需要 Docker Desktop

"@
}

# 主逻辑
switch ($Command) {
    "install" { Install-Tortoise }
    "start" { Start-Service }
    "stop" { Stop-Service }
    "restart" { Restart-Service }
    "status" { Get-Status }
    "uninstall" { Uninstall-Tortoise }
    "docker-install" { Docker-Install }
    "docker-start" { Docker-Start }
    "docker-stop" { Docker-Stop }
    "docker-restart" { Docker-Restart }
    "help" { Show-Help }
    default { Show-Help }
}
