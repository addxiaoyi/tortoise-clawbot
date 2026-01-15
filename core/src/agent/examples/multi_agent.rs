//! 示例代码 - 多代理协作

use anyhow::Result;
use tortoise_core::{
    agent::{
        multi_agent::{
            AgentRole, AgentInfo, CollaborationTask, MultiAgentManager,
            CollaborationResult,
        },
        AgentConfig, ChatOptions, Message, ModelProvider, ThinkMode, Agent,
    },
    init_logging,
};

#[tokio::main]
async fn main() -> Result<()> {
    init_logging!();

    // 创建多代理管理器
    let manager = MultiAgentManager::new();

    // 创建协调器
    let coordinator_config = AgentConfig {
        id: "coordinator".to_string(),
        name: "Coordinator".to_string(),
        model_provider: ModelProvider::Ollama {
            model: "llama3".to_string(),
            base_url: "http://localhost:11434".to_string(),
            temperature: None,
        },
        default_thinking: ThinkMode::Deep,
        system_prompt: Some(AgentRole::Coordinator.default_prompt().to_string()),
        ..Default::default()
    };

    let coordinator = tortoise_core::agent::create_agent(coordinator_config).await?;
    manager.register_agent(AgentInfo {
        id: coordinator.id().to_string(),
        name: coordinator.name().to_string(),
        role: AgentRole::Coordinator,
        description: AgentRole::Coordinator.default_prompt().to_string(),
        agent: coordinator,
    }).await;
    manager.set_coordinator("coordinator").await;

    // 创建专家代理
    let researcher_config = AgentConfig {
        id: "researcher".to_string(),
        name: "Researcher".to_string(),
        model_provider: ModelProvider::Ollama {
            model: "llama3".to_string(),
            base_url: "http://localhost:11434".to_string(),
            temperature: None,
        },
        default_thinking: ThinkMode::Research,
        system_prompt: Some(AgentRole::Researcher.default_prompt().to_string()),
        ..Default::default()
    };

    let researcher = tortoise_core::agent::create_agent(researcher_config).await?;
    manager.register_agent(AgentInfo {
        id: researcher.id().to_string(),
        name: researcher.name().to_string(),
        role: AgentRole::Researcher,
        description: AgentRole::Researcher.default_prompt().to_string(),
        agent: researcher,
    }).await;

    // 创建创意代理
    let creative_config = AgentConfig {
        id: "creative".to_string(),
        name: "Creative".to_string(),
        model_provider: ModelProvider::Ollama {
            model: "llama3".to_string(),
            base_url: "http://localhost:11434".to_string(),
            temperature: None,
        },
        default_thinking: ThinkMode::Creative,
        system_prompt: Some(AgentRole::Creative.default_prompt().to_string()),
        ..Default::default()
    };

    let creative = tortoise_core::agent::create_agent(creative_config).await?;
    manager.register_agent(AgentInfo {
        id: creative.id().to_string(),
        name: creative.name().to_string(),
        role: AgentRole::Creative,
        description: AgentRole::Creative.default_prompt().to_string(),
        agent: creative,
    }).await;

    println!("Multi-agent system initialized:");
    println!("  - {} agents registered", manager.count().await);
    println!("  - Coordinator: coordinator");
    println!("  - Specialist agents: researcher, creative");

    // 执行协作任务
    let task = CollaborationTask::Single(Message::user(
        "帮我研究一下量子计算的最新进展，并提出3个创新应用想法"
    ));

    println!("\nExecuting collaborative task...");
    match manager.collaborate(task).await {
        Ok(result) => {
            println!("Task completed!");
            match result {
                CollaborationResult::Single(msg) => {
                    println!("Result: {}", msg.content.to_string());
                }
                CollaborationResult::Combined(results) => {
                    println!("Combined results from {} agents", results.len());
                }
                _ => {}
            }
        }
        Err(e) => {
            eprintln!("Task failed: {}", e);
        }
    }

    // 列出所有代理
    println!("\nAll agents:");
    for agent in manager.list_agents().await {
        println!("  - {} ({:?}): {}", agent.name, agent.role, agent.description);
    }

    Ok(())
}
