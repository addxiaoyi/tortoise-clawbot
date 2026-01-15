//! Agent 核心引擎
//!
//! 负责代理的创建、消息处理、思维推理

use anyhow::{Context, Result};
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::{mpsc, RwLock};
use tokio::time::timeout;
use uuid::Uuid;

use super::context::{AgentContext, ContextManager};
use super::model::{ModelAdapter, ModelProvider, ModelResponse, create_model_adapter};
use super::thinking::{ThinkEngine, ThinkMode, ThinkResult};
use super::tool_call::{ToolCall, ToolCallExecutor, ToolRegistry};
use super::streaming::{StreamingResponse, EventSink};

/// 代理配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentConfig {
    /// 代理唯一 ID
    pub id: String,
    /// 代理名称
    pub name: String,
    /// 代理描述
    pub description: Option<String>,
    /// 模型提供商配置
    pub model_provider: ModelProvider,
    /// 默认思维模式
    pub default_thinking: ThinkMode,
    /// 最大上下文长度 (tokens)
    pub max_context: usize,
    /// 默认温度参数
    pub temperature: f32,
    /// 系统提示词
    pub system_prompt: Option<String>,
    /// 工具白名单
    pub allowed_tools: Option<Vec<String>>,
    /// 工具黑名单
    pub blocked_tools: Option<Vec<String>>,
    /// 超时配置
    pub timeout: Option<Duration>,
    /// 启用思维链
    pub enable_thinking: bool,
    /// 启用记忆
    pub enable_memory: bool,
    /// 启用主动推理
    pub enable_proactive_reasoning: bool,
}

impl Default for AgentConfig {
    fn default() -> Self {
        Self {
            id: "default".to_string(),
            name: "Tortoise".to_string(),
            description: Some("Tortoise AI Agent".to_string()),
            model_provider: ModelProvider::Ollama {
                model: "llama3".to_string(),
                base_url: "http://localhost:11434".to_string(),
            },
            default_thinking: ThinkMode::Balanced,
            max_context: 8192,
            temperature: 0.7,
            system_prompt: Some(Self::default_system_prompt()),
            allowed_tools: None,
            blocked_tools: None,
            timeout: Some(Duration::from_secs(60)),
            enable_thinking: true,
            enable_memory: true,
            enable_proactive_reasoning: false,
        }
    }
}

impl AgentConfig {
    /// 默认系统提示词
    pub fn default_system_prompt() -> String {
        r#"You are Tortoise, an advanced AI agent.

## 核心能力
- 多模型支持：可根据需要切换不同的 AI 模型
- 工具调用：可以调用各种工具来完成任务
- 记忆系统：可以记住之前的对话和重要信息
- 多代理协作：可以与其他代理协作解决问题

## 行为准则
1. helpful：尽你所能帮助用户
2. 安全优先：拒绝执行危险或有害的操作
3. 透明：诚实地告知你的能力和限制
4. 隐私保护：不泄露敏感信息

## 响应风格
- 简洁明了，避免冗长
- 主动思考，提供最佳解决方案
- 必要时主动调用工具"#.to_string()
    }

    /// 从环境变量加载 API key
    pub fn with_env_api_key(mut self) -> Self {
        match &self.model_provider {
            ModelProvider::OpenAI { api_key, .. } => {
                if api_key.starts_with("${") && api_key.ends_with("}") {
                    let env_var = &api_key[2..api_key.len() - 1];
                    if let Ok(key) = std::env::var(env_var) {
                        self.model_provider = match &self.model_provider {
                            ModelProvider::OpenAI { model, base_url, .. } => {
                                ModelProvider::OpenAI {
                                    model: model.clone(),
                                    api_key: key,
                                    base_url: base_url.clone(),
                                }
                            }
                            _ => self.model_provider,
                        };
                    }
                }
            }
            ModelProvider::Anthropic { api_key, .. } => {
                if api_key.starts_with("${") && api_key.ends_with("}") {
                    let env_var = &api_key[2..api_key.len() - 1];
                    if let Ok(key) = std::env::var(env_var) {
                        self.model_provider = match &self.model_provider {
                            ModelProvider::Anthropic { model, base_url, .. } => {
                                ModelProvider::Anthropic {
                                    model: model.clone(),
                                    api_key: key,
                                    base_url: base_url.clone(),
                                }
                            }
                            _ => self.model_provider,
                        };
                    }
                }
            }
            _ => {}
        }
        self
    }
}

/// 代理状态
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum AgentStatus {
    /// 空闲状态
    Idle,
    /// 处理中
    Busy,
    /// 错误状态
    Error(String),
    /// 初始化中
    Initializing,
}

/// 消息角色
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MessageRole {
    System,
    User,
    Assistant,
    Tool,
    /// 代理内部思考
    AssistantThinking,
}

impl MessageRole {
    pub fn as_str(&self) -> &'static str {
        match self {
            MessageRole::System => "system",
            MessageRole::User => "user",
            MessageRole::Assistant => "assistant",
            MessageRole::Tool => "tool",
            MessageRole::AssistantThinking => "assistant_thinking",
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "system" => Some(MessageRole::System),
            "user" => Some(MessageRole::User),
            "assistant" => Some(MessageRole::Assistant),
            "tool" => Some(MessageRole::Tool),
            "assistant_thinking" => Some(MessageRole::AssistantThinking),
            _ => None,
        }
    }
}

/// 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    /// 消息 ID
    pub id: String,
    /// 角色
    pub role: MessageRole,
    /// 内容
    pub content: Content,
    /// 工具调用
    #[serde(default)]
    pub tool_calls: Vec<ToolCall>,
    /// 工具结果
    #[serde(default)]
    pub tool_results: Vec<ToolResult>,
    /// 元数据
    #[serde(default)]
    pub metadata: MessageMetadata,
    /// 创建时间
    pub created_at: chrono::DateTime<chrono::Utc>,
}

impl Message {
    /// 创建用户消息
    pub fn user(content: impl Into<Content>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            role: MessageRole::User,
            content: content.into(),
            tool_calls: vec![],
            tool_results: vec![],
            metadata: MessageMetadata::default(),
            created_at: chrono::Utc::now(),
        }
    }

    /// 创建助手消息
    pub fn assistant(content: impl Into<Content>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            role: MessageRole::Assistant,
            content: content.into(),
            tool_calls: vec![],
            tool_results: vec![],
            metadata: MessageMetadata::default(),
            created_at: chrono::Utc::now(),
        }
    }

    /// 创建系统消息
    pub fn system(content: impl Into<Content>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            role: MessageRole::System,
            content: content.into(),
            tool_calls: vec![],
            tool_results: vec![],
            metadata: MessageMetadata::default(),
            created_at: chrono::Utc::now(),
        }
    }

    /// 创建工具结果消息
    pub fn tool_result(call_id: String, result: Result<String, String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            role: MessageRole::Tool,
            content: Content::Text(result.unwrap_or_else(|e| e)),
            tool_calls: vec![],
            tool_results: vec![ToolResult {
                call_id,
                success: result.is_ok(),
            }],
            metadata: MessageMetadata::default(),
            created_at: chrono::Utc::now(),
        }
    }
}

/// 内容类型
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Content {
    /// 纯文本
    Text(String),
    /// 结构化内容
    Structured {
        #[serde(rename = "type")]
        content_type: String,
        #[serde(flatten)]
        data: serde_json::Value,
    },
    /// 多模态内容
    Multimodal {
        text: Option<String>,
        images: Vec<ImageContent>,
        audio: Option<AudioContent>,
        video: Option<VideoContent>,
    },
}

impl From<String> for Content {
    fn from(s: String) -> Self {
        Content::Text(s)
    }
}

impl From<&str> for Content {
    fn from(s: &str) -> Self {
        Content::Text(s.to_string())
    }
}

/// 图像内容
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImageContent {
    pub url: Option<String>,
    pub base64: Option<String>,
    pub media_type: Option<String>,
}

/// 音频内容
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AudioContent {
    pub url: Option<String>,
    pub base64: Option<String>,
    pub media_type: Option<String>,
    pub duration_secs: Option<f32>,
}

/// 视频内容
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VideoContent {
    pub url: Option<String>,
    pub base64: Option<String>,
    pub media_type: Option<String>,
    pub duration_secs: Option<f32>,
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
    pub success: bool,
}

/// 消息元数据
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct MessageMetadata {
    /// 消息来源通道
    pub channel: Option<String>,
    /// 消息来源用户
    pub user_id: Option<String>,
    /// 会话 ID
    pub session_id: Option<String>,
    /// 父消息 ID
    pub parent_id: Option<String>,
    /// 思考时间 (ms)
    pub thinking_time_ms: Option<u64>,
    /// 额外数据
    #[serde(flatten)]
    pub extra: serde_json::Value,
}

/// 聊天选项
#[derive(Debug, Clone, Default)]
pub struct ChatOptions {
    /// 思维模式
    pub thinking_mode: Option<ThinkMode>,
    /// 系统提示词覆盖
    pub system_prompt: Option<String>,
    /// 最大 tokens
    pub max_tokens: Option<usize>,
    /// 温度参数
    pub temperature: Option<f32>,
    /// 停止序列
    pub stop: Option<Vec<String>>,
    /// 会话 ID
    pub session_id: Option<String>,
    /// 是否流式输出
    pub streaming: Option<bool>,
    /// 工具列表
    pub tools: Option<Vec<String>>,
    /// 上下文覆盖
    pub context_override: Option<Vec<Message>>,
}

/// 代理事件
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum AgentEvent {
    /// 开始思考
    ThinkingStarted {
        mode: ThinkMode,
    },
    /// 思考中 (流式)
    Thinking {
        content: String,
    },
    /// 思考完成
    ThinkingComplete {
        result: String,
    },
    /// 模型开始生成
    GenerationStarted,
    /// 模型生成中 (流式)
    Generation {
        content: String,
    },
    /// 工具调用请求
    ToolCall {
        call: ToolCall,
    },
    /// 工具执行开始
    ToolExecutionStarted {
        call_id: String,
        tool_name: String,
    },
    /// 工具执行完成
    ToolExecutionComplete {
        call_id: String,
        result: String,
    },
    /// 工具执行失败
    ToolExecutionFailed {
        call_id: String,
        error: String,
    },
    /// 响应完成
    ResponseComplete {
        content: String,
        tool_results: Vec<ToolResult>,
    },
    /// 错误
    Error {
        code: String,
        message: String,
    },
    /// 元数据更新
    MetadataUpdated {
        key: String,
        value: serde_json::Value,
    },
}

impl AgentEvent {
    /// 获取事件类型名称
    pub fn event_type(&self) -> &'static str {
        match self {
            AgentEvent::ThinkingStarted { .. } => "thinking_started",
            AgentEvent::Thinking { .. } => "thinking",
            AgentEvent::ThinkingComplete { .. } => "thinking_complete",
            AgentEvent::GenerationStarted { .. } => "generation_started",
            AgentEvent::Generation { .. } => "generation",
            AgentEvent::ToolCall { .. } => "tool_call",
            AgentEvent::ToolExecutionStarted { .. } => "tool_execution_started",
            AgentEvent::ToolExecutionComplete { .. } => "tool_execution_complete",
            AgentEvent::ToolExecutionFailed { .. } => "tool_execution_failed",
            AgentEvent::ResponseComplete { .. } => "response_complete",
            AgentEvent::Error { .. } => "error",
            AgentEvent::MetadataUpdated { .. } => "metadata_updated",
        }
    }
}

/// 代理接口
#[async_trait]
pub trait Agent: Send + Sync {
    /// 获取代理 ID
    fn id(&self) -> &str;
    
    /// 获取代理名称
    fn name(&self) -> &str;
    
    /// 获取当前状态
    fn status(&self) -> AgentStatus;
    
    /// 处理聊天请求
    async fn chat(&self, messages: Vec<Message>, options: ChatOptions) -> Result<StreamingResponse>;
    
    /// 同步聊天请求 (等待完整响应)
    async fn chat_sync(&self, messages: Vec<Message>, options: ChatOptions) -> Result<Message>;
    
    /// 获取配置
    fn config(&self) -> &AgentConfig;
    
    /// 更新配置
    async fn update_config(&self, config: AgentConfig) -> Result<()>;
    
    /// 重置代理
    async fn reset(&self) -> Result<()>;
    
    /// 获取统计信息
    async fn stats(&self) -> AgentStats;
}

/// 代理统计
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentStats {
    pub total_requests: u64,
    pub total_tokens: u64,
    pub total_tool_calls: u64,
    pub average_response_time_ms: f64,
    pub last_request_at: Option<chrono::DateTime<chrono::Utc>>,
    pub error_count: u64,
    pub uptime_seconds: u64,
}

/// Tortoise 代理实现
pub struct TortoiseAgent {
    config: RwLock<AgentConfig>,
    status: RwLock<AgentStatus>,
    model: Arc<dyn ModelAdapter>,
    tool_registry: Arc<ToolRegistry>,
    tool_executor: Arc<ToolCallExecutor>,
    context_manager: Arc<ContextManager>,
    think_engine: Arc<ThinkEngine>,
    stats: RwLock<AgentStats>,
    created_at: chrono::DateTime<chrono::Utc>,
}

impl TortoiseAgent {
    /// 创建新代理
    pub async fn new(config: AgentConfig) -> Result<Self> {
        tracing::info!("Creating agent: {} ({})", config.name, config.id);
        
        let config = config.with_env_api_key();
        
        // 创建模型适配器
        let model = create_model_adapter(&config.model_provider).await
            .context("Failed to create model adapter")?;
        
        // 创建工具注册表
        let tool_registry = Arc::new(ToolRegistry::new());
        
        // 创建工具执行器
        let tool_executor = Arc::new(ToolCallExecutor::new(Arc::clone(&tool_registry)));
        
        // 创建上下文管理器
        let context_manager = Arc::new(ContextManager::new(config.max_context));
        
        // 创建思维引擎
        let think_engine = Arc::new(ThinkEngine::new(
            config.default_thinking,
            config.enable_thinking,
        ));
        
        let now = chrono::Utc::now();
        
        Ok(Self {
            config: RwLock::new(config),
            status: RwLock::new(AgentStatus::Idle),
            model: Arc::from(model),
            tool_registry,
            tool_executor,
            context_manager,
            think_engine,
            stats: RwLock::new(AgentStats {
                total_requests: 0,
                total_tokens: 0,
                total_tool_calls: 0,
                average_response_time_ms: 0.0,
                last_request_at: None,
                error_count: 0,
                uptime_seconds: 0,
            }),
            created_at: now,
        })
    }

    /// 注册工具
    pub fn register_tool<T: super::tool::Tool + 'static>(&self, tool: T) -> Result<()> {
        self.tool_registry.register(Box::new(tool))
    }

    /// 处理聊天
    async fn process_chat(&self, messages: Vec<Message>, options: ChatOptions) -> Result<StreamingResponse> {
        // 更新状态
        *self.status.write().await = AgentStatus::Busy;
        
        // 获取配置
        let config = self.config.read().await;
        let thinking_mode = options.thinking_mode.unwrap_or(config.default_thinking);
        let system_prompt = options.system_prompt.as_ref().or(config.system_prompt.as_ref());
        let temperature = options.temperature.unwrap_or(config.temperature);
        let max_tokens = options.max_tokens.unwrap_or(4096);
        
        // 创建事件通道
        let (tx, rx) = mpsc::channel(100);
        let event_sink = EventSink::new(tx);
        
        // 异步处理
        let model = Arc::clone(&self.model);
        let tool_executor = Arc::clone(&self.tool_executor);
        let think_engine = Arc::clone(&self.think_engine);
        let context_manager = Arc::clone(&self.context_manager);
        
        tokio::spawn(async move {
            let result = Self::process_with_thinking(
                &model,
                &tool_executor,
                &think_engine,
                &context_manager,
                messages,
                system_prompt,
                thinking_mode,
                temperature,
                max_tokens,
                options.tools,
                event_sink,
            ).await;
            
            if let Err(e) = result {
                tracing::error!("Chat processing error: {:?}", e);
            }
        });
        
        Ok(StreamingResponse::new(rx))
    }

    /// 带思维处理的聊天流程
    async fn process_with_thinking(
        model: &Arc<dyn ModelAdapter>,
        tool_executor: &Arc<ToolCallExecutor>,
        think_engine: &Arc<ThinkEngine>,
        context_manager: &Arc<ContextManager>,
        messages: Vec<Message>,
        system_prompt: Option<&String>,
        thinking_mode: ThinkMode,
        temperature: f32,
        max_tokens: usize,
        allowed_tools: Option<Vec<String>>,
        event_sink: EventSink,
    ) -> Result<()> {
        // 1. 发送思考开始事件
        event_sink.send(AgentEvent::ThinkingStarted { mode: thinking_mode }).await?;
        
        // 2. 思维推理阶段
        let think_result = think_engine.think(
            &messages,
            thinking_mode,
            system_prompt,
            event_sink.clone(),
        ).await?;
        
        // 3. 发送生成开始事件
        event_sink.send(AgentEvent::GenerationStarted).await?;
        
        // 4. 模型调用阶段
        let mut full_response = String::new();
        let mut tool_calls = Vec::new();
        let mut tool_results = Vec::new();
        
        // 构建上下文
        let context_messages = context_manager.prepare_context(&messages, &think_result)?;
        
        // 调用模型 (流式)
        let mut stream = model.chat_stream(
            &context_messages,
            system_prompt,
            temperature,
            max_tokens,
            allowed_tools.as_ref(),
        ).await?;
        
        while let Some(chunk) = stream.next().await {
            match chunk? {
                ModelResponse::Content { text } => {
                    full_response.push_str(&text);
                    event_sink.send(AgentEvent::Generation { content: text }).await?;
                }
                ModelResponse::ToolCall { call } => {
                    tool_calls.push(call.clone());
                    event_sink.send(AgentEvent::ToolCall { call }).await?;
                }
                ModelResponse::Done => break,
                ModelResponse::Error(e) => {
                    event_sink.send(AgentEvent::Error {
                        code: "MODEL_ERROR".to_string(),
                        message: e,
                    }).await?;
                }
            }
        }
        
        // 5. 执行工具调用
        for call in &tool_calls {
            event_sink.send(AgentEvent::ToolExecutionStarted {
                call_id: call.id.clone(),
                tool_name: call.name.clone(),
            }).await?;
            
            match tool_executor.execute(&call).await {
                Ok(result) => {
                    tool_results.push(ToolResult {
                        call_id: call.id.clone(),
                        success: true,
                    });
                    event_sink.send(AgentEvent::ToolExecutionComplete {
                        call_id: call.id.clone(),
                        result,
                    }).await?;
                }
                Err(e) => {
                    tool_results.push(ToolResult {
                        call_id: call.id.clone(),
                        success: false,
                    });
                    event_sink.send(AgentEvent::ToolExecutionFailed {
                        call_id: call.id.clone(),
                        error: e.to_string(),
                    }).await?;
                }
            }
        }
        
        // 6. 发送完成事件
        event_sink.send(AgentEvent::ResponseComplete {
            content: full_response,
            tool_results,
        }).await?;
        
        Ok(())
    }
}

#[async_trait]
impl Agent for TortoiseAgent {
    fn id(&self) -> &str {
        // 暂时使用锁，之后优化
        "temp"
    }
    
    fn name(&self) -> &str {
        "Tortoise"
    }
    
    fn status(&self) -> AgentStatus {
        // 需要实现
        AgentStatus::Idle
    }
    
    async fn chat(&self, messages: Vec<Message>, options: ChatOptions) -> Result<StreamingResponse> {
        self.process_chat(messages, options).await
    }
    
    async fn chat_sync(&self, messages: Vec<Message>, options: ChatOptions) -> Result<Message> {
        let mut response = self.process_chat(messages, options).await?;
        
        let mut full_content = String::new();
        let mut tool_results = Vec::new();
        
        while let Some(event) = response.events.recv().await {
            match event {
                AgentEvent::Generation { content } => {
                    full_content.push_str(&content);
                }
                AgentEvent::ToolExecutionComplete { call_id, result } => {
                    tool_results.push(ToolResult {
                        call_id,
                        success: true,
                    });
                }
                AgentEvent::ResponseComplete { content, tool_results: tr } => {
                    full_content = content;
                    tool_results = tr;
                }
                _ => {}
            }
        }
        
        Ok(Message::assistant(full_content))
    }
    
    fn config(&self) -> &AgentConfig {
        // 需要实现
        &AgentConfig::default()
    }
    
    async fn update_config(&self, config: AgentConfig) -> Result<()> {
        *self.config.write().await = config;
        Ok(())
    }
    
    async fn reset(&self) -> Result<()> {
        self.context_manager.clear().await;
        *self.status.write().await = AgentStatus::Idle;
        Ok(())
    }
    
    async fn stats(&self) -> AgentStats {
        self.stats.read().await.clone()
    }
}

/// 创建代理
pub async fn create_agent(config: AgentConfig) -> Result<Arc<dyn Agent>> {
    let agent = TortoiseAgent::new(config).await?;
    Ok(Arc::new(agent))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_agent_config_default() {
        let config = AgentConfig::default();
        assert_eq!(config.id, "default");
        assert_eq!(config.name, "Tortoise");
        assert_eq!(config.default_thinking, ThinkMode::Balanced);
    }

    #[tokio::test]
    async fn test_message_creation() {
        let msg = Message::user("Hello, Tortoise!");
        assert_eq!(msg.role, MessageRole::User);
        
        let assistant = Message::assistant("Hello! How can I help you?");
        assert_eq!(assistant.role, MessageRole::Assistant);
    }

    #[tokio::test]
    async fn test_chat_options_default() {
        let options = ChatOptions::default();
        assert!(options.thinking_mode.is_none());
        assert!(options.system_prompt.is_none());
    }
}
