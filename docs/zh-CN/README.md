# Tortoise 文档

> Tortoise - 下一代 AI 代理框架

## 📚 文档目录

### 入门指南
- [快速开始](/quick-start) - 5 分钟快速上手
- [安装指南](/installation) - 详细安装步骤
- [配置说明](/configuration) - 配置文件详解

### 核心功能
- [会话管理](/sessions) - 会话创建、聊天、上下文
- [Skills 系统](/skills) - 内置 Skills 和自定义
- [记忆系统](/memory) - 长期记忆和遗忘机制
- [多代理系统](/agents) - Orchestrator 和工作流

### 渠道集成
- [Telegram](/channels/telegram) - Telegram Bot 集成
- [Discord](/channels/discord) - Discord 机器人
- [WhatsApp](/channels/whatsapp) - WhatsApp 消息
- [Slack](/channels/slack) - Slack 工作区
- [Signal](/channels/signal) - Signal 端到端加密
- [Microsoft Teams](/channels/teams) - Teams 集成
- [iMessage](/channels/imessage) - BlueBubbles
- [Matrix](/channels/matrix) - Matrix 协议
- [Email](/channels/email) - SMTP/IMAP

### 企业功能
- [LDAP 认证](/enterprise/ldap) - 企业目录集成
- [SAML SSO](/enterprise/saml) - 单点登录
- [OAuth/OIDC](/enterprise/oauth) - 第三方认证

### 插件系统
- [插件开发](/plugins/development) - 开发自己的插件
- [插件市场](/plugins/marketplace) - 安装和管理插件

### SDK 和 API
- [REST API](/api) - API 参考
- [Python SDK](/sdk/python) - Python 开发包
- [Rust SDK](/sdk/rust) - Rust 开发包
- [Go SDK](/sdk/go) - Go 开发包

### 部署
- [Docker 部署](/deployment/docker) - Docker 容器部署
- [无容器部署](/deployment/native) - 直接安装部署
- [集群部署](/deployment/cluster) - Gateway 集群

### 网络
- [Tailscale 集成](/network/tailscale) - 零信任网络
- [Gateway 集群](/network/cluster) - 分布式部署

## 🚀 快速开始

```bash
# 1. 安装
curl -fsSL https://tortoise.ai/install.sh | bash

# 2. 启动
tortoise start

# 3. 配置渠道
tortoise channel add telegram --token YOUR_TOKEN

# 4. 开始聊天
tortoise chat
```

## 📊 功能总览

| 功能 | 描述 |
|------|------|
| **9 大消息渠道** | Telegram, Discord, WhatsApp, Slack, Signal, Teams, iMessage, Matrix, Email |
| **8 个内置 Skills** | 搜索、计算、代码、文件、日历、单位转换、日期时间、文本处理 |
| **多代理系统** | Orchestrator + Specialist + Worker |
| **记忆系统** | 语义记忆 + 情景记忆 + 程序性记忆 + 遗忘机制 |
| **企业认证** | LDAP, SAML, OAuth/OIDC, MFA |
| **插件市场** | 社区插件生态 |
| **Gateway 集群** | Raft 领导者选举，分布式部署 |
| **Tailscale** | 零信任网络集成 |

## 🛠️ 技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go, Gin, WebSocket |
| **前端** | Flutter |
| **存储** | SQLite, Vector DB |
| **加密** | Signal Protocol, Olm/Megolm |
| **网络** | Tailscale, Raft |

## 📦 项目结构

```
tortoise/
├── server/           # Go 后端服务
│   ├── api/          # REST API
│   ├── handlers/     # 请求处理器
│   ├── services/     # 业务服务
│   └── internal/    # 内部模块
├── flutter/         # Flutter 跨平台应用
│   └── lib/         # Dart 代码
├── sdk/             # 多语言 SDK
│   ├── python/      # Python SDK
│   ├── rust/        # Rust SDK
│   └── go/          # Go SDK
└── docs/            # 文档
```

## 🔗 链接

- 🌐 官网: https://tortoise.ai
- 📚 文档: https://docs.tortoise.ai
- 💬 Discord: https://discord.gg/tortoise
- 🐙 GitHub: https://github.com/tortoise-ai/tortoise
