//! 渠道系统模块
//! 
//! 管理多渠道消息连接

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;

/// 渠道类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChannelType {
    Telegram,
    Discord,
    Slack,
    WhatsApp,
    Web,
    WeChat,
    LINE,
    Signal,
    SMS,
}

/// 渠道状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChannelState {
    Disconnected,
    Connecting,
    Connected,
    Error,
}

/// 渠道信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Channel {
    pub id: String,
    pub channel_type: ChannelType,
    pub name: String,
    pub state: ChannelState,
    pub config: HashMap<String, String>,
    pub credentials: HashMap<String, String>,
}

/// 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelMessage {
    pub id: String,
    pub channel_id: String,
    pub from: String,
    pub content: String,
    pub msg_type: MessageKind,
    pub timestamp: i64,
    pub metadata: HashMap<String, String>,
}

/// 消息类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MessageKind {
    Text,
    Image,
    Audio,
    Video,
    File,
    Location,
    Contact,
}

/// 渠道处理器
pub trait ChannelHandler: Send + Sync {
    fn channel_type(&self) -> ChannelType;
    fn connect(&self, credentials: &HashMap<String, String>, config: &HashMap<String, String>) -> Result<(), ChannelError>;
    fn disconnect(&self) -> Result<(), ChannelError>;
    fn send(&self, to: &str, content: &str) -> Result<String, ChannelError>;
    fn is_connected(&self) -> bool;
}

/// 渠道错误
#[derive(Debug, thiserror::Error)]
pub enum ChannelError {
    #[error("连接失败: {0}")]
    ConnectionFailed(String),
    
    #[error("认证失败: {0}")]
    AuthenticationFailed(String),
    
    #[error("消息发送失败: {0}")]
    SendFailed(String),
    
    #[error("已断开连接")]
    Disconnected,
    
    #[error("配置错误: {0}")]
    ConfigError(String),
}

/// 渠道管理器
pub struct ChannelManager {
    channels: Arc<RwLock<HashMap<String, Channel>>>,
    handlers: Arc<RwLock<HashMap<ChannelType, Box<dyn ChannelHandler>>>>,
}

impl ChannelManager {
    pub fn new() -> Self {
        Self {
            channels: Arc::new(RwLock::new(HashMap::new())),
            handlers: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    pub fn register_handler<H: ChannelHandler + 'static>(&self, handler: H) {
        let channel_type = handler.channel_type();
        let mut handlers = self.handlers.write();
        handlers.insert(channel_type, Box::new(handler));
    }
    
    pub fn connect(
        &self,
        channel_type: ChannelType,
        credentials: HashMap<String, String>,
        config: HashMap<String, String>,
    ) -> Result<Channel, ChannelError> {
        // 获取处理器
        let handler = {
            let handlers = self.handlers.read();
            handlers.get(&channel_type).cloned()
        };
        
        let handler = handler.ok_or_else(|| 
            ChannelError::ConfigError(format!("No handler for {:?}", channel_type))
        )?;
        
        // 连接
        handler.connect(&credentials, &config)?;
        
        // 创建渠道记录
        let channel = Channel {
            id: uuid::Uuid::new_v4().to_string(),
            channel_type,
            name: format!("{:?}", channel_type),
            state: ChannelState::Connected,
            config,
            credentials,
        };
        
        // 保存
        let id = channel.id.clone();
        let mut channels = self.channels.write();
        channels.insert(id, channel.clone());
        
        Ok(channel)
    }
    
    pub fn disconnect(&self, channel_id: &str) -> Result<(), ChannelError> {
        let channel_type = {
            let channels = self.channels.read();
            channels.get(channel_id).map(|c| c.channel_type)
        };
        
        if let Some(channel_type) = channel_type {
            let handler = {
                let handlers = self.handlers.read();
                handlers.get(&channel_type).cloned()
            };
            
            if let Some(handler) = handler {
                handler.disconnect()?;
            }
            
            let mut channels = self.channels.write();
            if let Some(channel) = channels.get_mut(channel_id) {
                channel.state = ChannelState::Disconnected;
            }
        }
        
        Ok(())
    }
    
    pub fn send(&self, channel_id: &str, to: &str, content: &str) -> Result<String, ChannelError> {
        let channel_type = {
            let channels = self.channels.read();
            channels.get(channel_id).map(|c| c.channel_type)
        };
        
        if let Some(channel_type) = channel_type {
            let handler = {
                let handlers = self.handlers.read();
                handlers.get(&channel_type).cloned()
            };
            
            if let Some(handler) = handler {
                return handler.send(to, content);
            }
        }
        
        Err(ChannelError::Disconnected)
    }
    
    pub fn get(&self, channel_id: &str) -> Option<Channel> {
        let channels = self.channels.read();
        channels.get(channel_id).cloned()
    }
    
    pub fn list(&self) -> Vec<Channel> {
        let channels = self.channels.read();
        channels.values().cloned().collect()
    }
    
    pub fn list_by_type(&self, channel_type: ChannelType) -> Vec<Channel> {
        let channels = self.channels.read();
        channels.values()
            .filter(|c| c.channel_type == channel_type)
            .cloned()
            .collect()
    }
}

impl Default for ChannelManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_channel_manager() {
        let manager = ChannelManager::new();
        let channels = manager.list();
        assert!(channels.is_empty());
    }
}
