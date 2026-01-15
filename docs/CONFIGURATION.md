# 配置指南

本文档介绍如何配置 Tortoise AI 框架以接入真实服务。

## 快速配置

### 1. 创建环境变量文件

```bash
cp .env.example .env
```

### 2. 配置 API Keys

编辑 `.env` 文件，填入你的 API Keys：

```env
# AI 模型 - 必须至少配置一个
OPENAI_API_KEY=sk-your-openai-api-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key

# 消息渠道 - 根据需要配置
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
DISCORD_BOT_TOKEN=your-discord-bot-token
```

### 3. 通过前端面板配置

访问 `http://localhost:18792` 打开管理面板：

1. **AI 配置** - 添加/编辑 API Keys，选择模型
2. **消息渠道** - 配置 Bot Tokens，测试连接
3. **数据库** - 选择存储方式
4. **安全设置** - 配置认证和限流

---

## 详细配置

### AI 模型配置

#### OpenAI

1. 访问 [OpenAI Platform](https://platform.openai.com/api-keys)
2. 创建新的 API Key
3. 填入配置：

```env
OPENAI_API_KEY=sk-...
```

支持模型：
- `gpt-4` - 最强模型
- `gpt-4-turbo` - 更快更便宜
- `gpt-3.5-turbo` - 性价比最高

#### Anthropic Claude

1. 访问 [Anthropic Console](https://console.anthropic.com/settings/keys)
2. 创建 API Key
3. 填入配置：

```env
ANTHROPIC_API_KEY=sk-ant-...
```

支持模型：
- `claude-3-opus-20240229` - 最强模型
- `claude-3-sonnet-20240229` - 平衡之选
- `claude-3-haiku-20240307` - 快速响应

#### Ollama (本地模型)

1. 安装 [Ollama](https://ollama.ai/)
2. 下载模型：`ollama pull llama2`
3. 配置：

```env
OLLAMA_BASE_URL=http://localhost:11434
```

---

### 消息渠道配置

#### Telegram

1. 在 Telegram 搜索 @BotFather
2. 发送 `/newbot` 创建机器人
3. 获取 Bot Token
4. 配置：

```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
# 可选：限制允许的 Chat ID
TELEGRAM_ALLOWED_CHATS=123456789,987654321
```

#### Discord

1. 访问 [Discord Developer Portal](https://discord.com/developers/applications)
2. 创建 Application
3. 创建 Bot，获取 Token
4. 启用 **Message Content Intent**：
   - Bot → Privileged Gateway Intents → 勾选 MESSAGE CONTENT INTENT
5. 配置：

```env
DISCORD_BOT_TOKEN=MTEwMTExMTExMTEx.exampl3.tOkEnHeRe
# 可选：指定服务器
DISCORD_GUILD_ID=123456789012345678
```

邀请 Bot 到服务器：
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=67456128&scope=bot
```

#### Slack

1. 访问 [Slack API](https://api.slack.com/apps)
2. 创建 App
3. 获取 Bot Token (xoxb-...)
4. 配置：

```env
SLACK_BOT_TOKEN=xoxb-your-slack-token
SLACK_SIGNING_SECRET=your-signing-secret
```

---

### 数据库配置

#### 内存存储（开发用）

```env
DB_TYPE=memory
```

#### SQLite（生产用）

```env
DB_TYPE=sqlite
SQLITE_PATH=./data/tortoise.db
```

#### Redis（生产用）

```env
DB_TYPE=redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

### 安全配置

#### API Key 认证

```env
REQUIRE_API_KEY=true
```

#### JWT Secret

```env
JWT_SECRET=your-super-secret-key-at-least-32-characters
```

#### 限流

```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPM=60
```

---

### 高级配置

```env
# 服务器
PORT=18792
HOST=0.0.0.0

# 会话
MAX_SESSIONS=100000
SESSION_TIMEOUT=86400

# 缓冲区
MESSAGE_BUFFER_SIZE=10000
WORKER_POOL_SIZE=100

# 监控
ENABLE_METRICS=true
ENABLE_TRACING=false
```

---

## 前端面板使用

### 访问面板

```
http://localhost:18792
```

### AI 配置页面

1. 切换到 **AI 配置** 标签
2. 启用需要的 AI 提供商
3. 填入 API Key
4. 选择模型和路由策略
5. 点击 **保存所有设置**

### 渠道配置页面

1. 切换到 **消息渠道** 标签
2. 启用需要的渠道（Telegram/Discord/Slack）
3. 填入 Bot Token
4. 点击 **保存所有设置**
5. 使用 **连接测试** 验证配置

### 安全设置页面

1. 切换到 **安全设置** 标签
2. 启用 API Key 认证（可选）
3. 管理 API Keys
4. 配置 JWT Secret
5. 设置限流规则

---

## 环境变量优先级

配置优先级（从高到低）：
1. 运行时配置（前端面板保存的配置文件）
2. 环境变量（`.env` 文件）
3. 默认值

这意味着：
- 可以先在 `.env` 设置基础配置
- 然后通过前端面板细调
- 前端面板的配置会覆盖环境变量

---

## 故障排查

### AI 模型连接失败

1. 检查 API Key 是否正确
2. 确认 API Key 有余额
3. 检查网络是否能访问 API

### 渠道连接失败

**Telegram:**
- 确认 Bot Token 正确
- 确认 Bot 已激活（与 Bot 开始对话）

**Discord:**
- 确认 Bot Token 正确
- 确认已启用 MESSAGE CONTENT INTENT
- 确认 Bot 已在目标服务器中

### 数据库连接失败

**Redis:**
- 确认 Redis 服务正在运行
- 确认 URL 格式正确：`redis://host:port`
- 检查防火墙设置

---

## 下一步

- 查看 [API 文档](API.md) 了解 REST API
- 查看 [开发指南](DEVELOPMENT.md) 开始开发
- 查看 [部署文档](DEPLOYMENT.md) 了解生产部署
