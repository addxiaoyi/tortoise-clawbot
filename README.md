# 🐢 Tortoise - 下一代 AI 代理框架

<p align="center">
  <img src="docs/images/tortoise-logo.png" alt="Tortoise Logo" width="200"/>
</p>

<p align="center">
  <a href="https://github.com/tortoise/tortoise/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"/>
  </a>
  <a href="https://github.com/tortoise/tortoise/actions">
    <img src="https://github.com/tortoise/tortoise/workflows/CI/badge.svg" alt="CI"/>
  </a>
</p>

---

## 📖 项目概述

Tortoise 是一个高性能、可扩展的 AI 代理框架，支持：

- 🤖 **多 AI 模型**: OpenAI GPT-4, Claude 3, Gemini 等
- 💬 **多消息渠道**: Telegram, Discord, Slack 等
- 🧠 **智能记忆**: 语义搜索，长期记忆
- 🔌 **插件系统**: 无限扩展功能
- 🌐 **跨平台**: Web, Desktop, Mobile

## 🏗️ 项目结构

```
tortoise/
├── flutter/           # Flutter 跨平台应用
│   ├── lib/
│   │   ├── core/     # 核心模块
│   │   └── features/ # 功能模块
│   └── web/          # Web 入口
├── server/           # Go 后端服务
│   ├── config/       # 配置
│   ├── database/     # 数据库
│   ├── handlers/     # 处理器
│   ├── middleware/   # 中间件
│   ├── services/     # 服务
│   └── websocket/    # WebSocket
└── docker-compose.yml # Docker 部署
```

## 🚀 快速开始

### 方式一: Docker 部署

```bash
# 克隆项目
git clone https://github.com/tortoise/tortoise.git
cd tortoise

# 配置环境变量
cp server/.env.example server/.env
# 编辑 .env 填入你的 API Keys

# 启动服务
docker-compose up -d
```

### 方式二: 无容器部署 (Linux/macOS)

```bash
# 1. 下载或构建二进制
./build.sh server

# 2. 安装 (需要 root)
sudo ./deploy.sh install

# 3. 编辑配置
sudo nano /etc/tortoise/env

# 4. 启动服务
sudo ./deploy.sh start

# 开机自启
sudo systemctl enable tortoise
```

### 方式三: 无容器部署 (Windows)

```powershell
# 1. 构建或下载 tortoise-server.exe

# 2. 以管理员运行 PowerShell
.\deploy.ps1 install

# 3. 编辑配置
notepad C:\Program Files\Tortoise\.env

# 4. 启动服务
.\deploy.ps1 start
```

### 方式四: 本地开发

**后端:**
```bash
cd server
go mod download
go run main.go
```

**前端 (Flutter):**
```bash
cd flutter
flutter pub get
flutter run
```

## ✨ 功能特性

### 🤖 AI 服务
- **OpenAI**: GPT-4, GPT-3.5-Turbo
- **Anthropic**: Claude 3 Opus, Sonnet, Haiku
- **Google**: Gemini Pro
- **本地模型**: Ollama 支持

### 💬 消息渠道
| 渠道 | 状态 | 说明 |
|------|------|------|
| Telegram | ✅ | Bot API |
| Discord | ✅ | Gateway |
| Slack | 🚧 | 开发中 |
| WhatsApp | 🚧 | 开发中 |
| Email | 🚧 | 开发中 |

### 🧠 记忆系统
- 语义搜索
- 长期记忆
- 上下文理解
- 自动摘要

### 🔌 插件系统
- 动态加载
- 热更新
- 插件市场

## 📡 API 文档

### 基础信息
- **Base URL**: `http://localhost:8080`
- **认证**: Bearer Token
- **格式**: JSON

### 主要端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /health | 健康检查 |
| GET | /api/v1/sessions | 会话列表 |
| POST | /api/v1/chat/completions | AI 聊天 |
| GET | /api/v1/channels | 渠道列表 |
| GET | /api/v1/memory | 记忆列表 |
| GET | /ws | WebSocket |

### 示例请求

```bash
# 创建会话
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"title": "新会话", "ai_provider": "openai", "model": "gpt-4"}'

# AI 聊天
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

## 🛠️ 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Flutter | 跨平台 UI |
| 后端 | Go + Gin | 高性能 API |
| 数据库 | SQLite | 本地持久化 |
| 实时 | WebSocket | 双向通信 |
| 容器 | Docker | 一键部署 |

## 📦 部署

### Docker Compose

```yaml
services:
  server:
    image: tortoise/server:latest
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    volumes:
      - ./data:/app/data
```

### 环境变量

```env
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx
TELEGRAM_BOT_TOKEN=xxx
DISCORD_BOT_TOKEN=xxx
```

## 📚 文档

- [快速开始](/docs/getting-started.md)
- [API 文档](/docs/api/)
- [插件开发](/docs/plugins/)
- [部署指南](/docs/deployment/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

Apache 2.0 - 详见 [LICENSE](LICENSE)

---

<p align="center">
  Copyright © 2026
</p>
