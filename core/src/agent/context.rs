//! 上下文管理
//!
//! 管理对话上下文和消息历史

use anyhow::Result;
use std::sync::Arc;
use tokio::sync::RwLock;

use super::engine::Message;
use super::thinking::ThinkResult;

/// 上下文配置
#[derive(Debug, Clone)]
pub struct ContextConfig {
    /// 最大上下文长度 (tokens)
    pub max_tokens: usize,
    /// 最大消息数
    pub max_messages: usize,
    /// 是否保留系统消息
    pub preserve_system: bool,
    /// 摘要阈值 (当超过此长度时生成摘要)
    pub summary_threshold: usize,
}

impl Default for ContextConfig {
    fn default() -> Self {
        Self {
            max_tokens: 128_000,
            max_messages: 100,
            preserve_system: true,
            summary_threshold: 100_000,
        }
    }
}

/// 代理上下文
#[derive(Debug, Clone)]
pub struct AgentContext {
    /// 会话 ID
    pub session_id: String,
    /// 消息历史
    messages: RwLock<Vec<Message>>,
    /// 当前思维结果
    current_think: RwLock<Option<ThinkResult>>,
    /// 上下文配置
    config: ContextConfig,
}

impl AgentContext {
    /// 创建新上下文
    pub fn new(session_id: String, config: ContextConfig) -> Self {
        Self {
            session_id,
            messages: RwLock::new(Vec::new()),
            current_think: RwLock::new(None),
            config,
        }
    }

    /// 添加消息
    pub async fn add_message(&self, message: Message) {
        let mut messages = self.messages.write().await;
        messages.push(message);
        
        // 裁剪旧消息
        self.trim_messages(&mut messages).await;
    }

    /// 获取消息
    pub async fn get_messages(&self) -> Vec<Message> {
        self.messages.read().await.clone()
    }

    /// 设置当前思维结果
    pub async fn set_think(&self, result: ThinkResult) {
        *self.current_think.write().await = Some(result);
    }

    /// 获取当前思维结果
    pub async fn get_think(&self) -> Option<ThinkResult> {
        self.current_think.read().await.clone()
    }

    /// 清除上下文
    pub async fn clear(&self) {
        let mut messages = self.messages.write().await;
        messages.clear();
        *self.current_think.write().await = None;
    }

    /// 裁剪消息
    async fn trim_messages(&self, messages: &mut Vec<Message>) {
        // 保留系统消息
        let system_messages: Vec<Message> = messages
            .iter()
            .filter(|m| m.role == super::engine::MessageRole::System)
            .cloned()
            .collect();
        
        // 计算消息数
        if messages.len() > self.config.max_messages {
            let to_remove = messages.len() - self.config.max_messages;
            messages.drain(0..to_remove);
            
            // 确保系统消息存在
            if !messages.iter().any(|m| m.role == super::engine::MessageRole::System) {
                messages.splice(0..0, system_messages);
            }
        }
    }
}

/// 上下文管理器
pub struct ContextManager {
    contexts: RwLock<std::collections::HashMap<String, Arc<AgentContext>>>,
    config: ContextConfig,
    default_session_id: String,
}

impl ContextManager {
    /// 创建新管理器
    pub fn new(max_tokens: usize) -> Self {
        Self {
            contexts: RwLock::new(std::collections::HashMap::new()),
            config: ContextConfig {
                max_tokens,
                ..Default::default()
            },
            default_session_id: "default".to_string(),
        }
    }

    /// 获取或创建上下文
    pub async fn get_context(&self, session_id: Option<&str>) -> Arc<AgentContext> {
        let id = session_id.unwrap_or(&self.default_session_id);
        
        {
            let contexts = self.contexts.read().await;
            if let Some(ctx) = contexts.get(id) {
                return Arc::clone(ctx);
            }
        }
        
        let ctx = Arc::new(AgentContext::new(
            id.to_string(),
            self.config.clone(),
        ));
        
        {
            let mut contexts = self.contexts.write().await;
            contexts.insert(id.to_string(), Arc::clone(&ctx));
        }
        
        ctx
    }

    /// 删除上下文
    pub async fn remove_context(&self, session_id: &str) {
        let mut contexts = self.contexts.write().await;
        contexts.remove(session_id);
    }

    /// 准备上下文消息
    pub fn prepare_context(
        &self,
        messages: &[Message],
        think_result: &ThinkResult,
    ) -> Result<Vec<Message>> {
        let mut context_messages = Vec::new();
        
        // 添加思维结果作为思考提示
        if !think_result.reasoning.is_empty() {
            context_messages.push(Message {
                id: uuid::Uuid::new_v4().to_string(),
                role: super::engine::MessageRole::AssistantThinking,
                content: super::engine::Content::Text(think_result.reasoning.clone()),
                tool_calls: vec![],
                tool_results: vec![],
                metadata: Default::default(),
                created_at: chrono::Utc::now(),
            });
        }
        
        // 添加历史消息
        context_messages.extend(messages.iter().cloned());
        
        Ok(context_messages)
    }

    /// 清除所有上下文
    pub async fn clear_all(&self) {
        let mut contexts = self.contexts.write().await;
        for ctx in contexts.values() {
            ctx.clear().await;
        }
    }

    /// 获取上下文统计
    pub async fn stats(&self) -> ContextStats {
        let contexts = self.contexts.read().await;
        ContextStats {
            total_sessions: contexts.len(),
            total_messages: contexts.values()
                .map(|ctx| ctx.messages.read().await.len())
                .sum(),
        }
    }
}

/// 上下文统计
#[derive(Debug, Clone, serde::Serialize)]
pub struct ContextStats {
    pub total_sessions: usize,
    pub total_messages: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_context_creation() {
        let manager = ContextManager::new(8000);
        let ctx = manager.get_context(Some("test")).await;
        
        assert_eq!(ctx.session_id, "test");
    }

    #[tokio::test]
    async fn test_add_message() {
        let manager = ContextManager::new(8000);
        let ctx = manager.get_context(None).await;
        
        let msg = Message::user("Hello");
        ctx.add_message(msg).await;
        
        let messages = ctx.get_messages().await;
        assert_eq!(messages.len(), 1);
    }
}
