# Tortoise 部署指南

## 目录

- [Docker 部署](#docker-部署)
- [无容器部署](#无容器部署-linuxmacos)
- [Windows 部署](#windows-部署)
- [反向代理配置](#反向代理配置)
- [环境变量](#环境变量)
- [安全配置](#安全配置)

---

## Docker 部署

### 快速部署

```bash
# 克隆项目
git clone https://github.com/tortoise/tortoise.git
cd tortoise

# 配置环境变量
cp server/.env.example server/.env
nano server/.env

# 启动服务
docker-compose up -d
```

### docker-compose.yml 配置

```yaml
version: '3.8'

services:
  tortoise:
    image: tortoise/server:latest
    container_name: tortoise-server
    ports:
      - "8080:8080"
    environment:
      - APP_SECRET_KEY=${APP_SECRET_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - tortoise-data:/app/data
      - tortoise-plugins:/app/plugins
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  tortoise-data:
  tortoise-plugins:

networks:
  default:
    name: tortoise-network
```

### 使用 Docker Run

```bash
docker run -d \
  --name tortoise \
  -p 8080:8080 \
  -e APP_SECRET_KEY=your-secret-key \
  -e OPENAI_API_KEY=sk-xxx \
  -v tortoise-data:/app/data \
  -v tortoise-plugins:/app/plugins \
  tortoise/server:latest
```

---

## 无容器部署 (Linux/macOS)

### 方式一: 下载二进制

```bash
# 下载最新版本
curl -L https://github.com/tortoise/tortoise/releases/latest/download/tortoise-server-linux-amd64 -o tortoise-server
chmod +x tortoise-server

# 运行
./tortoise-server
```

### 方式二: 从源码构建

```bash
# 安装 Go 1.22+
go version

# 克隆并构建
git clone https://github.com/tortoise/tortoise.git
cd tortoise/server
go mod download
go build -o tortoise-server ./cmd/api

# 运行
./tortoise-server
```

### 安装为系统服务 (Linux)

```bash
# 创建服务用户
sudo useradd -r -s /bin/false tortoise

# 创建目录
sudo mkdir -p /etc/tortoise
sudo mkdir -p /var/lib/tortoise
sudo chown -R tortoise:tortoise /var/lib/tortoise

# 复制二进制
sudo cp tortoise-server /usr/local/bin/tortoise-server

# 创建服务文件
sudo nano /etc/systemd/system/tortoise.service
```

```ini
[Unit]
Description=Tortoise AI Agent Framework
After=network.target

[Service]
Type=simple
User=tortoise
Group=tortoise
WorkingDirectory=/var/lib/tortoise
ExecStart=/usr/local/bin/tortoise-server
Restart=always
RestartSec=5
Environment=APP_SECRET_KEY=your-secret-key

[Install]
WantedBy=multi-user.target
```

```bash
# 启用并启动
sudo systemctl daemon-reload
sudo systemctl enable tortoise
sudo systemctl start tortoise

# 检查状态
sudo systemctl status tortoise
```

### 安装为 LaunchAgent (macOS)

```bash
# 创建目录
mkdir -p ~/Library/LaunchAgents

# 创建 plist
nano ~/Library/LaunchAgents/ai.tortoise.tortoise.plist
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>ai.tortoise.tortoise</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/tortoise-server</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>EnvironmentVariables</key>
    <dict>
        <key>APP_SECRET_KEY</key>
        <string>your-secret-key</string>
    </dict>
    <key>WorkingDirectory</key>
    <string>/var/lib/tortoise</string>
</dict>
</plist>
```

```bash
# 加载并启动
launchctl load ~/Library/LaunchAgents/ai.tortoise.tortoise.plist
```

---

## Windows 部署

### 使用 PowerShell 脚本

```powershell
# 下载最新版本
Invoke-WebRequest -Uri "https://github.com/tortoise/tortoise/releases/latest/download/tortoise-server-windows-amd64.exe" -OutFile "tortoise-server.exe"

# 创建配置目录
New-Item -ItemType Directory -Force -Path "C:\Program Files\Tortoise"
Move-Item tortoise-server.exe "C:\Program Files\Tortoise\"

# 创建配置文件
Copy-Item .env.example "$env:USERPROFILE\.tortoise\.env"

# 使用 nssm 安装为服务 (推荐)
nssm install Tortoise "C:\Program Files\Tortoise\tortoise-server.exe"
nssm set Tortoise AppDirectory "C:\Program Files\Tortoise"
nssm set Tortoise ObjectName "NT AUTHORITY\LocalService"
nssm set Tortoise Start SERVICE_AUTO_START

# 或直接运行
& "C:\Program Files\Tortoise\tortoise-server.exe"
```

---

## 反向代理配置

### Nginx

```nginx
server {
    listen 80;
    server_name tortoise.example.com;

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name tortoise.example.com;

    ssl_certificate /etc/ssl/certs/tortoise.crt;
    ssl_certificate_key /etc/ssl/private/tortoise.key;

    # WebSocket 支持
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddy
tortoise.example.com {
    reverse_proxy /ws* {
        to localhost:8080
        transport http {
            versions h2c
        }
    }

    reverse_proxy /* localhost:8080
}
```

---

## 环境变量

### 必需变量

```bash
APP_SECRET_KEY=your-secret-key-min-32-chars
```

### AI 配置

```bash
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx
```

### 数据库

```bash
DATABASE_TYPE=sqlite  # 或 postgresql
DATABASE_PATH=./data/tortoise.db
```

### 消息渠道

```bash
TELEGRAM_BOT_TOKEN=xxx
DISCORD_BOT_TOKEN=xxx
```

---

## 安全配置

### 防火墙设置

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp     # HTTP
sudo ufw allow 443/tcp    # HTTPS
sudo ufw enable

# firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### SSL/TLS 证书

使用 Let's Encrypt:

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d tortoise.example.com
```

### 安全检查清单

- [ ] 使用强 `APP_SECRET_KEY` (至少32字符)
- [ ] 启用 HTTPS
- [ ] 配置防火墙
- [ ] 定期更新 Docker 镜像或二进制
- [ ] 启用日志记录
- [ ] 配置速率限制
- [ ] 使用非 root 用户运行

---

## 故障排除

### 检查日志

```bash
# Docker
docker logs tortoise-server

# Systemd
journalctl -u tortoise -f

# 直接运行
./tortoise-server --verbose
```

### 常见问题

**服务无法启动**
```bash
# 检查端口占用
lsof -i :8080

# 检查配置
./tortoise-server --validate-config
```

**WebSocket 连接失败**
```bash
# 检查反向代理配置
# 确保启用了 WebSocket 支持
```

**数据库连接失败**
```bash
# 检查数据库文件权限
chmod 755 data/
chmod 644 data/tortoise.db
```
