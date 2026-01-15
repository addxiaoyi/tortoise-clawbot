//! 消息路由器
//!
//! 负责将消息路由到正确的通道和代理

use anyhow::Result;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

use super::{Channel, ChannelType, ChannelStatus, UnifiedMessage, Content, ContentType, Sender};
use crate::agent::{Agent, Message};

/// 消息路由器
pub struct MessageRouter {
    channels: RwLock<HashMap<ChannelType, Arc<dyn Channel>>>,
    agent: Arc<dyn Agent>,
}

impl MessageRouter {
    /// 创建新的消息路由器
    pub fn new(agent: Arc<dyn Agent>) -> Self {
        Self {
            channels: RwLock::new(HashMap::new()),
            agent,
        }
    }

    /// 注册通道
    pub async fn register_channel(&self, channel: Arc<dyn Channel>) {
        let channel_type = channel.channel_type();
        let mut channels = self.channels.write().await;
        channels.insert(channel_type, channel);
    }

    /// 注销通道
    pub async fn unregister_channel(&self, channel_type: ChannelType) {
        let mut channels = self.channels.write().await;
        channels.remove(&channel_type);
    }

    /// 获取通道
    pub async fn get_channel(&self, channel_type: ChannelType) -> Option<Arc<dyn Channel>> {
        let channels = self.channels.read().await;
        channels.get(&channel_type).cloned()
    }

    /// 列出所有通道
    pub async fn list_channels(&self) -> Vec<ChannelType> {
        let channels = self.channels.read().await;
        channels.keys().cloned().collect()
    }

    /// 路由消息到代理
    pub async fn route(&self, message: UnifiedMessage) -> Result<()> {
        let channel = self.channels.get(&message.channel)
            .ok_or_else(|| anyhow::anyhow!("Channel not found: {:?}", message.channel))?;

        // 构造代理消息
        let agent_messages = vec![
            Message {
                id: message.id.clone(),
                role: crate::agent::MessageRole::User,
                content: crate::agent::Content::Text(message.content.text.unwrap_or_default()),
                tool_calls: vec![],
                tool_results: vec![],
                metadata: crate::agent::MessageMetadata {
                    channel: Some(message.channel.name().to_string()),
                    user_id: Some(message.sender.id.clone()),
                    ..Default::default()
                },
                created_at: chrono::DateTime::from_timestamp(message.timestamp, 0)
                    .unwrap_or_else(chrono::Utc::now),
            }
        ];

        // 调用代理
        let response = self.agent.chat(
            agent_messages,
            crate::agent::ChatOptions::default()
        ).await?;

        // 处理响应
        self.process_response(response, &message, channel.as_ref()).await?;

        Ok(())
    }

    /// 处理响应
    async fn process_response(
        &self,
        mut response: crate::agent::StreamingResponse,
        original: &UnifiedMessage,
        channel: &dyn Channel,
    ) -> Result<()> {
        use crate::agent::AgentEvent;
        
        let mut full_content = String::new();
        
        while let Some(event) = response.events.recv().await {
            match event {
                AgentEvent::Generation { content } => {
                    full_content.push_str(&content);
                }
                AgentEvent::ResponseComplete { content, .. } => {
                    full_content = content;
                }
                _ => {}
            }
        }

        // 发送回复
        let reply = UnifiedMessage {
            id: uuid::Uuid::new_v4().to_string(),
            channel: original.channel,
            sender: Sender {
                id: "tortoise".to_string(),
                name: "Tortoise".to_string(),
                avatar: None,
                is_bot: true,
            },
            content: Content {
                content_type: ContentType::Text,
                text: Some(full_content),
                html: None,
                mentions: vec![],
            },
            reply_to: Some(original.id.clone()),
            attachments: vec![],
            metadata: Default::default(),
            timestamp: chrono::Utc::now().timestamp(),
        };

        channel.send(reply).await?;
        Ok(())
    }

    /// 广播消息到所有通道
    pub async fn broadcast(&self, content: &str) -> Result<()> {
        let channels = self.channels.read().await;
        
        let message = UnifiedMessage {
            id: uuid::Uuid::new_v4().to_string(),
            channel: ChannelType::Discord, // 默认通道
            sender: Sender {
                id: "tortoise".to_string(),
                name: "Tortoise".to_string(),
                avatar: None,
                is_bot: true,
            },
            content: Content {
                content_type: ContentType::Text,
                text: Some(content.to_string()),
                html: None,
                mentions: vec![],
            },
            reply_to: None,
            attachments: vec![],
            metadata: Default::default(),
            timestamp: chrono::Utc::now().timestamp(),
        };

        for channel in channels.values() {
            let _ = channel.send(message.clone()).await;
        }

        Ok(())
    }

    /// 获取所有通道状态
    pub async fn get_all_status(&self) -> HashMap<ChannelType, ChannelStatus> {
        let channels = self.channels.read().await;
        let mut status_map = HashMap::new();
        
        for (channel_type, channel) in channels.iter() {
            let status = channel.status().await;
            status_map.insert(*channel_type, status);
        }
        
        status_map
    }
}
