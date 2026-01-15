# Tortoise 开发指南

## 项目概述

Tortoise 是一个超级 AI 代理框架，结合了 OpenClaw 和 Hermes 的最佳特性。

## 技术栈

| 组件 | 语言 | 用途 |
|------|------|------|
| 核心运行时 | Rust | 高性能、安全 |
| 插件系统 | Go | 跨平台、快速开发 |
| 前端 UI | Flutter | 跨平台移动/桌面应用 |
| 嵌入式 SDK | C/C++ | IoT 和资源受限设备 |
| Python SDK | Python | 数据科学和脚本 |

## 项目结构

```
tortoise/
├── src/                    # Rust 核心
│   ├── agent/              # Agent 引擎
│   ├── memory/             # 三层记忆系统
│   ├── channel/           # 消息通道
│   ├── plugin/            # 插件系统
│   ├── skill/             # 技能系统
│   ├── tool/              # 工具系统
│   ├── security/          # 安全系统
│   └── network/           # P2P 网络
├── plugins/                # Go 插件
│   └── channels/           # 通道插件
├── ui/                    # Flutter 应用
│   └── lib/
├── embedded/              # C SDK
├── api/python/            # Python SDK
└── config/                # 配置文件
```

## 开发环境

### 前置要求

- Rust 1.75+
- Go 1.21+
- Flutter 3.16+
- Node.js 20+
- CMake (用于嵌入式)

### 安装

```bash
# 克隆仓库
git clone https://github.com/tortoise-ai/tortoise.git
cd tortoise

# 安装 Rust 依赖
cargo build --release

# 安装 Go 插件依赖
cd plugins/go
go mod download

# 安装 Flutter 依赖
cd ui
flutter pub get
```

## 构建

### Rust Core

```bash
cargo build --release
```

### Go Plugins

```bash
cd plugins/go
go build -o plugins.so ./...
```

### Flutter UI

```bash
cd ui
flutter build apk --release
flutter build ios --release
```

### C SDK

```bash
mkdir build && cd build
cmake ..
make
```

## 测试

### Rust 测试

```bash
cargo test
```

### Flutter 测试

```bash
cd ui
flutter test
```

## 配置

配置文件位于 `config/config.toml`：

```toml
[agent]
model = "gpt-4"
temperature = 0.7

[memory]
short_term_max_items = 100

[channels.discord]
token = "${DISCORD_TOKEN}"
```

## API 参考

### REST API

- `POST /api/chat` - 发送聊天消息
- `GET /api/memory` - 获取记忆
- `POST /api/memory` - 存储记忆
- `GET /api/status` - 获取状态

### WebSocket API

连接到 `ws://localhost:18789/ws` 进行流式响应。

## 贡献指南

1. Fork 仓库
2. 创建特性分支
3. 提交更改
4. 推送分支
5. 创建 Pull Request

## 许可证

AGPL-3.0
