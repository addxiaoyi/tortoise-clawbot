//! Agent Engine Implementations

use super::*;
use anyhow::Result;
use async_trait::async_trait;
use std::sync::Arc;
use tokio::sync::{RwLock, mpsc};
use flume::{Sender, Receiver, bounded};
use reqwest::Client;
use serde_json::json;

/// OpenAI Agent implementation
pub struct OpenAIAgent {
    config: AgentConfig,
    client: Client,
    http_client: Client,
    status: RwLock<AgentStatus>,
}

impl OpenAIAgent {
    pub async fn new(config: AgentConfig) -> Result<Self> {
        let client = Client::builder()
            .timeout(std::time::Duration::from_secs(120))
            .build()?;

        Ok(Self {
            config,
            client,
            http_client: Client::new(),
            status: RwLock::new(AgentStatus::Idle),
        })
    }

    fn get_base_url(&self) -> String {
        match &self.config.model_provider {
            ModelProvider::OpenAI { base_url, .. } => {
                base_url.clone().unwrap_or_else(|| "https://api.openai.com".to_string())
            }
            _ => "https://api.openai.com".to_string(),
        }
    }

    fn get_api_key(&self) -> String {
        match &self.config.model_provider {
            ModelProvider::OpenAI { api_key, .. } => api_key.clone(),
            ModelProvider::Anthropic { api_key, .. } => api_key.clone(),
            ModelProvider::Groq { api_key, .. } => api_key.clone(),
            ModelProvider::OpenRouter { api_key, .. } => api_key.clone(),
            _ => String::new(),
        }
    }

    fn get_model(&self) -> String {
        match &self.config.model_provider {
            ModelProvider::OpenAI { model, .. } => model.clone(),
            ModelProvider::Anthropic { model, .. } => model.clone(),
            ModelProvider::Ollama { model, .. } => model.clone(),
            ModelProvider::Groq { model, .. } => model.clone(),
            ModelProvider::OpenRouter { model, .. } => model.clone(),
            ModelProvider::Google { model, .. } => model.clone(),
            ModelProvider::Custom { name, .. } => name.clone(),
        }
    }

    async fn chat_internal(
        &self,
        messages: Vec<Message>,
        options: ChatOptions,
    ) -> Result<(String, Option<TokenUsage>)> {
        let thinking_mode = options.thinking_mode.unwrap_or(self.config.default_thinking);
        let temperature = options.temperature.unwrap_or_else(|| 
            thinking_mode.default_temperature()
        );
        let max_tokens = options.max_tokens.unwrap_or(4096);

        let base_url = self.get_base_url();
        let api_key = self.get_api_key();
        let model = self.get_model();

        // Convert messages to OpenAI format
        let mut openai_messages: Vec<serde_json::Value> = Vec::new();

        // Add system prompt
        let system_prompt = options.system_prompt
            .or_else(|| self.config.system_prompt.clone())
            .unwrap_or_else(|| self.default_system_prompt());

        openai_messages.push(serde_json::json!({
            "role": "system",
            "content": system_prompt
        }));

        for msg in &messages {
            let role = match msg.role {
                MessageRole::System => "system",
                MessageRole::User => "user",
                MessageRole::Assistant => "assistant",
                MessageRole::Tool => "tool",
            };

            openai_messages.push(serde_json::json!({
                "role": role,
                "content": msg.content
            }));
        }

        // Build request body
        let mut body = json!({
            "model": model,
            "messages": openai_messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
            "stream": true
        });

        // Add tools if provided
        if let Some(tools) = &options.tools {
            if !tools.is_empty() {
                body["tools"] = json!(tools.iter().map(|t| {
                    json!({
                        "type": "function",
                        "function": {
                            "name": t.name(),
                            "description": t.description(),
                            "parameters": t.parameters()
                        }
                    })
                }).collect::<Vec<_>>());
            }
        }

        // Send request
        let url = format!("{}/v1/chat/completions", base_url);
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .header("Content-Type", "application/json")
            .json(&body)
            .send()
            .await?;

        if !response.status().is_success() {
            let error_text = response.text().await?;
            anyhow::bail!("OpenAI API error: {}", error_text);
        }

        // For now, return a simple response (streaming implementation would be more complex)
        let response_json: serde_json::Value = response.json().await?;
        
        let content = response_json
            .get("choices")
            .and_then(|c| c.as_array())
            .and_then(|c| c.first())
            .and_then(|choice| choice.get("message"))
            .and_then(|msg| msg.get("content"))
            .and_then(|c| c.as_str())
            .unwrap_or("")
            .to_string();

        let usage = response_json.get("usage").map(|u| TokenUsage {
            prompt_tokens: u.get("prompt_tokens").and_then(|p| p.as_u64()).unwrap_or(0) as u32,
            completion_tokens: u.get("completion_tokens").and_then(|c| c.as_u64()).unwrap_or(0) as u32,
            total_tokens: u.get("total_tokens").and_then(|t| t.as_u64()).unwrap_or(0) as u32,
        });

        Ok((content, usage))
    }

    fn default_system_prompt(&self) -> String {
        format!(
            r#"You are Tortoise, a helpful AI assistant.

You are designed to be helpful, harmless, and honest.

Your capabilities:
- Text generation and analysis
- Code writing and debugging
- Research and information synthesis
- Creative tasks
- Problem solving

Respond clearly and concisely. Use formatting when helpful (Markdown, code blocks, etc.).
"#
        )
    }
}

#[async_trait]
impl Agent for OpenAIAgent {
    fn id(&self) -> &str {
        &self.config.id
    }

    fn name(&self) -> &str {
        &self.config.name
    }

    async fn status(&self) -> AgentStatus {
        self.status.read().await.clone()
    }

    async fn chat(
        &self,
        messages: Vec<Message>,
        options: ChatOptions,
    ) -> Result<StreamingResponse> {
        *self.status.write().await = AgentStatus::Busy;

        let (tx, rx) = bounded(100);

        let agent = self.clone();
        tokio::spawn(async move {
            let thinking_mode = options.thinking_mode.unwrap_or(agent.config.default_thinking);

            // Send thinking started event
            let _ = tx.send(AgentEvent::ThinkingStarted { mode: thinking_mode });

            match agent.chat_internal(messages, options).await {
                Ok((content, usage)) => {
                    // Send content chunks
                    for chunk in content.chars() {
                        let _ = tx.send(AgentEvent::Thinking {
                            content: chunk.to_string(),
                        }).await;
                    }

                    // Send completion event
                    let _ = tx.send(AgentEvent::ResponseComplete {
                        content,
                        usage,
                    }).await;
                }
                Err(e) => {
                    let _ = tx.send(AgentEvent::Error {
                        error: e.to_string(),
                    }).await;
                }
            }

            *agent.status.write().await = AgentStatus::Idle;
        });

        Ok(StreamingResponse {
            events: rx,
            _sender: tx,
        })
    }

    async fn reset(&self) -> Result<()> {
        *self.status.write().await = AgentStatus::Idle;
        Ok(())
    }

    fn config(&self) -> AgentConfig {
        self.config.clone()
    }
}

impl Clone for OpenAIAgent {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            client: self.client.clone(),
            http_client: self.http_client.clone(),
            status: RwLock::new(AgentStatus::Idle),
        }
    }
}

/// Anthropic Agent implementation
pub struct AnthropicAgent {
    config: AgentConfig,
    client: Client,
    status: RwLock<AgentStatus>,
}

impl AnthropicAgent {
    pub async fn new(config: AgentConfig) -> Result<Self> {
        let client = Client::builder()
            .timeout(std::time::Duration::from_secs(120))
            .build()?;

        Ok(Self {
            config,
            client,
            status: RwLock::new(AgentStatus::Idle),
        })
    }
}

#[async_trait]
impl Agent for AnthropicAgent {
    fn id(&self) -> &str {
        &self.config.id
    }

    fn name(&self) -> &str {
        &self.config.name
    }

    async fn status(&self) -> AgentStatus {
        self.status.read().await.clone()
    }

    async fn chat(
        &self,
        messages: Vec<Message>,
        options: ChatOptions,
    ) -> Result<StreamingResponse> {
        *self.status.write().await = AgentStatus::Busy;

        let (tx, rx) = bounded(100);

        tokio::spawn(async move {
            // Simplified Anthropic implementation
            let _ = tx.send(AgentEvent::ThinkingStarted {
                mode: options.thinking_mode.unwrap_or(ThinkMode::Balanced),
            }).await;

            let _ = tx.send(AgentEvent::ResponseComplete {
                content: "Anthropic agent not fully implemented yet".to_string(),
                usage: None,
            }).await;
        });

        Ok(StreamingResponse {
            events: rx,
            _sender: tx,
        })
    }

    async fn reset(&self) -> Result<()> {
        *self.status.write().await = AgentStatus::Idle;
        Ok(())
    }

    fn config(&self) -> AgentConfig {
        self.config.clone()
    }
}

/// Ollama Agent implementation (local models)
pub struct OllamaAgent {
    config: AgentConfig,
    client: Client,
    status: RwLock<AgentStatus>,
}

impl OllamaAgent {
    pub async fn new(config: AgentConfig) -> Result<Self> {
        let client = Client::builder()
            .timeout(std::time::Duration::from_secs(300))
            .build()?;

        Ok(Self {
            config,
            client,
            status: RwLock::new(AgentStatus::Idle),
        })
    }
}

#[async_trait]
impl Agent for OllamaAgent {
    fn id(&self) -> &str {
        &self.config.id
    }

    fn name(&self) -> &str {
        &self.config.name
    }

    async fn status(&self) -> AgentStatus {
        self.status.read().await.clone()
    }

    async fn chat(
        &self,
        messages: Vec<Message>,
        options: ChatOptions,
    ) -> Result<StreamingResponse> {
        *self.status.write().await = AgentStatus::Busy;

        let (tx, rx) = bounded(100);

        let model = match &self.config.model_provider {
            ModelProvider::Ollama { model, .. } => model.clone(),
            _ => "llama2".to_string(),
        };

        let base_url = match &self.config.model_provider {
            ModelProvider::Ollama { base_url, .. } => base_url.clone(),
            _ => "http://localhost:11434".to_string(),
        };

        let client = self.client.clone();
        tokio::spawn(async move {
            // Convert messages to Ollama format
            let ollama_messages: Vec<serde_json::Value> = messages
                .iter()
                .map(|m| {
                    let role = match m.role {
                        MessageRole::System => "system",
                        MessageRole::User => "user",
                        MessageRole::Assistant => "assistant",
                        MessageRole::Tool => "tool",
                    };
                    json!({
                        "role": role,
                        "content": m.content
                    })
                })
                .collect();

            let body = json!({
                "model": model,
                "messages": ollama_messages,
                "stream": true
            });

            let url = format!("{}/api/chat", base_url);
            
            match client.post(&url).json(&body).send().await {
                Ok(response) => {
                    if response.status().is_success() {
                        let _ = tx.send(AgentEvent::ThinkingStarted {
                            mode: options.thinking_mode.unwrap_or(ThinkMode::Balanced),
                        }).await;
                    }
                    let _ = tx.send(AgentEvent::ResponseComplete {
                        content: "Ollama agent streaming not fully implemented".to_string(),
                        usage: None,
                    }).await;
                }
                Err(e) => {
                    let _ = tx.send(AgentEvent::Error {
                        error: format!("Ollama error: {}", e),
                    }).await;
                }
            }
        });

        Ok(StreamingResponse {
            events: rx,
            _sender: tx,
        })
    }

    async fn reset(&self) -> Result<()> {
        *self.status.write().await = AgentStatus::Idle;
        Ok(())
    }

    fn config(&self) -> AgentConfig {
        self.config.clone()
    }
}
