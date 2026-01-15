//! Tool Module
//! 
//! Tool system for agent function calling.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Tool registry
pub struct ToolRegistry {
    tools: RwLock<HashMap<String, Arc<dyn crate::agent::Tool>>>,
}

impl ToolRegistry {
    /// Create a new tool registry
    pub fn new() -> Self {
        Self {
            tools: RwLock::new(HashMap::new()),
        }
    }

    /// Register a tool
    pub async fn register(&self, tool: Arc<dyn crate::agent::Tool>) {
        let mut tools = self.tools.write().await;
        tools.insert(tool.name().to_string(), tool);
    }

    /// Unregister a tool
    pub async fn unregister(&self, name: &str) {
        let mut tools = self.tools.write().await;
        tools.remove(name);
    }

    /// Get a tool by name
    pub async fn get(&self, name: &str) -> Option<Arc<dyn crate::agent::Tool>> {
        let tools = self.tools.read().await;
        tools.get(name).cloned()
    }

    /// List all tools
    pub async fn list(&self) -> Vec<ToolInfo> {
        let tools = self.tools.read().await;
        tools.values()
            .map(|t| ToolInfo {
                name: t.name().to_string(),
                description: t.description().to_string(),
                parameters: t.parameters(),
            })
            .collect()
    }

    /// Execute a tool
    pub async fn execute(&self, name: &str, arguments: serde_json::Value) -> Result<serde_json::Value> {
        let tool = self.get(name).await
            .ok_or_else(|| anyhow::anyhow!("Tool not found: {}", name))?;
        
        tool.execute(arguments).await
    }
}

impl Default for ToolRegistry {
    fn default() -> Self {
        Self::new()
    }
}

/// Tool information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolInfo {
    pub name: String,
    pub description: String,
    pub parameters: serde_json::Value,
}

/// Built-in tools

/// Calculator tool
pub struct CalculatorTool;

impl CalculatorTool {
    pub fn new() -> Self {
        Self
    }
}

impl crate::agent::Tool for CalculatorTool {
    fn name(&self) -> &str {
        "calculator"
    }

    fn description(&self) -> &str {
        "Perform mathematical calculations. Input should be a JSON object with 'expression' field containing a mathematical expression."
    }

    fn parameters(&self) -> serde_json::Value {
        serde_json::json!({
            "type": "object",
            "properties": {
                "expression": {
                    "type": "string",
                    "description": "The mathematical expression to evaluate (e.g., '2 + 2', 'sin(pi/2)', 'sqrt(16)')"
                }
            },
            "required": ["expression"]
        })
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value, String> {
        let expression = arguments.get("expression")
            .and_then(|v| v.as_str())
            .ok_or_else(|| "Missing 'expression' field".to_string())?;

        // Simplified evaluation - in production use proper math parser
        let result = evaluate_expression(expression)
            .map_err(|e| e.to_string())?;

        Ok(serde_json::json!({
            "expression": expression,
            "result": result
        }))
    }
}

fn evaluate_expression(expr: &str) -> Result<f64> {
    // Very simplified - just try to parse as number or handle basic operations
    // In production, use a proper expression parser like meval or rson
    if let Ok(num) = expr.trim().parse::<f64>() {
        return Ok(num);
    }

    // Basic operations
    if let Some(pos) = expr.find('+') {
        let a = evaluate_expression(expr[..pos].trim())?;
        let b = evaluate_expression(expr[pos + 1..].trim())?;
        return Ok(a + b);
    }

    if let Some(pos) = expr.rfind('-') {
        let a = evaluate_expression(expr[..pos].trim())?;
        let b = evaluate_expression(expr[pos + 1..].trim())?;
        return Ok(a - b);
    }

    if let Some(pos) = expr.find('*') {
        let a = evaluate_expression(expr[..pos].trim())?;
        let b = evaluate_expression(expr[pos + 1..].trim())?;
        return Ok(a * b);
    }

    if let Some(pos) = expr.find('/') {
        let a = evaluate_expression(expr[..pos].trim())?;
        let b = evaluate_expression(expr[pos + 1..].trim())?;
        if b == 0.0 {
            return Err(anyhow::anyhow!("Division by zero"));
        }
        return Ok(a / b);
    }

    Err(anyhow::anyhow!("Could not evaluate expression: {}", expr))
}

/// Current time tool
pub struct CurrentTimeTool;

impl CurrentTimeTool {
    pub fn new() -> Self {
        Self
    }
}

impl crate::agent::Tool for CurrentTimeTool {
    fn name(&self) -> &str {
        "current_time"
    }

    fn description(&self) -> &str {
        "Get the current date and time. Returns the current UTC time in ISO 8601 format."
    }

    fn parameters(&self) -> serde_json::Value {
        serde_json::json!({
            "type": "object",
            "properties": {},
            "required": []
        })
    }

    async fn execute(&self, _arguments: serde_json::Value) -> Result<serde_json::Value, String> {
        let now = chrono::Utc::now();
        Ok(serde_json::json!({
            "iso": now.to_rfc3339(),
            "unix": now.timestamp(),
            "timezone": "UTC"
        }))
    }
}

/// Web search tool
pub struct WebSearchTool {
    api_key: Option<String>,
}

impl WebSearchTool {
    pub fn new(api_key: Option<String>) -> Self {
        Self { api_key }
    }
}

impl crate::agent::Tool for WebSearchTool {
    fn name(&self) -> &str {
        "web_search"
    }

    fn description(&self) -> &str {
        "Search the web for information. Returns a list of search results with titles, URLs, and snippets."
    }

    fn parameters(&self) -> serde_json::Value {
        serde_json::json!({
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "The search query"
                },
                "limit": {
                    "type": "integer",
                    "description": "Maximum number of results to return",
                    "default": 5
                }
            },
            "required": ["query"]
        })
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value, String> {
        let query = arguments.get("query")
            .and_then(|v| v.as_str())
            .ok_or_else(|| "Missing 'query' field".to_string())?;

        let limit = arguments.get("limit")
            .and_then(|v| v.as_u64())
            .unwrap_or(5) as usize;

        // Placeholder - in production, call actual search API
        let results = vec![
            serde_json::json!({
                "title": format!("Result for: {}", query),
                "url": "https://example.com/result",
                "snippet": "This is a placeholder search result"
            })
        ];

        Ok(serde_json::json!({
            "query": query,
            "results": results.into_iter().take(limit).collect::<Vec<_>>()
        }))
    }
}

/// Weather tool
pub struct WeatherTool {
    api_key: Option<String>,
}

impl WeatherTool {
    pub fn new(api_key: Option<String>) -> Self {
        Self { api_key }
    }
}

impl crate::agent::Tool for WeatherTool {
    fn name(&self) -> &str {
        "weather"
    }

    fn description(&self) -> &str {
        "Get current weather information for a location. Returns temperature, conditions, and forecast."
    }

    fn parameters(&self) -> serde_json::Value {
        serde_json::json!({
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "City name or coordinates (lat,lon)"
                },
                "units": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "default": "celsius"
                }
            },
            "required": ["location"]
        })
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value, String> {
        let location = arguments.get("location")
            .and_then(|v| v.as_str())
            .ok_or_else(|| "Missing 'location' field".to_string())?;

        // Placeholder - in production, call actual weather API
        Ok(serde_json::json!({
            "location": location,
            "temperature": 22,
            "conditions": "Partly cloudy",
            "humidity": 65,
            "wind_speed": 10
        }))
    }
}

/// Code execution tool
pub struct CodeExecutionTool;

impl CodeExecutionTool {
    pub fn new() -> Self {
        Self
    }
}

impl crate::agent::Tool for CodeExecutionTool {
    fn name(&self) -> &str {
        "execute_code"
    }

    fn description(&self) -> &str {
        "Execute code in a sandboxed environment. Supports Python, JavaScript, and Bash."
    }

    fn parameters(&self) -> serde_json::Value {
        serde_json::json!({
            "type": "object",
            "properties": {
                "language": {
                    "type": "string",
                    "enum": ["python", "javascript", "bash"],
                    "description": "Programming language"
                },
                "code": {
                    "type": "string",
                    "description": "Code to execute"
                },
                "timeout": {
                    "type": "integer",
                    "description": "Timeout in seconds",
                    "default": 30
                }
            },
            "required": ["language", "code"]
        })
    }

    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value, String> {
        let language = arguments.get("language")
            .and_then(|v| v.as_str())
            .ok_or_else(|| "Missing 'language' field".to_string())?;

        let code = arguments.get("code")
            .and_then(|v| v.as_str())
            .ok_or_else(|| "Missing 'code' field".to_string())?;

        // Placeholder - in production, execute in sandbox
        Ok(serde_json::json!({
            "language": language,
            "stdout": format!("Executed {} code (placeholder)", language),
            "stderr": "",
            "exit_code": 0
        }))
    }
}
