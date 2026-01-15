//! Tortoise Core Library
//! 
//! A super AI agent framework that combines the best of OpenClaw and Hermes.

#![forbid(unsafe_code)]
#![warn(unused_crates_fundamental)]
#![warn(missing_docs)]
#![warn(missing_debug_implementations)]

pub mod agent;
pub mod memory;
pub mod channel;
pub mod plugin;
pub mod skill;
pub mod tool;
pub mod security;
pub mod network;

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

/// Tortoise configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    /// Agent configuration
    pub agent: agent::AgentConfig,
    /// Memory configuration
    pub memory: memory::MemoryConfig,
    /// Channel configuration
    pub channels: Vec<channel::ChannelConfig>,
    /// Plugin configuration
    pub plugins: plugin::PluginConfig,
    /// Security configuration
    pub security: security::SecurityConfig,
    /// Network configuration
    pub network: network::NetworkConfig,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            agent: AgentConfig::default(),
            memory: MemoryConfig::default(),
            channels: vec![],
            plugins: PluginConfig::default(),
            security: SecurityConfig::default(),
            network: NetworkConfig::default(),
        }
    }
}

/// Tortoise instance
pub struct Tortoise {
    config: RwLock<Config>,
    agent: Arc<dyn agent::Agent>,
    memory: Arc<memory::MemoryManager>,
    channels: RwLock<Vec<Arc<dyn channel::Channel>>>,
    plugins: Arc<plugin::PluginManager>,
    security: Arc<security::SecurityManager>,
    network: Option<Arc<network::P2PNode>>,
}

impl Tortoise {
    /// Create a new Tortoise instance
    pub async fn new(config: Config) -> Result<Self> {
        let agent = agent::create_agent(&config.agent).await?;
        let memory = memory::MemoryManager::new(&config.memory).await?;
        let plugins = plugin::PluginManager::new(&config.plugins)?;
        let security = security::SecurityManager::new(&config.security);

        let network = if config.network.enabled {
            Some(network::P2PNode::new(&config.network).await?)
        } else {
            None
        };

        Ok(Self {
            config: RwLock::new(config),
            agent,
            memory,
            channels: RwLock::new(vec![]),
            plugins,
            security,
            network,
        })
    }

    /// Start Tortoise
    pub async fn start(&self) -> Result<()> {
        tracing::info!("Starting Tortoise...");

        let channels = self.channels.read().await;
        for channel in channels.iter() {
            channel.start().await?;
        }

        if let Some(network) = &self.network {
            network.start().await?;
        }

        tracing::info!("Tortoise started successfully");
        Ok(())
    }

    /// Stop Tortoise
    pub async fn stop(&self) -> Result<()> {
        tracing::info!("Stopping Tortoise...");

        let channels = self.channels.read().await;
        for channel in channels.iter() {
            channel.stop().await?;
        }

        if let Some(network) = &self.network {
            network.stop().await?;
        }

        tracing::info!("Tortoise stopped");
        Ok(())
    }

    /// Register a channel
    pub async fn register_channel(&self, channel: Arc<dyn channel::Channel>) {
        let mut channels = self.channels.write().await;
        channels.push(channel);
    }

    /// Get agent reference
    pub fn agent(&self) -> Arc<dyn agent::Agent> {
        self.agent.clone()
    }

    /// Get memory manager reference
    pub fn memory(&self) -> Arc<memory::MemoryManager> {
        self.memory.clone()
    }

    /// Get plugin manager reference
    pub fn plugins(&self) -> Arc<plugin::PluginManager> {
        self.plugins.clone()
    }

    /// Get security manager reference
    pub fn security(&self) -> Arc<security::SecurityManager> {
        self.security.clone()
    }

    /// Get network node reference
    pub fn network(&self) -> Option<Arc<network::P2PNode>> {
        self.network.clone()
    }
}

pub use agent::{Agent, AgentConfig, AgentEvent, Message, ThinkMode, MessageRole, ToolCall, ToolResult};
pub use channel::{Channel, ChannelType, UnifiedMessage};
pub use memory::{MemoryManager, MemoryItem};
pub use plugin::{PluginManager, Plugin};
pub use security::{SecurityManager, TrustLevel};
pub use network::P2PNode;
