//! 工具调用系统
//!
//! 管理工具注册和执行

use anyhow::{anyhow, Result};
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;
use tokio::sync::RwLock;

use super::engine::ToolCall;

/// 工具元数据
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolMetadata {
    pub name: String,
    pub description: String,
    pub parameters: serde_json::Value,
    pub category: Option<String>,
    pub tags: Vec<String>,
    pub version: String,
}

/// 工具接口
#[async_trait]
pub trait Tool: Send + Sync {
    /// 获取工具元数据
    fn metadata(&self) -> ToolMetadata;
    
    /// 执行工具
    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value>;
    
    /// 验证参数
    fn validate_arguments(&self, arguments: &serde_json::Value) -> Result<()> {
        // 默认不验证
        let _ = arguments;
        Ok(())
    }
}

/// 内置工具注册表
pub struct ToolRegistry {
    tools: RwLock<HashMap<String, Arc<dyn Tool>>>,
}

impl ToolRegistry {
    /// 创建新的工具注册表
    pub fn new() -> Self {
        Self {
            tools: RwLock::new(HashMap::new()),
        }
    }

    /// 注册工具
    pub fn register(&self, tool: Arc<dyn Tool>) -> Result<()> {
        let name = tool.metadata().name.clone();
        
        let mut tools = self.tools.write().await;
        if tools.contains_key(&name) {
            return Err(anyhow!("Tool already registered: {}", name));
        }
        
        tools.insert(name, tool);
        tracing::info!("Registered tool: {}", name);
        Ok(())
    }

    /// 注册工具 (通过 Box)
    pub fn register_box(&self, tool: Box<dyn Tool>) -> Result<()> {
        let name = tool.metadata().name.clone();
        
        let mut tools = self.tools.write().await;
        if tools.contains_key(&name) {
            return Err(anyhow!("Tool already registered: {}", name));
        }
        
        tools.insert(name, Arc::from(tool));
        tracing::info!("Registered tool: {}", name);
        Ok(())
    }

    /// 取消注册工具
    pub fn unregister(&self, name: &str) -> bool {
        let mut tools = self.tools.write().await;
        tools.remove(name).is_some()
    }

    /// 获取工具
    pub async fn get(&self, name: &str) -> Option<Arc<dyn Tool>> {
        let tools = self.tools.read().await;
        tools.get(name).cloned()
    }

    /// 列出所有工具
    pub async fn list(&self) -> Vec<ToolMetadata> {
        let tools = self.tools.read().await;
        tools.values()
            .map(|t| t.metadata())
            .collect()
    }

    /// 检查工具是否存在
    pub async fn contains(&self, name: &str) -> bool {
        let tools = self.tools.read().await;
        tools.contains_key(name)
    }

    /// 获取工具数量
    pub async fn count(&self) -> usize {
        let tools = self.tools.read().await;
        tools.len()
    }
}

impl Default for ToolRegistry {
    fn default() -> Self {
        Self::new()
    }
}

/// 工具调用器
pub struct ToolCallExecutor {
    registry: Arc<ToolRegistry>,
    execution_stats: RwLock<HashMap<String, ToolStats>>,
}

impl ToolCallExecutor {
    /// 创建新的工具调用器
    pub fn new(registry: Arc<ToolRegistry>) -> Self {
        Self {
            registry,
            execution_stats: RwLock::new(HashMap::new()),
        }
    }

    /// 执行工具调用
    pub async fn execute(&self, call: &ToolCall) -> Result<String> {
        let start = Instant::now();
        
        let tool = self.registry.get(&call.name).await
            .ok_or_else(|| anyhow!("Tool not found: {}", call.name))?;
        
        tracing::info!("Executing tool: {} with args: {:?}", call.name, call.arguments);
        
        // 验证参数
        tool.validate_arguments(&call.arguments)?;
        
        // 执行工具
        let result = tool.execute(call.arguments.clone()).await?;
        
        // 记录统计
        let elapsed = start.elapsed();
        self.record_execution(&call.name, elapsed, true);
        
        tracing::info!("Tool {} completed in {:?}", call.name, elapsed);
        
        Ok(serde_json::to_string(&result)?)
    }

    /// 批量执行工具调用
    pub async fn execute_batch(&self, calls: &[ToolCall]) -> Vec<Result<String>> {
        let mut results = Vec::with_capacity(calls.len());
        
        for call in calls {
            results.push(self.execute(call).await);
        }
        
        results
    }

    /// 记录执行统计
    fn record_execution(&self, tool_name: &str, duration: std::time::Duration, success: bool) {
        let mut stats = self.execution_stats.blocking_write();
        let entry = stats.entry(tool_name.to_string()).or_insert_with(|| ToolStats {
            total_calls: 0,
            successful_calls: 0,
            failed_calls: 0,
            total_time_ms: 0,
        });
        
        entry.total_calls += 1;
        if success {
            entry.successful_calls += 1;
        } else {
            entry.failed_calls += 1;
        }
        entry.total_time_ms += duration.as_millis() as u64;
    }

    /// 获取执行统计
    pub async fn get_stats(&self) -> HashMap<String, ToolStats> {
        self.execution_stats.read().await.clone()
    }
}

/// 工具执行统计
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolStats {
    pub total_calls: u64,
    pub successful_calls: u64,
    pub failed_calls: u64,
    pub total_time_ms: u64,
}

impl ToolStats {
    pub fn success_rate(&self) -> f64 {
        if self.total_calls == 0 {
            return 0.0;
        }
        self.successful_calls as f64 / self.total_calls as f64
    }
    
    pub fn average_time_ms(&self) -> f64 {
        if self.total_calls == 0 {
            return 0.0;
        }
        self.total_time_ms as f64 / self.total_calls as f64
    }
}

// === 内置工具 ===

/// Web 搜索工具
pub struct WebSearchTool {
    description: String,
}

impl WebSearchTool {
    pub fn new() -> Self {
        Self {
            description: "Search the web for information".to_string(),
        }
    }
}

#[async_trait]
impl Tool for WebSearchTool {
    fn metadata(&self) -> ToolMetadata {
        ToolMetadata {
            name: "web_search".to_string(),
            description: self.description.clone(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The search query"
                    },
                    "limit": {
                        "type": "integer",
                        "description": "Maximum number of results",
                        "default": 5
                    }
                },
                "required": ["query"]
            }),
            category: Some("search".to_string()),
            tags: vec!["search".to_string(), "web".to_string()],
            version: "1.0.0".to_string(),
        }
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value> {
        let query = arguments["query"].as_str()
            .ok_or_else(|| anyhow!("Missing query parameter"))?;
        
        // 模拟搜索
        Ok(serde_json::json!({
            "query": query,
            "results": [
                {"title": "Result 1", "url": "https://example.com/1", "snippet": "Sample result 1"},
                {"title": "Result 2", "url": "https://example.com/2", "snippet": "Sample result 2"}
            ]
        }))
    }
}

/// 计算器工具
pub struct CalculatorTool {
    description: String,
}

impl CalculatorTool {
    pub fn new() -> Self {
        Self {
            description: "Perform mathematical calculations".to_string(),
        }
    }
}

#[async_trait]
impl Tool for CalculatorTool {
    fn metadata(&self) -> ToolMetadata {
        ToolMetadata {
            name: "calculator".to_string(),
            description: self.description.clone(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "The mathematical expression to evaluate"
                    }
                },
                "required": ["expression"]
            }),
            category: Some("math".to_string()),
            tags: vec!["math".to_string(), "calculator".to_string()],
            version: "1.0.0".to_string(),
        }
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value> {
        let expression = arguments["expression"].as_str()
            .ok_or_else(|| anyhow!("Missing expression parameter"))?;
        
        // 简单的表达式计算
        let result = meval::eval_str(expression)
            .map_err(|e| anyhow!("Calculation error: {}", e))?;
        
        Ok(serde_json::json!({
            "expression": expression,
            "result": result
        }))
    }
}

/// 文件操作工具
pub struct FileTool {
    description: String,
}

impl FileTool {
    pub fn new() -> Self {
        Self {
            description: "Read or write files".to_string(),
        }
    }
}

#[async_trait]
impl Tool for FileTool {
    fn metadata(&self) -> ToolMetadata {
        ToolMetadata {
            name: "file".to_string(),
            description: self.description.clone(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "action": {
                        "type": "string",
                        "enum": ["read", "write"],
                        "description": "The file action to perform"
                    },
                    "path": {
                        "type": "string",
                        "description": "The file path"
                    },
                    "content": {
                        "type": "string",
                        "description": "The content to write (for write action)"
                    }
                },
                "required": ["action", "path"]
            }),
            category: Some("filesystem".to_string()),
            tags: vec!["file".to_string(), "io".to_string()],
            version: "1.0.0".to_string(),
        }
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value> {
        let action = arguments["action"].as_str()
            .ok_or_else(|| anyhow!("Missing action parameter"))?;
        let path = arguments["path"].as_str()
            .ok_or_else(|| anyhow!("Missing path parameter"))?;
        
        match action {
            "read" => {
                let content = tokio::fs::read_to_string(path).await
                    .map_err(|e| anyhow!("Failed to read file: {}", e))?;
                Ok(serde_json::json!({
                    "path": path,
                    "content": content
                }))
            }
            "write" => {
                let content = arguments["content"].as_str()
                    .ok_or_else(|| anyhow!("Missing content parameter"))?;
                tokio::fs::write(path, content).await
                    .map_err(|e| anyhow!("Failed to write file: {}", e))?;
                Ok(serde_json::json!({
                    "path": path,
                    "success": true
                }))
            }
            _ => Err(anyhow!("Unknown action: {}", action)),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_tool_registry() {
        let registry = ToolRegistry::new();
        
        let tool = Arc::new(WebSearchTool::new());
        registry.register(tool).unwrap();
        
        assert!(registry.contains("web_search").await);
        assert!(!registry.contains("nonexistent").await);
        
        let tools = registry.list().await;
        assert_eq!(tools.len(), 1);
    }

    #[tokio::test]
    async fn test_tool_call_executor() {
        let registry = ToolRegistry::new();
        let tool = Arc::new(WebSearchTool::new());
        registry.register(tool).unwrap();
        
        let executor = ToolCallExecutor::new(Arc::clone(&registry));
        
        let call = ToolCall {
            id: "test-1".to_string(),
            name: "web_search".to_string(),
            arguments: serde_json::json!({
                "query": "test query",
                "limit": 5
            }),
        };
        
        let result = executor.execute(&call).await;
        assert!(result.is_ok());
    }
}
