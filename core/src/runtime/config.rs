//! Runtime configuration

use serde::{Deserialize, Serialize};

/// Runtime configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeConfig {
    /// Application name
    pub app_name: String,
    /// Application version
    pub version: String,
    /// Data directory for storage
    pub data_dir: String,
    /// Maximum concurrent sessions
    pub max_sessions: usize,
    /// Maximum message size in bytes
    pub max_message_size: usize,
    /// Request timeout in seconds
    pub request_timeout_secs: u64,
    /// Heartbeat interval in seconds
    pub heartbeat_interval_secs: u64,
    /// Enable debug mode
    pub debug: bool,
    /// Log level
    pub log_level: LogLevel,
    /// AI model configuration
    pub models: ModelConfig,
    /// Memory configuration
    pub memory: MemoryConfig,
    /// Plugin configuration
    pub plugins: PluginConfig,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            app_name: "Tortoise".into(),
            version: env!("CARGO_PKG_VERSION").into(),
            data_dir: "./data".into(),
            max_sessions: 10000,
            max_message_size: 10 * 1024 * 1024, // 10MB
            request_timeout_secs: 300, // 5 minutes
            heartbeat_interval_secs: 30,
            debug: false,
            log_level: LogLevel::Info,
            models: ModelConfig::default(),
            memory: MemoryConfig::default(),
            plugins: PluginConfig::default(),
        }
    }
}

/// Log level
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LogLevel {
    Trace,
    Debug,
    Info,
    Warn,
    Error,
}

/// Model configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelConfig {
    /// Default model to use
    pub default_model: String,
    /// Available models and their configs
    pub providers: Vec<ModelProvider>,
}

impl Default for ModelConfig {
    fn default() -> Self {
        Self {
            default_model: "gpt-4o".into(),
            providers: Vec::new(),
        }
    }
}

/// Model provider configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelProvider {
    /// Provider name (openai, anthropic, etc.)
    pub name: String,
    /// API endpoint
    pub endpoint: Option<String>,
    /// API key (or env var reference)
    pub api_key: String,
    /// Organization ID (for OpenAI)
    pub organization: Option<String>,
    /// Base URL override
    pub base_url: Option<String>,
    /// Rate limit settings
    pub rate_limit: RateLimitConfig,
    /// Fallback provider
    pub fallback: Option<String>,
}

impl ModelProvider {
    pub fn new(name: String, api_key: String) -> Self {
        Self {
            name,
            endpoint: None,
            api_key,
            organization: None,
            base_url: None,
            rate_limit: RateLimitConfig::default(),
            fallback: None,
        }
    }
}

/// Rate limit configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    /// Requests per minute
    pub rpm: u32,
    /// Tokens per minute
    pub tpm: u32,
    /// Concurrent requests
    pub concurrent: u32,
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            rpm: 60,
            tpm: 60000,
            concurrent: 10,
        }
    }
}

/// Memory configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryConfig {
    /// Working memory capacity (number of items)
    pub working_capacity: usize,
    /// Semantic memory vector dimension
    pub semantic_dimension: usize,
    /// Semantic memory max entries
    pub semantic_max_entries: usize,
    /// Vector store backend
    pub vector_backend: VectorBackend,
}

impl Default for MemoryConfig {
    fn default() -> Self {
        Self {
            working_capacity: 100,
            semantic_dimension: 1536,
            semantic_max_entries: 10000,
            vector_backend: VectorBackend::Local,
        }
    }
}

/// Vector store backend
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum VectorBackend {
    /// Local in-memory vector store
    Local,
    /// Qdrant vector database
    Qdrant,
    /// Pinecone vector database
    Pinecone,
    /// Weaviate vector database
    Weaviate,
}

/// Plugin configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginConfig {
    /// Enable plugin system
    pub enabled: bool,
    /// Plugin directory
    pub plugin_dir