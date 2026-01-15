//! 多代理系统
//!
//! 支持多个代理协作的架构

use anyhow::Result;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{mpsc, RwLock};

use super::engine::{Agent, AgentConfig, AgentEvent, AgentStatus, ChatOptions, Message, StreamingResponse};

/// 代理角色
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AgentRole {
    /// 主代理
    Coordinator,
    /// 专家代理
    Specialist,
    /// 研究代理
    Researcher,
    /// 创意代理
    Creative,
    /// 执行代理
    Executor,
    /// 审核代理
    Reviewer,
}

impl AgentRole {
    pub fn default_prompt(&self) -> &'static str {
        match self {
            AgentRole::Coordinator => "You are the coordinator. Your job is to break down tasks and delegate to appropriate specialists.",
            AgentRole::Specialist => "You are a specialist agent. You have deep knowledge in your domain.",
            AgentRole::Researcher => "You are a researcher. You gather and analyze information thoroughly.",
            AgentRole::Creative => "You are a creative agent. You generate innovative ideas and solutions.",
            AgentRole::Executor => "You are an executor agent. You complete tasks efficiently and accurately.",
            AgentRole::Reviewer => "You are a reviewer agent. You evaluate and provide feedback on work.",
        }
    }
}

/// 代理信息
#[derive(Debug, Clone)]
pub struct AgentInfo {
    pub id: String,
    pub name: String,
    pub role: AgentRole,
    pub description: String,
    pub agent: Arc<dyn Agent>,
}

/// 多代理管理器
pub struct MultiAgentManager {
    agents: RwLock<HashMap<String, AgentInfo>>,
    coordinator_id: RwLock<Option<String>>,
    collaboration_channel: mpsc::Sender<CollaborationMessage>,
}

impl MultiAgentManager {
    /// 创建新的多代理管理器
    pub fn new() -> Self {
        let (tx, _rx) = mpsc::channel(1000);
        Self {
            agents: RwLock::new(HashMap::new()),
            coordinator_id: RwLock::new(None),
            collaboration_channel: tx,
        }
    }

    /// 注册代理
    pub async fn register_agent(&self, agent: AgentInfo) {
        let mut agents = self.agents.write().await;
        agents.insert(agent.id.clone(), agent);
    }

    /// 注册代理 (通过配置)
    pub async fn register_from_config(
        &self,
        config: AgentConfig,
        role: AgentRole,
    ) -> Result<()> {
        let agent = super::create_agent(config).await?;
        
        let info = AgentInfo {
            id: agent.id().to_string(),
            name: agent.name().to_string(),
            role,
            description: role.default_prompt().to_string(),
            agent,
        };
        
        self.register_agent(info).await;
        Ok(())
    }

    /// 设置协调器
    pub async fn set_coordinator(&self, agent_id: &str) {
        let mut coordinator = self.coordinator_id.write().await;
        *coordinator = Some(agent_id.to_string());
    }

    /// 获取协调器
    pub async fn get_coordinator(&self) -> Option<AgentInfo> {
        let coordinator_id = self.coordinator_id.read().await;
        let agents = self.agents.read().await;
        
        coordinator_id.as_ref().and_then(|id| {
            agents.get(id).cloned()
        })
    }

    /// 获取代理
    pub async fn get_agent(&self, id: &str) -> Option<AgentInfo> {
        let agents = self.agents.read().await;
        agents.get(id).cloned()
    }

    /// 按角色获取代理
    pub async fn get_agents_by_role(&self, role: AgentRole) -> Vec<AgentInfo> {
        let agents = self.agents.read().await;
        agents.values()
            .filter(|a| a.role == role)
            .cloned()
            .collect()
    }

    /// 列出所有代理
    pub async fn list_agents(&self) -> Vec<AgentInfo> {
        let agents = self.agents.read().await;
        agents.values().cloned().collect()
    }

    /// 协作处理任务
    pub async fn collaborate(
        &self,
        task: CollaborationTask,
    ) -> Result<CollaborationResult> {
        let coordinator = self.get_coordinator().await
            .ok_or_else(|| anyhow::anyhow!("No coordinator set"))?;
        
        let mut results = HashMap::new();
        let mut pending_tasks = vec![(coordinator.clone(), task)];
        
        while let Some((agent_info, current_task)) = pending_tasks.pop() {
            match current_task {
                CollaborationTask::Single(message) => {
                    let response = agent_info.agent.chat_sync(
                        vec![message],
                        ChatOptions::default(),
                    ).await?;
                    
                    results.insert(agent_info.id.clone(), CollaborationResult::Single(response));
                }
                CollaborationTask::Parallel(tasks) => {
                    let mut subtasks = Vec::new();
                    
                    for task in tasks {
                        // 查找合适的代理
                        let agent = self.find_agent_for_task(&task).await;
                        if let Some(a) = agent {
                            subtasks.push((a, task));
                        }
                    }
                    
                    // 并行执行
                    let handles: Vec<_> = subtasks.into_iter()
                        .map(|(agent, task)| {
                            let agent = agent.clone();
                            tokio::spawn(async move {
                                match task {
                                    CollaborationTask::Single(msg) => {
                                        agent.agent.chat_sync(vec![msg], ChatOptions::default()).await
                                    }
                                    _ => Ok(Message::assistant("Complex task not supported")),
                                }
                            })
                        })
                        .collect();
                    
                    let mut outputs = Vec::new();
                    for handle in handles {
                        if let Ok(result) = handle.await {
                            if let Ok(msg) = result {
                                outputs.push(msg);
                            }
                        }
                    }
                    
                    results.insert(agent_info.id.clone(), CollaborationResult::Parallel(outputs));
                }
                CollaborationTask::Sequential(tasks) => {
                    let mut outputs = Vec::new();
                    let mut current_agent = Some(agent_info.clone());
                    
                    for task in tasks {
                        if let Some(a) = current_agent {
                            match task {
                                CollaborationTask::Single(msg) => {
                                    let response = a.agent.chat_sync(
                                        vec![msg],
                                        ChatOptions::default(),
                                    ).await?;
                                    outputs.push(response);
                                }
                                _ => {}
                            }
                            current_agent = self.find_agent_for_task(&task).await;
                        }
                    }
                    
                    results.insert(agent_info.id.clone(), CollaborationResult::Sequential(outputs));
                }
            }
        }
        
        Ok(CollaborationResult::Combined(results))
    }

    /// 查找适合任务的代理
    async fn find_agent_for_task(&self, task: &CollaborationTask) -> Option<AgentInfo> {
        let agents = self.agents.read().await;
        
        // 简单策略：返回第一个可用的代理
        agents.values().next().cloned()
    }

    /// 移除代理
    pub async fn remove_agent(&self, id: &str) -> bool {
        let mut agents = self.agents.write().await;
        agents.remove(id).is_some()
    }

    /// 获取代理数量
    pub async fn count(&self) -> usize {
        let agents = self.agents.read().await;
        agents.len()
    }
}

impl Default for MultiAgentManager {
    fn default() -> Self {
        Self::new()
    }
}

/// 协作任务
#[derive(Debug, Clone)]
pub enum CollaborationTask {
    /// 单个任务
    Single(Message),
    /// 并行任务
    Parallel(Vec<CollaborationTask>),
    /// 顺序任务
    Sequential(Vec<CollaborationTask>),
}

/// 协作结果
#[derive(Debug, Clone)]
pub enum CollaborationResult {
    /// 单个结果
    Single(Message),
    /// 并行结果
    Parallel(Vec<Message>),
    /// 顺序结果
    Sequential(Vec<Message>),
    /// 组合结果
    Combined(HashMap<String, CollaborationResult>),
}

/// 协作消息
#[derive(Debug, Clone)]
pub struct CollaborationMessage {
    pub from: String,
    pub to: Option<String>,
    pub content: String,
    pub message_type: CollaborationMessageType,
}

/// 协作消息类型
#[derive(Debug, Clone)]
pub enum CollaborationMessageType {
    Request,
    Response,
    Status,
    Error,
}

/// 代理通信通道
pub struct AgentChannel {
    pub from: String,
    pub to: String,
    pub sender: mpsc::Sender<CollaborationMessage>,
}

impl AgentChannel {
    pub async fn send(&self, content: String) -> Result<()> {
        self.sender.send(CollaborationMessage {
            from: self.from.clone(),
            to: Some(self.to.clone()),
            content,
            message_type: CollaborationMessageType::Response,
        }).await?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_multi_agent_manager() {
        let manager = MultiAgentManager::new();
        
        assert_eq!(manager.count().await, 0);
    }

    #[test]
    fn test_agent_role_default_prompt() {
        assert!(!AgentRole::Coordinator.default_prompt().is_empty());
        assert!(!AgentRole::Researcher.default_prompt().is_empty());
    }
}
