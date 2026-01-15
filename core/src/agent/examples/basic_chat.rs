//! 示例代码 - 基础聊天

use anyhow::Result;
use tortoise_core::{
    agent::{Agent, AgentConfig, ChatOptions, Message, ModelProvider, ThinkMode},
    init_logging,
};

#[tokio::main]
async fn main() -> Result<()> {
    init_logging!();

    // 创建代理配置
    let config = AgentConfig {
        id: "my-agent".to_string(),
        name: "My Tortoise Agent".to_string(),
        model_provider: ModelProvider::Ollama {
            model: "llama3".to_string(),
            base_url: "http://localhost:11434".to_string(),
            temperature: None,
        },
        default_thinking: ThinkMode::Balanced,
        max_context: 8192,
        temperature: 0.7,
        system_prompt: Some(
            "You are a helpful assistant named Tortoise. \
             You are patient, knowledgeable, and always ready to help.".to_string(),
        ),
        ..Default::default()
    };

    // 创建代理
    println!("Creating agent...");
    let agent = tortoise_core::agent::create_agent(config).await?;
    println!("Agent created: {} ({})", agent.name(), agent.id());

    // 发送消息
    let messages = vec![
        Message::user("Hello! What can you do?"),
    ];

    println!("Sending message...");
    let options = ChatOptions {
        thinking_mode: Some(ThinkMode::Balanced),
        ..Default::default()
    };

    let mut response = agent.chat(messages, options).await?;
    println!("Response received:");

    // 处理流式响应
    while let Some(event) = response.events.recv().await {
        match event {
            tortoise_core::agent::AgentEvent::Thinking { content } => {
                println!("  [Thinking] {}", content);
            }
            tortoise_core::agent::AgentEvent::Generation { content } => {
                print!("{}", content);
            }
            tortoise_core::agent::AgentEvent::ResponseComplete { content, .. } => {
                println!("\n[Complete] Response: {} chars", content.len());
            }
            tortoise_core::agent::AgentEvent::Error { code, message } => {
                println!("\n[Error] {}: {}", code, message);
            }
            _ => {}
        }
    }

    // 获取统计
    let stats = agent.stats().await;
    println!("\nAgent Stats:");
    println!("  Total requests: {}", stats.total_requests);
    println!("  Total tokens: {}", stats.total_tokens);
    println!("  Avg response time: {:.2}ms", stats.average_response_time_ms);

    Ok(())
}
