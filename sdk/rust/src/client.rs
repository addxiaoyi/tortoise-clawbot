//! Tortoise Client Implementation.

use crate::error::{Result, TortoiseError};
use crate::models::*;

/// Client configuration.
#[derive(Debug, Clone)]
pub struct ClientConfig {
    /// Gateway URL.
    pub url: String,
    /// API key for authentication.
    pub api_key: Option<String>,
    /// Request timeout in seconds.
    pub timeout: u64,
    /// Enable WebSocket for real-time events.
    pub enable_websocket: bool,
}

impl Default for ClientConfig {
    fn default() -> Self {
        Self {
            url: "http://localhost:8080".to_string(),
            api_key: None,
            timeout: 60,
            enable_websocket: true,
        }
    }
}

impl ClientConfig {
    /// Set gateway URL.
    pub fn with_url(mut self, url: impl Into<String>) -> Self {
        self.url = url.into();
        self
    }
    
    /// Set API key.
    pub fn with_api_key(mut self, api_key: impl Into<String>) -> Self {
        self.api_key = Some(api_key.into());
        self
    }
    
    /// Set timeout.
    pub fn with_timeout(mut self, timeout: u64) -> Self {
        self.timeout = timeout;
        self
    }
    
    /// Enable/disable WebSocket.
    pub fn with_websocket(mut self, enable: bool) -> Self {
        self.enable_websocket = enable;
        self
    }
}

/// Tortoise Gateway Client.
#[derive(Debug)]
pub struct Client {
    config: ClientConfig,
    http_client: reqwest::Client,
    connected: bool,
}

impl Client {
    /// Create a new client.
    pub fn new(config: ClientConfig) -> Self {
        let http_client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(config.timeout))
            .build()
            .expect("Failed to create HTTP client");
        
        Self {
            config,
            http_client,
            connected: false,
        }
    }
    
    /// Connect to the gateway.
    pub async fn connect(&mut self) -> Result<()> {
        let health = self.health_check().await?;
        
        if health.status != "healthy" {
            return Err(TortoiseError::ConnectionError(
                format!("Gateway unhealthy: {}", health.status)
            ));
        }
        
        self.connected = true;
        Ok(())
    }
    
    /// Disconnect from the gateway.
    pub fn disconnect(&mut self) {
        self.connected = false;
    }
    
    /// Check if connected.
    pub fn is_connected(&self) -> bool {
        self.connected
    }
    
    /// Send a chat message.
    pub async fn chat(&self, message: impl Into<String>) -> Result<ChatMessage> {
        let request = ChatRequest::new(
            "gpt-4",
            vec![ChatMessage::user(message)]
        );
        
        self.chat_request(request).await
    }
    
    /// Send a chat request.
    pub async fn chat_request(&self, request: ChatRequest) -> Result<ChatMessage> {
        let url = format!("{}/api/v1/chat/completions", self.config.url);
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&request);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                let chat_response: ChatResponse = response.json().await?;
                Ok(chat_response.choices.into_iter().next()
                    .map(|c| c.message)
                    .unwrap_or_else(|| ChatMessage::assistant("")))
            },
            401 => Err(TortoiseError::AuthError("Invalid API key".to_string())),
            status => {
                let message = response.text().await.unwrap_or_default();
                Err(TortoiseError::HttpError { status, message })
            }
        }
    }
    
    /// Stream chat messages.
    pub async fn chat_stream(
        &self,
        message: impl Into<String>,
    ) -> Result<impl futures_util::Stream<Item = Result<String>>> {
        let url = format!("{}/api/v1/chat/completions", self.config.url);
        
        let request = ChatRequest::new(
            "gpt-4",
            vec![ChatMessage::user(message)]
        ).with_stream();
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&request);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        if !response.status().is_success() {
            return Err(TortoiseError::HttpError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_default(),
            });
        }
        
        let stream = response.bytes_stream()
            .map(|chunk| {
                match chunk {
                    Ok(bytes) => {
                        match String::from_utf8(bytes.to_vec()) {
                            Ok(text) => Ok(text),
                            Err(e) => Err(TortoiseError::ConnectionError(e.to_string())),
                        }
                    },
                    Err(e) => Err(TortoiseError::HttpClientError(e)),
                }
            });
        
        Ok(stream)
    }
    
    /// Health check.
    pub async fn health_check(&self) -> Result<HealthResponse> {
        let url = format!("{}/health", self.config.url);
        
        let mut req_builder = self.http_client.get(&url);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                let health: HealthResponse = response.json().await?;
                Ok(health)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// Get all sessions.
    pub async fn get_sessions(&self) -> Result<Vec<Session>> {
        let url = format!("{}/api/v1/sessions", self.config.url);
        
        let mut req_builder = self.http_client.get(&url);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                #[derive(serde::Deserialize)]
                struct SessionsResponse { items: Vec<Session> }
                
                let sessions: SessionsResponse = response.json().await?;
                Ok(sessions.items)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// Create a new session.
    pub async fn create_session(&self, title: impl Into<String>) -> Result<Session> {
        let url = format!("{}/api/v1/sessions", self.config.url);
        
        #[derive(serde::Serialize)]
        struct CreateRequest { title: String }
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&CreateRequest { title: title.into() });
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 | 201 => {
                let session: Session = response.json().await?;
                Ok(session)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// List plugins.
    pub async fn list_plugins(&self) -> Result<Vec<Plugin>> {
        let url = format!("{}/api/v1/plugins", self.config.url);
        
        let mut req_builder = self.http_client.get(&url);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                #[derive(serde::Deserialize)]
                struct PluginsResponse { plugins: Vec<Plugin> }
                
                let plugins: PluginsResponse = response.json().await?;
                Ok(plugins.plugins)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// List channels.
    pub async fn list_channels(&self) -> Result<Vec<Channel>> {
        let url = format!("{}/api/v1/channels", self.config.url);
        
        let mut req_builder = self.http_client.get(&url);
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                #[derive(serde::Deserialize)]
                struct ChannelsResponse { channels: Vec<Channel> }
                
                let channels: ChannelsResponse = response.json().await?;
                Ok(channels.channels)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// Send channel message.
    pub async fn send_channel_message(
        &self,
        channel: &str,
        to: &str,
        content: &str,
    ) -> Result<()> {
        let url = format!("{}/api/v1/channels/{}/send", self.config.url, channel);
        
        #[derive(serde::Serialize)]
        struct SendRequest<'a> {
            to: &'a str,
            content: &'a str,
        }
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&SendRequest { to, content });
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => Ok(()),
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// Search memory.
    pub async fn memory_search(&self, query: &str, limit: usize) -> Result<Vec<MemoryEntry>> {
        let url = format!("{}/api/v1/memory/search", self.config.url);
        
        #[derive(serde::Serialize)]
        struct SearchRequest<'a> {
            query: &'a str,
            limit: usize,
        }
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&SearchRequest { query, limit });
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 => {
                #[derive(serde::Deserialize)]
                struct SearchResponse { results: Vec<MemoryEntry> }
                
                let results: SearchResponse = response.json().await?;
                Ok(results.results)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
    
    /// Add memory entry.
    pub async fn memory_add(&self, content: &str, memory_type: &str) -> Result<MemoryEntry> {
        let url = format!("{}/api/v1/memory", self.config.url);
        
        #[derive(serde::Serialize)]
        struct AddRequest<'a> {
            content: &'a str,
            #[serde(rename = "type")]
            memory_type: &'a str,
        }
        
        let mut req_builder = self.http_client
            .post(&url)
            .json(&AddRequest { content, memory_type });
        
        if let Some(ref api_key) = self.config.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }
        
        let response = req_builder.send().await?;
        
        match response.status().as_u16() {
            200..=299 | 201 => {
                let entry: MemoryEntry = response.json().await?;
                Ok(entry)
            },
            status => Err(TortoiseError::HttpError {
                status,
                message: response.text().await.unwrap_or_default(),
            }),
        }
    }
}
