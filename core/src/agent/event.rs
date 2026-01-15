//! Agent Events
//! 
//! Event types for agent streaming and status updates.

use crate::agent::ThinkMode;
use flume::Receiver;
use serde::{Deserialize, Serialize};

/// Agent event types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data", rename_all = "snake_case")]
pub enum AgentEvent {
    /// Thinking started
    ThinkingStarted {
        mode: ThinkMode,
    },
    /// Thinking in progress
    Thinking {
        content: String,
    },
    /// Tool call initiated
    ToolCall {
        call_id: String,
        tool: String,
        arguments: serde_json::Value,
    },
    /// Tool result received
    ToolResult {
        call_id: String,
        success: bool,
        result: String,
        error: Option<String>,
    },
    /// Content chunk received
    ContentChunk {
        content: String,
    },
    /// Response completed
    ResponseComplete {
        content: String,
        finish_reason: Option<String>,
    },
    /// Error occurred
    Error {
        error: String,
    },
    /// Agent status update
    StatusUpdate {
        status: String,
    },
}

/// Streaming response wrapper
pub struct StreamingResponse {
    receiver: Receiver<AgentEvent>,
}

impl StreamingResponse {
    /// Create a new streaming response
    pub fn new(receiver: Receiver<AgentEvent>) -> Self {
        Self { receiver }
    }

    /// Get the receiver
    pub fn into_inner(self) -> Receiver<AgentEvent> {
        self.receiver
    }

    /// Get receiver reference
    pub fn receiver(&self) -> &Receiver<AgentEvent> {
        &self.receiver
    }
}

/// Chat response (non-streaming)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatResponse {
    /// Response ID
    pub id: String,
    /// Response content
    pub content: String,
    /// Finish reason
    pub finish_reason: String,
    /// Token usage
    pub usage: Usage,
}

/// Token usage information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Usage {
    /// Prompt tokens
    pub prompt_tokens: u32,
    /// Completion tokens
    pub completion_tokens: u32,
    /// Total tokens
    pub total_tokens: u32,
}

/// Tool call
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    /// Call ID
    pub id: String,
    /// Tool name
    pub name: String,
    /// Arguments
    pub arguments: serde_json::Value,
}

impl ToolCall {
    /// Create a new tool call
    pub fn new(name: String, arguments: serde_json::Value) -> Self {
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            name,
            arguments,
        }
    }
}

/// Tool result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    /// Call ID
    pub call_id: String,
    /// Result content
    pub result: String,
    /// Whether the call succeeded
    pub success: bool,
}

/// Agent status
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AgentStatus {
    /// Agent is idle
    Idle,
    /// Agent is processing
    Busy,
    /// Agent encountered an error
    Error(String),
}

impl Default for AgentStatus {
    fn default() -> Self {
        AgentStatus::Idle
    }
}
