//! 协议层模块
//! 
//! 实现 Tortoise Protocol 和 OpenClaw 协议兼容

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Tortoise 消息类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TortoiseMessage {
    pub id: String,
    pub session_id: String,
    pub role: MessageRole,
    pub content: String,
    pub msg_type: MessageType,
    pub timestamp: i64,
    pub metadata: HashMap<String, String>,
    pub tool_calls: Vec<ToolCall>,
}

/// 消息角色
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum MessageRole {
    System,
    User,
    Assistant,
    Tool,
}

/// 消息类型
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

/// 工具调用
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub tool_name: String,
    pub arguments: HashMap<String, serde_json::Value>,
    pub state: ToolCallState,
}

/// 工具调用状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ToolCallState {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
}

/// 会话信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TortoiseSession {
    pub id: String,
    pub user_id: String,
    pub state: SessionState,
    pub created_at: i64,
    pub updated_at: i64,
    pub metadata: HashMap<String, String>,
}

/// 会话状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum SessionState {
    Active,
    Paused,
    Closed,
}

/// 协议版本
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub struct ProtocolVersion {
    pub major: u16,
    pub minor: u16,
    pub patch: u16,
}

impl ProtocolVersion {
    pub const CURRENT: Self = Self {
        major: 0,
        minor: 1,
        patch: 0,
    };
    
    pub fn to_string(&self) -> String {
        format!("{}.{}.{}", self.major, self.minor, self.patch)
    }
}

/// 协议能力
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ProtocolCapabilities {
    pub streaming: bool,
    pub tools: bool,
    pub memory: bool,
    pub multi_modal: bool,
    pub channels: Vec<String>,
}

impl ProtocolCapabilities {
    pub fn basic() -> Self {
        Self {
            streaming: true,
            tools: true,
            memory: true,
            multi_modal: false,
            channels: vec!["web".to_string()],
        }
    }
    
    pub fn full() -> Self {
        Self {
            streaming: true,
            tools: true,
            memory: true,
            multi_modal: true,
            channels: vec![
                "telegram".to_string(),
                "discord".to_string(),
                "slack".to_string(),
                "whatsapp".to_string(),
                "web".to_string(),
                "wechat".to_string(),
                "line".to_string(),
                "signal".to_string(),
                "sms".to_string(),
            ],
        }
    }
}

/// 协议握手请求
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandshakeRequest {
    pub version: ProtocolVersion,
    pub client_name: String,
    pub capabilities: ProtocolCapabilities,
}

/// 协议握手响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandshakeResponse {
    pub version: ProtocolVersion,
    pub server_name: String,
    pub capabilities: ProtocolCapabilities,
    pub session_id: Option<String>,
}

/// 错误响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorResponse {
    pub code: i32,
    pub message: String,
    pub details: Option<String>,
}

impl ErrorResponse {
    pub fn new(code: i32, message: impl Into<String>) -> Self {
        Self {
            code,
            message: message.into(),
            details: None,
        }
    }
    
    pub fn with_details(mut self, details: impl Into<String>) -> Self {
        self.details = Some(details.into());
        self
    }
}

/// 序列化工具
pub fn serialize<T: Serialize>(value: &T) -> Result<String, serde_json::Error> {
    serde_json::to_string(value)
}

pub fn deserialize<T: for<'de> Deserialize<'de>>(data: &str) -> Result<T, serde_json::Error> {
    serde_json::from_str(data)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version() {
        let version = ProtocolVersion::CURRENT;
        assert_eq!(version.to_string(), "0.1.0");
    }

    #[test]
    fn test_message_serialization() {
        let msg = TortoiseMessage {
            id: "msg123".to_string(),
            session_id: "sess456".to_string(),
            role: MessageRole::User,
            content: "Hello".to_string(),
            msg_type: MessageType::Text,
            timestamp: 1234567890,
            metadata: HashMap::new(),
            tool_calls: Vec::new(),
        };
        
        let json = serialize(&msg).unwrap();
        assert!(json.contains("Hello"));
    }
}
