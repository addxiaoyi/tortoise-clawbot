//! Tools module - Plugin and tool management

use std::sync::Arc;
use tokio::sync::RwLock;
use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use uuid::Uuid;

/// Tool definition
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Tool {
    pub name: String,
    pub description: String,
    pub parameters: Value,
    pub enabled: bool,
}

impl Tool {
    pub fn new(name: String, description: String) -> Self {
        Self {
            name,
            description,
            parameters: Value::Object(serde_json::Map::new()),
            enabled: true,
        }
    }

    pub fn with_parameters(mut self, params: Value) -> Self {
        self.parameters = params;
        self
    }
}

/// Tool call request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    pub arguments: Value,
    pub timeout_ms: Option<u64>,
}

impl ToolCall {
    pub fn new(name: String, arguments: Value) -> Self {
        Self {
            id: format!("call_{}", Uuid::new_v4()),
            name,
            arguments,
            timeout_ms: None,
        }
    }
}

/// Tool call result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    pub call_id: String,
    pub success: bool,
    pub result: Value,
    pub error: Option<String>,
    pub execution_time_ms: u64,
}

impl ToolResult {
    pub fn success(call_id: String, result: Value, execution_time_ms: u64) -> Self {
        Self {
            call_id,
            success: true,
            result,
            error: None,
            execution_time_ms,
        }
    }

    pub fn failure(call_id: String, error: String, execution_time_ms: u64) -> Self {
        Self {
            call_id,
            success: false,
            result: Value::Null,
            error: Some(error),
            execution_time_ms,
        }
    }
}

/// Tool executor trait
#[async_trait::async_trait]
pub trait ToolExecutor: Send + Sync {
    async fn execute(&self, call: ToolCall) -> ToolResult;
}

/// Tool registry
pub struct ToolRegistry {
    tools: Arc<DashMap<String, Tool>>,
    executors: Arc<DashMap<String, Box<dyn ToolExecutor>>>,
}

impl ToolRegistry {
    pub fn new() -> Self {
        let registry = Self {
            tools: Arc::new(DashMap::new()),
            executors: Arc::new(DashMap::new()),
        };
        registry.register_builtin_tools();
        registry
    }

    /// Register a tool
    pub fn register(&self, tool: Tool, executor: Box<dyn ToolExecutor>) {
        self.tools.insert(tool.name.clone(), tool);
        self.executors.insert(tool.name.clone(), executor);
    }

    /// Register multiple tools
    pub fn register_all(&self, tools: Vec<(Tool, Box<dyn ToolExecutor>)>) {
        for (tool, executor) in tools {
            self.register(tool, executor);
        }
    }

    /// Get a tool by name
    pub fn get(&self, name: &str) -> Option<Tool> {
        self.tools.get(name).map(|t| t.clone())
    }

    /// List all tools
    pub fn list(&self) -> Vec<Tool> {
        self.tools.iter().map(|t| t.clone()).collect()
    }

    /// List enabled tools
    pub fn list_enabled(&self) -> Vec<Tool> {
        self.tools
            .iter()
            .filter(|t| t.enabled)
            .map(|t| t.clone())
            .collect()
    }

    /// Execute a tool call
    pub async fn execute(&self, call: ToolCall) -> ToolResult {
        let start = std::time::Instant::now();

        if let Some(executor) = self.executors.get(&call.name) {
            executor.execute(call).await
        } else {
            ToolResult::failure(
                call.id,
                format!("Tool '{}' not found", call.name),
                start.elapsed().as_millis() as u64,
            )
        }
    }

    /// Enable a tool
    pub fn enable(&self, name: &str) -> bool {
        if let Some(mut tool) = self.tools.get_mut(name) {
            tool.enabled = true;
            return true;
        }
        false
    }

    /// Disable a tool
    pub fn disable(&self, name: &str) -> bool {
        if let Some(mut tool) = self.tools.get_mut(name) {
            tool.enabled = false;
            return true;
        }
        false
    }

    /// Remove a tool
    pub fn unregister(&self, name: &str) -> bool {
        self.tools.remove(name);
        self.executors.remove(name);
        true
    }

    /// Get tool count
    pub fn len(&self) -> usize {
        self.tools.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.tools.is_empty()
    }

    /// Register built-in tools
    fn register_builtin_tools(&self) {
        // Web Search tool
        let web_search = Tool {
            name: "web_search".to_string(),
            description: "Search the web for information".to_string(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The search query"
                    },
                    "num_results": {
                        "type": "integer",
                        "description": "Number of results to return",
                        "default": 5
                    }
                },
                "required": ["query"]
            }),
            enabled: true,
        };

        let web_search_executor = WebSearchExecutor {};
        self.register(web_search, Box::new(web_search_executor));

        // Calculator tool
        let calculator = Tool {
            name: "calculator".to_string(),
            description: "Perform mathematical calculations".to_string(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "Mathematical expression to evaluate"
                    }
                },
                "required": ["expression"]
            }),
            enabled: true,
        };

        let calculator_executor = CalculatorExecutor {};
        self.register(calculator, Box::new(calculator_executor));

        // File Read tool
        let file_read = Tool {
            name: "file_read".to_string(),
            description: "Read file contents".to_string(),
            parameters: serde_json::json!({
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to the file"
                    }
                },
                "required": ["path"]
            }),
            enabled: true,
        };

        let file_read_executor = FileReadExecutor {};
        self.register(file_read, Box::new(file_read_executor));
    }
}

impl Default for ToolRegistry {
    fn default() -> Self {
        Self::new()
    }
}

/// Web search executor
struct WebSearchExecutor;

#[async_trait::async_trait]
impl ToolExecutor for WebSearchExecutor {
    async fn execute(&self, call: ToolCall) -> ToolResult {
        let start = std::time::Instant::now();

        let query = match call.arguments.get("query") {
            Some(Value::String(s)) => s.clone(),
            _ => {
                return ToolResult::failure(
                    call.id,
                    "Missing or invalid 'query' parameter".to_string(),
                    start.elapsed().as_millis() as u64,
                );
            }
        };

        let num_results = call.arguments
            .get("num_results")
            .and_then(|v| v.as_u64())
            .unwrap_or(5) as usize;

        // Placeholder - actual implementation would call a search API
        let results = serde_json::json!({
            "query": query,
            "results": (0..num_results).map(|i| {
                serde_json::json!({
                    "title": format!("Result {} for '{}'", i + 1, query),
                    "url": format!("https://example.com/result/{}", i + 1),
                    "snippet": format!("This is a placeholder result for the query '{}'.", query)
                })
            }).collect::<Vec<_>>()
        });

        ToolResult::success(call.id, results, start.elapsed().as_millis() as u64)
    }
}

/// Calculator executor
struct CalculatorExecutor;

#[async_trait::async_trait]
impl ToolExecutor for CalculatorExecutor {
    async fn execute(&self, call: ToolCall) -> ToolResult {
        let start = std::time::Instant::now();

        let expression = match call.arguments.get("expression") {
            Some(Value::String(s)) => s.clone(),
            _ => {
                return ToolResult::failure(
                    call.id,
                    "Missing or invalid 'expression' parameter".to_string(),
                    start.elapsed().as_millis() as u64,
                );
            }
        };

        // Simple expression evaluation (placeholder)
        // In production, use a safe expression evaluator
        let result = expression.parse::<f64>().unwrap_or(0.0);

        ToolResult::success(
            call.id,
            serde_json::json!({
                "expression": expression,
                "result": result
            }),
            start.elapsed().as_millis() as u64,
        )
    }
}

/// File read executor
struct FileReadExecutor;

#[async_trait::async_trait]
impl ToolExecutor for FileReadExecutor {
    async fn execute(&self, call: ToolCall) -> ToolResult {
        let start = std::time::Instant::now();

        let path = match call.arguments.get("path") {
            Some(Value::String(s)) => s.clone(),
            _ => {
                return ToolResult::failure(
                    call.id,
                    "Missing or invalid 'path' parameter".to_string(),
                    start.elapsed().as_millis() as u64,
                );
            }
        };

        // Placeholder - actual implementation would read from filesystem
        // with proper sandboxing
        ToolResult::success(
            call.id,
            serde_json::json!({
                "path": path,
                "content": "[File content placeholder]",
                "size": 0
            }),
            start.elapsed().as_millis() as u64,
        )
    }
}
