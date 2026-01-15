//! 工具系统模块
//! 
//! 管理 Agent 可调用的工具

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;

/// 工具调用结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolResult {
    pub success: bool,
    pub result: Option<String>,
    pub error: Option<String>,
    pub duration_ms: u64,
}

impl ToolResult {
    pub fn success(result: impl Into<String>) -> Self {
        Self {
            success: true,
            result: Some(result.into()),
            error: None,
            duration_ms: 0,
        }
    }
    
    pub fn failure(error: impl Into<String>) -> Self {
        Self {
            success: false,
            result: None,
            error: Some(error.into()),
            duration_ms: 0,
        }
    }
    
    pub fn with_duration(mut self, ms: u64) -> Self {
        self.duration_ms = ms;
        self
    }
}

/// 工具接口
pub trait Tool: Send + Sync {
    /// 工具名称
    fn name(&self) -> &str;
    
    /// 工具描述
    fn description(&self) -> &str;
    
    /// 执行工具
    fn execute(&self, args: HashMap<String, serde_json::Value>) -> ToolResult;
    
    /// 获取参数定义
    fn parameters(&self) -> Vec<ToolParameter> {
        Vec::new()
    }
}

/// 工具参数定义
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolParameter {
    pub name: String,
    pub param_type: String,
    pub description: String,
    pub required: bool,
    pub default: Option<String>,
}

/// 工具注册表
pub struct ToolRegistry {
    tools: Arc<RwLock<HashMap<String, Arc<dyn Tool>>>>,
}

impl ToolRegistry {
    pub fn new() -> Self {
        Self {
            tools: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    pub fn register<T: Tool + 'static>(&self, tool: T) {
        let name = tool.name().to_string();
        let mut tools = self.tools.write();
        tools.insert(name, Arc::new(tool));
    }
    
    pub fn unregister(&self, name: &str) -> bool {
        let mut tools = self.tools.write();
        tools.remove(name).is_some()
    }
    
    pub fn get(&self, name: &str) -> Option<Arc<dyn Tool>> {
        let tools = self.tools.read();
        tools.get(name).cloned()
    }
    
    pub fn list(&self) -> Vec<String> {
        let tools = self.tools.read();
        tools.keys().cloned().collect()
    }
    
    pub fn execute(
        &self,
        name: &str,
        args: HashMap<String, serde_json::Value>,
    ) -> ToolResult {
        match self.get(name) {
            Some(tool) => tool.execute(args),
            None => ToolResult::failure(format!("Tool not found: {}", name)),
        }
    }
}

impl Default for ToolRegistry {
    fn default() -> Self {
        Self::new()
    }
}

/// 内置工具实现

/// 计算器工具
pub struct CalculatorTool;

impl CalculatorTool {
    pub fn new() -> Self {
        Self
    }
}

impl Tool for CalculatorTool {
    fn name(&self) -> &str {
        "calculator"
    }
    
    fn description(&self) -> &str {
        "Perform mathematical calculations"
    }
    
    fn execute(&self, args: HashMap<String, serde_json::Value>) -> ToolResult {
        if let Some(expr) = args.get("expression") {
            if let Some(expr_str) = expr.as_str() {
                // 简单的表达式计算（实际应使用专门的表达式解析库）
                let result = calculate_simple(expr_str);
                return ToolResult::success(format!("{}", result));
            }
        }
        ToolResult::failure("Missing expression parameter")
    }
    
    fn parameters(&self) -> Vec<ToolParameter> {
        vec![
            ToolParameter {
                name: "expression".to_string(),
                param_type: "string".to_string(),
                description: "Mathematical expression to evaluate".to_string(),
                required: true,
                default: None,
            }
        ]
    }
}

fn calculate_simple(expr: &str) -> f64 {
    // 简化的计算（实际应使用专门的表达式解析库）
    let expr = expr.replace(" ", "");
    if let Ok(num) = expr.parse::<f64>() {
        return num;
    }
    
    // 简单加法
    if let Some(pos) = expr.find('+') {
        let left: f64 = expr[..pos].parse().unwrap_or(0.0);
        let right: f64 = expr[pos+1..].parse().unwrap_or(0.0);
        return left + right;
    }
    
    // 简单减法
    if let Some(pos) = expr.find('-') {
        let left: f64 = expr[..pos].parse().unwrap_or(0.0);
        let right: f64 = expr[pos+1..].parse().unwrap_or(0.0);
        return left - right;
    }
    
    // 简单乘法
    if let Some(pos) = expr.find('*') {
        let left: f64 = expr[..pos].parse().unwrap_or(0.0);
        let right: f64 = expr[pos+1..].parse().unwrap_or(0.0);
        return left * right;
    }
    
    // 简单除法
    if let Some(pos) = expr.find('/') {
        let left: f64 = expr[..pos].parse().unwrap_or(0.0);
        let right: f64 = expr[pos+1..].parse().unwrap_or(1.0);
        if right != 0.0 {
            return left / right;
        }
    }
    
    0.0
}

/// 日期时间工具
pub struct DateTimeTool;

impl DateTimeTool {
    pub fn new() -> Self {
        Self
    }
}

impl Tool for DateTimeTool {
    fn name(&self) -> &str {
        "datetime"
    }
    
    fn description(&self) -> &str {
        "Get current date and time"
    }
    
    fn execute(&self, args: HashMap<String, serde_json::Value>) -> ToolResult {
        let now = chrono::Utc::now();
        let format = args.get("format")
            .and_then(|v| v.as_str())
            .unwrap_or("%Y-%m-%d %H:%M:%S UTC");
        
        ToolResult::success(now.format(format).to_string())
    }
    
    fn parameters(&self) -> Vec<ToolParameter> {
        vec![
            ToolParameter {
                name: "format".to_string(),
                param_type: "string".to_string(),
                description: "DateTime format string".to_string(),
                required: false,
                default: Some("%Y-%m-%d %H:%M:%S UTC".to_string()),
            }
        ]
    }
}

/// 随机数工具
pub struct RandomTool;

impl RandomTool {
    pub fn new() -> Self {
        Self
    }
}

impl Tool for RandomTool {
    fn name(&self) -> &str {
        "random"
    }
    
    fn description(&self) -> &str {
        "Generate random numbers"
    }
    
    fn execute(&self, args: HashMap<String, serde_json::Value>) -> ToolResult {
        let min = args.get("min")
            .and_then(|v| v.as_i64())
            .unwrap_or(0) as i32;
        let max = args.get("max")
            .and_then(|v| v.as_i64())
            .unwrap_or(100) as i32;
        
        use std::time::{SystemTime, UNIX_EPOCH};
        let seed = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos() as u32;
        
        let range = (max - min + 1) as u32;
        let random = (seed.wrapping_mul(1103515245).wrapping_add(12345) % (1 << 31)) % range;
        let result = (min as u32 + random) as i32;
        
        ToolResult::success(format!("{}", result))
    }
    
    fn parameters(&self) -> Vec<ToolParameter> {
        vec![
            ToolParameter {
                name: "min".to_string(),
                param_type: "number".to_string(),
                description: "Minimum value".to_string(),
                required: false,
                default: Some("0".to_string()),
            },
            ToolParameter {
                name: "max".to_string(),
                param_type: "number".to_string(),
                description: "Maximum value".to_string(),
                required: false,
                default: Some("100".to_string()),
            }
        ]
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_calculator() {
        let tool = CalculatorTool::new();
        let mut args = HashMap::new();
        args.insert("expression".to_string(), serde_json::json!("10 + 20"));
        
        let result = tool.execute(args);
        assert!(result.success);
    }

    #[test]
    fn test_builtin_tools() {
        let registry = ToolRegistry::new();
        registry.register(CalculatorTool::new());
        registry.register(DateTimeTool::new());
        registry.register(RandomTool::new());
        
        assert_eq!(registry.list().len(), 3);
    }
}
