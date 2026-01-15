//! Channel Module
//! 
//! Unified messaging channel system supporting multiple platforms.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Channel configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    /// Channel type
    pub channel_type: ChannelType,
    /// Whether enabled
    pub enabled: bool,
    /// Channel-specific credentials (JSON)
    pub credentials: serde_json::Value,
    /// Channel options
    pub options: ChannelOptions,
}

/// Channel options
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelOptions {
    /// Rate limiting
    pub rate_limit: Option<RateLimit>,
    /// Proxy URL
    pub proxy: Option<String>,
    /// Timeout in ms
    pub timeout_ms: Option<u64>,
    /// Retry count
    pub retry_count: Option<u32>,
}

/// Rate limiting configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimit {
    /// Requests per second
    pub requests_per_second: f32,
    /// Requests per minute
    pub requests_per_minute: u32,
    /// Burst size
    pub burst_size: u32,
}

/// Channel type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
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
}

impl std::fmt::Display for ChannelType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.name())
    }
}

/// Sender information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sender {
    /// Sender ID
    pub id: String,
    /// Display name
    pub name: String,
    /// Avatar URL
    pub avatar: Option<String>,
    /// Whether this is a bot
    pub is_bot: bool,
}

/// Content type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
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

/// Message content
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Content {
    /// Content type
    pub content_type: ContentType,
    /// Text content
    pub text: Option<String>,
    /// HTML content
    pub html: Option<String>,
    /// Mentioned user IDs
    pub mentions: Vec<String>,
}

/// Attachment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attachment {
    /// Attachment ID
    pub id: String,
    /// Attachment type
    pub attachment_type: AttachmentType,
    /// URL (if external)
    pub url: Option<String>,
    /// Raw data (if inline)
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
pub enum AttachmentType {
    Image,
    Audio,
    Video,
    Document,
    Archive,
}

/// Reaction
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Reaction {
    /// Emoji
    pub emoji: String,
    /// Count
    pub count: u32,
    /// User IDs
    pub users: Vec<String>,
}

/// Message metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMetadata {
    /// Thread ID
    pub thread_id: Option<String>,
    /// Guild/server ID
    pub guild_id: Option<String>,
    /// Channel ID
    pub channel_id: Option<String>,
    /// Reactions
    pub reactions: Vec<Reaction>,
    /// Edited timestamp
    pub edited_at: Option<i64>,
    /// Original message for forwarded messages
    pub forwarded_from: Option<String>,
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
    pub content: Content,
    /// Original message ID (for replies)
    pub reply_to: Option<String>,
    /// Attachments
    pub attachments: Vec<Attachment>,
    /// Metadata
    pub metadata: MessageMetadata,
    /// Timestamp
    pub timestamp: i64,
}

impl UnifiedMessage {
    /// Create a simple text message
    pub fn text(channel: ChannelType, sender: Sender, text: String) -> Self {
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            channel,
            sender,
            content: Content {
                content_type: ContentType::Text,
                text: Some(text),
                html: None,
                mentions: vec![],
            },
            reply_to: None,
            attachments: vec![],
            metadata: MessageMetadata::default(),
            timestamp: chrono::Utc::now().timestamp(),
        }
    }
}

impl Default for MessageMetadata {
    fn default() -> Self {
        Self {
            thread_id: None,
            guild_id: None,
            channel_id: None,
            reactions: vec![],
            edited_at: None,
            forwarded_from: None,
        }
    }
}

/// Channel status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChannelStatus {
    Connected,
    Connecting,
    Disconnected,
    Error(String),
}

/// Channel trait
#[async_trait::async_trait]
pub trait Channel: Send + Sync {
    /// Get channel type
    fn channel_type(&self) -> ChannelType;

    /// Start the channel
    async fn start(&self) -> Result<()>;

    /// Stop the channel
    async fn stop(&self) -> Result<()>;

    /// Send a message
    async fn send(&self, message: UnifiedMessage) -> Result<String>;

    /// Edit a message
    async fn edit(&self, message_id: &str, content: &str) -> Result<()>;

    /// Delete a message
    async fn delete(&self, message_id: &str) -> Result<()>;

    /// Add a reaction
    async fn react(&self, message_id: &str, emoji: &str) -> Result<()>;

    /// Create a thread
    async fn create_thread(&self, message_id: &str, name: &str) -> Result<String>;

    /// Get channel status
    async fn status(&self) -> ChannelStatus;
}

/// Message router
pub struct MessageRouter {
    channels: RwLock<HashMap<ChannelType, Arc<dyn Channel>>>,
    agent: Arc<dyn crate::agent::Agent>,
}

impl MessageRouter {
    /// Create a new message router
    pub fn new(agent: Arc<dyn crate::agent::Agent>) -> Self {
        Self {
            channels: RwLock::new(HashMap::new()),
            agent,
        }
    }

    /// Register a channel
    pub async fn register_channel(&self, channel: Arc<dyn Channel>) {
        let mut channels = self.channels.write().await;
        channels.insert(channel.channel_type(), channel);
    }

    /// Unregister a channel
    pub async fn unregister_channel(&self, channel_type: ChannelType) {
        let mut channels = self.channels.write().await;
        channels.remove(&channel_type);
    }

    /// Route a message to the agent and send response
    pub async fn route(&self, message: UnifiedMessage) -> Result<()> {
        let channel = {
            let channels = self.channels.read().await;
            channels.get(&message.channel)
                .cloned()
                .ok_or_else(|| anyhow::anyhow!("Channel not found: {:?}", message.channel))?
        };

        // Build agent messages
        let mut messages = vec![
            crate::agent::Message {
                role: crate::agent::MessageRole::System,
                content: "You are Tortoise, a helpful AI assistant.".to_string(),
                tool_calls: None,
                tool_results: None,
                metadata: Default::default(),
            }
        ];

        messages.push(crate::agent::Message {
            role: crate::agent::MessageRole::User,
            content: format!(
                "{}: {}",
                message.sender.name,
                message.content.text.as_deref().unwrap_or("")
            ),
            tool_calls: None,
            tool_results: None,
            metadata: Default::default(),
        });

        // Call agent
        let response = self.agent.chat(messages, Default::default()).await?;

        // Process response
        self.process_response(response, &message, channel.as_ref()).await?;

        Ok(())
    }

    /// Process agent response
    async fn process_response(
        &self,
        mut response: crate::agent::StreamingResponse,
        original: &UnifiedMessage,
        channel: &dyn Channel,
    ) -> Result<()> {
        use crate::agent::AgentEvent;

        let mut full_content = String::new();

        // Collect streaming response
        while let Some(event) = response.events.recv().await {
            match event {
                AgentEvent::Thinking(content) => {
                    full_content.push_str(&content);
                }
                AgentEvent::ResponseComplete { content, .. } => {
                    full_content = content;
                }
                _ => {}
            }
        }

        // Send reply
        if !full_content.is_empty() {
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
                metadata: MessageMetadata {
                    thread_id: original.metadata.thread_id.clone(),
                    guild_id: original.metadata.guild_id.clone(),
                    channel_id: original.metadata.channel_id.clone(),
                    reactions: vec![],
                    edited_at: None,
                    forwarded_from: None,
                },
                timestamp: chrono::Utc::now().timestamp(),
            };

            channel.send(reply).await?;
        }

        Ok(())
    }

    /// List all registered channels
    pub async fn list_channels(&self) -> Vec<ChannelType> {
        let channels = self.channels.read().await;
        channels.keys().cloned().collect()
    }
}

/// Session management
pub struct SessionManager {
    sessions: RwLock<HashMap<String, Session>>,
}

impl SessionManager {
    /// Create a new session manager
    pub fn new() -> Self {
        Self {
            sessions: RwLock::new(HashMap::new()),
        }
    }

    /// Create a new session
    pub async fn create_session(&self, channel: ChannelType, sender_id: &str) -> String {
        let session_id = uuid::Uuid::new_v4().to_string();
        let session = Session {
            id: session_id.clone(),
            channel,
            sender_id: sender_id.to_string(),
            created_at: chrono::Utc::now().timestamp(),
            last_activity: chrono::Utc::now().timestamp(),
            message_count: 0,
            context: HashMap::new(),
        };

        let mut sessions = self.sessions.write().await;
        sessions.insert(session_id.clone(), session);
        
        session_id
    }

    /// Get a session
    pub async fn get_session(&self, session_id: &str) -> Option<Session> {
        let sessions = self.sessions.read().await;
        sessions.get(session_id).cloned()
    }

    /// Update session activity
    pub async fn touch_session(&self, session_id: &str) -> Result<()> {
        let mut sessions = self.sessions.write().await;
        if let Some(session) = sessions.get_mut(session_id) {
            session.last_activity = chrono::Utc::now().timestamp();
            session.message_count += 1;
        }
        Ok(())
    }

    /// Delete a session
    pub async fn delete_session(&self, session_id: &str) {
        let mut sessions = self.sessions.write().await;
        sessions.remove(session_id);
    }
}

impl Default for SessionManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Session structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Session {
    /// Session ID
    pub id: String,
    /// Channel type
    pub channel: ChannelType,
    /// Sender ID
    pub sender_id: String,
    /// Created timestamp
    pub created_at: i64,
    /// Last activity timestamp
    pub last_activity: i64,
    /// Message count
    pub message_count: u32,
    /// Session context
    pub context: HashMap<String, serde_json::Value>,
}
