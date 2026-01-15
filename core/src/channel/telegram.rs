//! Telegram Channel Implementation
//! 
//! Example implementation of Telegram bot integration.

use crate::channel::{
    Attachment, AttachmentType, Channel, ChannelConfig, ChannelStatus, ChannelType,
    MessageContent, MessageMetadata, RateLimit, Sender, UnifiedMessage,
};
use anyhow::{anyhow, Result};
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

/// Telegram bot configuration
#[derive(Debug, Clone)]
pub struct TelegramConfig {
    /// Bot token from BotFather
    pub bot_token: String,
    /// Allowed chat IDs (empty = all)
    pub allowed_chats: Vec<i64>,
    /// Enable group support
    pub group_enabled: bool,
    /// Admin user IDs
    pub admins: Vec<i64>,
}

impl TelegramConfig {
    /// Create from generic config
    pub fn from_config(config: &ChannelConfig) -> Result<Self> {
        let bot_token = config.credentials["bot_token"]
            .as_str()
            .ok_or_else(|| anyhow!("Missing bot_token"))?
            .to_string();

        let allowed_chats: Vec<i64> = config.credentials["allowed_chats"]
            .as_array()
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_i64())
                    .collect()
            })
            .unwrap_or_default();

        let group_enabled = config.credentials["group_enabled"]
            .as_bool()
            .unwrap_or(true);

        let admins: Vec<i64> = config.credentials["admins"]
            .as_array()
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_i64())
                    .collect()
            })
            .unwrap_or_default();

        Ok(Self {
            bot_token,
            allowed_chats,
            group_enabled,
            admins,
        })
    }
}

/// Telegram channel implementation
pub struct TelegramChannel {
    /// Configuration
    config: TelegramConfig,
    /// Channel status
    status: RwLock<ChannelStatus>,
    /// HTTP client
    client: reqwest::Client,
}

impl TelegramChannel {
    /// Create a new Telegram channel
    pub fn new(config: TelegramConfig) -> Self {
        Self {
            config,
            status: RwLock::new(ChannelStatus::Disconnected),
            client: reqwest::Client::new(),
        }
    }

    /// Get API base URL
    fn api_url(&self, method: &str) -> String {
        format!(
            "https://api.telegram.org/bot{}/{}",
            self.config.bot_token, method
        )
    }

    /// Send a request to Telegram API
    async fn api_call<T: serde::de::DeserializeOwned>(
        &self,
        method: &str,
        body: Option<serde_json::Value>,
    ) -> Result<T> {
        let url = self.api_url(method);
        let mut request = self.client.post(&url);

        if let Some(body) = body {
            request = request.json(&body);
        }

        let response = request.send().await?;

        if !response.status().is_success() {
            let error = response.text().await?;
            return Err(anyhow!("Telegram API error: {}", error));
        }

        let result: TelegramResponse<T> = response.json().await?;

        if !result.ok {
            return Err(anyhow!("Telegram error: {:?}", result));
        }

        result.result.ok_or_else(|| anyhow!("No result"))
    }

    /// Parse update from Telegram
    fn parse_update(&self, update: TelegramUpdate) -> Option<UnifiedMessage> {
        let message = update.message?;
        let chat = message.chat;

        // Check if it's a private chat or group
        let is_group = chat.id < 0;
        if is_group && !self.config.group_enabled {
            return None;
        }

        // Check allowed chats
        if !self.config.allowed_chats.is_empty()
            && !self.config.allowed_chats.contains(&chat.id)
        {
            return None;
        }

        // Parse text content
        let text = message.text.or(message.caption).unwrap_or_default();

        Some(UnifiedMessage {
            id: message.id.to_string(),
            channel: ChannelType::Telegram,
            sender: Sender {
                id: message.from.as_ref().map(|f| f.id.to_string()).unwrap_or_default(),
                name: message.from.as_ref().map(|f| f.first_name.clone()).unwrap_or_default(),
                avatar: None,
                is_bot: message.from.as_ref().map(|f| f.is_bot).unwrap_or(false),
            },
            content: MessageContent {
                content_type: crate::channel::ContentType::Text,
                text: Some(text),
                html: None,
                mentions: Vec::new(),
            },
            reply_to: message.reply_to_message.map(|m| m.id.to_string()),
            attachments: self.parse_attachments(&message),
            metadata: MessageMetadata {
                thread_id: None,
                guild_id: None,
                channel_id: Some(chat.id.to_string()),
                reactions: Vec::new(),
                edited_at: message.edit_date.map(|t| t as i64),
                forwarded_from: message.forward_from.map(|f| f.id.to_string()),
            },
            timestamp: message.date as i64,
        })
    }

    /// Parse attachments from message
    fn parse_attachments(&self, message: &TelegramMessage) -> Vec<Attachment> {
        let mut attachments = Vec::new();

        // Photos
        for photo in &message.photo {
            attachments.push(Attachment {
                id: photo.file_id.clone(),
                attachment_type: AttachmentType::Image,
                url: None, // Would need file_id -> file_url
                data: None,
                name: None,
                size: photo.file_size,
                mime_type: Some("image/jpeg".to_string()),
            });
        }

        // Documents
        if let Some(doc) = &message.document {
            attachments.push(Attachment {
                id: doc.file_id.clone(),
                attachment_type: AttachmentType::Document,
                url: None,
                data: None,
                name: doc.file_name.clone(),
                size: doc.file_size,
                mime_type: doc.mime_type.clone(),
            });
        }

        // Videos
        if let Some(video) = &message.video {
            attachments.push(Attachment {
                id: video.file_id.clone(),
                attachment_type: AttachmentType::Video,
                url: None,
                data: None,
                name: None,
                size: video.file_size,
                mime_type: Some("video/mp4".to_string()),
            });
        }

        // Voice
        if let Some(voice) = &message.voice {
            attachments.push(Attachment {
                id: voice.file_id.clone(),
                attachment_type: AttachmentType::Audio,
                url: None,
                data: None,
                name: None,
                size: voice.file_size,
                mime_type: Some("audio/ogg".to_string()),
            });
        }

        attachments
    }
}

#[async_trait]
impl Channel for TelegramChannel {
    fn channel_type(&self) -> ChannelType {
        ChannelType::Telegram
    }

    async fn start(&self) -> Result<()> {
        *self.status.write().await = ChannelStatus::Connecting;

        // Verify bot token
        let _: TelegramResponse<TelegramUser> = self
            .api_call("getMe", None)
            .await?;

        *self.status.write().await = ChannelStatus::Connected;
        tracing::info!("Telegram channel started");
        Ok(())
    }

    async fn stop(&self) -> Result<()> {
        *self.status.write().await = ChannelStatus::Disconnected;
        tracing::info!("Telegram channel stopped");
        Ok(())
    }

    async fn send(&self, message: UnifiedMessage) -> Result<String> {
        let chat_id = message.metadata.channel_id
            .as_ref()
            .ok_or_else(|| anyhow!("No channel_id"))?;

        let chat_id: i64 = chat_id.parse()?;

        let text = message.content.text.as_ref()
            .ok_or_else(|| anyhow!("No content"))?;

        #[derive(Serialize)]
        struct SendMessageParams {
            chat_id: i64,
            text: String,
            reply_to_message_id: Option<u32>,
            #[serde(skip_serializing_if = "Option::is_none")]
            reply_markup: Option<serde_json::Value>,
        }

        let params = SendMessageParams {
            chat_id,
            text: text.clone(),
            reply_to_message_id: message.reply_to.as_ref().map(|m| m.parse().unwrap_or(0)),
            reply_markup: None,
        };

        #[derive(Deserialize)]
        struct SendResponse {
            message_id: u32,
        }

        let response: TelegramResponse<SendResponse> = self
            .api_call("sendMessage", Some(serde_json::to_value(params)?))
            .await?;

        Ok(response.result.message_id.to_string())
    }

    async fn edit(&self, message_id: &str, content: &str) -> Result<()> {
        // Would need channel_id, simplified for example
        Ok(())
    }

    async fn delete(&self, message_id: &str) -> Result<()> {
        // Would need channel_id, simplified for example
        Ok(())
    }

    async fn react(&self, message_id: &str, emoji: &str) -> Result<()> {
        // Telegram uses emoji as reactions
        Ok(())
    }

    async fn unreact(&self, message_id: &str, emoji: &str) -> Result<()> {
        Ok(())
    }

    async fn create_thread(&self, _message_id: &str, _name: &str) -> Result<String> {
        // Telegram uses topics instead of threads
        Ok(String::new())
    }

    async fn start_typing(&self) -> Result<()> {
        Ok(())
    }

    async fn stop_typing(&self) -> Result<()> {
        Ok(())
    }

    async fn status(&self) -> ChannelStatus {
        self.status.read().await.clone()
    }

    fn config(&self) -> &ChannelConfig {
        unimplemented!()
    }
}

// Telegram API types
#[derive(Debug, Deserialize)]
struct TelegramResponse<T> {
    ok: bool,
    #[serde(default)]
    result: TelegramResult<T>,
    #[serde(default)]
    description: Option<String>,
}

type TelegramResult<T> = Option<T>;

#[derive(Debug, Deserialize)]
struct TelegramUpdate {
    #[serde(default)]
    update_id: u64,
    #[serde(default)]
    message: Option<TelegramMessage>,
}

#[derive(Debug, Deserialize)]
struct TelegramMessage {
    message_id: u32,
    from: Option<TelegramUser>,
    chat: TelegramChat,
    date: u64,
    text: Option<String>,
    caption: Option<String>,
    reply_to_message: Option<Box<TelegramMessage>>,
    #[serde(default)]
    photo: Vec<TelegramPhoto>,
    document: Option<TelegramDocument>,
    video: Option<TelegramVideo>,
    voice: Option<TelegramVoice>,
    #[serde(default)]
    edit_date: Option<u64>,
    forward_from: Option<TelegramUser>,
}

#[derive(Debug, Deserialize)]
struct TelegramUser {
    id: i64,
    is_bot: bool,
    first_name: String,
    #[serde(default)]
    last_name: Option<String>,
    #[serde(default)]
    username: Option<String>,
}

#[derive(Debug, Deserialize)]
struct TelegramChat {
    id: i64,
    #[serde(default)]
    title: Option<String>,
    #[serde(default)]
    username: Option<String>,
    #[serde(default)]
    type_: String,
}

#[derive(Debug, Deserialize)]
struct TelegramPhoto {
    file_id: String,
    file_size: Option<u64>,
}

#[derive(Debug, Deserialize)]
struct TelegramDocument {
    file_id: String,
    file_name: Option<String>,
    mime_type: Option<String>,
    file_size: Option<u64>,
}

#[derive(Debug, Deserialize)]
struct TelegramVideo {
    file_id: String,
    file_size: Option<u64>,
}

#[derive(Debug, Deserialize)]
struct TelegramVoice {
    file_id: String,
    file_size: Option<u64>,
}
