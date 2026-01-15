# Tortoise 多代理系统设计

## 概述

Tortoise 多代理系统支持多个专业代理协同工作，处理复杂任务。每个代理可以扮演不同角色，通过消息传递和状态共享实现协作。

## 代理角色定义

```rust
// src/agent/multi_agent.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// 代理角色
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AgentRole {
    /// 主控代理 - 负责任务分解和结果汇总
    Orchestrator,
    /// 专业代理 - 专注于特定领域
    Specialist,
    /// 助手代理 - 提供日常辅助
    Assistant,
    /// 批评代理 - 审查和提出改进建议
    Critic,
    /// 研究代理 - 深入研究和分析
    Researcher,
    /// 编码代理 - 负责代码开发和调试
    Coder,
    /// 测试代理 - 负责测试和质量保证
    Tester,
    /// 文档代理 - 负责文档编写
    Documenter,
}

impl AgentRole {
    pub fn system_prompt(&self) -> &'static str {
        match self {
            AgentRole::Orchestrator => r#"You are the Orchestrator, responsible for:
- Breaking down complex tasks into subtasks
- Assigning tasks to specialized agents
- Collecting and synthesizing results
- Managing workflow and dependencies
- Ensuring quality and consistency

You think step by step and communicate clearly."#,
            AgentRole::Specialist => r#"You are a Specialist agent, focused on:
- Deep expertise in your domain
- High-quality, accurate outputs
- Thorough analysis and solutions
- Clear explanations of complex topics

You are the expert in your field."#,
            AgentRole::Assistant => r#"You are a helpful Assistant, providing:
- Quick and accurate responses
- Friendly and approachable tone
- Concise and actionable information
- Proactive suggestions

You make users' lives easier."#,
            AgentRole::Critic => r#"You are a Critical thinker, responsible for:
- Reviewing and critiquing work
- Identifying flaws and improvements
- Ensuring quality standards
- Providing constructive feedback
- Validating correctness

You are thorough and unbiased."#,
            AgentRole::Researcher => r#"You are a Research expert, focused on:
- Deep investigation and analysis
- Finding credible sources
- Comprehensive understanding
- Evidence-based conclusions
- Clear documentation of findings

You leave no stone unturned."#,
            AgentRole::Coder => r#"You are an Expert Coder, skilled in:
- Clean, maintainable code
- Best practices and design patterns
- Performance optimization
- Testing and debugging
- Code reviews

You write code that others want to maintain."#,
            AgentRole::Tester => r#"You are a Quality Assurance expert, responsible for:
- Comprehensive test coverage
- Finding edge cases
- Validating functionality
- Bug reporting with details
- Quality metrics

You ensure software quality."#,
            AgentRole::Documenter => r#"You are a Technical Writer, skilled in:
- Clear, concise documentation
- Proper formatting and structure
- Examples and tutorials
- API documentation
- User guides

You make complex things understandable."#,
        }
    }
}

/// 代理实例
pub struct AgentInstance {
    pub id: String,
    pub name: String,
    pub role: AgentRole,
    pub description: String,
    pub agent: Arc<dyn super::Agent>,
    pub status: AgentStatus,
    pub capabilities: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AgentStatus {
    Idle,
    Busy,
    Waiting,
    Error(String),
}

/// 任务
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub id: String,
    pub description: String,
    pub assigned_to: Option<String>,
    pub status: TaskStatus,
    pub dependencies: Vec<String>,
    pub result: Option<TaskResult>,
    pub created_at: i64,
    pub completed_at: Option<i64>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaskStatus {
    Pending,
    InProgress,
    Completed,
    Failed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskResult {
    pub success: bool,
    pub output: String,
    pub artifacts: Vec<Artifact>,
    pub metrics: HashMap<String, f64>,
}

/// 产物
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Artifact {
    pub name: String,
    pub artifact_type: ArtifactType,
    pub content: String,
    pub path: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ArtifactType {
    Text,
    Code,
    Document,
    Image,
    Data,
    Other,
}

/// 多代理系统
pub struct MultiAgentSystem {
    agents: RwLock<HashMap<String, Arc<AgentInstance>>>,
    tasks: RwLock<HashMap<String, Task>>,
    orchestrator: Arc<AgentInstance>,
    message_bus: Arc<MessageBus>,
}

impl MultiAgentSystem {
    pub fn new(orchestrator: Arc<dyn super::Agent>) -> Self {
        let message_bus = Arc::new(MessageBus::new());
        
        let orchestrator_instance = Arc::new(AgentInstance {
            id: "orchestrator".to_string(),
            name: "Orchestrator".to_string(),
            role: AgentRole::Orchestrator,
            description: "Main coordinator for task delegation".to_string(),
            agent: orchestrator,
            status: AgentStatus::Idle,
            capabilities: vec![
                "task_decomposition".to_string(),
                "workflow_planning".to_string(),
                "result_synthesis".to_string(),
            ],
        });

        Self {
            agents: RwLock::new(HashMap::new()),
            tasks: RwLock::new(HashMap::new()),
            orchestrator: orchestrator_instance,
            message_bus,
        }
    }

    /// 注册代理
    pub async fn register_agent(&self, agent: AgentInstance) {
        let mut agents = self.agents.write().await;
        agents.insert(agent.id.clone(), Arc::new(agent));
    }

    /// 创建任务
    pub async fn create_task(&self, description: String) -> String {
        let id = uuid::Uuid::new_v4().to_string();
        let task = Task {
            id: id.clone(),
            description,
            assigned_to: None,
            status: TaskStatus::Pending,
            dependencies: vec![],
            result: None,
            created_at: chrono::Utc::now().timestamp(),
            completed_at: None,
        };
        
        let mut tasks = self.tasks.write().await;
        tasks.insert(id.clone(), task);
        id
    }

    /// 执行复杂任务
    pub async fn execute_task(&self, description: String) -> Result<TaskResult> {
        // 创建主任务
        let task_id = self.create_task(description).await;
        
        // 更新状态
        {
            let mut tasks = self.tasks.write().await;
            if let Some(task) = tasks.get_mut(&task_id) {
                task.status = TaskStatus::InProgress;
            }
        }

        // 调用 Orchestrator 分析任务
        let plan = self.plan_task(&task_id).await?;

        // 执行子任务
        let results = self.execute_plan(plan).await?;

        // 汇总结果
        let final_result = self.synthesize_results(results).await?;

        // 更新任务状态
        {
            let mut tasks = self.tasks.write().await;
            if let Some(task) = tasks.get_mut(&task_id) {
                task.status = TaskStatus::Completed;
                task.result = Some(final_result.clone());
                task.completed_at = Some(chrono::Utc::now().timestamp());
            }
        }

        Ok(final_result)
    }

    /// 规划任务
    async fn plan_task(&self, task_id: &str) -> Result<ExecutionPlan> {
        let tasks = self.tasks.read().await;
        let task = tasks.get(task_id)
            .ok_or_else(|| anyhow::anyhow!("Task not found"))?;

        // 构建提示词让 Orchestrator 分析
        let prompt = format!(
            r#"Analyze this task and create an execution plan:

Task: {}

Available agents:
{}

Create a plan with:
1. Subtasks to complete
2. Dependencies between subtasks
3. Best agent for each subtask
4. Estimated complexity

Return a JSON plan."#,
            task.description,
            self.list_available_agents().await
        );

        let messages = vec![
            super::Message {
                role: super::MessageRole::System,
                content: AgentRole::Orchestrator.system_prompt().to_string(),
                tool_calls: None,
                tool_results: None,
            },
            super::Message {
                role: super::MessageRole::User,
                content: prompt,
                tool_calls: None,
                tool_results: None,
            },
        ];

        let response = self.orchestrator.agent
            .chat(messages, Default::default())
            .await?;

        // 解析响应为执行计划
        // 简化版本，实际应该解析 JSON
        Ok(ExecutionPlan {
            task_id: task_id.to_string(),
            subtasks: vec![
                SubTask {
                    id: uuid::Uuid::new_v4().to_string(),
                    description: task.description.clone(),
                    assigned_role: AgentRole::Specialist,
                    dependencies: vec![],
                }
            ],
        })
    }

    /// 执行计划
    async fn execute_plan(&self, plan: ExecutionPlan) -> Result<Vec<SubTaskResult>> {
        let mut results = Vec::new();
        
        // 按依赖顺序执行
        for subtask in &plan.subtasks {
            let result = self.execute_subtask(subtask).await?;
            results.push(result);
        }

        Ok(results)
    }

    /// 执行子任务
    async fn execute_subtask(&self, subtask: &SubTask) -> Result<SubTaskResult> {
        // 查找合适的代理
        let agent = self.find_agent_for_role(subtask.assigned_role).await
            .ok_or_else(|| anyhow::anyhow!("No agent available for role: {:?}", subtask.assigned_role))?;

        // 构建提示词
        let prompt = format!(
            "{}\n\nTask: {}",
            agent.role.system_prompt(),
            subtask.description
        );

        let messages = vec![
            super::Message {
                role: super::MessageRole::System,
                content: prompt,
                tool_calls: None,
                tool_results: None,
            },
        ];

        // 执行
        let response = agent.agent
            .chat(messages, Default::default())
            .await?;

        // 收集结果
        let mut output = String::new();
        use tokio::sync::mpsc;
        let (tx, mut rx) = mpsc::channel(100);
        
        tokio::spawn(async move {
            while let Some(event) = response.events.recv().await {
                if let super::AgentEvent::Thinking(content) = event {
                    let _ = tx.send(content).await;
                }
            }
        });

        while let Some(chunk) = rx.recv().await {
            output.push_str(&chunk);
        }

        Ok(SubTaskResult {
            subtask_id: subtask.id.clone(),
            success: true,
            output,
            artifacts: vec![],
        })
    }

    /// 查找代理
    async fn find_agent_for_role(&self, role: AgentRole) -> Option<Arc<AgentInstance>> {
        let agents = self.agents.read().await;
        agents.values()
            .find(|a| a.role == role)
            .cloned()
    }

    /// 列出可用代理
    async fn list_available_agents(&self) -> String {
        let agents = self.agents.read().await;
        let mut output = String::new();
        
        for agent in agents.values() {
            output.push_str(&format!(
                "- {} ({:?}): {}\n",
                agent.name, agent.role, agent.description
            ));
        }
        
        if output.is_empty() {
            output = "No additional agents registered.".to_string();
        }
        
        output
    }

    /// 汇总结果
    async fn synthesize_results(&self, results: Vec<SubTaskResult>) -> Result<TaskResult> {
        let combined_output: String = results.iter()
            .map(|r| format!("[{}] {}\n", r.subtask_id, r.output))
            .collect();

        let prompt = format!(
            r#"Synthesize the following results into a coherent final answer:

{}

Create a clear, well-structured response that addresses the original task."#,
            combined_output
        );

        let messages = vec![
            super::Message {
                role: super::MessageRole::System,
                content: AgentRole::Orchestrator.system_prompt().to_string(),
                tool_calls: None,
                tool_results: None,
            },
            super::Message {
                role: super::MessageRole::User,
                content: prompt,
                tool_calls: None,
                tool_results: None,
            },
        ];

        let response = self.orchestrator.agent
            .chat(messages, Default::default())
            .await?;

        let mut final_output = String::new();
        use tokio::sync::mpsc;
        let (tx, mut rx) = mpsc::channel(100);
        
        tokio::spawn(async move {
            while let Some(event) = response.events.recv().await {
                if let super::AgentEvent::Thinking(content) = event {
                    let _ = tx.send(content).await;
                }
            }
        });

        while let Some(chunk) = rx.recv().await {
            final_output.push_str(&chunk);
        }

        Ok(TaskResult {
            success: true,
            output: final_output,
            artifacts: results.into_iter()
                .flat_map(|r| r.artifacts)
                .collect(),
            metrics: HashMap::new(),
        })
    }
}

/// 执行计划
#[derive(Debug, Clone)]
pub struct ExecutionPlan {
    pub task_id: String,
    pub subtasks: Vec<SubTask>,
}

/// 子任务
#[derive(Debug, Clone)]
pub struct SubTask {
    pub id: String,
    pub description: String,
    pub assigned_role: AgentRole,
    pub dependencies: Vec<String>,
}

/// 子任务结果
#[derive(Debug, Clone)]
pub struct SubTaskResult {
    pub subtask_id: String,
    pub success: bool,
    pub output: String,
    pub artifacts: Vec<Artifact>,
}

/// 消息总线
pub struct MessageBus {
    channels: RwLock<HashMap<String, tokio::sync::mpsc::Sender<AgentMessage>>>,
}

impl MessageBus {
    pub fn new() -> Self {
        Self {
            channels: RwLock::new(HashMap::new()),
        }
    }

    pub async fn subscribe(&self, agent_id: &str) -> tokio::sync::mpsc::Receiver<AgentMessage> {
        let (tx, rx) = tokio::sync::mpsc::channel(100);
        let mut channels = self.channels.write().await;
        channels.insert(agent_id.to_string(), tx);
        rx
    }

    pub async fn publish(&self, target: &str, message: AgentMessage) -> Result<()> {
        let channels = self.channels.read().await;
        if let Some(tx) = channels.get(target) {
            tx.send(message).await?;
        }
        Ok(())
    }

    pub async fn broadcast(&self, message: AgentMessage) -> Result<()> {
        let channels = self.channels.read().await;
        for tx in channels.values() {
            let _ = tx.send(message.clone()).await;
        }
        Ok(())
    }
}

/// 代理消息
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum AgentMessage {
    TaskAssigned { task_id: String, from: String, to: String },
    TaskCompleted { task_id: String, result: String },
    NeedHelp { from: String, question: String },
    Feedback { from: String, to: String, content: String },
    StatusUpdate { agent_id: String, status: AgentStatus },
}
```

## 使用示例

```rust
// 示例：使用多代理系统

use tortoise_core::agent::multi_agent::*;

#[tokio::main]
async fn main() -> Result<()> {
    // 创建基础代理
    let base_agent = create_openai_agent("gpt-4").await?;
    
    // 创建多代理系统
    let multi_agent = MultiAgentSystem::new(base_agent);
    
    // 注册专业代理
    multi_agent.register_agent(AgentInstance {
        id: "coder-1".to_string(),
        name: "Code Assistant".to_string(),
        role: AgentRole::Coder,
        description: "Expert in code development".to_string(),
        agent: create_openai_agent("gpt-4").await?,
        status: AgentStatus::Idle,
        capabilities: vec!["python".to_string(), "rust".to_string()],
    }).await;
    
    multi_agent.register_agent(AgentInstance {
        id: "researcher-1".to_string(),
        name: "Research Assistant".to_string(),
        role: AgentRole::Researcher,
        description: "Expert in research and analysis".to_string(),
        agent: create_openai_agent("gpt-4").await?,
        status: AgentStatus::Idle,
        capabilities: vec!["web_search".to_string(), "data_analysis".to_string()],
    }).await;
    
    // 执行复杂任务
    let result = multi_agent.execute_task(
        "Research the latest developments in AI and write a Python script to analyze trends".to_string()
    ).await?;
    
    println!("Result: {}", result.output);
    
    Ok(())
}
```
