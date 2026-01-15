//! 会话管理模块
//! 
//! 管理用户会话、上下文和状态

use crate::error::{Error, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use uuid::Uuid;
use chrono::{DateTime, Utc};

mod error;
pub use error::Error;

/// 会话状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum SessionState {
    Active,
    Paused,
    Closed,
}

/// 会话信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Session {
    pub id: String,
    pub user_id: String,
    pub state: SessionState,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub metadata: HashMap<String, String>,
    pub context: SessionContext,
}

impl Session {
    /// 创建新会话
    pub fn new(user_id: impl Into<String>) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            user_id: user_id.into(),
            state: SessionState::Active,
            created_at: now,
            updated_at: now,
            metadata: HashMap::new(),
            context: SessionContext::default(),
        }
    }
    
    /// 更新会话
    pub fn touch(&mut self) {
        self.updated_at = Utc::now();
    }
    
    /// 添加消息到上下文
    pub fn add_message(&mut self, message: Message) {
        self.context.add_message(message);
        self.touch();
    }
    
    /// 获取上下文消息
    pub fn get_context(&self) -> &[Message] {
        &self.context.messages
    }
    
    /// 关闭会话
    pub fn close(&mut self) {
        self.state = SessionState::Closed;
        self.touch();
    }
    
    /// 暂停会话
    pub fn pause(&mut self) {
        self.state = SessionState::Paused;
        self.touch();
    }
    
    /// 恢复会话
    pub fn resume(&mut self) {
        self.state = SessionState::Active;
        self.touch();
    }
}

/// 消息格式
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MessageType {
    Text,
    Image,
    Audio,
    Video,
    File,
    Location,
    Contact,
}

/// 消息内容格式
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ContentFormat {
    Plain,
    Markdown,
    Html,
    Json,
}

/// 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub id: String,
    pub session_id: String,
    pub role: String,
    pub content: String,
    pub format: ContentFormat,
    pub msg_type: MessageType,
    pub timestamp: DateTime<Utc>,
    pub metadata: HashMap<String, String>,
    pub parent_id: Option<String>,
}

impl Message {
    /// 创建新消息
    pub fn new(
        session_id: impl Into<String>,
        role: impl Into<String>,
        content: impl Into<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            session_id: session_id.into(),
            role: role.into(),
            content: content.into(),
            format: ContentFormat::Plain,
            msg_type: MessageType::Text,
            timestamp: Utc::now(),
            metadata: HashMap::new(),
            parent_id: None,
        }
    }
    
    /// 创建用户消息
    pub fn user(content: impl Into<String>) -> Self {
        Self::new("default", "user", content)
    }
    
    /// 创建助手消息
    pub fn assistant(session_id: impl Into<String>, content: impl Into<String>) -> Self {
        Self::new(session_id, "assistant", content)
    }
    
    /// 创建系统消息
    pub fn system(session_id: impl Into<String>, content: impl Into<String>) -> Self {
        Self::new(session_id, "system", content)
    }
}

/// 会话上下文
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SessionContext {
    pub messages: Vec<Message>,
    pub max_context: usize,
    pub total_tokens: usize,
}

impl Default for Session {
    fn default() -> Self {
        Self::new("default")
    }
}

impl Default for Message {
    fn default() -> Self {
        Self::new("default", "user", "")
    }
}

impl SessionContext {
    /// 创建新上下文
    pub fn new(max_context: usize) -> Self {
        Self {
            messages: Vec::with_capacity(max_context),
            max_context,
            total_tokens: 0,
        }
    }
    
    /// 添加消息
    pub fn add_message(&mut self, message: Message) {
        // 估算 token 数（简单按字符数 / 4 计算）
        let tokens = message.content.len() / 4;
        self.total_tokens += tokens;
        
        self.messages.push(message);
        
        // 如果超出上下文限制，移除最早的消息
        while self.messages.len() > self.max_context {
            if let Some(old) = self.messages.first() {
                self.total_tokens -= old.content.len() / 4;
                self.messages.remove(0);
            }
        }
    }
    
    /// 清空上下文
    pub fn clear(&mut self) {
        self.messages.clear();
        self.total_tokens = 0;
    }
    
    /// 获取消息数量
    pub fn len(&self) -> usize {
        self.messages.len()
    }
    
    /// 是否为空
    pub fn is_empty(&self) -> bool {
        self.messages.is_empty()
    }
}

/// 会话管理器
pub struct SessionManager {
    sessions: Arc<RwLock<HashMap<String, Session>>>,
    max_sessions: usize,
}

impl SessionManager {
    /// 创建新的会话管理器
    pub fn new(max_sessions: usize) -> Self {
        Self {
            sessions: Arc::new(RwLock::new(HashMap::new())),
            max_sessions,
        }
    }
    
    /// 创建会话
    pub fn create(&self, user_id: impl Into<String>) -> Session {
        let session = Session::new(user_id);
        let id = session.id.clone();
        
        let mut sessions = self.sessions.write();
        sessions.insert(id, session.clone());
        
        // 如果超出限制，清理最旧的会话
        while sessions.len() > self.max_sessions {
            if let Some(oldest) = sessions.values()
                .min_by_key(|s| s.updated_at)
                .map(|s| s.id.clone())
            {
                sessions.remove(&oldest);
            }
        }
        
        session
    }
    
    /// 获取会话
    pub fn get(&self, id: &str) -> Option<Session> {
        let sessions = self.sessions.read();
        sessions.get(id).cloned()
    }
    
    /// 获取会话（可变）
    pub fn get_mut(&self, id: &str) -> Option<parking_lot::RwLockWriteGuard<'_, HashMap<String, Session>>> {
        let sessions = self.sessions.write();
        if sessions.contains_key(id) {
            drop(sessions);
            Some(self.sessions.write())
        }
        None
    }
    
    /// 删除会话
    pub fn delete(&self, id: &str) -> bool {
        let mut sessions = self.sessions.write();
        sessions.remove(id).is_some()
    }
    
    /// 列出所有会话
    pub fn list(&self) -> Vec<Session> {
        let sessions = self.sessions.read();
        sessions.values().cloned().collect()
    }
    
    /// 获取会话数量
    pub fn count(&self) -> usize {
        let sessions = self.sessions.read();
        sessions.len()
    }
    
    /// 清理过期会话
    pub fn cleanup_expired(&self, max_idle_secs: i64) -> usize {
        let now = Utc::now();
        let mut sessions = self.sessions.write();
        let before = sessions.len();
        
        sessions.retain(|_, session| {
            let duration = now.signed_duration_since(session.updated_at);
            duration.num_seconds() < max_idle_secs
        });
        
        before - sessions.len()
    }
}

impl Default for SessionManager {
    fn default() -> Self {
        Self::new(10000) // 默认最大 10000 会话
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_create_session() {
        let manager = SessionManager::new(100);
        let session = manager.create("user123");
        
        assert_eq!(session.user_id, "user123");
        assert_eq!(session.state, SessionState::Active);
        assert!(!session.id.is_empty());
    }
    
    #[test]
    fn test_get_session() {
        let manager = SessionManager::new(100);
        let session = manager.create("user123");
        let id = session.id.clone();
        
        let retrieved = manager.get(&id);
        assert!(retrieved.is_some());
        assert_eq!(retrieved.unwrap().user_id, "user123");
    }
    
    #[test]
    fn test_delete_session() {
        let manager = SessionManager::new(100);
        let session = manager.create("user123");
        let id = session.id.clone();
        
        assert!(manager.delete(&id));
        assert!(manager.get(&id).is_none());
    }
    
    #[test]
    fn test_context_management() {
        let mut ctx = SessionContext::new(3);
        
        ctx.add_message(Message::user("Hello"));
        ctx.add_message(Message::assistant("default", "Hi"));
        ctx.add_message(Message::user("How are you?"));
        ctx.add_message(Message::assistant("default", "I'm fine"));
        
        // 超出限制，应该移除最早的消息
        assert_eq!(ctx.len(), 3);
    }
}
