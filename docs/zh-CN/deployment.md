# Tortoise 部署指南

本文档详细介绍 Tortoise 的各种部署方式。

## 目录

- [Docker 部署](#docker-部署)
- [Linux/macOS 部署](#linuxmacos-部署)
- [Windows 部署](#windows-部署)
- [反向代理配置](#反向代理配置)
- [SSL 证书配置](#ssl-证书配置)
- [常见问题](#常见问题)

---

## Docker 部署

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+

### 快速启动

```bash
# 克隆项目
git clone https://github.com/tortoise/tortoise.git
cd tortoise

# 配置环境变量
cp server/.env.example server/.env
nano server/.env  # 编辑填入 API Keys

# 启动服务
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f server
```

### 手动构建

```bash
# 构建镜像
docker build -t tortoise/server ./server

# 运行
docker run -d \
  --name tortoise \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e OPENAI_API_KEY=your-key \
  -e ANTHROPIC_API_KEY=your-key \
  tortoise/server
```

---

## Linux/macOS 部署

### 前置要求

- Go 1.22+ (如需从源码构建)
- systemd (Linux)

### 方法一: 使用安装脚本

```bash
# 1. 构建或下载二进制文件
./build.sh server

# 2. 安装 (需要 root)
sudo ./deploy.sh install

# 3. 编辑配置
sudo nano /etc/tortoise/env

# 4. 启动服务
sudo ./deploy.sh start

# 5. 设置开机自启
sudo systemctl enable tortoise

# 6. 查看状态
sudo ./deploy.sh status

# 7. 查看日志
sudo journalctl -u tortoise -f
```

### 方法二: 手动安装

```bash
# 1. 创建用户和目录
sudo useradd -r -s /bin/false -d /opt/tortoise tortoise
sudo mkdir -p /opt/tortoise /var/lib/tortoise /var/log/tortoise

# 2. 复制二进制和配置
sudo cp tortoise-server /opt/tortoise/
sudo cp config.yaml /opt/tortoise/

# 3. 设置权限
sudo chown -R tortoise:tortoise /opt/tortoise /var/lib/tortoise /var/log/tortoise

# 4. 创建 Systemd 服务
sudo nano /etc/systemd/system/tortoise.service
```

Systemd 服务文件内容:

```ini
[Unit]
Description=Tortoise AI Agent Server
After=network.target

[Service]
Type=simple
User=tortoise
Group=tortoise
WorkingDirectory=/opt/tortoise
ExecStart=/opt/tortoise/tortoise-server
Restart=always
RestartSec=10
StandardOutput=append:/var/log/tortoise/stdout.log
StandardOutput=append:/var/log/tortoise/stderr.log

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/tortoise /var/log/tortoise

[Install]
WantedBy=multi-user.target
```

```bash
# 5. 启用并启动
sudo systemctl daemon-reload
sudo systemctl enable tortoise
sudo systemctl start tortoise

# 6. 检查状态
sudo systemctl status tortoise
```

### 环境变量配置

```bash
# 编辑环境变量文件
sudo nano /etc/tortoise/env

# 内容示例
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx
TELEGRAM_BOT_TOKEN=xxx
DISCORD_BOT_TOKEN=xxx
APP_SECRET_KEY=your-secret-key
```

### 常用命令

```bash
# 启动
sudo systemctl start tortoise

# 停止
sudo systemctl stop tortoise

# 重启
sudo systemctl restart tortoise

# 查看状态
sudo systemctl status tortoise

# 查看日志
sudo journalctl -u tortoise -f

# 卸载
sudo ./deploy.sh uninstall
```

---

## Windows 部署

### 前置要求

- Windows 10/11
- PowerShell 5.1+
- Go 1.22+ (如需从源码构建)

### 方法一: 使用安装脚本

```powershell
# 1. 以管理员身份打开 PowerShell

# 2. 构建或复制二进制文件到项目根目录
# 二进制文件应命名为: tortoise-server.exe

# 3. 运行安装脚本
.\deploy.ps1 install

# 4. 编辑配置
notepad C:\Program Files\Tortoise\.env

# 5. 启动服务
.\deploy.ps1 start

# 6. 查看状态
.\deploy.ps1 status
```

### 方法二: 手动安装

```powershell
# 1. 创建目录
New-Item -ItemType Directory -Force -Path C:\Program Files\Tortoise
New-Item -ItemType Directory -Force -Path "$env:APPDATA\Tortoise"

# 2. 复制文件
Copy-Item tortoise-server.exe C:\Program Files\Tortoise\
Copy-Item config.yaml C:\Program Files\Tortoise\

# 3. 创建服务 (需要管理员)
sc.exe create TortoiseService binPath= "C:\Program Files\Tortoise\tortoise-server.exe" start= auto

# 4. 配置服务
sc.exe config TortoiseService obj= "NT AUTHORITY\LocalService"

# 5. 启动服务
Start-Service TortoiseService

# 6. 检查状态
Get-Service TortoiseService
```

### 直接运行 (无需服务)

```powershell
# 直接运行
cd C:\Program Files\Tortoise
.\tortoise-server.exe

# 或使用环境变量
$env:OPENAI_API_KEY="sk-xxx"
.\tortoise-server.exe
```

### 常用命令

```powershell
# 启动
.\deploy.ps1 start

# 停止
.\deploy.ps1 stop

# 重启
.\deploy.ps1 restart

# 状态
.\deploy.ps1 status

# 卸载
.\deploy.ps1 uninstall
```

---

## 反向代理配置

### Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # WebSocket 支持
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # API 代理
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddy
your-domain.com {
    reverse_proxy /ws/* 127.0.0.1:8080 {
        header_up Upgrade {header.Connection}
        header_up Connection "upgrade"
    }
    
    reverse_proxy /* 127.0.0.1:8080
}
```

---

## SSL 证书配置

### Let's Encrypt (Certbot)

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期测试
sudo certbot renew --dry-run
```

### 自签名证书 (测试用)

```bash
# 生成私钥
openssl genrsa -out key.pem 2048

# 生成证书
openssl req -new -x509 -key key.pem -out cert.pem -days 365

# 转换为 PFX (Windows)
openssl pkcs12 -export -out certificate.pfx -inkey key.pem -in cert.pem
```

---

## 常见问题

### 1. 服务启动失败

```bash
# 检查日志
sudo journalctl -u tortoise -n 50

# 检查端口占用
sudo lsof -i :8080

# 检查配置文件
cat /etc/tortoise/env
```

### 2. 权限问题

```bash
# 修复权限
sudo chown -R tortoise:tortoise /var/lib/tortoise /var/log/tortoise
sudo chmod 755 /opt/tortoise/tortoise-server
```

### 3. 内存不足

```bash
# 检查内存
free -h

# 限制内存 (Systemd)
# 编辑服务文件添加:
MemoryMax=512M
```

### 4. 持久化存储

```bash
# 确保数据目录存在
sudo mkdir -p /var/lib/tortoise
sudo chown tortoise:tortoise /var/lib/tortoise
```

### 5. 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

---

## 下一步

- [配置 AI 提供商](/docs/configuration.md)
- [设置消息渠道](/docs/channels.md)
- [安装插件](/docs/plugins.md)
