//! Channel Trait
//! 
//! Trait interface for channel implementations.

use crate::channel::{ChannelConfig, ChannelStatus, ChannelType, UnifiedMessage};
use anyhow::Result;
use async_trait::async_trait;

/// Channel trait for messaging platform integrations
#[async_trait]
pub trait Channel: Send + Sync {
    /// Get channel type
    fn channel_type(&self) -> ChannelType;
    
    /// Start the channel connection
    async fn start(&self) -> Result<()>;
    
    /// Stop the channel connection
    async fn stop(&self) -> Result<()>;
    
    /// Send a message
    async fn send(&self, message: UnifiedMessage) -> Result<String>;
    
    /// Edit an existing message
    async fn edit(&self, message_id: &str, content: &str) -> Result<()>;
    
    /// Delete a message
    async fn delete(&self, message_id: &str) -> Result<()>;
    
    /// Add a reaction to a message
    async fn react(&self, message_id: &str, emoji: &str) -> Result<()>;
    
    /// Remove a reaction from a message
    async fn unreact(&self, message_id: &str, emoji: &str) -> Result<()>;
    
    /// Create a thread
    async fn create_thread(&self, message_id: &str, name: &str) -> Result<String>;
    
    /// Send a typing indicator
    async fn start_typing(&self) -> Result<()>;
    
    /// Stop typing indicator
    async fn stop_typing(&self) -> Result<()>;
    
    /// Get channel status
    async fn status(&self) -> ChannelStatus;
    
    /// Get channel configuration
    fn config(&self) -> &ChannelConfig;
}
