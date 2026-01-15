# Tortoise Embedded C SDK

一个轻量级的 C SDK，用于嵌入式设备和物联网集成。

## 特性

- 轻量级设计，最小依赖
- 跨平台支持 (Linux, Windows, macOS, RTOS)
- 简洁的 C API
- TLS/SSL 支持
- 异步操作
- 低内存占用

## 支持的平台

- Linux (x86, ARM, RISC-V)
- Windows (x86, x64)
- macOS
- FreeRTOS
- Zephyr
- Arduino

## 快速开始

### 构建

```bash
mkdir build && cd build
cmake ..
make
```

### 使用

```c
#include "tortoise.h"

void log_callback(const char* level, const char* message) {
    printf("[%s] %s\n", level, message);
}

int main() {
    tortoise_config_t config = {
        .gateway_url = "http://localhost:18789",
        .api_key = "your-api-key",
        .log_callback = log_callback
    };

    tortoise_ctx_t ctx = tortoise_init(&config);
    tortoise_connect(ctx);

    tortoise_message_t msg = { TORTOISE_ROLE_USER, "Hello!", 6 };
    tortoise_response_t response;
    tortoise_chat(ctx, &msg, 1, TORTOISE_THINK_BALANCED, &response);

    printf("Response: %s\n", response.content);

    tortoise_response_free(&response);
    tortoise_disconnect(ctx);
    tortoise_destroy(ctx);
    return 0;
}
```

## API 参考

### 初始化

```c
tortoise_ctx_t tortoise_init(tortoise_config_t* config);
```

### 连接

```c
tortoise_error_t tortoise_connect(tortoise_ctx_t ctx);
tortoise_error_t tortoise_disconnect(tortoise_ctx_t ctx);
tortoise_state_t tortoise_get_state(tortoise_ctx_t ctx);
```

### 聊天

```c
tortoise_error_t tortoise_chat(
    tortoise_ctx_t ctx,
    tortoise_message_t* messages,
    size_t message_count,
    tortoise_think_mode_t mode,
    tortoise_response_t* response
);
```

### 记忆

```c
char* tortoise_remember(tortoise_ctx_t ctx, const char* content, float importance);
size_t tortoise_recall(tortoise_ctx_t ctx, const char* query, tortoise_memory_t* memories, size_t max_count);
```

## 配置选项

| 选项 | 类型 | 描述 |
|------|------|------|
| gateway_url | char* | Tortoise 网关 URL |
| api_key | char* | API 密钥 |
| device_id | char* | 设备 ID |
| port | uint16_t | 端口 (默认 18789) |
| timeout_ms | uint32_t | 超时毫秒 |
| max_memory_kb | size_t | 最大内存使用 |
| auto_reconnect | bool | 自动重连 |
| log_callback | void(*)(const char*, const char*) | 日志回调 |

## 错误代码

| 代码 | 描述 |
|------|------|
| TORTOISE_OK | 成功 |
| TORTOISE_ERROR_INVALID_PARAM | 无效参数 |
| TORTOISE_ERROR_NO_MEMORY | 内存不足 |
| TORTOISE_ERROR_NOT_CONNECTED | 未连接 |
| TORTOISE_ERROR_TIMEOUT | 超时 |
| TORTOISE_ERROR_PROTOCOL | 协议错误 |
| TORTOISE_ERROR_AUTH | 认证失败 |

## 许可证

Apache-2.0
