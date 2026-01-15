//! 预导入模块
//!
//! 提供常用类型的便捷导入

pub use anyhow::{Result, Context, anyhow};
pub use async_trait::async_trait;
pub use serde::{Deserialize, Serialize};
pub use serde_json;

// Agent 模块
pub use crate::agent::{
    Agent, AgentConfig, AgentEvent, AgentStatus, ChatOptions,
    Message, MessageRole, Content, ToolCall, ThinkMode,
};

// Memory 模块
pub use crate::memory::{
    MemoryManager, MemoryItem, MemoryType, MemoryStore,
    MemoryConfig, MemoryQuery, MemoryStats,
};

// Channel 模块
pub use crate::channel::{
    Channel, ChannelConfig, ChannelType, ChannelStatus,
    UnifiedMessage, Content as ChannelContent, ContentType,
};

// Plugin 模块
pub use crate::plugin::{
    Plugin, PluginManager, PluginType, PluginState,
    PluginConfig, PluginMetadata,
};

// Security 模块
pub use crate::security::{
    SecurityManager, SecurityConfig, ThreatLevel, TrustScore, TrustLevel,
    AuditEvent,
};

// Network 模块
pub use crate::network::{
    P2PNode, NetworkConfig, PeerId,
};
