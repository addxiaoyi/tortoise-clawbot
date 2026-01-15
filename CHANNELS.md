# Tortoise 渠道集成

## 概述

Tortoise 支持多种消息渠道，可以同时连接多个平台。

## 支持的渠道

| 渠道 | 状态 | 描述 |
|------|------|------|
| Web | ✅ 稳定 | Web 界面 |
| WebSocket | ✅ 稳定 | WebSocket 连接 |
| Telegram | ✅ 稳定 | Telegram Bot |
| Discord | ✅ 稳定 | Discord Bot |
| Slack | 🧪 Beta | Slack App |
| WhatsApp | 🧪 Beta | WhatsApp Business |
| Matrix | 🔜 计划中 | Matrix Protocol |
| Signal | 🔜 计划中 | Signal Messenger |

## 配置

### Telegram

```yaml
channels:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    api_id: "${TELEGRAM_API_ID}"
    api_hash: "${TELEGRAM_API_HASH}"
    # 可选: 限制用户
    allowed_users:
      - 123456789
    # 可选: 群组白名单
    allowed_groups:
      - -1001234567890
```

### Discord

```yaml
channels:
  discord:
    enabled: true
    bot_token: "${DISCORD_BOT_TOKEN}"
    # 可选: 服务器白名单
    allowed_guilds:
      - 123456789012345678
    # 可选: 频道白名单  
    allowed_channels:
      - 123456789012345678
```

### Slack

```yaml
channels:
  slack:
    enabled: true
    app_token: "${SLACK_APP_TOKEN}"
    bot_token: "${SLACK_BOT_TOKEN}"
    signing_secret: "${SLACK_SIGNING_SECRET}"
```

## 自定义渠道

### 创建渠道处理器

```go
package channels

import (
    "context"
    
    "github.com/tortoise/server/channel"
)

type MyChannel struct {
    // 渠道配置
    config *Config
}

func (c *MyChannel) Name() string {
    return "my-channel"
}

func (c *MyChannel) Connect(ctx context.Context) error {
    // 连接逻辑
    return nil
}

func (c *MyChannel) Disconnect() error {
    // 断开连接
    return nil
}

func (c *MyChannel) Send(ctx context.Context, msg *channel.Message) error {
    // 发送消息
    return nil
}

func (c *MyChannel) Events() <-chan *channel.Event {
    // 返回事件通道
    return c.eventCh
}
```

### 注册渠道

```go
import "github.com/tortoise/server/channel"

channel.Register("my-channel", func(cfg *Config) (channel.Handler, error) {
    return NewMyChannel(cfg)
})
```

## 消息格式

### 统一消息结构

```go
type Message struct {
    ID          string                 // 消息 ID
    Channel     string                // 渠道名称
    ChannelID   string                // 渠道消息 ID
    UserID      string                // 用户 ID
    UserName    string                // 用户名
    Content     string                // 消息内容
    Type        MessageType           // 消息类型
    Attachments []Attachment          // 附件
    Metadata    map[string]interface{} // 元数据
    Timestamp   time.Time             // 时间戳
}

type MessageType string

const (
    TextMessage    MessageType = "text"
    ImageMessage   MessageType = "image"
    AudioMessage   MessageType = "audio"
    VideoMessage   MessageType = "video"
    FileMessage   MessageType = "file"
)
```

## 事件

### 事件类型

```go
const (
    EventConnect    = "channel.connect"
    EventDisconnect = "channel.disconnect"
    EventMessage    = "channel.message"
    EventTyping     = "channel.typing"
    EventReaction   = "channel.reaction"
)
```

### 处理事件

```go
handler := func(event *channel.Event) {
    switch event.Type {
    case channel.EventMessage:
        msg := event.Data.(*channel.Message)
        log.Info().Str("content", msg.Content).Msg("Received message")
    case channel.EventTyping:
        user := event.Data.(string)
        log.Info().Str("user", user).Msg("User is typing")
    }
}
```
