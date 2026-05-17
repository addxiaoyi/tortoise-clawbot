//! Tortoise Rust SDK - 渠道示例
//! 演示如何通过不同渠道发送和接收消息

use tortoise_sdk::{Client, Config, Channel, Message};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new(Config {
        api_key: Some("your-api-key".to_string()),
        ..Default::default()
    });
    
    // ============ Telegram 示例 ============
    telegram_example(&client).await?;
    
    // ============ Discord 示例 ============
    discord_example(&client).await?;
    
    // ============ 消息路由示例 ============
    message_routing_example(&client).await?;
    
    Ok(())
}

async fn telegram_example(client: &Client) -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Telegram 渠道示例 ===");
    
    // 连接到 Telegram
    client.connect_channel(Channel::Telegram {
        token: "your-telegram-bot-token".to_string(),
    }).await?;
    
    // 发送消息
    client.send_message(Channel::Telegram {
        token: "your-telegram-bot-token".to_string(),
    }, "Hello from Tortoise!").await?;
    
    // 监听消息
    let mut receiver = client.subscribe_channel(Channel::Telegram {
        token: "your-telegram-bot-token".to_string(),
    }).await?;
    
    while let Some(message) = receiver.recv().await {
        println!("收到 Telegram 消息: {:?}", message);
    }
    
    Ok(())
}

async fn discord_example(client: &Client) -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Discord 渠道示例 ===");
    
    // 连接到 Discord
    client.connect_channel(Channel::Discord {
        token: "your-discord-bot-token".to_string(),
    }).await?;
    
    // 监听消息
    let mut receiver = client.subscribe_channel(Channel::Discord {
        token: "your-discord-bot-token".to_string(),
    }).await?;
    
    while let Some(message) = receiver.recv().await {
        // 处理 Discord 消息
        println!("收到 Discord 消息: {:?}", message);
        
        // 回复消息
        if message.content.contains("!help") {
            client.send_message(Channel::Discord {
                token: "your-discord-bot-token".to_string(),
            }).to_channel(message.channel_id)
            .content("Available commands: !help, !info, !status")
            .send()
            .await?;
        }
    }
    
    Ok(())
}

async fn message_routing_example(client: &Client) -> Result<(), Box<dyn std::error::Error>> {
    println!("=== 消息路由示例 ===");
    
    // 创建消息路由器
    let router = client.create_router();
    
    // 添加路由规则
    router.add_rule(
        |msg: &Message| -> Option<String> {
            // 根据内容路由到不同渠道
            if msg.content.starts_with("!telegram") {
                Some("telegram".to_string())
            } else if msg.content.starts_with("!discord") {
                Some("discord".to_string())
            } else {
                None // 默认渠道
            }
        },
    );
    
    // 监听所有渠道的消息并路由
    let mut receiver = router.subscribe().await?;
    
    while let Some((channel, message)) = receiver.recv().await {
        println!("消息从 {:?} 路由到 {}", channel, message.destination);
    }
    
    Ok(())
}

// 自定义渠道处理器
struct CustomChannelHandler;

impl CustomChannelHandler {
    async fn handle_message(&self, message: &Message) -> Result<(), Box<dyn std::error::Error>> {
        // 自定义消息处理逻辑
        println!("处理自定义消息: {}", message.content);
        
        // 可以在这里调用 AI 服务
        // let response = ai_service.chat(&message.content).await?;
        
        Ok(())
    }
}
