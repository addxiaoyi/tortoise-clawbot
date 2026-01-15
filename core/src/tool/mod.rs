//! Tool Module
//! 
//! Tool system for extending agent capabilities.

use anyhow::Result;
use async_trait::async_trait;

/// Tool trait for extending agent capabilities
#[async_trait]
pub trait Tool: Send + Sync {
    /// Get tool name
    fn name(&self) -> &str;
    
    /// Get tool description
    fn description(&self) -> &str;
    
    /// Get parameter schema (JSON Schema)
    fn parameters(&self) -> serde_json::Value;
    
    /// Execute the tool
    async fn execute(&self, arguments: serde_json::Value) -> Result<serde_json::Value>;
}
