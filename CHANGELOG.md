# Tortoise 变更日志

## [0.3.0] - 2026-05-18

### 新增

#### 核心渠道系统完善
- **Matrix 渠道** - 端到端加密通讯协议
  - Homeserver 连接
  - 房间消息收发
  - E2E 加密支持
  - 富文本格式 (HTML)
  - 正在输入指示器
  
- **Email 渠道** - SMTP/IMAP 邮件集成
  - SMTP 发送邮件
  - IMAP 轮询接收
  - HTML 富文本支持
  - 附件处理
  - 发送者白名单过滤

#### 插件沙箱系统
- 权限隔离 (网络、文件系统、环境变量)
- 沙箱化日志、存储、事件总线
- 网络访问限制 (白名单)
- 资源限制 (内存、CPU、超时)
- 权限预设 (minimal/standard/trusted)

#### 多代理协调器
- Agent 注册与管理
- 任务调度与依赖
- 并行/串行/管道/分层执行
- 智能代理选择
- 失败重试机制
- 任务超时控制

#### 语义记忆系统
- 语义搜索 (嵌入向量相似度)
- 智能遗忘 (基于重要性和访问频率)
- 记忆类型分类
- TTL 过期机制
- 会话作用域
- 自动衰减

#### Tailscale 网络集成
- 设备发现服务
- 服务自动扫描
- ACL 策略管理
- DNS 配置

#### Flutter UI 完善
- 语音唤醒页面 UI (唤醒词、灵敏度、语音活动)
- 插件市场页面 (搜索、分类、详情、安装)
- 渠道配置页面

---

## [0.2.0] - 2026-05-17

### 新增

#### 渠道集成
- **WhatsApp Baileys 协议支持** - 完整的 WhatsApp Web 协议实现，支持原生 WhatsApp 连接
  - 二维码登录认证
  - 会话持久化
  - 文本、图片、视频、音频、文档消息
  - 群组消息支持
  - 表情回应、回复消息
  - 阅后即焚消息
  - 位置和联系人消息
  - Webhook 事件处理
  
- **Signal 安全通讯渠道** - 端到端加密的隐私通讯
  - Signal Protocol 端到端加密实现
  - 预密钥和身份密钥管理
  - 会话建立和管理
  - 文本、媒体消息发送
  - 正在输入指示器
  - 已读回执
  - 用户资料获取

- **Slack 企业级集成增强**
  - Webhook 消息发送
  - 线程消息回复
  - 签名验证

- **Microsoft Teams 集成**
  - Bot Framework 集成
  - 活动事件处理
  - 消息回复

#### Voice Wake 语音唤醒
- 语音唤醒服务 - 支持多种唤醒词
- 自定义唤醒词配置
- 灵敏度调节
- 平台特定实现 (iOS/Android/Desktop)

#### Flutter 桌面端增强
- 桌面窗口管理 (window_manager)
- 系统托盘支持 (system_tray)
- 全局快捷键 (hotkey_manager)
- 跨平台媒体支持
- 安全存储
- 通知系统
- 系统信息获取

#### 无容器部署支持
- **Windows PowerShell 部署脚本** (deploy.ps1)
  - 完整的服务安装/卸载
  - Windows 服务注册
  - 环境变量配置
  - 日志管理
  - 状态检查
  - Docker 支持

- **Linux/macOS Bash 部署脚本** (deploy.sh)
  - Systemd 服务管理
  - 用户权限隔离
  - 安全配置

---

## [0.1.0] - 规划中

### 新增

#### 核心功能
- Rust Core Runtime 基础架构
- Go Gateway Server 框架
- Flutter Desktop/Mobile UI 基础
- Tortoise Protocol 二进制协议定义
- REST API 规范

#### SDK
- TypeScript SDK 基础实现
- Go SDK 基础实现
- Python SDK 基础实现

#### 文档
- 架构设计文档 (ARCHITECTURE.md)
- 协议规范 (PROTOCOL.md)
- API 参考 (API.md)
- 插件开发指南 (PLUGIN_DEVELOPMENT.md)
- 渠道集成文档 (CHANNELS.md)
- 性能优化指南 (PERFORMANCE.md)
- 安全指南 (SECURITY.md)
- 开发指南 (CONTRIBUTING.md)

#### 支持的渠道
- Web
- WebSocket
- Telegram (计划)
- Discord (计划)

#### 支持的模型
- OpenAI (GPT-4, GPT-4o)
- Anthropic (Claude 3.5)
- Google (Gemini)
- Ollama (本地)

---

## OpenClaw 能力覆盖对比

| 功能 | OpenClaw | Tortoise | 状态 |
|------|----------|----------|------|
| 多渠道消息 | ✅ | ✅ | 完成 |
| - Telegram | ✅ | ✅ | 完成 |
| - Discord | ✅ | ✅ | 完成 |
| - WhatsApp | ✅ | ✅ | 完成 |
| - Slack | ✅ | ✅ | 完成 |
| - Signal | ❌ | ✅ | 新增 |
| - Teams | ❌ | ✅ | 新增 |
| 插件系统 | ✅ | ✅ | 进行中 |
| Gateway 管理 | ✅ | ✅ | 进行中 |
| Pi Session | ✅ | ✅ | 进行中 |
| 设备发现 | ✅ | ✅ | 完成 |
| Skills 配置 | ✅ | ✅ | 完成 |
| Voice Wake | ❌ | ✅ | 新增 |
| 无容器部署 | ❌ | ✅ | 新增 |
| 桌面应用 | ⚠️ | ✅ | 完成 |

## Hermes 能力覆盖对比

| 功能 | Hermes | Tortoise | 状态 |
|------|--------|----------|------|
| 智能对话 | ✅ | ✅ | 完成 |
| 多模型路由 | ✅ | ✅ | 完成 |
| 上下文理解 | ✅ | ✅ | 完成 |
| 工具调用 | ✅ | ✅ | 进行中 |
| 记忆系统 | ✅ | ✅ | 进行中 |
| 主动推送 | ✅ | ✅ | 完成 |
| 隐私通讯 | ❌ | ✅ | 新增 |
| 跨平台 UI | ❌ | ✅ | 完成 |
| 端到端加密 | ❌ | ✅ | 新增 |
| 插件市场 | ❌ | ✅ | 计划中 |
