//! 示例代码 - 流式聊天与工具调用

use anyhow::Result;
use tortoise_core::{
    agent::{
        Agent, AgentConfig, ChatOptions, Message, ModelProvider, ThinkMode, ToolCall,
    },
    tool::{Tool, ToolMetadata},
    init_logging,
};
use async_trait::async_trait;
use serde_json::json;

#[tokio::main]
async fn main() -> Result<()> {
    init_logging!();

    // 创建代理
    let config = AgentConfig {
        id: "tool-agent".to_string(),
        name: "Tool Agent".to_string(),
        model_provider: ModelProvider::OpenAI {
            model: "gpt-4".to_string(),
            api_key: std::env::var("OPENAI_API_KEY")
                .unwrap_or_else(|_| "sk-test".to_string()),
            base_url: None,
            organization: None,
        },
        default_thinking: ThinkMode::Balanced,
        max_context: 128_000,
        temperature: 0.7,
        ..Default::default()
    };

    let agent = tortoise_core::agent::create_agent(config).await?;

    // 注册自定义工具
    let calculator = CalculatorTool;
    agent.register_tool(calculator)?;

    // 发送带工具调用的请求
    let messages = vec![
        Message::user("请计算 123 * 456 + 789 等于多少？"),
    ];

    let options = ChatOptions {
        tools: Some(vec!["calculator".to_string()]),
        ..Default::default()
    };

    println!("Sending message with tool support...");
    let mut response = agent.chat(messages, options).await?;

    println!("Response stream:\n");

    while let Some(event) = response.events.recv().await {
        match event {
            tortoise_core::agent::AgentEvent::ThinkingStarted { mode } => {
                println!("[Thinking Started] Mode: {:?}", mode);
            }
            tortoise_core::agent::AgentEvent::Thinking { content } => {
                println!("[Thinking] {}", content);
            }
            tortoise_core::agent::AgentEvent::Generation { content } => {
                print!("{}", content);
            }
            tortoise_core::agent::AgentEvent::ToolCall { call } => {
                println!("\n[Tool Call] {}", call.name);
                println!("  Arguments: {:?}", call.arguments);
            }
            tortoise_core::agent::AgentEvent::ToolExecutionStarted { call_id, tool_name } => {
                println!("[Tool Started] {} ({})", tool_name, call_id);
            }
            tortoise_core::agent::AgentEvent::ToolExecutionComplete { call_id, result } => {
                println!("[Tool Complete] {}: {}", call_id, result);
            }
            tortoise_core::agent::AgentEvent::ToolExecutionFailed { call_id, error } => {
                println!("[Tool Failed] {}: {}", call_id, error);
            }
            tortoise_core::agent::AgentEvent::ResponseComplete { content, tool_results } => {
                println!("\n\n[Complete]");
                println!("  Content length: {} chars", content.len());
                println!("  Tool results: {}", tool_results.len());
            }
            tortoise_core::agent::AgentEvent::Error { code, message } => {
                println!("\n[Error] {}: {}", code, message);
            }
            _ => {}
        }
    }

    Ok(())
}

// 计算器工具
struct CalculatorTool;

#[async_trait]
impl Tool for CalculatorTool {
    fn metadata(&self) -> ToolMetadata {
        ToolMetadata {
            name: "calculator".to_string(),
            description: "Perform mathematical calculations".to_string(),
            parameters: json!({
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "The mathematical expression to evaluate"
                    }
                },
                "required": ["expression"]
            }),
            category: Some("math".to_string()),
            tags: vec!["math".to_string(), "calculator".to_string()],
            version: "1.0.0".to_string(),
        }
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value> {
        let expression = arguments["expression"]
            .as_str()
            .ok_or_else(|| anyhow::anyhow!("Missing expression"))?;

        // 简单的计算
        let result = meval::eval_str(expression)
            .map_err(|e| anyhow::anyhow!("Calculation error: {}", e))?;

        Ok(json!({
            "expression": expression,
            "result": result,
            "formatted": format!("{} = {}", expression, result)
        }))
    }
}

mod tool {
    use super::*;

    #[async_trait]
    pub trait Tool: Send + Sync {
        fn metadata(&self) -> ToolMetadata;
        async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value>;
    }

    pub struct ToolMetadata {
        pub name: String,
        pub description: String,
        pub parameters: serde_json::Value,
        pub category: Option<String>,
        pub tags: Vec<String>,
        pub version: String,
    }
}
