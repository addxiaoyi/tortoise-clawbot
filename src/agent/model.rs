//! Model abstractions and utilities

use super::*;

/// Model capabilities
#[derive(Debug, Clone, Default)]
pub struct ModelCapabilities {
    /// Maximum context length
    pub max_context: usize,
    /// Supports function calling
    pub supports_function_calling: bool,
    /// Supports streaming
    pub supports_streaming: bool,
    /// Supports vision/images
    pub supports_vision: bool,
    /// Supports audio input
    pub supports_audio_input: bool,
    /// Supports audio output
    pub supports_audio_output: bool,
    /// Maximum output tokens
    pub max_output_tokens: usize,
}

/// Known model configurations
pub mod models {
    use super::*;

    pub fn gpt_4() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 128_000,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: true,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 8192,
        }
    }

    pub fn gpt_4_turbo() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 128_000,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: true,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn gpt_35_turbo() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 16_385,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: false,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn claude_3_opus() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 200_000,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: true,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn claude_3_sonnet() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 200_000,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: true,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn claude_3_haiku() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 200_000,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: true,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn llama_3_70b() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 8192,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: false,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    pub fn llama_3_8b() -> ModelCapabilities {
        ModelCapabilities {
            max_context: 8192,
            supports_function_calling: true,
            supports_streaming: true,
            supports_vision: false,
            supports_audio_input: false,
            supports_audio_output: false,
            max_output_tokens: 4096,
        }
    }

    /// Get capabilities for a model name
    pub fn get_capabilities(model_name: &str) -> ModelCapabilities {
        let model_lower = model_name.to_lowercase();
        
        if model_lower.contains("gpt-4-32k") || model_lower == "gpt-4" {
            gpt_4()
        } else if model_lower.contains("gpt-4-turbo") || model_lower.contains("gpt-4o") {
            gpt_4_turbo()
        } else if model_lower.contains("gpt-3.5-turbo") {
            gpt_35_turbo()
        } else if model_lower.contains("claude-3-opus") {
            claude_3_opus()
        } else if model_lower.contains("claude-3-sonnet") {
            claude_3_sonnet()
        } else if model_lower.contains("claude-3-haiku") {
            claude_3_haiku()
        } else if model_lower.contains("llama-3-70b") || model_lower.contains("llama3-70b") {
            llama_3_70b()
        } else if model_lower.contains("llama-3-8b") || model_lower.contains("llama3-8b") {
            llama_3_8b()
        } else {
            // Default capabilities
            ModelCapabilities::default()
        }
    }
}

/// Message builder utility
pub struct MessageBuilder {
    role: MessageRole,
    content: String,
    tool_calls: Option<Vec<ToolCall>>,
    tool_results: Option<Vec<ToolResult>>,
    metadata: MessageMetadata,
}

impl MessageBuilder {
    /// Create a new builder for a user message
    pub fn user(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::User,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
            metadata: MessageMetadata::default(),
        }
    }

    /// Create a new builder for a system message
    pub fn system(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::System,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
            metadata: MessageMetadata::default(),
        }
    }

    /// Create a new builder for an assistant message
    pub fn assistant(content: impl Into<String>) -> Self {
        Self {
            role: MessageRole::Assistant,
            content: content.into(),
            tool_calls: None,
            tool_results: None,
            metadata: MessageMetadata::default(),
        }
    }

    /// Create a new builder for a tool message
    pub fn tool(content: impl Into<String>, call_id: impl Into<String>) -> Self {
        Self {
            role: MessageRole::Tool,
            content: content.into(),
            tool_calls: None,
            tool_results: Some(vec![ToolResult {
                call_id: call_id.into(),
                result: Ok(String::new()),
            }]),
            metadata: MessageMetadata::default(),
        }
    }

    /// Add tool calls
    pub fn with_tool_calls(mut self, calls: Vec<ToolCall>) -> Self {
        self.tool_calls = Some(calls);
        self
    }

    /// Add tool results
    pub fn with_tool_results(mut self, results: Vec<ToolResult>) -> Self {
        self.tool_results = Some(results);
        self
    }

    /// Add metadata
    pub fn with_metadata(mut self, metadata: MessageMetadata) -> Self {
        self.metadata = metadata;
        self
    }

    /// Set sender ID
    pub fn sender_id(mut self, id: impl Into<String>) -> Self {
        self.metadata.sender_id = Some(id.into());
        self
    }

    /// Set channel
    pub fn channel(mut self, channel: impl Into<String>) -> Self {
        self.metadata.channel = Some(channel.into());
        self
    }

    /// Build the message
    pub fn build(self) -> Message {
        Message {
            role: self.role,
            content: self.content,
            tool_calls: self.tool_calls,
            tool_results: self.tool_results,
            metadata: self.metadata,
        }
    }
}

impl From<String> for Message {
    fn from(content: String) -> Self {
        MessageBuilder::user(content).build()
    }
}

impl From<&str> for Message {
    fn from(content: &str) -> Self {
        MessageBuilder::user(content.to_string()).build()
    }
}
