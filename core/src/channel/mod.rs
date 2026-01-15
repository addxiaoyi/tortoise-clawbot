//! Channel management module - Multi-channel support

use std::sync::Arc;
use tokio::sync::RwLock;
use dashmap::DashMap;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

/// Channel type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChannelType {
    Web,
    WebSocket,
    Telegram,
    Discord,
    Slack,
    WhatsApp,
    Matrix,
    Signal,
}

/// Channel status
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ChannelStatus {
    Connected,
    Connecting,
    Disconnected,
    Error,
}

/// Channel configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    pub channel_type: ChannelType,
    pub enabled: bool,
    pub rate_limit: Option<RateLimitConfig>,
    pub allowed_users: Option<Vec<String>>,
    pub custom: serde_json::Value,
}

/// Rate limit configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    pub requests_per_minute: u32,
    pub burst: u32,
}

/// Channel message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelMessage {
    pub id: String,
    pub channel_type: ChannelType,
    pub channel_id: String,
    pub user_id: String,
    pub user_name: String,
    pub content: String,
    pub message_type: MessageType,
    pub attachments: Vec<Attachment>,
    pub metadata: serde_json::Value,
    pub timestamp: DateTime<Utc>,
}

/// Message type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum MessageType {
    Text,
    Image,
    Audio,
    Video,
    File,
    Command,
}

/// Attachment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attachment {
    pub url: String,
    pub mime_type: String,
    pub size: u64,
    pub thumbnail: Option<String>,
}

/// Channel event
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelEvent {
    pub event_type: EventType,
    pub channel_type: ChannelType,
    pub channel_id: String,
    pub user_id: Option<String>,
    pub data: serde_json::Value,
    pub timestamp: DateTime<Utc>,
}

/// Event type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EventType {
    Connect,
    Disconnect,
    Message,
    Typing,
    Reaction,
    Error,
}

/// Channel handler trait
#[async_trait]
pub trait ChannelHandler: Send + Sync {
    /// Get channel type
    fn channel_type(&self) -> ChannelType;
    
    /// Get channel name/identifier
    fn name(&self) -> String;
    
    /// Connect to the channel
    async fn connect(&self) -> Result<(), ChannelError>;
    
    /// Disconnect from the channel
    async fn disconnect(&self) -> Result<(), ChannelError>;
    
    /// Get connection status
    fn status(&self) -> ChannelStatus;
    
    /// Send a message
    async fn send(&self, msg: ChannelMessage) -> Result<(), ChannelError>;
    
    /// Send a typing indicator
    async fn send_typing(&self, user_id: &str, typing: bool) -> Result<(), ChannelError>;
    
    /// Subscribe to events
    fn subscribe(&self) -> tokio::sync::mpsc::Receiver<ChannelEvent>;
}

/// Channel error
#[derive(Debug, thiserror::Error)]
pub enum ChannelError {
    #[error("Channel not connected")]
    NotConnected,
    
    #[error("Channel error: {0}")]
    Connection(String),
    
    #[error("Send failed: {0}")]
    SendFailed(String),
    
    #[error("Rate limited")]
    RateLimited,
    
    #[error("Permission denied")]
    PermissionDenied,
    
    #[error("Invalid message: {0}")]
    InvalidMessage(String),
}

/// Channel manager
pub struct ChannelManager {
    channels: Arc<DashMap<String, Arc<Box<dyn ChannelHandler>>>>,
    status: Arc<DashMap<String, ChannelStatus>>,
}

impl ChannelManager {
    pub fn new() -> Self {
        Self {
            channels: Arc::new(DashMap::new()),
            status: Arc::new(DashMap::new()),
        }
    }

    /// Register a channel handler
    pub fn register<H: ChannelHandler + 'static>(&self, handler: H) -> Result<(), ChannelError> {
        let name = handler.name();
        let status = handler.status();
        
        self.channels.insert(name.clone(), Arc::new(Box::new(handler)));
        self.status.insert(name, status);
        
        Ok(())
    }

    /// Unregister a channel
    pub fn unregister(&self, name: &str) -> bool {
        self.channels.remove(name);
        self.status.remove(name);
        true
    }

    /// Get a channel handler
    pub fn get(&self, name: &str) -> Option<Arc<Box<dyn ChannelHandler>>> {
        self.channels.get(name).map(|c| c.clone())
    }

    /// List all channels
    pub fn list(&self) -> Vec<ChannelInfo> {
        self.channels
            .iter()
            .map(|c| ChannelInfo {
                name: c.name(),
                channel_type: c.channel_type(),
                status: self.status.get(&c.name()).map(|s| *s).unwrap_or(ChannelStatus::Disconnected),
            })
            .collect()
    }

    /// Connect all channels
    pub async fn connect_all(&self) -> Vec<(String, Result<(), ChannelError>)> {
        let mut results = Vec::new();
        
        for entry in self.channels.iter() {
            let name = entry.name();
            let result = entry.connect().await;
            self.status.insert(name.clone(), entry.status());
            results.push((name, result));
        }
        
        results
    }

    /// Disconnect all channels
    pub async fn disconnect_all(&self) -> Vec<(String, Result<(), ChannelError>)> {
        let mut results = Vec::new();
        
        for entry in self.channels.iter() {
            let name = entry.name();
            let result = entry.disconnect().await;
            self.status.insert(name.clone(), ChannelStatus::Disconnected);
            results.push((name, result));
        }
        
        results
    }

    /// Send to a specific channel
    pub async fn send_to(&self, channel_name: &str, msg: ChannelMessage) -> Result<(), ChannelError> {
        if let Some(channel) = self.channels.get(channel_name) {
            channel.send(msg).await
        } else {
            Err(ChannelError::NotConnected)
        }
    }

    /// Broadcast to all connected channels
    pub async fn broadcast(&self, msg: ChannelMessage) -> Vec<(String, Result<(), ChannelError>)> {
        let mut results = Vec::new();
        
        for entry in self.channels.iter() {
            if entry.status() == ChannelStatus::Connected {
                let name = entry.name();
                let result = entry.send(msg.clone()).await;
                results.push((name, result));
            }
        }
        
        results
    }

    /// Get channel count
    pub fn len(&self) -> usize {
        self.channels.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.channels.is_empty()
    }

    /// Get connected channels count
    pub fn connected_count(&self) -> usize {
        self.status
            .iter()
            .filter(|s| *s == ChannelStatus::Connected)
            .count()
    }
}

impl Default for ChannelManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Channel information
#[derive(Debug, Clone)]
pub struct ChannelInfo {
    pub name: String,
    pub channel_type: ChannelType,
    pub status: ChannelStatus,
}
