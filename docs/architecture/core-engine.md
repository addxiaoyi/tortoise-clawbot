# Tortoise 核心引擎架构设计

## 概述

Tortoise Core Engine 是整个框架的核心运行时，采用 Rust 语言实现，确保高性能、安全和零成本抽象。

## 模块结构

```
core/
├── src/
│   ├── lib.rs                 # 库入口
│   ├── main.rs                # CLI 入口
│   ├── agent/                 # 代理引擎
│   │   ├── mod.rs
│   │   ├── engine.rs          # 核心引擎
│   │   ├── context.rs        # 上下文管理
│   │   ├── model.rs          # 模型抽象
│   │   ├── thinking.rs       # 思维引擎
│   │   └── multi_agent.rs    # 多代理系统
│   ├── memory/                # 记忆系统
│   │   ├── mod.rs
│   │   ├── short_term.rs     # 短期记忆
│   │   ├── medium_term.rs    # 中期记忆
│   │   ├── long_term.rs      # 长期记忆
│   │   └── vector_store.rs   # 向量存储
│   ├── skill/                 # 技能系统
│   │   ├── mod.rs
│   │   ├── registry.rs       # 技能注册表
│   │   ├── executor.rs       # 技能执行器
│   │   └── sandbox.rs        # 技能沙盒
│   ├── tool/                  # 工具系统
│   │   ├── mod.rs
│   │   ├── registry.rs       # 工具注册表
│   │   └── executor.rs       # 工具执行器
│   ├── channel/               # 通道系统
│   │   ├── mod.rs
│   │   ├── router.rs         # 消息路由
│   │   ├── session.rs        # 会话管理
│   │   └── adapter.rs        # 通道适配器
│   ├── plugin/                # 插件系统
│   │   ├── mod.rs
│   │   ├── loader.rs         # 插件加载器
│   │   ├── runtime.rs        # 插件运行时
│   │   └── sandbox.rs        # 插件沙盒
│   ├── security/              # 安全系统
│   │   ├── mod.rs
│   │   ├── trust.rs          # 信任管理
│   │   ├── crypto.rs         # 加密模块
│   │   └── audit.rs          # 审计日志
│   └── network/               # 网络系统
│       ├── mod.rs
│       ├── p2p.rs            # P2P 通信
│       ├── dht.rs            # DHT 网络
│       └── discovery.rs      # 节点发现
```

## 核心 API 设计

### Agent Engine

```rust
// src/agent/engine.rs

use anyhow::Result;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

/// 代理配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentConfig {
    /// 代理 ID
    pub id: String,
    /// 代理名称
    pub name: String,
    /// 模型提供商
    pub model_provider: ModelProvider,
    /// 默认思维模式
    pub default_thinking: ThinkMode,
    /// 最大上下文长度
    pub max_context: usize,
    /// 温度参数
    pub temperature: f32,
}

/// 模型提供商
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ModelProvider {
    OpenAI {
        model: String,
        api_key: String,
        base_url: Option<String>,
    },
    Anthropic {
        model: String,
        api_key: String,
        base_url: Option<String>,
    },
    Google {
        model: String,
        api_key: String,
    },
    Ollama {
        model: String,
        base_url: String,
    },
    Groq {
        model: String,
        api_key: String,
    },
    OpenRouter {
        model: String,
        api_key: String,
    },
    Custom {
        name: String,
        base_url: String,
        api_key: Option<String>,
    },
}

/// 思维模式
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ThinkMode {
    /// 快速响应 (< 100ms)
    Fast,
    /// 平衡模式
    Balanced,
    /// 深度思考
    Deep,
    /// 研究模式
    Research,
    /// 创意发散
    Creative,
}

impl ThinkMode {
    pub fn timeout_ms(&self) -> u64 {
        match self {
            ThinkMode::Fast => 100,
            ThinkMode::Balanced => 500,
            ThinkMode::Deep => 2000,
            ThinkMode::Research => 5000,
            ThinkMode::Creative => 10000,
        }
    }

    pub fn default_temperature(&self) -> f32 {
        match self {
            ThinkMode::Fast => 0.0,
            ThinkMode::Balanced => 0.5,
            ThinkMode::Deep => 0.7,
            ThinkMode::Research => 0.6,
            ThinkMode::Creative => 1.0,
        }
    }
}

/// 消息角色
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MessageRole {
    System,
    User,
    Assistant,
    Tool,
}

/// 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub role: MessageRole,
    pub content: String,
    pub tool_calls: Option<Vec<ToolCall>>,
    pub tool_results: Option<Vec<ToolResult>>,
}

/// 工具调用
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    pub arguments: serde_json::Value,
}

/// 工具结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    pub call_id: String,
    pub result: Result<String, String>,
}

/// 代理事件
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum AgentEvent {
    /// 开始思考
    ThinkingStarted { mode: ThinkMode },
    /// 思考中 (流式)
    Thinking { content: String },
    /// 工具调用
    ToolCall { call: ToolCall },
    /// 工具结果
    ToolResult { result: ToolResult },
    /// 响应完成
    ResponseComplete { content: String },
    /// 错误
    Error { error: String },
}

/// 代理接口
#[async_trait]
pub trait Agent: Send + Sync {
    /// 获取代理 ID
    fn id(&self) -> &str;
    
    /// 获取代理名称
    fn name(&self) -> &str;
    
    /// 处理消息
    async fn chat(
        &self,
        messages: Vec<Message>,
        options: ChatOptions,
    ) -> Result<StreamingResponse>;
    
    /// 获取代理状态
    async fn status(&self) -> AgentStatus;
    
    /// 重置代理
    async fn reset(&self) -> Result<()>;
}

/// 聊天选项
#[derive(Debug, Clone, Default)]
pub struct ChatOptions {
    pub thinking_mode: Option<ThinkMode>,
    pub system_prompt: Option<String>,
    pub tools: Option<Vec<Box<dyn Tool>>>,
    pub max_tokens: Option<usize>,
    pub temperature: Option<f32>,
    pub stop: Option<Vec<String>>,
}

/// 流式响应
pub struct StreamingResponse {
    pub events: Receiver<AgentEvent>,
}

impl StreamingResponse {
    pub async fn next(&mut self) -> Option<AgentEvent> {
        self.events.recv().await.ok()
    }
}

/// 代理状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AgentStatus {
    Idle,
    Busy,
    Error(String),
}

/// 工具接口
#[async_trait]
pub trait Tool: Send + Sync {
    /// 获取工具名称
    fn name(&self) -> &str;
    
    /// 获取工具描述
    fn description(&self) -> &str;
    
    /// 获取工具参数 schema
    fn parameters(&self) -> serde_json::Value;
    
    /// 执行工具
    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value>;
}
```

### Memory System

```rust
// src/memory/mod.rs

use anyhow::Result;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

/// 记忆项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryItem {
    pub id: String,
    pub content: String,
    pub memory_type: MemoryType,
    pub importance: f32,
    pub created_at: i64,
    pub last_accessed: i64,
    pub access_count: u32,
    pub metadata: serde_json::Value,
}

/// 记忆类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MemoryType {
    ShortTerm,
    MediumTerm,
    LongTerm,
}

/// 记忆查询
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryQuery {
    pub query: String,
    pub memory_type: Option<MemoryType>,
    pub limit: Option<usize>,
    pub threshold: Option<f32>,
}

/// 记忆接口
#[async_trait]
pub trait MemoryStore: Send + Sync {
    /// 存储记忆
    async fn store(&self, item: MemoryItem) -> Result<String>;
    
    /// 检索记忆
    async fn retrieve(&self, query: MemoryQuery) -> Result<Vec<MemoryItem>>;
    
    /// 更新记忆
    async fn update(&self, id: &str, item: MemoryItem) -> Result<()>;
    
    /// 删除记忆
    async fn delete(&self, id: &str) -> Result<()>;
    
    /// 获取记忆统计
    async fn stats(&self) -> Result<MemoryStats>;
}

/// 记忆统计
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStats {
    pub short_term_count: usize,
    pub medium_term_count: usize,
    pub long_term_count: usize,
    pub total_size_bytes: u64,
}

/// 记忆管理器
pub struct MemoryManager {
    short_term: Arc<dyn MemoryStore>,
    medium_term: Arc<dyn MemoryStore>,
    long_term: Arc<dyn MemoryStore>,
    vector_store: Arc<dyn VectorStore>,
}

impl MemoryManager {
    pub fn new(
        short_term: Arc<dyn MemoryStore>,
        medium_term: Arc<dyn MemoryStore>,
        long_term: Arc<dyn MemoryStore>,
        vector_store: Arc<dyn VectorStore>,
    ) -> Self {
        Self {
            short_term,
            medium_term,
            long_term,
            vector_store,
        }
    }

    /// 存储新记忆
    pub async fn remember(&self, content: String, importance: f32) -> Result<String> {
        let embedding = self.vector_store.embed(&content).await?;
        
        let item = MemoryItem {
            id: uuid::Uuid::new_v4().to_string(),
            content,
            memory_type: MemoryType::ShortTerm,
            importance,
            created_at: chrono::Utc::now().timestamp(),
            last_accessed: chrono::Utc::now().timestamp(),
            access_count: 0,
            metadata: serde_json::json!({ "embedding": embedding }),
        };

        let id = self.short_term.store(item).await?;
        
        // 如果重要性高，转移到长期记忆
        if importance > 0.8 {
            self.promote_to_long_term(&id).await?;
        }
        
        Ok(id)
    }

    /// 检索相关记忆
    pub async fn recall(&self, query: &str) -> Result<Vec<MemoryItem>> {
        let query_embedding = self.vector_store.embed(query).await?;
        
        let mut results = Vec::new();
        
        // 查询长期记忆
        let long_term_results = self.long_term.retrieve(MemoryQuery {
            query: query.to_string(),
            memory_type: Some(MemoryType::LongTerm),
            limit: Some(10),
            threshold: Some(0.7),
        }).await?;
        results.extend(long_term_results);
        
        // 查询中期记忆
        let medium_term_results = self.medium_term.retrieve(MemoryQuery {
            query: query.to_string(),
            memory_type: Some(MemoryType::MediumTerm),
            limit: Some(5),
            threshold: Some(0.7),
        }).await?;
        results.extend(medium_term_results);
        
        // 去重并排序
        results.sort_by(|a, b| b.importance.partial_cmp(&a.importance).unwrap());
        results.truncate(15);
        
        Ok(results)
    }

    /// 提升到长期记忆
    pub async fn promote_to_long_term(&self, id: &str) -> Result<()> {
        // 从各层记忆查找
        if let Ok(item) = self.medium_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await?.first().cloned() {
            self.long_term.store(item).await?;
            self.medium_term.delete(id).await?;
        }
        Ok(())
    }

    /// 遗忘低重要性记忆
    pub async fn forget_low_importance(&self, threshold: f32) -> Result<usize> {
        let mut count = 0;
        
        let items = self.medium_term.retrieve(MemoryQuery {
            query: String::new(),
            memory_type: Some(MemoryType::MediumTerm),
            limit: Some(1000),
            ..Default::default()
        }).await?;
        
        for item in items {
            if item.importance < threshold {
                self.medium_term.delete(&item.id).await?;
                count += 1;
            }
        }
        
        Ok(count)
    }
}

/// 向量存储接口
#[async_trait]
pub trait VectorStore: Send + Sync {
    /// 生成嵌入向量
    async fn embed(&self, text: &str) -> Result<Vec<f32>>;
    
    /// 批量嵌入
    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>>;
    
    /// 近似搜索
    async fn search(&self, query: &[f32], limit: usize) -> Result<Vec<SearchResult>>;
}

/// 搜索结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchResult {
    pub id: String,
    pub score: f32,
    pub content: String,
}
```

### Channel System

```rust
// src/channel/mod.rs

use anyhow::Result;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

/// 通道配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    pub channel_type: ChannelType,
    pub enabled: bool,
    pub credentials: serde_json::Value,
    pub options: ChannelOptions,
}

/// 通道类型
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

/// 通道选项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelOptions {
    pub rate_limit: Option<RateLimit>,
    pub proxy: Option<String>,
    pub timeout_ms: Option<u64>,
    pub retry_count: Option<u32>,
}

/// 速率限制
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimit {
    pub requests_per_second: f32,
    pub requests_per_minute: u32,
    pub burst_size: u32,
}

/// 统一消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnifiedMessage {
    pub id: String,
    pub channel: ChannelType,
    pub sender: Sender,
    pub content: Content,
    pub reply_to: Option<String>,
    pub attachments: Vec<Attachment>,
    pub metadata: MessageMetadata,
    pub timestamp: i64,
}

/// 发送者
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sender {
    pub id: String,
    pub name: String,
    pub avatar: Option<String>,
    pub is_bot: bool,
}

/// 消息内容
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Content {
    pub content_type: ContentType,
    pub text: Option<String>,
    pub html: Option<String>,
    pub mentions: Vec<String>,
}

/// 内容类型
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

/// 附件
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attachment {
    pub id: String,
    pub attachment_type: AttachmentType,
    pub url: Option<String>,
    pub data: Option<Vec<u8>>,
    pub name: Option<String>,
    pub size: Option<u64>,
    pub mime_type: Option<String>,
}

/// 附件类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AttachmentType {
    Image,
    Audio,
    Video,
    Document,
    Archive,
}

/// 消息元数据
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMetadata {
    pub thread_id: Option<String>,
    pub guild_id: Option<String>,
    pub channel_id: Option<String>,
    pub reactions: Vec<Reaction>,
    pub edited_at: Option<i64>,
    pub forwarded_from: Option<String>,
}

/// 表情回应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Reaction {
    pub emoji: String,
    pub count: u32,
    pub users: Vec<String>,
}

/// 通道接口
#[async_trait]
pub trait Channel: Send + Sync {
    /// 获取通道类型
    fn channel_type(&self) -> ChannelType;
    
    /// 启动通道
    async fn start(&self) -> Result<()>;
    
    /// 停止通道
    async fn stop(&self) -> Result<()>;
    
    /// 发送消息
    async fn send(&self, message: UnifiedMessage) -> Result<String>;
    
    /// 编辑消息
    async fn edit(&self, message_id: &str, content: &str) -> Result<()>;
    
    /// 删除消息
    async fn delete(&self, message_id: &str) -> Result<()>;
    
    /// 添加反应
    async fn react(&self, message_id: &str, emoji: &str) -> Result<()>;
    
    /// 创建线程
    async fn create_thread(&self, message_id: &str, name: &str) -> Result<String>;
    
    /// 获取状态
    async fn status(&self) -> ChannelStatus;
}

/// 通道状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChannelStatus {
    Connected,
    Connecting,
    Disconnected,
    Error(String),
}

/// 消息路由器
pub struct MessageRouter {
    channels: HashMap<ChannelType, Arc<dyn Channel>>,
    agent: Arc<dyn super::agent::Agent>,
}

impl MessageRouter {
    pub fn new(agent: Arc<dyn super::agent::Agent>) -> Self {
        Self {
            channels: HashMap::new(),
            agent,
        }
    }

    /// 注册通道
    pub fn register_channel(&mut self, channel: Arc<dyn Channel>) {
        self.channels.insert(channel.channel_type(), channel);
    }

    /// 路由消息到代理
    pub async fn route(&self, message: UnifiedMessage) -> Result<()> {
        let channel = self.channels.get(&message.channel)
            .ok_or_else(|| anyhow::anyhow!("Channel not found: {:?}", message.channel))?;

        // 构造代理消息
        let agent_messages = vec![
            super::agent::Message {
                role: super::agent::MessageRole::User,
                content: message.content.text.unwrap_or_default(),
                tool_calls: None,
                tool_results: None,
            }
        ];

        // 调用代理
        let response = self.agent.chat(agent_messages, Default::default()).await?;

        // 处理响应
        self.process_response(response, &message, channel).await?;

        Ok(())
    }

    async fn process_response(
        &self,
        response: super::agent::StreamingResponse,
        original: &UnifiedMessage,
        channel: Arc<dyn Channel>,
    ) -> Result<()> {
        use tokio::sync::mpsc;
        use tokio::stream::StreamExt;

        let mut sender = mpsc::Sender::new(64);
        let tx = sender.clone();
        
        // 后台任务收集响应
        tokio::spawn(async move {
            let mut full_content = String::new();
            while let Some(event) = response.events.recv().await {
                match event {
                    super::agent::AgentEvent::Thinking(content) => {
                        full_content.push_str(&content);
                    }
                    super::agent::AgentEvent::ResponseComplete { content } => {
                        full_content = content;
                    }
                    _ => {}
                }
            }
            
            let _ = tx.send(full_content).await;
        });

        let content = sender.recv().await.unwrap_or_default();

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
                text: Some(content),
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
        Ok(())
    }
}
```

### Plugin System

```rust
// src/plugin/mod.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;

/// 插件元数据
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginMetadata {
    pub id: String,
    pub name: String,
    pub version: String,
    pub description: String,
    pub author: String,
    pub license: String,
    pub repository: Option<String>,
    pub keywords: Vec<String>,
    pub dependencies: HashMap<String, String>,
}

/// 插件类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PluginType {
    Channel,
    Skill,
    Tool,
    Integration,
    Custom,
}

/// 插件状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PluginState {
    Loaded,
    Running,
    Stopped,
    Error(String),
}

/// 插件实例
pub struct Plugin {
    pub metadata: PluginMetadata,
    pub plugin_type: PluginType,
    pub state: PluginState,
    runtime: Box<dyn PluginRuntime>,
}

impl Plugin {
    pub fn new(
        metadata: PluginMetadata,
        plugin_type: PluginType,
        runtime: Box<dyn PluginRuntime>,
    ) -> Self {
        Self {
            metadata,
            plugin_type,
            state: PluginState::Loaded,
            runtime,
        }
    }

    /// 启动插件
    pub async fn start(&mut self) -> Result<()> {
        self.runtime.start().await?;
        self.state = PluginState::Running;
        Ok(())
    }

    /// 停止插件
    pub async fn stop(&mut self) -> Result<()> {
        self.runtime.stop().await?;
        self.state = PluginState::Stopped;
        Ok(())
    }

    /// 调用插件方法
    pub async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        self.runtime.call(method, args).await
    }
}

/// 插件运行时接口
#[async_trait::async_trait]
pub trait PluginRuntime: Send + Sync {
    /// 启动
    async fn start(&self) -> Result<()>;
    
    /// 停止
    async fn stop(&self) -> Result<()>;
    
    /// 调用方法
    async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value>;
    
    /// 获取资源使用
    fn resource_usage(&self) -> ResourceUsage;
}

/// 资源使用
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResourceUsage {
    pub memory_mb: f64,
    pub cpu_percent: f32,
    pub uptime_seconds: u64,
}

/// WASM 沙盒运行时
pub struct WasmSandbox {
    instance: wasmer::Instance,
    memory_limit_mb: u32,
}

impl WasmSandbox {
    pub fn new(wasm_bytes: &[u8], memory_limit_mb: u32) -> Result<Self> {
        let store = wasmer::Store::default();
        let module = wasmer::Module::new(&store, wasm_bytes)?;
        
        // 导入 WASI
        let wasi_env = wasmer_wasi::WasiState::new("tortoise-plugin")
            .memory_limits(memory_limit_mb)
            .finalize(&mut wasmer::ImportObject::default());
        
        let instance = wasmer::Instance::new(&module, &wasi_env.import_object())?;
        
        Ok(Self {
            instance,
            memory_limit_mb,
        })
    }
}

#[async_trait::async_trait]
impl PluginRuntime for WasmSandbox {
    async fn start(&self) -> Result<()> {
        if let Ok(start) = self.instance.exports.get_function("_start") {
            start.call(&[])?;
        }
        Ok(())
    }

    async fn stop(&self) -> Result<()> {
        // WASM 清理
        Ok(())
    }

    async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        let func = self.instance.exports.get_function(method)?;
        let result = func.call(&[wasmer::Value::I32(0)])?;
        Ok(serde_json::json!(result))
    }

    fn resource_usage(&self) -> ResourceUsage {
        ResourceUsage {
            memory_mb: 0.0, // 从内存限制估算
            cpu_percent: 0.0,
            uptime_seconds: 0,
        }
    }
}

/// 插件管理器
pub struct PluginManager {
    plugins: HashMap<String, Arc<Mutex<Plugin>>>,
    sandbox_enabled: bool,
}

impl PluginManager {
    pub fn new(sandbox_enabled: bool) -> Self {
        Self {
            plugins: HashMap::new(),
            sandbox_enabled,
        }
    }

    /// 加载插件
    pub async fn load_plugin(&self, path: &str) -> Result<String> {
        let metadata = self.load_metadata(path)?;
        let id = metadata.id.clone();
        
        let runtime: Box<dyn PluginRuntime> = if path.ends_with(".wasm") {
            let wasm_bytes = std::fs::read(path)?;
            Box::new(WasmSandbox::new(&wasm_bytes, 128)?)
        } else if path.ends_with(".so") || path.ends_with(".dll") {
            // 动态加载 Rust/Go 编译的插件
            Box::new(DynamicPlugin::load(path)?)
        } else {
            anyhow::bail!("Unsupported plugin format: {}", path);
        };

        let plugin = Plugin::new(
            metadata,
            PluginType::Custom,
            runtime,
        );

        self.plugins.insert(id.clone(), Arc::new(Mutex::new(plugin)));
        Ok(id)
    }

    fn load_metadata(&self, path: &str) -> Result<PluginMetadata> {
        // 从插件目录读取 plugin.json
        let meta_path = format!("{}/plugin.json", path);
        let content = std::fs::read_to_string(meta_path)?;
        let metadata: PluginMetadata = serde_json::from_str(&content)?;
        Ok(metadata)
    }

    /// 启用插件
    pub async fn enable_plugin(&self, id: &str) -> Result<()> {
        let plugin = self.plugins.get(id)
            .ok_or_else(|| anyhow::anyhow!("Plugin not found: {}", id))?;
        
        let mut p = plugin.lock().await;
        p.start().await?;
        Ok(())
    }

    /// 禁用插件
    pub async fn disable_plugin(&self, id: &str) -> Result<()> {
        let plugin = self.plugins.get(id)
            .ok_or_else(|| anyhow::anyhow!("Plugin not found: {}", id))?;
        
        let mut p = plugin.lock().await;
        p.stop().await?;
        Ok(())
    }

    /// 列出所有插件
    pub fn list_plugins(&self) -> Vec<PluginMetadata> {
        self.plugins.values()
            .map(|p| p.lock().unwrap().metadata.clone())
            .collect()
    }
}
```

## 性能目标

| 指标 | 目标值 |
|------|--------|
| 冷启动时间 | < 100ms |
| 热响应时间 | < 10ms |
| 内存占用 (空闲) | < 50MB |
| 内存占用 (满载) | < 500MB |
| 并发连接数 | > 10,000 |
| 消息吞吐量 | > 10,000 msg/s |
