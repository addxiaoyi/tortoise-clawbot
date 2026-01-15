//! Error types for Tortoise SDK.

use thiserror::Error;

/// Result type alias using TortoiseError.
pub type Result<T> = std::result::Result<T, TortoiseError>;

/// Main error type for Tortoise SDK.
#[derive(Error, Debug)]
pub enum TortoiseError {
    /// Connection error.
    #[error("Failed to connect to gateway: {0}")]
    ConnectionError(String),
    
    /// Authentication error.
    #[error("Authentication failed: {0}")]
    AuthError(String),
    
    /// Request timeout.
    #[error("Request timeout: {0}")]
    TimeoutError(String),
    
    /// HTTP error.
    #[error("HTTP error {status}: {message}")]
    HttpError { status: u16, message: String },
    
    /// Validation error.
    #[error("Validation error: {0}")]
    ValidationError(String),
    
    /// Resource not found.
    #[error("Not found: {0}")]
    NotFoundError(String),
    
    /// Rate limit exceeded.
    #[error("Rate limit exceeded: {0}")]
    RateLimitError(String),
    
    /// Channel error.
    #[error("Channel error: {0}")]
    ChannelError(String),
    
    /// Plugin error.
    #[error("Plugin error: {0}")]
    PluginError(String),
    
    /// Memory error.
    #[error("Memory error: {0}")]
    MemoryError(String),
    
    /// WebSocket error.
    #[error("WebSocket error: {0}")]
    WebSocketError(String),
    
    /// Serialization error.
    #[error("Serialization error: {0}")]
    SerializationError(#[from] serde_json::Error),
    
    /// HTTP client error.
    #[error("HTTP client error: {0}")]
    HttpClientError(#[from] reqwest::Error),
    
    /// WebSocket stream error.
    #[error("WebSocket stream error: {0}")]
    WebSocketStreamError(#[from] tokio_tungstenite::tungstenite::Error),
    
    /// URL parse error.
    #[error("URL parse error: {0}")]
    UrlParseError(#[from] url::ParseError),
    
    /// IO error.
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}

impl TortoiseError {
    /// Check if error is retryable.
    pub fn is_retryable(&self) -> bool {
        matches!(
            self,
            TortoiseError::TimeoutError(_) 
                | TortoiseError::RateLimitError(_)
                | TortoiseError::ConnectionError(_)
        )
    }
    
    /// Get HTTP status code if applicable.
    pub fn status_code(&self) -> Option<u16> {
        match self {
            TortoiseError::HttpError { status, .. } => Some(*status),
            _ => None,
        }
    }
}
