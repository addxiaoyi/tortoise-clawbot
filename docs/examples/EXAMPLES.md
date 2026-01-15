# Tortoise 使用示例

## Python SDK 示例

```python
import asyncio
from tortoise import TortoiseClient, Message, ChatOptions, ThinkMode

async def main():
    async with TortoiseClient("http://localhost:18789") as client:
        # 简单聊天
        messages = [Message.user("Hello, Tortoise!")]
        response = await client.chat(messages)
        print(f"Response: {response}")

        # 带选项的聊天
        messages = [
            Message.user("Explain quantum computing"),
            Message.assistant("Quantum computing is...")
        ]
        options = ChatOptions(
            thinking_mode=ThinkMode.RESEARCH,
            temperature=0.7
        )
        response = await client.chat(messages, options)
        print(f"Detailed response: {response}")

        # 记住信息
        memory_id = await client.remember(
            "Python asyncio is powerful",
            importance=0.8
        )
        print(f"Memory ID: {memory_id}")

        # 回忆相关信息
        memories = await client.recall("async programming", limit=5)
        for memory in memories:
            print(f"- {memory['content']}")

        # 获取状态
        status = await client.status()
        print(f"Memory stats: {status['memory_stats']}")

if __name__ == "__main__":
    asyncio.run(main())
```

## Go SDK 示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tortoise-ai/tortoise/plugins"
)

func main() {
    ctx := context.Background()

    // Create plugin manager
    manager := plugins.NewPluginManager()

    // Create Discord plugin
    discord := channels.NewDiscordPlugin()

    // Register plugin
    if err := manager.Register(discord); err != nil {
        log.Fatalf("Failed to register: %v", err)
    }

    // Start plugin
    config := []byte(`{"token": "your-token"}`)
    if err := manager.Load(ctx, "discord-channel", config); err != nil {
        log.Fatalf("Failed to load: %v", err)
    }

    if err := manager.Start(ctx, "discord-channel"); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }

    // Send message
    args := channels.SendMessageArgs{
        ChannelID: "123456789",
        Content:   "Hello from Tortoise!",
    }
    argsJSON, _ := json.Marshal(args)
    result, err := manager.Execute(ctx, "discord-channel", "send", argsJSON)
    if err != nil {
        log.Fatalf("Failed to execute: %v", err)
    }
    fmt.Printf("Result: %s\n", result)
}
```

## C SDK 示例

```c
#include "tortoise.h"
#include <stdio.h>

void log_callback(const char* level, const char* message) {
    printf("[%s] %s\n", level, message);
}

int main() {
    tortoise_config_t config = {
        .gateway_url = "http://localhost:18789",
        .api_key = "your-api-key",
        .device_id = "my-device",
        .log_callback = log_callback
    };

    tortoise_ctx_t ctx = tortoise_init(&config);
    if (!ctx) {
        fprintf(stderr, "Init failed\n");
        return 1;
    }

    // Connect
    tortoise_connect(ctx);

    // Chat
    tortoise_message_t messages[] = {
        { TORTOISE_ROLE_USER, "Hello!", 6 }
    };
    tortoise_response_t response;
    tortoise_chat(ctx, messages, 1, TORTOISE_THINK_BALANCED, &response);
    printf("Response: %s\n", response.content);
    tortoise_response_free(&response);

    // Remember
    char* memory_id = tortoise_remember(ctx, "Important fact", 0.9);
    printf("Memory ID: %s\n", memory_id);
    free(memory_id);

    tortoise_disconnect(ctx);
    tortoise_destroy(ctx);
    return 0;
}
```

## Rust Core 示例

```rust
use tortoise_core::{Config, Tortoise, Message, MessageBuilder, ThinkMode};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let config = Config::default();
    let tortoise = Tortoise::new(config).await?;

    tortoise.start().await?;

    let messages = vec![
        MessageBuilder::user("Hello, Tortoise!").build(),
    ];

    let options = ChatOptions {
        thinking_mode: Some(ThinkMode::Balanced),
        temperature: Some(0.7),
        ..Default::default()
    };

    let response = tortoise.agent().chat(messages, options).await?;

    // Process streaming response
    while let Some(event) = response.events.recv().await {
        match event {
            AgentEvent::Thinking(content) => print!("{}", content),
            AgentEvent::ResponseComplete { content, .. } => println!("\nFinal: {}", content),
            _ => {}
        }
    }

    tortoise.stop().await?;
    Ok(())
}
```

## Flutter 示例

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/chat/bloc/chat_bloc.dart';

class ChatScreen extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final chatState = ref.watch(chatBlocProvider);
    final chatBloc = ref.read(chatBlocProvider.notifier);

    return Scaffold(
      appBar: AppBar(title: Text('Tortoise')),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: chatState.messages.length,
              itemBuilder: (context, index) {
                return MessageBubble(message: chatState.messages[index]);
              },
            ),
          ),
          InputBar(
            onSend: (text) => chatBloc.sendMessage(text),
          ),
        ],
      ),
    );
  }
}
```
