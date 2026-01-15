# Tortoise 快速启动

> 2周冲刺目标：让代码真正可用

## 🎯 核心目标

1. **渠道系统** - Telegram/Discord 真正可用
2. **插件沙箱** - 安全隔离执行
3. **Skills** - Calculator 真正可用

---

## 🚀 快速启动

### 1. Server (Go)

```bash
cd server

# 安装依赖
go mod tidy

# 配置 (创建 config.yaml)
cat > config.yaml << EOF
app:
  address: ":8080"
  secret_key: "your-secret-key"
  
ai:
  providers:
    - name: openai
      base_url: https://api.openai.com/v1
      api_key: ${OPENAI_API_KEY}
      models:
        - gpt-4
        - gpt-3.5-turbo

channels:
  telegram:
    bot_token: ${TELEGRAM_BOT_TOKEN}
    allowed_chats: []

plugins:
  sandbox:
    max_memory_mb: 256
    max_duration_seconds: 30
    allow_network: false
EOF

# 运行
go run main.go
```

### 2. Flutter App

```bash
cd flutter

# 安装依赖
flutter pub get

# 运行
flutter run
```

---

## ✅ 已完成并可用的功能

### 渠道系统

| 渠道 | 状态 | 配置 |
|------|------|------|
| Telegram | ✅ 可用 | `TELEGRAM_BOT_TOKEN` |
| Discord | ✅ 可用 | `DISCORD_BOT_TOKEN` |

### Skills

| Skill | 状态 | 用法 |
|-------|------|------|
| Calculator | ✅ 可用 | `{"expression": "2+2*3"}` |
| Web Search | ⚠️ 待测 | DuckDuckGo API |

### 插件沙箱

| 功能 | 状态 |
|------|------|
| JS 执行 | ✅ 可用 |
| Python 执行 | ⚠️ 待测 |
| 资源限制 | ✅ 可用 |

---

## 🧪 测试清单

### 1. Telegram 测试

```bash
# 设置环境变量
export TELEGRAM_BOT_TOKEN="your-bot-token"

# 启动服务器
cd server && go run main.go

# 测试 Bot
# 1. 在 Telegram 搜索你的 Bot
# 2. 发送 /start
# 3. 发送任意消息
```

### 2. Calculator 测试

```bash
# API 测试
curl -X POST http://localhost:8080/api/v1/skills/calculator \
  -H "Content-Type: application/json" \
  -d '{"expression": "2+2*3"}'
```

### 3. 插件沙箱测试

```bash
# JS 执行测试
curl -X POST http://localhost:8080/api/v1/plugins/sandbox/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "javascript", "code": "console.log(1+1)"}'
```

---

## 📋 2周冲刺待办

### Week 1: 核心功能可用

- [ ] Telegram Bot 真正响应
- [ ] Discord Bot 真正响应
- [ ] Calculator Skill 真正工作
- [ ] 插件沙箱安全执行

### Week 2: 完善和测试

- [ ] Web Search Skill 集成
- [ ] Memory System 集成
- [ ] Flutter UI 完整
- [ ] 端到端测试

---

## 🔧 常见问题

### 编译错误

```bash
# 清理并重新安装依赖
cd server
go clean -modcache
go mod tidy
```

### 连接 Telegram 超时

检查网络和 `TELEGRAM_BOT_TOKEN` 是否正确

### AI 服务无响应

确保 `OPENAI_API_KEY` 环境变量已设置

---

## 📁 关键文件

```
server/
├── main.go              # 入口
├── config/             # 配置
├── services/ai/        # AI 服务
├── services/channel/   # 渠道服务
├── services/plugin/    # 插件服务
└── internal/
    ├── channel/        # 渠道实现
    ├── skill/          # Skills
    └── plugin/         # 插件系统
        └── sandbox.go  # 沙箱隔离
```

---

## 下一步

1. 设置 Telegram Bot Token
2. 运行 `go run main.go`
3. 测试 `/start` 命令
4. 测试计算器 `2+2`
