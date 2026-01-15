//! AI Engine - 多模型 AI 引擎
//! 
//! 支持 OpenAI、Anthropic、Google、本地模型等

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use async_trait::async_trait;

/// 模型提供商
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ModelProvider {
    OpenAI,
    Anthropic,
    Google,
    Ollama,
    Local,
}

/// 模型配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelConfig {
    pub provider: ModelProvider,
    pub model: String,
    pub api_key: Option<String>,
    pub base_url: Option<String>,
    pub temperature: f32,
    pub max_tokens: usize,
    pub top_p: Option<f32>,
    pub top_k: Option<usize>,
    pub stop: Vec<String>,
    pub extra: HashMap<String, String>,
}

impl Default for ModelConfig {
    fn default() -> Self {
        Self {
            provider: ModelProvider::OpenAI,
            model: "gpt-4".to_string(),
            api_key: None,
            base_url: None,
            temperature: 0.7,
            max_tokens: 4096,
            top_p: None,
            top_k: None,
            stop: Vec::new(),
            extra: HashMap::new(),
        }
    }
}

/// 聊天消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub role: MessageRole,
    pub content: String,
    pub name: Option<String>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum MessageRole {
    System,
    User,
    Assistant,
    Tool,
}

/// 聊天请求
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatRequest {
    pub messages: Vec<ChatMessage>,
    pub model: Option<String>,
    pub temperature: Option<f32>,
    pub max_tokens: Option<usize>,
    pub stream: bool,
    pub tools: Vec<ToolDefinition>,
}

impl Default for ChatRequest {
    fn default() -> Self {
        Self {
            messages: Vec::new(),
            model: None,
            temperature: None,
            max_tokens: None,
            stream: false,
            tools: Vec::new(),
        }
    }
}

/// 聊天响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatResponse {
    pub id: String,
    pub model: String,
    pub content: String,
    pub role: MessageRole,
    pub finish_reason: FinishReason,
    pub usage: Usage,
    pub tool_calls: Vec<ToolCall>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FinishReason {
    Stop,
    Length,
    ContentFilter,
    ToolCalls,
}

/// Token 使用量
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Usage {
    pub prompt_tokens: usize,
    pub completion_tokens: usize,
    pub total_tokens: usize,
}

/// 工具定义
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolDefinition {
    pub name: String,
    pub description: String,
    pub parameters: serde_json::Value,
}

/// 工具调用
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    pub arguments: String,
}

/// AI Provider 接口
#[async_trait]
pub trait AIProvider: Send + Sync {
    /// 获取提供商名称
    fn name(&self) -> ModelProvider;
    
    /// 获取支持的模型列表
    fn supported_models(&self) -> Vec<String>;
    
    /// 聊天完成
    async fn chat(&self, request: ChatRequest) -> Result<ChatResponse, AIError>;
    
    /// 流式聊天完成
    async fn chat_stream(
        &self,
        request: ChatRequest,
    ) -> Result<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>, AIError>;
    
    /// 检查健康状态
    async fn health_check(&self) -> bool;
}

/// AI 错误
#[derive(Debug, thiserror::Error)]
pub enum AIError {
    #[error("API 请求失败: {0}")]
    ApiError(String),
    
    #[error("认证失败: {0}")]
    AuthenticationError(String),
    
    #[error("速率限制: {0}")]
    RateLimitError(String),
    
    #[error("配额超限: {0}")]
    QuotaExceededError(String),
    
    #[error("模型不可用: {0}")]
    ModelNotAvailable(String),
    
    #[error("上下文超限: {0}")]
    ContextLengthError(String),
    
    #[error("无效请求: {0}")]
    InvalidRequest(String),
    
    #[error("网络错误: {0}")]
    NetworkError(String),
    
    #[error("超时: {0}")]
    Timeout(String),
}

impl AIError {
    pub fn is_retryable(&self) -> bool {
        matches!(
            self,
            AIError::RateLimitError(_)
            | AIError::NetworkError(_)
            | AIError::Timeout(_)
        )
    }
    
    pub fn code(&self) -> i32 {
        match self {
            AIError::ApiError(_) => 5001,
            AIError::AuthenticationError(_) => 5002,
            AIError::RateLimitError(_) => 5003,
            AIError::QuotaExceededError(_) => 5004,
            AIError::ModelNotAvailable(_) => 5005,
            AIError::ContextLengthError(_) => 5006,
            AIError::InvalidRequest(_) => 5007,
            AIError::NetworkError(_) => 5008,
            AIError::Timeout(_) => 5009,
        }
    }
}

/// 流式响应块
#[derive(Debug, Clone)]
pub struct ChatChunk {
    pub id: String,
    pub content: String,
    pub is_final: bool,
    pub finish_reason: Option<FinishReason>,
}

/// AI Engine - AI 引擎管理器
pub struct AIEngine {
    /// 默认模型配置
    default_config: Arc<RwLock<ModelConfig>>,
    /// 模型路由配置
    routing: Arc<RwLock<ModelRouting>>,
    /// 熔断器状态
    circuit_breakers: Arc<RwLock<HashMap<String, CircuitBreaker>>>,
    /// 成本追踪
    cost_tracker: Arc<RwLock<CostTracker>>,
}

/// 模型路由
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelRouting {
    pub default_strategy: RoutingStrategy,
    pub rules: Vec<RoutingRule>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RoutingStrategy {
    Default,
    CostOptimized,
    LatencyOptimized,
    QualityOptimized,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoutingRule {
    pub condition: RoutingCondition,
    pub model: String,
    pub provider: ModelProvider,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoutingCondition {
    pub keywords: Vec<String>,
    pub min_complexity: Option<f32>,
    pub max_tokens: Option<usize>,
}

impl Default for ModelRouting {
    fn default() -> Self {
        Self {
            default_strategy: RoutingStrategy::Default,
            rules: Vec::new(),
        }
    }
}

/// 熔断器
#[derive(Debug, Clone)]
pub struct CircuitBreaker {
    pub failures: usize,
    pub last_failure: Option<std::time::Instant>,
    pub state: CircuitState,
    pub threshold: usize,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CircuitState {
    Closed,    // 正常
    Open,      // 熔断
    HalfOpen,  // 半开
}

impl Default for CircuitBreaker {
    fn default() -> Self {
        Self {
            failures: 0,
            last_failure: None,
            state: CircuitState::Closed,
            threshold: 5,
        }
    }
}

impl CircuitBreaker {
    pub fn record_success(&mut self) {
        self.failures = 0;
        self.state = CircuitState::Closed;
    }
    
    pub fn record_failure(&mut self) {
        self.failures += 1;
        self.last_failure = Some(std::time::Instant::now());
        
        if self.failures >= self.threshold {
            self.state = CircuitState::Open;
        }
    }
    
    pub fn can_execute(&self) -> bool {
        match self.state {
            CircuitState::Closed => true,
            CircuitState::Open => {
                // 如果打开超过 30 秒，尝试半开
                if let Some(last) = self.last_failure {
                    if last.elapsed() > std::time::Duration::from_secs(30) {
                        return true;
                    }
                }
                false
            }
            CircuitState::HalfOpen => true,
        }
    }
}

/// 成本追踪器
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct CostTracker {
    pub total_cost: f64,
    pub prompt_tokens_used: usize,
    pub completion_tokens_used: usize,
    pub requests_count: usize,
    pub by_model: HashMap<String, ModelCost>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelCost {
    pub requests: usize,
    pub prompt_tokens: usize,
    pub completion_tokens: usize,
    pub cost: f64,
}

impl AIEngine {
    pub fn new() -> Self {
        Self {
            default_config: Arc::new(RwLock::new(ModelConfig::default())),
            routing: Arc::new(RwLock::new(ModelRouting::default())),
            circuit_breakers: Arc::new(RwLock::new(HashMap::new())),
            cost_tracker: Arc::new(RwLock::new(CostTracker::default())),
        }
    }
    
    /// 设置默认模型配置
    pub fn set_default_config(&self, config: ModelConfig) {
        let mut default = self.default_config.write();
        *default = config;
    }
    
    /// 获取默认配置
    pub fn get_default_config(&self) -> ModelConfig {
        self.default_config.read().clone()
    }
    
    /// 设置路由策略
    pub fn set_routing(&self, routing: ModelRouting) {
        let mut r = self.routing.write();
        *r = routing;
    }
    
    /// 路由请求到合适的模型
    pub fn route_request(&self, request: &ChatRequest) -> ModelConfig {
        let routing = self.routing.read();
        let mut config = self.default_config.read().clone();
        
        // 根据策略选择模型
        match routing.default_strategy {
            RoutingStrategy::CostOptimized => {
                // 选择最便宜的模型
                config.model = "gpt-3.5-turbo".to_string();
                config.provider = ModelProvider::OpenAI;
            }
            RoutingStrategy::QualityOptimized => {
                // 选择最高质量的模型
                config.model = "gpt-4".to_string();
                config.provider = ModelProvider::OpenAI;
            }
            RoutingStrategy::LatencyOptimized => {
                // 选择延迟最低的模型
                config.model = "gpt-3.5-turbo".to_string();
                config.provider = ModelProvider::OpenAI;
            }
            RoutingStrategy::Default => {
                // 使用默认配置
            }
        }
        
        // 应用请求中的覆盖
        if let Some(model) = &request.model {
            config.model = model.clone();
        }
        if let Some(temp) = request.temperature {
            config.temperature = temp;
        }
        if let Some(max) = request.max_tokens {
            config.max_tokens = max;
        }
        
        config
    }
    
    /// 检查熔断器状态
    pub fn check_circuit(&self, model: &str) -> bool {
        let breakers = self.circuit_breakers.read();
        match breakers.get(model) {
            Some(cb) => cb.can_execute(),
            None => true,
        }
    }
    
    /// 记录成功
    pub fn record_success(&self, model: &str, usage: &Usage, cost: f64) {
        let mut tracker = self.cost_tracker.write();
        tracker.total_cost += cost;
        tracker.prompt_tokens_used += usage.prompt_tokens;
        tracker.completion_tokens_used += usage.completion_tokens;
        tracker.requests_count += 1;
        
        // 更新模型成本
        let model_cost = tracker.by_model.entry(model.to_string()).or_default();
        model_cost.requests += 1;
        model_cost.prompt_tokens += usage.prompt_tokens;
        model_cost.completion_tokens += usage.completion_tokens;
        model_cost.cost += cost;
        
        // 重置熔断器
        let mut breakers = self.circuit_breakers.write();
        if let Some(cb) = breakers.get_mut(model) {
            cb.record_success();
        }
    }
    
    /// 记录失败
    pub fn record_failure(&self, model: &str) {
        let mut breakers = self.circuit_breakers.write();
        let cb = breakers.entry(model.to_string()).or_default();
        cb.record_failure();
    }
    
    /// 获取成本追踪
    pub fn get_cost_tracker(&self) -> CostTracker {
        self.cost_tracker.read().clone()
    }
    
    /// 估算成本
    pub fn estimate_cost(&self, config: &ModelConfig, tokens: usize) -> f64 {
        match config.provider {
            ModelProvider::OpenAI => {
                match config.model.as_str() {
                    "gpt-4" => tokens as f64 * 0.03 / 1000.0, // $0.03/1K tokens
                    "gpt-4-32k" => tokens as f64 * 0.06 / 1000.0,
                    "gpt-3.5-turbo" => tokens as f64 * 0.002 / 1000.0,
                    "gpt-3.5-turbo-16k" => tokens as f64 * 0.004 / 1000.0,
                    _ => tokens as f64 * 0.002 / 1000.0,
                }
            }
            ModelProvider::Anthropic => {
                match config.model.as_str() {
                    "claude-3-opus" => tokens as f64 * 0.015 / 1000.0,
                    "claude-3-sonnet" => tokens as f64 * 0.003 / 1000.0,
                    "claude-3-haiku" => tokens as f64 * 0.00025 / 1000.0,
                    _ => tokens as f64 * 0.003 / 1000.0,
                }
            }
            _ => 0.0,
        }
    }
}

impl Default for AIEngine {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_circuit_breaker() {
        let mut cb = CircuitBreaker::default();
        assert!(cb.can_execute());
        
        // 模拟失败
        for _ in 0..4 {
            cb.record_failure();
        }
        assert!(cb.can_execute()); // 还没达到阈值
        
        cb.record_failure(); // 第5次失败
        assert!(!cb.can_execute()); // 应该打开
    }

    #[test]
    fn test_cost_estimation() {
        let engine = AIEngine::new();
        let config = ModelConfig {
            provider: ModelProvider::OpenAI,
            model: "gpt-4".to_string(),
            ..Default::default()
        };
        
        let cost = engine.estimate_cost(&config, 1000);
        assert!((cost - 0.03).abs() < 0.001);
    }
}
