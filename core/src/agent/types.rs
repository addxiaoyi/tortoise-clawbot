//! Message Types
//! 
//! Core message structures for agent communication.

use serde::{Deserialize, Serialize};

/// Message role
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MessageRole {
    /// System message
    System,
    /// User message
    User,
    /// Assistant message
    Assistant,
    /// Tool message
    Tool,
}

/// Chat message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    /// Message role
    pub role: MessageRole,
    /// Message content
    pub content: String,
    /// Tool calls (for assistant messages with function calls)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_calls: Option<Vec<ToolCall>>,
    /// Tool results (for messages containing tool responses)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_results: Option<Vec<ToolResult>>,
}

impl Message {
    /// Create a new user message
    pub fn user(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::User,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
        }
    }

    /// Create a new assistant message
    pub fn assistant(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::Assistant,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
        }
    }

    /// Create a new system message
    pub fn system(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::System,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
        }
    }

    /// Create a new tool message
    pub fn tool(content: impl Into<String>, tool_call_id: &str) -> Self {
        Self {
            role: MessageRole::Tool,
            content: content.into(),
            tool_calls: None,
            tool_results: Some(vec![ToolResult {
                call_id: tool_call_id.to_string(),
                result: Ok(content.into()),
            }]),
        }
    }
}

/// Tool call
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    /// Call ID
    pub id: String,
    /// Tool name
    pub name: String,
    /// Arguments as JSON
    pub arguments: serde_json::Value,
}

impl ToolCall {
    /// Create a new tool call
    pub fn new(name: impl Into<String>, arguments: serde_json::Value) -> Self {
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            name: name.into(),
            arguments,
        }
    }

    /// Parse arguments as a specific type
    pub fn parse_args<T: serde::de::DeserializeOwned>(&self) -> Option<T> {
        serde_json::from_value(self.arguments.clone()).ok()
    }
}

/// Tool result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    /// Call ID this result is for
    pub call_id: String,
    /// Result (Ok) or error (Err)
    pub result: Result<String, String>,
}

impl ToolResult {
    /// Create a successful result
    pub fn success(call_id: impl Into<String>, content: impl Into<String>) -> Self {
        Self {
            call_id: call_id.into(),
            result: Ok(content.into()),
        }
    }

    /// Create an error result
    pub fn error(call_id: impl Into<String>, error: impl Into<String>) -> Self {
        Self {
            call_id: call_id.into(),
            result: Err(error.into()),
        }
    }
}

/// Chat options
#[derive(Debug, Clone, Default)]
pub struct ChatOptions {
    /// Thinking mode override
    pub thinking_mode: Option<ThinkMode>,
    /// System prompt override
    pub system_prompt: Option<String>,
    /// Available tools
    pub tools: Option<Vec<Box<dyn crate::tool::Tool>>>,
    /// Max tokens
    pub max_tokens: Option<usize>,
    /// Temperature
    pub temperature: Option<f32>,
    /// Stop sequences
    pub stop: Option<Vec<String>>,
}

impl ChatOptions {
    /// Create new options
    pub fn new() -> Self {
        Self::default()
    }

    /// Set thinking mode
    pub fn with_thinking_mode(mut self, mode: ThinkMode) -> Self {
        self.thinking_mode = Some(mode);
        self
    }

    /// Set system prompt
    pub fn with_system_prompt(mut self, prompt: impl Into<String>) -> Self {
        self.system_prompt = Some(prompt.into());
        self
    }

    /// Set temperature
    pub fn with_temperature(mut self, temp: f32) -> Self {
        self.temperature = Some(temp);
        self
    }

    /// Set max tokens
    pub fn with_max_tokens(mut self, max: usize) -> Self {
        self.max_tokens = Some(max);
        self
    }
}

pub use crate::agent::ThinkMode;
