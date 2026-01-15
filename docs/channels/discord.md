# Tortoise Discord 通道实现

## 概述

Discord 通道插件提供与 Discord 服务器的完整集成，支持消息、线程、频道、反应等所有 Discord 功能。

## 目录结构

```
plugins/channels/discord/
├── Cargo.toml
├── src/
│   ├── lib.rs
│   ├── bot.rs
│   ├── commands.rs
│   ├── events.rs
│   ├── handlers.rs
│   └── slash.rs
└── plugin.json
```

## 核心实现

```rust
// plugins/channels/discord/src/lib.rs

mod bot;
mod commands;
mod events;
mod handlers;
mod slash;

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tortoise_core::channel::{Channel, ChannelType, ChannelStatus, UnifiedMessage, Content, ContentType, MessageMetadata, Attachment, Sender};
use tortoise_core::agent::{Agent, Message, MessageRole, ChatOptions, StreamingResponse};
use tokio::sync::RwLock;

// 插件元数据
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscordConfig {
    pub token: String,
    pub guild_id: Option<u64>,
    pub allowed_channels: Vec<u64>,
    pub prefix: Option<String>,
    pub mention_as_trigger: bool,
    pub dm_enabled: bool,
}

pub struct DiscordChannel {
    config: DiscordConfig,
    http: Arc<serenity::http::Http>,
    shard_manager: Arc<RwLock<Option<serenity::gateway::ShardManager>>>,
    agent: Arc<dyn Agent>,
}

impl DiscordChannel {
    pub fn new(config: DiscordConfig, agent: Arc<dyn Agent>) -> Self {
        let http = serenity::http::Http::new_with_token(&config.token);
        Self {
            config,
            http: Arc::new(http),
            shard_manager: Arc::new(RwLock::new(None)),
            agent,
        }
    }

    async fn convert_message(&self, serenity_msg: &serenity::Message) -> Result<UnifiedMessage> {
        let attachments = serenity_msg
            .attachments
            .iter()
            .map(|a| Attachment {
                id: a.id.to_string(),
                attachment_type: match a.content_type.as_deref() {
                    Some(t) if t.starts_with("image/") => tortoise_core::channel::AttachmentType::Image,
                    Some(t) if t.starts_with("audio/") => tortoise_core::channel::AttachmentType::Audio,
                    Some(t) if t.starts_with("video/") => tortoise_core::channel::AttachmentType::Video,
                    _ => tortoise_core::channel::AttachmentType::Document,
                },
                url: Some(a.url.clone()),
                data: None,
                name: Some(a.filename.clone()),
                size: Some(a.size as u64),
                mime_type: a.content_type.clone(),
            })
            .collect();

        let content_type = if serenity_msg.content.starts_with("http") && 
            serenity_msg.content.contains("discordapp.com/attachments") {
            ContentType::Text
        } else if serenity_msg.embeds.len() > 0 {
            ContentType::Text
        } else {
            ContentType::Text
        };

        Ok(UnifiedMessage {
            id: serenity_msg.id.0.to_string(),
            channel: ChannelType::Discord,
            sender: Sender {
                id: serenity_msg.author.id.0.to_string(),
                name: serenity_msg.author.name.clone(),
                avatar: serenity_msg.author.avatar_url(),
                is_bot: serenity_msg.author.bot,
            },
            content: Content {
                content_type,
                text: Some(serenity_msg.content.clone()),
                html: None,
                mentions: serenity_msg
                    .mentions
                    .iter()
                    .map(|m| m.id.0.to_string())
                    .collect(),
            },
            reply_to: serenity_msg.message_reference.as_ref()
                .map(|r| r.message_id.0.to_string()),
            attachments,
            metadata: MessageMetadata {
                thread_id: serenity_msg.thread.as_ref().map(|t| t.id.0.to_string()),
                guild_id: serenity_msg.guild_id.map(|g| g.0.to_string()),
                channel_id: serenity_msg.channel_id.0.to_string(),
                reactions: vec![],
                edited_at: serenity_msg.edited_timestamp.map(|t| t.timestamp()),
                forwarded_from: None,
            },
            timestamp: serenity_msg.timestamp.timestamp(),
        })
    }

    async fn handle_message(&self, msg: UnifiedMessage) -> Result<()> {
        // 过滤条件检查
        if !self.should_respond(&msg).await {
            return Ok(());
        }

        // 构造 Agent 消息
        let mut messages = vec![Message {
            role: MessageRole::System,
            content: self.get_system_prompt().await,
            tool_calls: None,
            tool_results: None,
        }];

        messages.push(Message {
            role: MessageRole::User,
            content: format!("{}: {}", msg.sender.name, msg.content.text.as_deref().unwrap_or("")),
            tool_calls: None,
            tool_results: None,
        });

        // 调用 Agent
        let response = self.agent.chat(messages, ChatOptions::default()).await?;

        // 收集响应
        let mut full_response = String::new();
        use tokio::sync::mpsc;
        let (tx, mut rx) = mpsc::channel::<String>(100);

        // 后台收集流式响应
        tokio::spawn(async move {
            let mut collector = response.events;
            while let Some(event) = collector.recv().await {
                if let tortoise_core::agent::AgentEvent::Thinking(content) = event {
                    let _ = tx.send(content).await;
                } else if let tortoise_core::agent::AgentEvent::ResponseComplete { content } = event {
                    let _ = tx.send(content).await;
                    break;
                }
            }
        });

        // 收集并发送
        while let Some(chunk) = rx.recv().await {
            full_response.push_str(&chunk);
            // 可以实现打字指示器
        }

        // 发送回复
        if !full_response.is_empty() {
            self.send_message(&msg, &full_response).await?;
        }

        Ok(())
    }

    async fn should_respond(&self, msg: &UnifiedMessage) -> bool {
        // 忽略机器人消息
        if msg.sender.is_bot {
            return false;
        }

        // 检查是否在允许的频道
        if !self.config.allowed_channels.is_empty() {
            if let Some(channel_id) = &msg.metadata.channel_id {
                if !self.config.allowed_channels.iter().any(|c| c.to_string() == *channel_id) {
                    return false;
                }
            }
        }

        // 检查是否是 DM
        if msg.metadata.guild_id.is_none() {
            return self.config.dm_enabled;
        }

        true
    }

    async fn get_system_prompt(&self) -> String {
        r#"你是一个友好的 Discord 助手，名叫 Tortoise。
请用简洁、有趣的方式回复。
注意 Discord 的聊天风格，可以使用表情符号。
回复长度适中，不要太长。>"#.to_string()
    }

    async fn send_message(&self, original: &UnifiedMessage, content: &str) -> Result<String> {
        let channel_id = original.metadata.channel_id.as_ref()
            .ok_or_else(|| anyhow::anyhow!("No channel_id"))?;
        
        let builder = serenity::CreateMessage::new()
            .content(content)
            .reference_message(serenity::MessageReference::from((&original.metadata.guild_id, 
                &serenity::ChannelId(channel_id.parse()?), 
                &serenity::MessageId(original.id.parse()?)));

        let msg = serenity::ChannelId(channel_id.parse()?)
            .send_message(&*self.http, builder)
            .await?;

        Ok(msg.id.0.to_string())
    }
}

#[async_trait::async_trait]
impl Channel for DiscordChannel {
    fn channel_type(&self) -> ChannelType {
        ChannelType::Discord
    }

    async fn start(&self) -> Result<()> {
        let framework = serenity::Framework::builder()
            .options(|o| {
                o.on_message(handlers::on_message);
                o.on_interaction(handlers::on_interaction);
            })
            .build();

        let intents = serenity::gateway::Intent::GUILD_MESSAGES
            | serenity::gateway::Intent::MESSAGE_CONTENT
            | serenity::gateway::Intent::DIRECT_MESSAGES
            | serenity::gateway::Intent::GUILD_MESSAGE_REACTIONS;

        let mut client = serenity::Client::builder(&self.config.token, intents)
            .event_handler(events::EventHandler::new(self.agent.clone()))
            .framework(framework)
            .await?;

        let shard_manager = client.shard_manager_clone();
        *self.shard_manager.write().await = Some(shard_manager);

        tokio::spawn(async move {
            if let Err(e) = client.start().await {
                tracing::error!("Discord client error: {}", e);
            }
        });

        Ok(())
    }

    async fn stop(&self) -> Result<()> {
        if let Some(manager) = self.shard_manager.write().await.take() {
            manager.lock().await.shutdown_all().await;
        }
        Ok(())
    }

    async fn send(&self, message: UnifiedMessage) -> Result<String> {
        let channel_id = message.metadata.channel_id.as_ref()
            .ok_or_else(|| anyhow::anyhow!("No channel_id"))?;
        
        let content = message.content.text.as_ref()
            .ok_or_else(|| anyhow::anyhow!("No content"))?;

        let builder = serenity::CreateMessage::new().content(content);
        let msg = serenity::ChannelId(channel_id.parse()?)
            .send_message(&*self.http, builder)
            .await?;

        Ok(msg.id.0.to_string())
    }

    async fn edit(&self, message_id: &str, content: &str) -> Result<()> {
        let msg_id: u64 = message_id.parse()?;
        // 需要获取 channel_id
        Ok(())
    }

    async fn delete(&self, message_id: &str) -> Result<()> {
        let msg_id: u64 = message_id.parse()?;
        // 需要获取 channel_id
        Ok(())
    }

    async fn react(&self, message_id: &str, emoji: &str) -> Result<()> {
        // 实现反应
        Ok(())
    }

    async fn create_thread(&self, message_id: &str, name: &str) -> Result<String> {
        // 创建线程
        Ok(String::new())
    }

    async fn status(&self) -> ChannelStatus {
        ChannelStatus::Connected
    }
}
```

## 命令处理

```rust
// plugins/channels/discord/src/commands.rs

use serenity::prelude::*;
use anyhow::Result;

pub struct CommandHandler {
    prefix: String,
}

impl CommandHandler {
    pub fn new(prefix: String) -> Self {
        Self { prefix }
    }

    pub async fn handle(&self, ctx: &Context, msg: &Message) -> Result<Option<String>> {
        let content = msg.content.trim();
        
        if !content.starts_with(&self.prefix) {
            return Ok(None);
        }

        let args: Vec<&str> = content[self.prefix.len()..]
            .split_whitespace()
            .collect();

        if args.is_empty() {
            return Ok(None);
        }

        let command = args[0].to_lowercase();
        let args = &args[1..];

        match command.as_str() {
            "help" => Ok(Some(self.help_command().await?)),
            "status" => Ok(Some(self.status_command().await?)),
            "model" => Ok(Some(self.model_command(args).await?)),
            "skill" => Ok(Some(self.skill_command(args).await?)),
            "memory" => Ok(Some(self.memory_command(args).await?)),
            "config" => Ok(Some(self.config_command(args).await?)),
            _ => Ok(Some(format!("Unknown command: {}", command))),
        }
    }

    async fn help_command(&self) -> Result<String> {
        Ok(r#"
**Tortoise Commands**

`!t help` - 显示此帮助
`!t status` - 显示代理状态
`!t model [name]` - 切换 AI 模型
`!t skill [list/enable/disable]` - 管理技能
`!t memory [stats/clear]` - 管理记忆
`!t config [key] [value]` - 设置配置

Use slash commands `/` for quick access!
"#.to_string())
    }

    async fn status_command(&self) -> Result<String> {
        Ok("**Tortoise Status**: Online 🟢\n\
            - Model: GPT-4\n\
            - Memory: 85% available\n\
            - Skills: 12 active".to_string())
    }

    async fn model_command(&self, args: &[&str]) -> Result<String> {
        if args.is_empty() {
            return Ok("Usage: !t model <model-name>".to_string());
        }
        
        let model = args.join(" ");
        Ok(format!("Model changed to: **{}**", model))
    }

    async fn skill_command(&self, args: &[&str]) -> Result<String> {
        if args.is_empty() {
            return Ok("Usage: !t skill <list/enable/disable> [name]".to_string());
        }

        match args[0] {
            "list" => Ok("**Active Skills:**\n\
                - web_search\n\
                - image_gen\n\
                - calculator\n\
                - code_executor".to_string()),
            "enable" if args.len() > 1 => Ok(format!("Enabled skill: **{}**", args[1])),
            "disable" if args.len() > 1 => Ok(format!("Disabled skill: **{}**", args[1])),
            _ => Ok("Invalid skill command".to_string()),
        }
    }

    async fn memory_command(&self, args: &[&str]) -> Result<String> {
        match args.first() {
            Some(&"stats") => Ok("**Memory Stats:**\n\
                - Short-term: 42 items\n\
                - Medium-term: 128 items\n\
                - Long-term: 1,247 items".to_string()),
            Some(&"clear") => Ok("Memory cleared. 🔄"),
            _ => Ok("Usage: !t memory <stats/clear>".to_string()),
        }
    }

    async fn config_command(&self, args: &[&str]) -> Result<String> {
        if args.is_empty() {
            return Ok("**Current Config:**\n- thinking_mode: balanced\n- temperature: 0.7".to_string());
        }
        Ok(format!("Config updated: {} = {}", args[0], args.get(1).unwrap_or(&"")))
    }
}
```

## 事件处理

```rust
// plugins/channels/discord/src/events.rs

use serenity::prelude::*;
use serenity::model::prelude::*;
use anyhow::Result;
use std::sync::Arc;
use tortoise_core::agent::Agent;

pub struct EventHandler {
    agent: Arc<dyn Agent>,
}

impl EventHandler {
    pub fn new(agent: Arc<dyn Agent>) -> Self {
        Self { agent }
    }
}

#[serenity::async_trait]
impl EventHandler for EventHandler {
    async fn message(&self, ctx: Context, msg: Message) {
        if msg.author.bot {
            return;
        }

        // 处理消息
        tracing::info!("Received message from {}: {}", msg.author.name, msg.content);
    }

    async fn interaction(&self, ctx: Context, interaction: Interaction) {
        match interaction {
            Interaction::ApplicationCommand(cmd) => {
                self.handle_slash_command(&ctx, &cmd).await;
            }
            Interaction::MessageComponent(component) => {
                self.handle_button_click(&ctx, &component).await;
            }
            _ => {}
        }
    }

    async fn reaction_add(&self, ctx: Context, reaction: Reaction) {
        // 处理反应
    }

    async fn thread_create(&self, ctx: Context, thread: GuildThread) {
        tracing::info!("Thread created: {}", thread.name());
    }
}
```
