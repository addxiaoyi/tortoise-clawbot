# 🐢 Tortoise Flutter App

全平台 AI Agent 应用 - Flutter 实现

## 功能特性

### AI 引擎
- **OpenAI** - GPT-4, GPT-3.5-turbo
- **Anthropic Claude** - Claude 3 Opus, Sonnet, Haiku
- **Ollama** - 本地开源模型支持
- 多模型路由和负载均衡

### 消息渠道
- **Telegram** - Bot API 集成
- **Discord** - Gateway 集成
- **Slack** - WebSocket 支持
- **WebSocket** - 自定义渠道

### 平台支持
| 平台 | 状态 |
|------|------|
| Android | ✅ |
| iOS | ✅ |
| Windows | ✅ |
| macOS | ✅ |
| Linux | ✅ |
| Web | ✅ |

## 快速开始

### 环境要求
- Flutter SDK 3.2+

### 安装运行
```bash
cd flutter
flutter pub get
flutter run
```

### 构建发布
```bash
flutter build apk --release          # Android
flutter build ios --release          # iOS
flutter build windows --release      # Windows
flutter build web --release          # Web
```

## 项目结构

```
flutter/
├── lib/
│   ├── core/
│   │   ├── ai/               # AI 引擎
│   │   ├── channel/          # 渠道管理
│   │   ├── storage/          # 数据持久化
│   │   ├── discovery/        # 设备发现
│   │   ├── session/          # 会话管理
│   │   ├── config/           # 配置管理
│   │   ├── plugin/           # 插件系统
│   │   ├── analytics/        # 统计分析
│   │   ├── notification/     # 本地通知
│   │   ├── platform/         # 平台检测
│   │   ├── theme/            # 主题配置
│   │   ├── di/               # 依赖注入
│   │   └── errors/           # 异常处理
│   └── app/
│       ├── pages/            # 页面
│       ├── providers/        # 状态管理
│       └── widgets/          # 组件
├── test/
└── pubspec.yaml
```

## 许可证

Apache 2.0
