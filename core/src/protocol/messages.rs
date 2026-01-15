//! Protocol message definitions

use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

/// Protocol version
pub const PROTOCOL_VERSION: &str = "1.0.0";

/// Protocol magic bytes
pub const PROTOCOL_MAGIC: &[u8; 4] = b"TORT";

/// Message flags
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Flags(u16);

impl Flags {
    pub const COMPRESSED: Self = Flags(1 << 0);
    pub const ENCRYPTED: Self = Flags(1 << 1);
    pub const STREAMING: Self = Flags(1 << 2);
    pub const MULTIPART: Self = Flags(1 << 3);
    
    #[inline]
    pub fn is_compressed(&self) -> bool {
        self.0 & Self::COMPRESSED.0 != 0
    }
    
    #[inline]
    pub fn is_encrypted(&self) -> bool {
        self.0 & Self::ENCRYPTED.0 != 0
    }
    
    #[inline]
    pub fn is_streaming(&self) -> bool {
        self.0 & Self::STREAMING.0 != 0
    }
}

/// Protocol message frame
#[derive(Debug, Clone)]
pub struct MessageFrame {
    pub version: u16,
    pub msg_type: u16,
    pub flags: Flags,
    pub payload: Vec<u8>,
}

impl MessageFrame {
    pub fn new(msg_type: u16, payload: Vec<u8>) -> Self {
        Self {
            version: 1,
            msg_type,
            flags: Flags::default(),
            payload,
        }
    }
    
    pub fn with_flags(mut self, flags: Flags) -> Self {
        self.flags = flags;
        self
    }
}

impl Default for Flags {
    fn default() -> Self {
        Flags(0)
    }
}

/// Handshake request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandshakeRequest {
    pub client_id: String,
    pub client_version: String,
    pub protocol_version: String,
    pub auth_token: Option<String>,
    pub capabilities: Vec<String>,
}

/// Handshake response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandshakeResponse {
    pub server_version: String,
    pub session_id: String,
    pub server_capabilities: Vec<String>,
    pub config: serde_json::Value,
}

/// Agent request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentRequest {
    pub request_id: String,
    pub session_id: String,
    pub content: String,
    #[serde(default)]
    pub content_type: ContentType,
    pub metadata: RequestMetadata,
    #[serde(default]
    pub tools: Vec<ToolDefinition>,
    pub context: RequestContext,
}

impl AgentRequest {
    pub fn new(session_id: String, content: String) -> Self {
        Self {
            request_id: Uuid::new_v4().to_string(),
            session_id,
            content,
            content_type: ContentType::Text,
            metadata: RequestMetadata::default(),
            tools: Vec::new(),
            context: RequestContext::default(),
        }
    }
}

/// Content type enumeration
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ContentType {
    #[default]
    Text,
    Image,
    Audio,
    Video,
    File,
}

/// Request metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestMetadata {
    pub user_id: Option<String>,
    pub channel: Option<String>,
    #[serde(default)]
    pub timestamp: DateTime<Utc>,
}

impl Default for RequestMetadata {
    fn default() -> Self {
        Self {
            user_id: None,
            channel: None,
            timestamp: Utc::now(),
        }
    }
}

/// Tool definition for requests
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolDefinition {
    pub name: String,
    pub description: String,
    #[serde(default)]
    pub parameters: serde_json::Value,
}

/// Request context
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestContext {
    #[serde(default)]
    pub system_prompt: Option<String>,
    #[serde(default)]
    pub conversation_history: Vec<Message>,
    #[serde(default)]
    pub attachments: Vec<Attachment>,
}

impl Default for RequestContext {
    fn default() -> Self {
        Self {
            system_prompt: None,
            conversation_history: Vec::new(),
            attachments: Vec::new(),
        }
    }
}

/// Message in conversation history
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub role: Role,
    pub content: String,
    #[serde(default)]
    pub attachments: Vec<Attachment>,
    #[serde(default)]
    pub timestamp: DateTime<Utc>,
}

/// Message role
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Role {
    User,
    Assistant,
    System,
}

/// Attachment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attachment {
    pub r#type: ContentType,
    pub url: String,
    pub mime_type: Option<String>,
    #[serde(default)]
    pub size: Option<u64>,
}

/// Agent response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentResponse {
    pub request_id: String,
    pub session_id: String,
    pub content: String,
    #[serde(default)]
    pub content_type: ContentType,
    #[serde(default)]
    pub tool_calls: Vec<ToolCall>,
    pub metadata: ResponseMetadata,
}

/// Response metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResponseMetadata {
    pub model: String,
    pub tokens: TokenUsage,
    #[serde(default)]
    pub latency_ms: Option<u64>,
}

/// Token usage
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenUsage {
    pub prompt: u32,
    pub completion: u32,
}

/// Tool call
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    #[serde(default)]
    pub arguments: serde_json::Value,
    #[serde(default)]
    pub timeout_ms: Option<u64>,
}

impl ToolCall {
    pub fn new(name: String, arguments: serde_json::Value) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            arguments,
            timeout_ms: None,
        }
    }
}

/// Tool result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    pub call_id: String,
    pub success: bool,
    #[serde(default)]
    pub result: serde_json::Value,
    #[serde(default)]
    pub error: Option<String>,
    pub execution_time_ms: u64,
}

/// Error message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorMessage {
    pub code: String,
    pub message: String,
    #[serde(default)]
    pub details: serde_json::Value,
    pub request_id: Option<String>,
}

impl ErrorMessage {
    pub fn new(code: String, message: String) -> Self {
        Self {
            code,
            message,
            details: serde_json::Value::Null,
            request_id: None,
        }
    }
    
    pub fn with_details(mut self, details: serde_json::Value) -> Self {
        self.details = details;
        self
    }
    
    pub fn with_request_id(mut self, request_id: String) -> Self {
        self.request_id = Some(request_id);
        self
    }
}

/// Stream chunk
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamChunk {
    pub request_id: String,
    pub delta: String,
    #[serde(default)]
    pub is_final: bool,
}

/// Heartbeat
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Heartbeat {
    pub timestamp: DateTime<Utc>,
    #[serde(default)]
    pub latency_ms: Option<u64>,
}

impl Default for Heartbeat {
    fn default() -> Self {
        Self {
            timestamp: Utc::now(),
            latency_ms: None,
        }
    }
}
