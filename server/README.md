# Tortoise Server

> Tortoise 后端服务器 - Go 语言实现的高性能 AI 代理后端

## 功能特性

- 🚀 **高性能**: 基于 Go + Gin 框架
- 🤖 **多 AI 提供商**: OpenAI, Anthropic Claude 等
- 💬 **多消息渠道**: Telegram, Discord 等
- 🔌 **插件系统**: 支持动态加载插件
- 💾 **数据持久化**: SQLite 数据库
- 🔄 **WebSocket**: 实时通信支持
- 📊 **RESTful API**: 完整的 API 接口

## 快速开始

### 1. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的 API Keys:

```env
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx
TELEGRAM_BOT_TOKEN=xxx
DISCORD_BOT_TOKEN=xxx
```

### 2. 使用 Docker 运行

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f server
```

### 3. 本地运行

```bash
# 安装依赖
go mod download

# 运行
go run main.go
```

## API 端点

### 健康检查
```
GET /health
```

### 会话管理
```
GET    /api/v1/sessions           # 列出会话
POST   /api/v1/sessions           # 创建会话
GET    /api/v1/sessions/:id       # 获取会话
DELETE /api/v1/sessions/:id       # 删除会话
GET    /api/v1/sessions/:id/messages  # 获取消息
```

### AI 聊天
```
POST /api/v1/chat/completions         # 聊天补全
POST /api/v1/chat/completions/stream  # 流式聊天
```

### 渠道管理
```
GET    /api/v1/channels        # 列出渠道
POST   /api/v1/channels        # 创建渠道
GET    /api/v1/channels/:id    # 获取渠道
PUT    /api/v1/channels/:id    # 更新渠道
DELETE /api/v1/channels/:id    # 删除渠道
POST   /api/v1/channels/:id/connect    # 连接
POST   /api/v1/channels/:id/disconnect  # 断开
```

### 记忆管理
```
GET    /api/v1/memory         # 列出记忆
POST   /api/v1/memory         # 创建记忆
GET    /api/v1/memory/:id     # 获取记忆
PUT    /api/v1/memory/:id     # 更新记忆
DELETE /api/v1/memory/:id     # 删除记忆
GET    /api/v1/memory/search  # 搜索记忆
```

### 插件管理
```
GET    /api/v1/plugins              # 列出插件
POST   /api/v1/plugins/install      # 安装插件
POST   /api/v1/plugins/:id/enable   # 启用插件
POST   /api/v1/plugins/:id/disable  # 禁用插件
DELETE /api/v1/plugins/:id          # 卸载插件
```

### WebSocket
```
GET /ws  # WebSocket 连接
```

## 认证

API 使用 Bearer Token 认证:

```bash
curl -H "Authorization: Bearer your-secret-key" http://localhost:8080/api/v1/sessions
```

## 配置

配置文件位于 `config.yaml`:

```yaml
app:
  address: ":8080"
  secret_key: "your-secret-key"
  debug: false

database:
  type: "sqlite"
  path: "./data/tortoise.db"

ai:
  providers:
    - name: "openai"
      base_url: "https://api.openai.com/v1"
      api_key: "${OPENAI_API_KEY}"
      models:
        - "gpt-4-turbo"
        - "gpt-4"

channels:
  channels:
    - type: "telegram"
      token: "${TELEGRAM_BOT_TOKEN}"
```

## 开发

```bash
# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 构建
go build -o tortoise-server .
```

## 许可证

Apache 2.0
