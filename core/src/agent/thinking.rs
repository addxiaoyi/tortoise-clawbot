//! 思维引擎
//!
//! 负责代理的思维推理，包括快速响应、深度思考、研究模式等

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

use super::engine::{Message, AgentEvent};
use super::streaming::EventSink;

/// 思维模式
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ThinkMode {
    /// 快速响应 (< 100ms)
    /// 适用于简单查询和日常对话
    Fast,
    /// 平衡模式 (500ms)
    /// 默认模式，在速度和深度之间取得平衡
    Balanced,
    /// 深度思考 (2s)
    /// 适用于复杂问题分析
    Deep,
    /// 研究模式 (5s)
    /// 适用于需要深入研究和推理的问题
    Research,
    /// 创意发散 (10s)
    /// 适用于需要创意和创新的任务
    Creative,
}

impl ThinkMode {
    /// 获取超时时间
    pub fn timeout(&self) -> Duration {
        match self {
            ThinkMode::Fast => Duration::from_millis(100),
            ThinkMode::Balanced => Duration::from_millis(500),
            ThinkMode::Deep => Duration::from_secs(2),
            ThinkMode::Research => Duration::from_secs(5),
            ThinkMode::Creative => Duration::from_secs(10),
        }
    }

    /// 获取超时毫秒
    pub fn timeout_ms(&self) -> u64 {
        self.timeout().as_millis() as u64
    }

    /// 获取默认温度
    pub fn default_temperature(&self) -> f32 {
        match self {
            ThinkMode::Fast => 0.0,
            ThinkMode::Balanced => 0.5,
            ThinkMode::Deep => 0.7,
            ThinkMode::Research => 0.6,
            ThinkMode::Creative => 1.0,
        }
    }

    /// 获取思考深度级别
    pub fn depth_level(&self) -> u8 {
        match self {
            ThinkMode::Fast => 1,
            ThinkMode::Balanced => 2,
            ThinkMode::Deep => 3,
            ThinkMode::Research => 4,
            ThinkMode::Creative => 5,
        }
    }

    /// 获取描述
    pub fn description(&self) -> &'static str {
        match self {
            ThinkMode::Fast => "Quick response mode for simple queries",
            ThinkMode::Balanced => "Balanced mode between speed and depth",
            ThinkMode::Deep => "Deep thinking for complex analysis",
            ThinkMode::Research => "Research mode for thorough investigation",
            ThinkMode::Creative => "Creative mode for innovative tasks",
        }
    }

    /// 获取思维提示
    pub fn prompt_suffix(&self) -> &'static str {
        match self {
            ThinkMode::Fast => "Respond concisely and directly.",
            ThinkMode::Balanced => "Provide a balanced response.",
            ThinkMode::Deep => "Think step by step and explain your reasoning.",
            ThinkMode::Research => "Conduct thorough research and cite sources if possible.",
            ThinkMode::Creative => "Think creatively and explore unconventional ideas.",
        }
    }
}

impl Default for ThinkMode {
    fn default() -> Self {
        ThinkMode::Balanced
    }
}

impl std::fmt::Display for ThinkMode {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ThinkMode::Fast => write!(f, "fast"),
            ThinkMode::Balanced => write!(f, "balanced"),
            ThinkMode::Deep => write!(f, "deep"),
            ThinkMode::Research => write!(f, "research"),
            ThinkMode::Creative => write!(f, "creative"),
        }
    }
}

/// 思维结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThinkResult {
    /// 思维模式
    pub mode: ThinkMode,
    /// 思考过程
    pub reasoning: String,
    /// 思考时间 (ms)
    pub thinking_time_ms: u64,
    /// 置信度
    pub confidence: f32,
    /// 关键要点
    pub key_points: Vec<String>,
    /// 需要调用的工具
    pub required_tools: Vec<String>,
}

/// 思维引擎
pub struct ThinkEngine {
    mode: RwLock<ThinkMode>,
    enabled: bool,
    cache: RwLock<lru::LruCache<String, ThinkResult>>,
}

impl ThinkEngine {
    /// 创建新的思维引擎
    pub fn new(default_mode: ThinkMode, enabled: bool) -> Self {
        Self {
            mode: RwLock::new(default_mode),
            enabled,
            cache: RwLock::new(lru::LruCache::new(100)),
        }
    }

    /// 设置思维模式
    pub async fn set_mode(&self, mode: ThinkMode) {
        *self.mode.write().await = mode;
    }

    /// 获取当前思维模式
    pub async fn get_mode(&self) -> ThinkMode {
        *self.mode.read().await
    }

    /// 执行思维推理
    pub async fn think(
        &self,
        messages: &[Message],
        mode: ThinkMode,
        system_prompt: Option<&String>,
        event_sink: EventSink,
    ) -> Result<ThinkResult> {
        let start = Instant::now();
        
        // 如果禁用思维，直接返回
        if !self.enabled {
            return Ok(ThinkResult {
                mode,
                reasoning: String::new(),
                thinking_time_ms: 0,
                confidence: 1.0,
                key_points: vec![],
                required_tools: vec![],
            });
        }

        // 根据模式决定是否需要深度思考
        if mode == ThinkMode::Fast {
            return Ok(ThinkResult {
                mode,
                reasoning: String::new(),
                thinking_time_ms: start.elapsed().as_millis() as u64,
                confidence: 1.0,
                key_points: vec![],
                required_tools: vec![],
            });
        }

        // 生成缓存键
        let cache_key = self.generate_cache_key(messages, system_prompt);
        
        // 检查缓存
        {
            let cache = self.cache.read().await;
            if let Some(cached) = cache.get(&cache_key) {
                if cached.mode == mode {
                    tracing::debug!("Using cached think result");
                    return Ok(cached.clone());
                }
            }
        }

        // 执行思维推理
        let reasoning = self.perform_reasoning(messages, mode, system_prompt, &event_sink).await?;
        
        // 分析结果
        let (confidence, key_points, required_tools) = self.analyze_result(&reasoning);
        
        let result = ThinkResult {
            mode,
            reasoning: reasoning.clone(),
            thinking_time_ms: start.elapsed().as_millis() as u64,
            confidence,
            key_points,
            required_tools,
        };

        // 缓存结果
        {
            let mut cache = self.cache.write().await;
            cache.put(cache_key, result.clone());
        }

        // 发送思考完成事件
        event_sink.send(AgentEvent::ThinkingComplete {
            result: reasoning,
        }).await?;

        Ok(result)
    }

    /// 生成缓存键
    fn generate_cache_key(&self, messages: &[Message], system_prompt: Option<&String>) -> String {
        let mut key = String::new();
        
        if let Some(prompt) = system_prompt {
            key.push_str(prompt);
            key.push('\n');
        }
        
        for msg in messages {
            key.push_str(&format!("{:?}:{:?}\n", msg.role, msg.content));
        }
        
        use std::collections::hash_map::DefaultHasher;
        use std::hash::{Hash, Hasher};
        let mut hasher = DefaultHasher::new();
        key.hash(&mut hasher);
        format!("{:x}", hasher.finish())
    }

    /// 执行推理
    async fn perform_reasoning(
        &self,
        messages: &[Message],
        mode: ThinkMode,
        system_prompt: Option<&String>,
        event_sink: &EventSink,
    ) -> Result<String> {
        let reasoning_steps = match mode {
            ThinkMode::Fast => vec!["Quick check".to_string()],
            ThinkMode::Balanced => vec![
                "Understanding request".to_string(),
                "Identifying key points".to_string(),
                "Formulating response".to_string(),
            ],
            ThinkMode::Deep => vec![
                "Understanding request in depth".to_string(),
                "Analyzing context and constraints".to_string(),
                "Identifying key factors".to_string(),
                "Evaluating options".to_string(),
                "Formulating detailed response".to_string(),
            ],
            ThinkMode::Research => vec![
                "Understanding research question".to_string(),
                "Breaking down into sub-problems".to_string(),
                "Gathering relevant information".to_string(),
                "Analyzing evidence".to_string(),
                "Synthesizing findings".to_string(),
                "Drawing conclusions".to_string(),
            ],
            ThinkMode::Creative => vec![
                "Understanding creative brief".to_string(),
                "Brainstorming initial ideas".to_string(),
                "Exploring unconventional approaches".to_string(),
                "Combining disparate concepts".to_string(),
                "Refining and iterating".to_string(),
            ],
        };

        let mut reasoning = String::new();
        
        for (i, step) in reasoning_steps.iter().enumerate() {
            // 发送思考进度
            event_sink.send(AgentEvent::Thinking {
                content: format!("[{}] {}", i + 1, step),
            }).await?;
            
            reasoning.push_str(&format!("Step {}: {}\n", i + 1, step));
            
            // 模拟思考延迟
            tokio::time::sleep(Duration::from_millis(50)).await;
        }

        Ok(reasoning)
    }

    /// 分析结果
    fn analyze_result(&self, reasoning: &str) -> (f32, Vec<String>, Vec<String>) {
        let confidence = if reasoning.contains("conclusion") || reasoning.contains("final") {
            0.9
        } else if reasoning.contains("analysis") || reasoning.contains("evaluating") {
            0.7
        } else {
            0.5
        };

        let key_points: Vec<String> = reasoning
            .lines()
            .filter(|line| !line.starts_with("Step"))
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .take(5)
            .collect();

        let required_tools: Vec<String> = if reasoning.contains("search") {
            vec!["web_search".to_string()]
        } else if reasoning.contains("calculate") || reasoning.contains("compute") {
            vec!["calculator".to_string()]
        } else {
            vec![]
        };

        (confidence, key_points, required_tools)
    }

    /// 清除缓存
    pub async fn clear_cache(&self) {
        self.cache.write().await.clear();
    }
}

// === LRU 缓存 (简化实现) ===

mod lru {
    use std::collections::{HashMap, VecDeque};
    use std::hash::Hash;

    pub struct LruCache<K, V> {
        capacity: usize,
        cache: HashMap<K, V>,
        order: VecDeque<K>,
    }

    impl<K: Hash + Eq, V> LruCache<K, V> {
        pub fn new(capacity: usize) -> Self {
            Self {
                capacity,
                cache: HashMap::new(),
                order: VecDeque::new(),
            }
        }

        pub fn get(&self, key: &K) -> Option<&V> {
            self.cache.get(key)
        }

        pub fn put(&mut self, key: K, value: V) {
            if self.cache.contains_key(&key) {
                // 更新现有键
                self.order.retain(|k| k != &key);
            } else if self.cache.len() >= self.capacity {
                // 移除最旧的
                if let Some(oldest) = self.order.pop_front() {
                    self.cache.remove(&oldest);
                }
            }
            
            self.cache.insert(key.clone(), value);
            self.order.push_back(key);
        }

        pub fn clear(&mut self) {
            self.cache.clear();
            self.order.clear();
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_think_mode_timeout() {
        assert_eq!(ThinkMode::Fast.timeout_ms(), 100);
        assert_eq!(ThinkMode::Balanced.timeout_ms(), 500);
        assert_eq!(ThinkMode::Deep.timeout_ms(), 2000);
        assert_eq!(ThinkMode::Research.timeout_ms(), 5000);
        assert_eq!(ThinkMode::Creative.timeout_ms(), 10000);
    }

    #[test]
    fn test_think_mode_temperature() {
        assert_eq!(ThinkMode::Fast.default_temperature(), 0.0);
        assert_eq!(ThinkMode::Balanced.default_temperature(), 0.5);
        assert_eq!(ThinkMode::Creative.default_temperature(), 1.0);
    }

    #[tokio::test]
    async fn test_think_engine_mode() {
        let engine = ThinkEngine::new(ThinkMode::Balanced, true);
        
        assert_eq!(engine.get_mode().await, ThinkMode::Balanced);
        
        engine.set_mode(ThinkMode::Deep).await;
        assert_eq!(engine.get_mode().await, ThinkMode::Deep);
    }
}
