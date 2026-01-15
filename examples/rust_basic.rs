// Tortoise Framework - Rust Usage Examples

use tortoise_core::{Runtime, RuntimeConfig, AgentConfig, ModelConfig};
use tortoise_core::memory::{MemoryService, MemoryConfig, MemoryType};
use tortoise_core::mcp::{McpServer, ToolDefinition};
use tortoise_core::plugin::{PluginRegistry, PluginConfig, PluginType};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    println!("Tortoise Framework - Rust Examples");
    println!("================================\n");
    
    // Example 1: Basic Runtime
    basic_runtime_example().await?;
    
    // Example 2: Memory Service
    memory_example().await?;
    
    // Example 3: MCP Server
    mcp_example().await?;
    
    // Example 4: Plugin Registry
    plugin_example().await?;
    
    Ok(())
}

// Example 1: Basic Runtime Usage
async fn basic_runtime_example() -> anyhow::Result<()> {
    println!("Example 1: Basic Runtime");
    println!("------------------------");
    
    // Create runtime
    let config = RuntimeConfig::default();
    let runtime = Runtime::new(config);
    
    // Create agent configuration
    let agent_config = AgentConfig {
        id: None,
        name: "my-agent".into(),
        model: ModelConfig {
            provider: "openai".into(),
            model: "gpt-4".into(),
            api_key: Some("sk-...".into()),
            base_url: None,
            max_tokens: Some(2048),
            temperature: Some(0.7),
        },
        skills: vec!["github".into(), "code-review".into()],
        memory: tortoise_core::runtime::MemoryConfig {
            episodic_ttl_hours: 48,
            semantic_enabled: true,
            procedural_enabled: true,
        },
        permissions: vec![],
        metadata: serde_json::json!({}),
    };
    
    // Create agent
    let agent_id = runtime.create_agent(agent_config).await?;
    println!("Created agent: {}", agent_id);
    
    // List agents
    let agents = runtime.list_agents();
    println!("Total agents: {}", agents.len());
    
    // Get agent
    if let Some(agent) = runtime.get_agent(agent_id) {
        let agent = agent.read();
        println!("Agent name: {}", agent.name);
        println!("Agent state: {}", agent.state);
    }
    
    // Delete agent
    runtime.remove_agent(agent_id).await?;
    println!("Deleted agent: {}", agent_id);
    
    println!("");
    Ok(())
}

// Example 2: Memory Service
async fn memory_example() -> anyhow::Result<()> {
    println!("Example 2: Memory Service");
    println!("--------------------------");
    
    // Create memory service
    let config = MemoryConfig::default();
    let memory = MemoryService::new(config);
    
    // Store episodic memory (short-term)
    memory.remember(
        "conversation-1".into(),
        serde_json::json!({
            "user": "John",
            "message": "Hello!",
            "timestamp": chrono::Utc::now()
        }),
        MemoryType::Episodic,
    ).await?;
    println!("Stored episodic memory");
    
    // Store semantic memory (knowledge)
    memory.remember(
        "user-preferences".into(),
        serde_json::json!({
            "theme": "dark",
            "language": "en-US"
        }),
        MemoryType::Semantic,
    ).await?;
    println!("Stored semantic memory");
    
    // Store procedural memory (skills)
    memory.remember(
        "code-review-skill".into(),
        serde_json::json!({
            "name": "Code Review",
            "steps": [
                "Check formatting",
                "Run tests",
                "Review logic",
                "Check security"
            ]
        }),
        MemoryType::Procedural,
    ).await?;
    println!("Stored procedural memory");
    
    // Recall memory
    if let Some(entry) = memory.recall("user-preferences", MemoryType::Semantic).await {
        println!("Recalled: {:?}", entry.content);
    }
    
    // Search memories
    let results = memory.search("preferences", None, 10).await;
    println!("Search results: {} items", results.len());
    
    // Get stats
    let stats = memory.get_stats().await;
    println!("Memory stats: {:?}", stats);
    
    println!("");
    Ok(())
}

// Example 3: MCP Server
async fn mcp_example() -> anyhow::Result<()> {
    println!("Example 3: MCP Server");
    println!("----------------------");
    
    let mut mcp = McpServer::new();
    
    // Register a tool
    let tool = ToolDefinition::new(
        "hello_world",
        "Say hello to the world",
        serde_json::json!({
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Name to greet"
                }
            }
        }),
        Arc::new(|args| async move {
            let name = args.get("name")
                .and_then(|v| v.as_str())
                .unwrap_or("World");
            Ok(format!("Hello, {}!", name))
        }),
    );
    
    mcp.register_tool(tool);
    
    // List tools
    let tools = mcp.list_tools();
    println!("Registered {} tools", tools.len());
    for tool in &tools {
        println!("  - {}: {}", tool.name, tool.description);
    }
    
    println!("");
    Ok(())
}

// Example 4: Plugin Registry
async fn plugin_example() -> anyhow::Result<()> {
    println!("Example 4: Plugin Registry");
    println!("---------------------------");
    
    let registry = PluginRegistry::new();
    
    // Register a channel plugin
    let discord_config = PluginConfig::new(
        "discord-channel",
        "Discord Channel Plugin",
        PluginType::Channel,
    );
    
    registry.register(discord_config).await?;
    println!("Registered Discord plugin");
    
    // Register a tool plugin
    let github_config = PluginConfig::new(
        "github-tools",
        "GitHub Tools Plugin",
        PluginType::Tool,
    );
    
    registry.register(github_config).await?;
    println!("Registered GitHub plugin");
    
    // List all plugins
    let plugins = registry.list();
    println!("Total plugins: {}", plugins.len());
    
    // List by type
    let channels = registry.list_by_type(PluginType::Channel);
    println!("Channel plugins: {}", channels.len());
    
    let tools = registry.list_by_type(PluginType::Tool);
    println!("Tool plugins: {}", tools.len());
    
    println!("");
    Ok(())
}

// Helper for async closures
use std::future::Future;
use std::pin::Pin;
use std::sync::Arc;
use std::collections::HashMap;
use serde_json::Value;
use async_trait::async_trait;

struct AsyncClosure<F, Fut>
where
    F: Fn(HashMap<String, Value>) -> Fut,
    Fut: Future<Output = Result<String, Box<dyn std::error::Error + Send + Sync>>> + Send,
{
    func: F,
}

impl<F, Fut> AsyncClosure<F, Fut>
where
    F: Fn(HashMap<String, Value>) -> Fut + Send + Sync + 'static,
    Fut: Future<Output = Result<String, Box<dyn std::error::Error + Send + Sync>>> + Send,
{
    fn new(func: F) -> Self {
        Self { func }
    }
}

#[async_trait]
impl<F, Fut> tortoise_core::mcp::ToolHandler for AsyncClosure<F, Fut>
where
    F: Fn(HashMap<String, Value>) -> Fut + Send + Sync + 'static,
    Fut: Future<Output = Result<String, Box<dyn std::error::Error + Send + Sync>>> + Send,
{
    async fn call(&self, arguments: &HashMap<String, Value>) -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
        (self.func)(arguments.clone()).await
    }
}
