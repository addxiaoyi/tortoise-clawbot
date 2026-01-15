//! 模型适配器
//!
//! 支持多种 AI 模型提供商的统一接口

use anyhow::{anyhow, Context, Result};
use async_trait::async_trait;
use futures::Stream;
use serde::{Deserialize, Serialize};
use std::pin::Pin;
use std::sync::Arc;
use tokio::sync::RwLock;

use super::engine::{Content, Message, MessageRole};

/// 模型提供商
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "provider", rename_all = "snake_case")]
pub enum ModelProvider {
    /// OpenAI (GPT-4, GPT-3.5)
    OpenAI {
        model: String,
        api_key: String,
        base_url: Option<String>,
        organization: Option<String>,
    },
    /// Anthropic (Claude)
    Anthropic {
        model: String,
        api_key: String,
        base_url: Option<String>,
        version: Option<String>,
    },
    /// Google (Gemini)
    Google {
        model: String,
        api_key: String,
        base_url: Option<String>,
    },
    /// Ollama (本地模型)
    Ollama {
        model: String,
        base_url: String,
        temperature: Option<f32>,
    },
    /// Groq
    Groq {
        model: String,
        api_key: String,
        base_url: Option<String>,
    },
    /// OpenRouter
    OpenRouter {
        model: String,
        api_key: String,
        base_url: Option<String>,
    },
    /// LM Studio
    LMStudio {
        model: String,
        base_url: String,
    },
    /// 通用 OpenAI 兼容接口
    Custom {
        name: String,
        base_url: String,
        api_key: Option<String>,
        headers: Option<std::collections::HashMap<String, String>>,
    },
}

impl ModelProvider {
    /// 获取模型名称
    pub fn model_name(&self) -> &str {
        match self {
            ModelProvider::OpenAI { model, .. } => model,
            ModelProvider::Anthropic { model, .. } => model,
            ModelProvider::Google { model, .. } => model,
            ModelProvider::Ollama { model, .. } => model,
            ModelProvider::Groq { model, .. } => model,
            ModelProvider::OpenRouter { model, .. } => model,
            ModelProvider::LMStudio { model, .. } => model,
            ModelProvider::Custom { name, .. } => name,
        }
    }

    /// 检查是否需要 API key
    pub fn requires_api_key(&self) -> bool {
        match self {
            ModelProvider::Ollama { .. } | ModelProvider::LMStudio { .. } => false,
            ModelProvider::Custom { api_key, .. } => api_key.is_some(),
            _ => true,
        }
    }
}

/// 模型响应
#[derive(Debug)]
pub enum ModelResponse {
    /// 内容块
    Content { text: String },
    /// 工具调用
    ToolCall { call: super::engine::ToolCall },
    /// 流结束
    Done,
    /// 错误
    Error(String),
}

/// 模型适配器接口
#[async_trait]
pub trait ModelAdapter: Send + Sync {
    /// 获取提供商名称
    fn provider_name(&self) -> &str;
    
    /// 获取模型名称
    fn model_name(&self) -> &str;
    
    /// 检查是否支持工具调用
    fn supports_tools(&self) -> bool;
    
    /// 检查是否支持流式
    fn supports_streaming(&self) -> bool;
    
    /// 聊天 (同步)
    async fn chat(
        &self,
        messages: &[Message],
        system_prompt: Option<&String>,
        temperature: f32,
        max_tokens: usize,
        tools: Option<&[String]>,
    ) -> Result<String>;
    
    /// 聊天 (流式)
    async fn chat_stream(
        &self,
        messages: &[Message],
        system_prompt: Option<&String>,
        temperature: f32,
        max_tokens: usize,
        tools: Option<&[String]>,
    ) -> Result<Box<dyn Stream<Item = Result<ModelResponse>> + Send + Unpin>>;
    
    /// 获取上下文窗口大小
    fn context_window(&self) -> usize;
    
    /// 获取模型信息
    async fn get_model_info(&self) -> Result<ModelInfo>;
}

/// 模型信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelInfo {
    pub id: String,
    pub provider: String,
    pub name: String,
    pub context_window: usize,
    pub supports_tools: bool,
    pub supports_streaming: bool,
    pub input_cost_per_token: Option<f64>,
    pub output_cost_per_token: Option<f64>,
}

/// 创建模型适配器
pub async fn create_model_adapter(provider: &ModelProvider) -> Result<Box<dyn ModelAdapter>> {
    match provider {
        ModelProvider::OpenAI { .. } => {
            Ok(Box::new(OpenAIAdapter::new(provider.clone()).await?))
        }
        ModelProvider::Anthropic { .. } => {
            Ok(Box::new(AnthropicAdapter::new(provider.clone()).await?))
        }
        ModelProvider::Ollama { .. } => {
            Ok(Box::new(OllamaAdapter::new(provider.clone()).await?))
        }
        ModelProvider::Groq { .. } => {
            Ok(Box::new(GroqAdapter::new(provider.clone()).await?))
        }
        ModelProvider::OpenRouter { .. } => {
            Ok(Box::new(OpenRouterAdapter::new(provider.clone()).await?))
        }
        ModelProvider::LMStudio { .. } => {
            Ok(Box::new(LMStudioAdapter::new(provider.clone()).await?))
        }
        ModelProvider::Google { .. } => {
            Ok(Box::new(GoogleAdapter::new(provider.clone()).await?))
        }
        ModelProvider::Custom { .. } => {
            Ok(Box::new(CustomAdapter::new(provider.clone()).await?))
        }
    }
}

// === OpenAI 适配器 ===

pub struct OpenAIAdapter {
    config: ModelProvider,
    client: reqwest::Client,
}

impl OpenAIAdapter {
    pub async fn new(config: ModelProvider) -> Result<Self> {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(120))
            .build()?;
        Ok(Self { config, client })
    }

    fn base_url(&self) -> &str {
        match &self.config {
            ModelProvider::OpenAI { base_url, .. } => {
                base_url.as_deref().unwrap_or("https://api.openai.com")
            }
            _ => "https://api.openai.com",
        }
    }

    fn api_key(&self) -> &str {
        match &self.config {
            ModelProvider::OpenAI { api_key, .. } => api_key,
            _ => "",
        }
    }

    fn model(&self) -> &str {
        match &self.config {
            ModelProvider::OpenAI { model, .. } => model,
            _ => "",
        }
    }
}

#[async_trait]
impl ModelAdapter for OpenAIAdapter {
    fn provider_name(&self) -> &str {
        "openai"
    }

    fn model_name(&self) -> &str {
        self.model()
    }

    fn supports_tools(&self) -> bool {
        true
    }

    fn supports_streaming(&self) -> bool {
        true
    }

    async fn chat(
        &self,
        messages: &[Message],
        _system_prompt: Option<&String>,
        temperature: f32,
        max_tokens: usize,
        tools: Option<&[String]>,
    ) -> Result<String> {
        let url = format!("{}/v1/chat/completions", self.base_url());
        
        #[derive(Serialize)]
        struct ChatRequest {
            model: String,
            messages: Vec<serde_json::Value>,
            temperature: f32,
            max_tokens: usize,
            stream: bool,
            tools: Option<Vec<serde_json::Value>>,
        }

        let mut req_messages = Vec::new();
        for msg in messages {
            let role = match msg.role {
                MessageRole::System => "system",
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                MessageRole::Tool => "tool",
                MessageRole::AssistantThinking => "assistant",
            };
            
            let content = match &msg.content {
                Content::Text(s) => s.clone(),
                Content::Structured { data, .. } => serde_json::to_string(data)?,
                Content::Multimodal { text, .. } => text.clone().unwrap_or_default(),
            };

            let mut msg_obj = serde_json::json!({
                "role": role,
                "content": content,
            });

            if !msg.tool_calls.is_empty() {
                msg_obj["tool_calls"] = serde_json::json!(msg.tool_calls);
            }
            if let Some(results) = &msg.tool_results {
                msg_obj["tool_call_id"] = serde_json::json!(results.first().map(|r| &r.call_id));
            }

            req_messages.push(msg_obj);
        }

        let mut request = ChatRequest {
            model: self.model().to_string(),
            messages: req_messages,
            temperature,
            max_tokens,
            stream: false,
            tools: None,
        };

        // TODO: 添加工具支持

        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", self.api_key()))
            .json(&request)
            .send()
            .await?
            .error_for_status()?
            .json::<serde_json::Value>()
            .await?;

        let content = response["choices"][0]["message"]["content"]
            .as_str()
            .unwrap_or("")
            .to_string();

        Ok(content)
    }

    async fn chat_stream(
        &self,
        messages: &[Message],
        _system_prompt: Option<&String>,
        temperature: f32,
        max_tokens: usize,
        tools: Option<&[String]>,
    ) -> Result<Box<dyn Stream<Item = Result<ModelResponse>> + Send + Unpin>> {
        let url = format!("{}/v1/chat/completions", self.base_url());
        
        #[derive(Serialize)]
        struct ChatRequest {
            model: String,
            messages: Vec<serde_json::Value>,
            temperature: f32,
            max_tokens: usize,
            stream: bool,
        }

        let mut req_messages = Vec::new();
        for msg in messages {
            let role = match msg.role {
                MessageRole::System => "system",
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                MessageRole::Tool => "tool",
                MessageRole::AssistantThinking => "assistant",
            };
            
            let content = match &msg.content {
                Content::Text(s) => s.clone(),
                Content::Structured { data, .. } => serde_json::to_string(data)?,
                Content::Multimodal { text, .. } => text.clone().unwrap_or_default(),
            };

            req_messages.push(serde_json::json!({
                "role": role,
                "content": content,
            }));
        }

        let request = ChatRequest {
            model: self.model().to_string(),
            messages: req_messages,
            temperature,
            max_tokens,
            stream: true,
        };

        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", self.api_key()))
            .header("Content-Type", "application/json")
            .json(&request)
            .send()
            .await?;

        let stream = response.bytes_stream();
        let model = self.model().to_string();
        let tools = tools.map(|t| t.to_vec());

        Ok(Box::new(OpenAIStream::new(stream, tools)))
    }

    fn context_window(&self) -> usize {
        match self.model() {
            "gpt-4" | "gpt-4-turbo" => 128_000,
            "gpt-4-32k" => 32_000,
            "gpt-3.5-turbo" | "gpt-3.5-turbo-16k" => 16_000,
            _ => 8_000,
        }
    }

    async fn get_model_info(&self) -> Result<ModelInfo> {
        Ok(ModelInfo {
            id: self.model().to_string(),
            provider: "openai".to_string(),
            name: self.model().to_string(),
            context_window: self.context_window(),
            supports_tools: true,
            supports_streaming: true,
            input_cost_per_token: Some(0.00001),
            output_cost_per_token: Some(0.00003),
        })
    }
}

struct OpenAIStream {
    inner: futures::stream::Map<
        reqwest::Result bytes::Bytes>,
        fn(Result<bytes::Bytes, reqwest::Error>) -> Result<String, ()>,
    >,
    tools: Option<Vec<String>>,
}

impl OpenAIStream {
    fn new(stream: impl Stream<Item = Result<bytes::Bytes, reqwest::Error>> + Send + Unpin + 'static, tools: Option<Vec<String>>) -> Self {
        Self {
            inner: stream.map(|result| {
                result.map_err(|e| tracing::error!("Stream error: {}", e))
            }),
            tools,
        }
    }
}

impl Stream for OpenAIStream {
    type Item = Result<ModelResponse>;

    fn poll_next(mut self: Pin<&mut Self>, cx: &mut std::task::Context<'_>) -> std::task::Poll<Option<Self::Item>> {
        use futures::StreamExt;
        self.inner.poll_next_unpin(cx).map(|opt| {
            opt.map(|result| {
                match result {
                    Ok(bytes) => {
                        // 解析 SSE 数据
                        let text = String::from_utf8_lossy(&bytes);
                        for line in text.lines() {
                            if let Some(data) = line.strip_prefix("data: ") {
                                if data == "[DONE]" {
                                    return Ok(ModelResponse::Done);
                                }
                                if let Ok(json) = serde_json::from_str::<serde_json::Value>(data) {
                                    if let Some(content) = json["choices"][0]["delta"]["content"].as_str() {
                                        return Ok(ModelResponse::Content { text: content.to_string() });
                                    }
                                }
                            }
                        }
                        Ok(ModelResponse::Content { text: String::new() })
                    }
                    Err(e) => Ok(ModelResponse::Error(e.to_string())),
                }
            })
        })
    }
}

// === Anthropic 适配器 ===

pub struct AnthropicAdapter {
    config: ModelProvider,
    client: reqwest::Client,
}

impl AnthropicAdapter {
    pub async fn new(config: ModelProvider) -> Result<Self> {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(120))
            .build()?;
        Ok(Self { config, client })
    }
}

#[async_trait]
impl ModelAdapter for AnthropicAdapter {
    fn provider_name(&self) -> &str {
        "anthropic"
    }

    fn model_name(&self) -> &str {
        match &self.config {
            ModelProvider::Anthropic { model, .. } => model,
            _ => "",
        }
    }

    fn supports_tools(&self) -> bool {
        true
    }

    fn supports_streaming(&self) -> bool {
        true
    }

    async fn chat(&self, messages: &[Message], system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<String> {
        let base_url = match &self.config {
            ModelProvider::Anthropic { base_url, .. } => {
                base_url.as_deref().unwrap_or("https://api.anthropic.com")
            }
            _ => "https://api.anthropic.com",
        };

        let api_key = match &self.config {
            ModelProvider::Anthropic { api_key, .. } => api_key,
            _ => "",
        };

        let model = match &self.config {
            ModelProvider::Anthropic { model, .. } => model,
            _ => "",
        };

        #[derive(Serialize)]
        struct Request<'a> {
            model: &'a str,
            messages: Vec<serde_json::Value>,
            system: Option<&'a String>,
            temperature: f32,
            max_tokens: usize,
        }

        let mut req_messages = Vec::new();
        for msg in messages {
            let role = match msg.role {
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                _ => continue,
            };

            let content = match &msg.content {
                Content::Text(s) => s.clone(),
                _ => serde_json::to_string(&msg.content)?,
            };

            req_messages.push(serde_json::json!({
                "role": role,
                "content": content,
            }));
        }

        let request = Request {
            model,
            messages: req_messages,
            system: system_prompt,
            temperature,
            max_tokens,
        };

        let response = self.client
            .post(&format!("{}/v1/messages", base_url))
            .header("x-api-key", api_key)
            .header("anthropic-version", "2023-06-01")
            .header("Content-Type", "application/json")
            .json(&request)
            .send()
            .await?
            .error_for_status()?
            .json::<serde_json::Value>()
            .await?;

        let content = response["content"][0]["text"]
            .as_str()
            .unwrap_or("")
            .to_string();

        Ok(content)
    }

    async fn chat_stream(&self, messages: &[Message], system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<Box<dyn Stream<Item = Result<ModelResponse>> + Send + Unpin>> {
        // TODO: 实现 Anthropic 流式
        let text = self.chat(messages, system_prompt, temperature, max_tokens, None).await?;
        Ok(Box::new(futures::stream::once(async { Ok(ModelResponse::Content { text }) })))
    }

    fn context_window(&self) -> usize {
        match self.model_name() {
            "claude-3-5-sonnet" | "claude-3-5-sonnet-20241022" => 200_000,
            "claude-3-opus" | "claude-3-opus-20240229" => 200_000,
            "claude-3-sonnet" | "claude-3-sonnet-20240229" => 200_000,
            "claude-3-haiku" | "claude-3-haiku-20240307" => 200_000,
            _ => 100_000,
        }
    }

    async fn get_model_info(&self) -> Result<ModelInfo> {
        Ok(ModelInfo {
            id: self.model_name().to_string(),
            provider: "anthropic".to_string(),
            name: self.model_name().to_string(),
            context_window: self.context_window(),
            supports_tools: true,
            supports_streaming: true,
            input_cost_per_token: Some(0.000003),
            output_cost_per_token: Some(0.000015),
        })
    }
}

// === Ollama 适配器 ===

pub struct OllamaAdapter {
    config: ModelProvider,
    client: reqwest::Client,
}

impl OllamaAdapter {
    pub async fn new(config: ModelProvider) -> Result<Self> {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(300))
            .build()?;
        Ok(Self { config, client })
    }

    fn base_url(&self) -> &str {
        match &self.config {
            ModelProvider::Ollama { base_url, .. } => base_url,
            _ => "http://localhost:11434",
        }
    }

    fn model(&self) -> &str {
        match &self.config {
            ModelProvider::Ollama { model, .. } => model,
            _ => "llama3",
        }
    }
}

#[async_trait]
impl ModelAdapter for OllamaAdapter {
    fn provider_name(&self) -> &str {
        "ollama"
    }

    fn model_name(&self) -> &str {
        self.model()
    }

    fn supports_tools(&self) -> bool {
        false
    }

    fn supports_streaming(&self) -> bool {
        true
    }

    async fn chat(&self, messages: &[Message], _system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<String> {
        let url = format!("{}/api/chat", self.base_url());

        #[derive(Serialize)]
        struct Request<'a> {
            model: &'a str,
            messages: Vec<serde_json::Value>,
            stream: bool,
            temperature: f32,
            options: serde_json::Value,
        }

        let mut req_messages = Vec::new();
        for msg in messages {
            let role = match msg.role {
                MessageRole::System => "system",
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                MessageRole::Tool => "tool",
                _ => "user",
            };

            let content = match &msg.content {
                Content::Text(s) => s.clone(),
                _ => serde_json::to_string(&msg.content)?,
            };

            req_messages.push(serde_json::json!({
                "role": role,
                "content": content,
            }));
        }

        let request = Request {
            model: self.model(),
            messages: req_messages,
            stream: false,
            temperature,
            options: serde_json::json!({
                "num_predict": max_tokens,
            }),
        };

        let response = self.client
            .post(&url)
            .json(&request)
            .send()
            .await?
            .error_for_status()?
            .json::<serde_json::Value>()
            .await?;

        let content = response["message"]["content"]
            .as_str()
            .unwrap_or("")
            .to_string();

        Ok(content)
    }

    async fn chat_stream(&self, messages: &[Message], _system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<Box<dyn Stream<Item = Result<ModelResponse>> + Send + Unpin>> {
        let url = format!("{}/api/chat", self.base_url());

        #[derive(Serialize)]
        struct Request<'a> {
            model: &'a str,
            messages: Vec<serde_json::Value>,
            stream: bool,
            temperature: f32,
            options: serde_json::Value,
        }

        let mut req_messages = Vec::new();
        for msg in messages {
            let role = match msg.role {
                MessageRole::System => "system",
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                MessageRole::Tool => "tool",
                _ => "user",
            };

            let content = match &msg.content {
                Content::Text(s) => s.clone(),
                _ => serde_json::to_string(&msg.content)?,
            };

            req_messages.push(serde_json::json!({
                "role": role,
                "content": content,
            }));
        }

        let request = Request {
            model: self.model(),
            messages: req_messages,
            stream: true,
            temperature,
            options: serde_json::json!({
                "num_predict": max_tokens,
            }),
        };

        let response = self.client
            .post(&url)
            .header("Content-Type", "application/json")
            .json(&request)
            .send()
            .await?;

        Ok(Box::new(OllamaStream::new(response.bytes_stream())))
    }

    fn context_window(&self) -> usize {
        // Ollama 模型默认值
        8_000
    }

    async fn get_model_info(&self) -> Result<ModelInfo> {
        Ok(ModelInfo {
            id: self.model().to_string(),
            provider: "ollama".to_string(),
            name: self.model().to_string(),
            context_window: self.context_window(),
            supports_tools: false,
            supports_streaming: true,
            input_cost_per_token: None,
            output_cost_per_token: None,
        })
    }
}

struct OllamaStream {
    inner: futures::stream::Map<
        reqwest::Result<bytes::Bytes>,
        fn(Result<bytes::Bytes, reqwest::Error>) -> Result<ModelResponse, ()>,
    >,
}

impl OllamaStream {
    fn new(stream: impl Stream<Item = Result<bytes::Bytes, reqwest::Error>> + Send + Unpin + 'static) -> Self {
        Self {
            inner: stream.map(|result| {
                match result {
                    Ok(bytes) => {
                        if let Ok(json) = serde_json::from_slice::<serde_json::Value>(&bytes) {
                            if let Some(content) = json["message"]["content"].as_str() {
                                return Ok(ModelResponse::Content { text: content.to_string() });
                            }
                            if json["done"].as_bool() == Some(true) {
                                return Ok(ModelResponse::Done);
                            }
                        }
                        Ok(ModelResponse::Content { text: String::new() })
                    }
                    Err(e) => Ok(ModelResponse::Error(e.to_string())),
                }
            }),
        }
    }
}

impl Stream for OllamaStream {
    type Item = Result<ModelResponse>;

    fn poll_next(mut self: Pin<&mut Self>, cx: &mut std::task::Context<'_>) -> std::task::Poll<Option<Self::Item>> {
        use futures::StreamExt;
        self.inner.poll_next_unpin(cx)
    }
}

// === 其他适配器的占位实现 ===

macro_rules! impl_simple_adapter {
    ($name:ident, $provider:expr) => {
        pub struct $name {
            config: ModelProvider,
            client: reqwest::Client,
        }

        impl $name {
            pub async fn new(config: ModelProvider) -> Result<Self> {
                let client = reqwest::Client::builder()
                    .timeout(std::time::Duration::from_secs(120))
                    .build()?;
                Ok(Self { config, client })
            }
        }

        #[async_trait]
        impl ModelAdapter for $name {
            fn provider_name(&self) -> &str {
                $provider
            }

            fn model_name(&self) -> &str {
                match &self.config {
                    ModelProvider::$name { model, .. } => model,
                    _ => "",
                }
            }

            fn supports_tools(&self) -> bool {
                true
            }

            fn supports_streaming(&self) -> bool {
                true
            }

            async fn chat(&self, messages: &[Message], _system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<String> {
                // 使用 OpenAI 兼容 API
                let base_url = match &self.config {
                    ModelProvider::$name { base_url, .. } => {
                        base_url.as_deref().unwrap_or("https://api.openai.com/v1")
                    }
                    _ => "https://api.openai.com/v1",
                };

                let api_key = match &self.config {
                    ModelProvider::$name { api_key, .. } => api_key.as_str(),
                    _ => "",
                };

                let model = match &self.config {
                    ModelProvider::$name { model, .. } => model,
                    _ => "",
                };

                #[derive(Serialize)]
                struct Request<'a> {
                    model: &'a str,
                    messages: Vec<serde_json::Value>,
                    temperature: f32,
                    max_tokens: usize,
                    stream: bool,
                }

                let mut req_messages = Vec::new();
                for msg in messages {
                    let role = match msg.role {
                        MessageRole::System => "system",
                        MessageRole::User => "user",
                        MessageRole::Assistant => "assistant",
                        MessageRole::Tool => "tool",
                        _ => "user",
                    };

                    let content = match &msg.content {
                        Content::Text(s) => s.clone(),
                        _ => serde_json::to_string(&msg.content)?,
                    };

                    req_messages.push(serde_json::json!({
                        "role": role,
                        "content": content,
                    }));
                }

                let request = Request {
                    model,
                    messages: req_messages,
                    temperature,
                    max_tokens,
                    stream: false,
                };

                let url = format!("{}/chat/completions", base_url);
                let response = self.client
                    .post(&url)
                    .header("Authorization", format!("Bearer {}", api_key))
                    .header("Content-Type", "application/json")
                    .json(&request)
                    .send()
                    .await?
                    .error_for_status()?
                    .json::<serde_json::Value>()
                    .await?;

                let content = response["choices"][0]["message"]["content"]
                    .as_str()
                    .unwrap_or("")
                    .to_string();

                Ok(content)
            }

            async fn chat_stream(&self, messages: &[Message], _system_prompt: Option<&String>, temperature: f32, max_tokens: usize, _tools: Option<&[String]>) -> Result<Box<dyn Stream<Item = Result<ModelResponse>> + Send + Unpin>> {
                let text = self.chat(messages, None, temperature, max_tokens, None).await?;
                Ok(Box::new(futures::stream::once(async { Ok(ModelResponse::Content { text }) })))
            }

            fn context_window(&self) -> usize {
                8_000
            }

            async fn get_model_info(&self) -> Result<ModelInfo> {
                Ok(ModelInfo {
                    id: self.model_name().to_string(),
                    provider: $provider.to_string(),
                    name: self.model_name().to_string(),
                    context_window: self.context_window(),
                    supports_tools: true,
                    supports_streaming: true,
                    input_cost_per_token: None,
                    output_cost_per_token: None,
                })
            }
        }
    };
}

impl_simple_adapter!(GroqAdapter, "groq");
impl_simple_adapter!(OpenRouterAdapter, "openrouter");
impl_simple_adapter!(LMStudioAdapter, "lmstudio");
impl_simple_adapter!(GoogleAdapter, "google");
impl_simple_adapter!(CustomAdapter, "custom");

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_provider_model_name() {
        let provider = ModelProvider::OpenAI {
            model: "gpt-4".to_string(),
            api_key: "test".to_string(),
            base_url: None,
            organization: None,
        };
        assert_eq!(provider.model_name(), "gpt-4");
    }

    #[test]
    fn test_provider_requires_api_key() {
        let ollama = ModelProvider::Ollama {
            model: "llama3".to_string(),
            base_url: "http://localhost:11434".to_string(),
            temperature: None,
        };
        assert!(!ollama.requires_api_key());

        let openai = ModelProvider::OpenAI {
            model: "gpt-4".to_string(),
            api_key: "test".to_string(),
            base_url: None,
            organization: None,
        };
        assert!(openai.requires_api_key());
    }
}
