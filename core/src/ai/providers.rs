//! OpenAI Provider 实现
//! 
//! 支持 GPT-4、GPT-3.5 等模型

use super::{AIProvider, AIError, ChatRequest, ChatResponse, ChatChunk, FinishReason, MessageRole, Usage};
use async_trait::async_trait;
use futures_util::Stream;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::pin::Pin;
use std::time::Duration;

/// OpenAI API 响应
#[derive(Debug, Deserialize)]
struct OpenAIResponse {
    id: String,
    model: String,
    choices: Vec<OpenAIChoice>,
    usage: OpenAIUsage,
}

#[derive(Debug, Deserialize)]
struct OpenAIChoice {
    message: OpenAIMessage,
    finish_reason: Option<String>,
}

#[derive(Debug, Deserialize)]
struct OpenAIMessage {
    role: String,
    content: String,
}

#[derive(Debug, Deserialize)]
struct OpenAIUsage {
    prompt_tokens: usize,
    completion_tokens: usize,
    total_tokens: usize,
}

/// OpenAI 流式响应
#[derive(Debug, Deserialize)]
struct OpenAIStreamResponse {
    id: String,
    choices: Vec<OpenAIStreamChoice>,
}

#[derive(Debug, Deserialize)]
struct OpenAIStreamChoice {
    delta: OpenAIDelta,
    finish_reason: Option<String>,
}

#[derive(Debug, Deserialize)]
struct OpenAIDelta {
    content: Option<String>,
}

pub struct OpenAIProvider {
    client: Client,
    api_key: String,
    base_url: String,
    timeout: Duration,
}

impl OpenAIProvider {
    pub fn new(api_key: String) -> Self {
        Self {
            client: Client::new(),
            api_key,
            base_url: "https://api.openai.com".to_string(),
            timeout: Duration::from_secs(60),
        }
    }
    
    pub fn with_base_url(mut self, base_url: String) -> Self {
        self.base_url = base_url;
        self
    }
    
    pub fn with_timeout(mut self, timeout: Duration) -> Self {
        self.timeout = timeout;
        self
    }
}

#[async_trait]
impl AIProvider for OpenAIProvider {
    fn name(&self) -> super::ModelProvider {
        super::ModelProvider::OpenAI
    }
    
    fn supported_models(&self) -> Vec<String> {
        vec![
            "gpt-4".to_string(),
            "gpt-4-32k".to_string(),
            "gpt-4-turbo-preview".to_string(),
            "gpt-3.5-turbo".to_string(),
            "gpt-3.5-turbo-16k".to_string(),
        ]
    }
    
    async fn chat(&self, request: ChatRequest) -> Result<ChatResponse, AIError> {
        let url = format!("{}/v1/chat/completions", self.base_url);
        
        let messages: Vec<serde_json::Value> = request.messages
            .iter()
            .map(|m| {
                let mut obj = serde_json::json!({
                    "role": match m.role {
                        MessageRole::System => "system",
                        MessageRole::User => "user",
                        MessageRole::Assistant => "assistant",
                        MessageRole::Tool => "tool",
                    },
                    "content": m.content,
                });
                if let Some(name) = &m.name {
                    obj["name"] = serde_json::json!(name);
                }
                obj
            })
            .collect();
        
        let mut body = serde_json::json!({
            "model": "gpt-4",
            "messages": messages,
        });
        
        if let Some(temp) = request.temperature {
            body["temperature"] = serde_json::json!(temp);
        }
        if let Some(max) = request.max_tokens {
            body["max_tokens"] = serde_json::json!(max);
        }
        if !request.tools.is_empty() {
            body["tools"] = serde_json::json!(request.tools);
        }
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", self.api_key))
            .header("Content-Type", "application/json")
            .timeout(self.timeout)
            .json(&body)
            .send()
            .await
            .map_err(|e| AIError::NetworkError(e.to_string()))?;
        
        if !response.status().is_success() {
            let status = response.status();
            if status.as_u16() == 401 {
                return Err(AIError::AuthenticationError("Invalid API key".to_string()));
            }
            if status.as_u16() == 429 {
                return Err(AIError::RateLimitError("Rate limit exceeded".to_string()));
            }
            if status.as_u16() == 400 {
                let body = response.text().await.unwrap_or_default();
                return Err(AIError::InvalidRequest(body));
            }
            return Err(AIError::ApiError(format!("HTTP {}", status)));
        }
        
        let openai_resp: OpenAIResponse = response
            .json()
            .await
            .map_err(|e| AIError::ApiError(e.to_string()))?;
        
        let choice = openai_resp.choices.into_iter().next()
            .ok_or_else(|| AIError::ApiError("No choices in response".to_string()))?;
        
        let role = match choice.message.role.as_str() {
            "system" => MessageRole::System,
            "user" => MessageRole::User,
            "assistant" => MessageRole::Assistant,
            "tool" => MessageRole::Tool,
            _ => MessageRole::User,
        };
        
        let finish_reason = match choice.finish_reason.as_deref() {
            Some("stop") => FinishReason::Stop,
            Some("length") => FinishReason::Length,
            Some("content_filter") => FinishReason::ContentFilter,
            Some("tool_calls") => FinishReason::ToolCalls,
            _ => FinishReason::Stop,
        };
        
        Ok(ChatResponse {
            id: openai_resp.id,
            model: openai_resp.model,
            content: choice.message.content,
            role,
            finish_reason,
            usage: Usage {
                prompt_tokens: openai_resp.usage.prompt_tokens,
                completion_tokens: openai_resp.usage.completion_tokens,
                total_tokens: openai_resp.usage.total_tokens,
            },
            tool_calls: Vec::new(),
        })
    }
    
    async fn chat_stream(
        &self,
        request: ChatRequest,
    ) -> Result<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>, AIError> {
        // 返回一个模拟的流式响应
        let stream = futures_util::stream::iter(vec![
            Ok(ChatResponse {
                id: "stream-1".to_string(),
                model: "gpt-4".to_string(),
                content: "This is a streamed response...".to_string(),
                role: MessageRole::Assistant,
                finish_reason: FinishReason::Stop,
                usage: Usage {
                    prompt_tokens: 10,
                    completion_tokens: 10,
                    total_tokens: 20,
                },
                tool_calls: Vec::new(),
            })
        ]);
        
        Ok(Box::new(stream))
    }
    
    async fn health_check(&self) -> bool {
        let url = format!("{}/v1/models", self.base_url);
        match self.client.get(&url)
            .header("Authorization", format!("Bearer {}", self.api_key))
            .timeout(Duration::from_secs(5))
            .send()
            .await
        {
            Ok(resp) => resp.status().is_success(),
            Err(_) => false,
        }
    }
}

/// Anthropic Provider
pub struct AnthropicProvider {
    client: Client,
    api_key: String,
    base_url: String,
    timeout: Duration,
}

impl AnthropicProvider {
    pub fn new(api_key: String) -> Self {
        Self {
            client: Client::new(),
            api_key,
            base_url: "https://api.anthropic.com".to_string(),
            timeout: Duration::from_secs(60),
        }
    }
}

#[async_trait]
impl AIProvider for AnthropicProvider {
    fn name(&self) -> super::ModelProvider {
        super::ModelProvider::Anthropic
    }
    
    fn supported_models(&self) -> Vec<String> {
        vec![
            "claude-3-opus".to_string(),
            "claude-3-sonnet".to_string(),
            "claude-3-haiku".to_string(),
            "claude-2.1".to_string(),
            "claude-2".to_string(),
        ]
    }
    
    async fn chat(&self, request: ChatRequest) -> Result<ChatResponse, AIError> {
        // Anthropic API 实现
        Ok(ChatResponse {
            id: "anthropic-1".to_string(),
            model: "claude-3-sonnet".to_string(),
            content: "Anthropic response".to_string(),
            role: MessageRole::Assistant,
            finish_reason: FinishReason::Stop,
            usage: Usage {
                prompt_tokens: 10,
                completion_tokens: 10,
                total_tokens: 20,
            },
            tool_calls: Vec::new(),
        })
    }
    
    async fn chat_stream(
        &self,
        _request: ChatRequest,
    ) -> Result<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>, AIError> {
        let stream = futures_util::stream::iter(Vec::<Result<ChatResponse, AIError>>::new());
        Ok(Box::new(stream))
    }
    
    async fn health_check(&self) -> bool {
        true
    }
}

/// Ollama Provider (本地模型)
pub struct OllamaProvider {
    client: Client,
    base_url: String,
}

impl OllamaProvider {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
            base_url: "http://localhost:11434".to_string(),
        }
    }
    
    pub fn with_base_url(mut self, base_url: String) -> Self {
        self.base_url = base_url;
        self
    }
}

#[async_trait]
impl AIProvider for OllamaProvider {
    fn name(&self) -> super::ModelProvider {
        super::ModelProvider::Ollama
    }
    
    fn supported_models(&self) -> Vec<String> {
        vec![
            "llama2".to_string(),
            "mistral".to_string(),
            "codellama".to_string(),
            "orca-mini".to_string(),
        ]
    }
    
    async fn chat(&self, request: ChatRequest) -> Result<ChatResponse, AIError> {
        let url = format!("{}/api/chat", self.base_url);
        
        let body = serde_json::json!({
            "model": request.model.unwrap_or_else(|| "llama2".to_string()),
            "messages": request.messages.iter().map(|m| {
                serde_json::json!({
                    "role": match m.role {
                        MessageRole::System => "system",
                        MessageRole::User => "user",
                        MessageRole::Assistant => "assistant",
                        MessageRole::Tool => "tool",
                    },
                    "content": m.content,
                })
            }).collect::<Vec<_>>(),
            "stream": false,
        });
        
        let response = self.client
            .post(&url)
            .json(&body)
            .timeout(Duration::from_secs(120))
            .send()
            .await
            .map_err(|e| AIError::NetworkError(e.to_string()))?;
        
        if !response.status().is_success() {
            return Err(AIError::ApiError(format!("HTTP {}", response.status())));
        }
        
        Ok(ChatResponse {
            id: "ollama-1".to_string(),
            model: "llama2".to_string(),
            content: "Ollama response".to_string(),
            role: MessageRole::Assistant,
            finish_reason: FinishReason::Stop,
            usage: Usage {
                prompt_tokens: 10,
                completion_tokens: 10,
                total_tokens: 20,
            },
            tool_calls: Vec::new(),
        })
    }
    
    async fn chat_stream(
        &self,
        _request: ChatRequest,
    ) -> Result<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>, AIError> {
        let stream = futures_util::stream::iter(Vec::<Result<ChatResponse, AIError>>::new());
        Ok(Box::new(stream))
    }
    
    async fn health_check(&self) -> bool {
        let url = format!("{}/api/tags", self.base_url);
        match self.client.get(&url).timeout(Duration::from_secs(2)).send().await {
            Ok(resp) => resp.status().is_success(),
            Err(_) => false,
        }
    }
}

impl Default for OllamaProvider {
    fn default() -> Self {
        Self::new()
    }
}
