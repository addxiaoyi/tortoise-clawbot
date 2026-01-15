//! Channel Types
//! 
//! Core types for the channel system.

use serde::{Deserialize, Serialize};

/// Channel type enumeration
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChannelType {
    Discord,
    Telegram,
    WhatsApp,
    Slack,
    Matrix,
    Signal,
    SMS,
    Email,
    IMessage,
    WebWidget,
    VoiceCall,
    VideoCall,
    Custom(String),
}

impl ChannelType {
    /// Get channel name
    pub fn name(&self) -> &str {
        match self {
            ChannelType::Discord => "discord",
            ChannelType::Telegram => "telegram",
            ChannelType::WhatsApp => "whatsapp",
            ChannelType::Slack => "slack",
            ChannelType::Matrix => "matrix",
            ChannelType::Signal => "signal",
            ChannelType::SMS => "sms",
            ChannelType::Email => "email",
            ChannelType::IMessage => "imessage",
            ChannelType::WebWidget => "web",
            ChannelType::VoiceCall => "voice",
            ChannelType::VideoCall => "video",
            ChannelType::Custom(name) => name,
        }
    }

    /// Check if channel supports threads
    pub fn supports_threads(&self) -> bool {
        matches!(
            self,
            ChannelType::Discord 
            | ChannelType::Slack 
            | ChannelType::Telegram
        )
    }

    /// Check if channel supports reactions
    pub fn supports_reactions(&self) -> bool {
        matches!(
            self,
            ChannelType::Discord
            | ChannelType::Slack
            | ChannelType::Telegram
            | ChannelType::Matrix
        )
    }

    /// Check if channel supports markdown
    pub fn supports_markdown(&self) -> bool {
        matches!(
            self,
            ChannelType::Discord
            | ChannelType::Slack
            | ChannelType::Telegram
            | ChannelType::Matrix
            | ChannelType::Email
        )
    }

    /// Get max message length
    pub fn max_message_length(&self) -> usize {
        match self {
            ChannelType::Discord => 2000,
            ChannelType::Telegram => 4096,
            ChannelType::Slack => 40000,
            ChannelType::SMS => 160,
            ChannelType::Email => 50000,
            _ => 10000,
        }
    }
}

impl std::fmt::Display for ChannelType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.name())
    }
}

/// Channel configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    /// Channel ID
    pub id: String,
    /// Channel type
    pub channel_type: ChannelType,
    /// Enabled status
    pub enabled: bool,
    /// Credentials (channel-specific)
    pub credentials: serde_json::Value,
    /// Rate limit configuration
    pub rate_limit: Option<RateLimitConfig>,
    /// Proxy configuration
    pub proxy: Option<String>,
    /// Timeout in milliseconds
    pub timeout_ms: Option<u64>,
}

/// Rate limit configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    /// Requests per second
    pub requests_per_second: f32,
    /// Requests per minute
    pub requests_per_minute: u32,
    /// Burst size
    pub burst_size: u32,
}

/// Unified message structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnifiedMessage {
    /// Message ID
    pub id: String,
    /// Channel type
    pub channel: ChannelType,
    /// Sender information
    pub sender: Sender,
    /// Message content
    pub content: MessageContent,
    /// Reply to message ID
    pub reply_to: Option<String>,
    /// Attachments
    pub attachments: Vec<Attachment>,
    /// Metadata
    pub metadata: MessageMetadata,
    /// Timestamp
    pub timestamp: i64,
}

/// Sender information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sender {
    /// Sender ID
    pub id: String,
    /// Sender name
    pub name: String,
    /// Avatar URL
    pub avatar: Option<String>,
    /// Is bot flag
    pub is_bot: bool,
}

/// Message content
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageContent {
    /// Content type
    pub content_type: ContentType,
    /// Text content
    pub text: Option<String>,
    /// HTML content
    pub html: Option<String>,
    /// Mentioned user IDs
    pub mentions: Vec<String>,
}

impl MessageContent {
    /// Create text content
    pub fn text(content: impl Into<String>) -> Self {
        Self {
            content_type: ContentType::Text,
            text: Some(content.into()),
            html: None,
            mentions: Vec::new(),
        }
    }

    /// Get plain text
    pub fn plain_text(&self) -> String {
        self.text.clone().unwrap_or_default()
    }
}

/// Content type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ContentType {
    Text,
    Image,
    Audio,
    Video,
    File,
    Location,
    Contact,
    Sticker,
    Template,
}

/// Attachment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attachment {
    /// Attachment ID
    pub id: String,
    /// Attachment type
    pub attachment_type: AttachmentType,
    /// URL
    pub url: Option<String>,
    /// Raw data
    pub data: Option<Vec<u8>>,
    /// File name
    pub name: Option<String>,
    /// File size
    pub size: Option<u64>,
    /// MIME type
    pub mime_type: Option<String>,
}

/// Attachment type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AttachmentType {
    Image,
    Audio,
    Video,
    Document,
    Archive,
}

/// Message metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMetadata {
    /// Thread ID
    pub thread_id: Option<String>,
    /// Guild/Server ID
    pub guild_id: Option<String>,
    /// Channel ID
    pub channel_id: Option<String>,
    /// Reactions
    pub reactions: Vec<Reaction>,
    /// Edited timestamp
    pub edited_at: Option<i64>,
    /// Forwarded from
    pub forwarded_from: Option<String>,
}

/// Reaction
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Reaction {
    /// Emoji
    pub emoji: String,
    /// Count
    pub count: u32,
    /// User IDs who reacted
    pub users: Vec<String>,
}

/// Channel status
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChannelStatus {
    Connected,
    Connecting,
    Disconnected,
    Error(String),
}

impl Default for ChannelStatus {
    fn default() -> Self {
        ChannelStatus::Disconnected
    }
}
