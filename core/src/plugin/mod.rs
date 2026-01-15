//! Plugin system module

use std::sync::Arc;
use tokio::sync::RwLock;
use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use serde_json::Value;

/// Plugin metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginInfo {
    pub id: String,
    pub name: String,
    pub version: String,
    pub description: String,
    pub author: Option<String>,
    pub enabled: bool,
    pub dependencies: Vec<String>,
}

/// Plugin interface
#[async_trait::async_trait]
pub trait Plugin: Send + Sync {
    /// Get plugin info
    fn info(&self) -> PluginInfo;
    
    /// Initialize the plugin
    async fn init(&self, config: Value) -> Result<(), PluginError>;
    
    /// Cleanup resources
    async fn cleanup(&self) -> Result<(), PluginError>;
    
    /// Handle a message
    async fn handle(&self, message: PluginMessage) -> Result<PluginResponse, PluginError>;
}

/// Plugin message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginMessage {
    pub plugin_id: String,
    pub message_type: String,
    pub payload: Value,
    pub context: PluginContext,
}

/// Plugin context
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginContext {
    pub session_id: Option<String>,
    pub user_id: Option<String>,
    pub channel: Option<String>,
    pub metadata: Value,
}

/// Plugin response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginResponse {
    pub success: bool,
    pub data: Value,
    pub error: Option<String>,
}

/// Plugin error
#[derive(Debug, thiserror::Error)]
pub enum PluginError {
    #[error("Plugin not found: {0}")]
    NotFound(String),
    
    #[error("Plugin error: {0}")]
    Execution(String),
    
    #[error("Plugin initialization failed: {0}")]
    InitFailed(String),
    
    #[error("Plugin disabled")]
    Disabled,
}

/// Plugin host - manages plugin lifecycle
pub struct PluginHost {
    plugins: Arc<DashMap<String, Arc<Box<dyn Plugin>>>>,
    enabled: Arc<DashMap<String, bool>>,
}

impl PluginHost {
    pub fn new() -> Self {
        Self {
            plugins: Arc::new(DashMap::new()),
            enabled: Arc::new(DashMap::new()),
        }
    }

    /// Register a plugin
    pub fn register<P: Plugin + 'static>(&self, plugin: P) -> Result<(), PluginError> {
        let info = plugin.info();
        let plugin_id = info.id.clone();
        
        // Initialize plugin
        plugin.init(Value::Null).await?;
        
        self.plugins.insert(plugin_id.clone(), Arc::new(Box::new(plugin)));
        self.enabled.insert(plugin_id, true);
        
        Ok(())
    }

    /// Unregister a plugin
    pub async fn unregister(&self, plugin_id: &str) -> Result<(), PluginError> {
        if let Some(plugin) = self.plugins.get(plugin_id) {
            plugin.cleanup().await?;
        }
        
        self.plugins.remove(plugin_id);
        self.enabled.remove(plugin_id);
        
        Ok(())
    }

    /// Get a plugin by ID
    pub fn get(&self, plugin_id: &str) -> Option<Arc<Box<dyn Plugin>>> {
        self.plugins.get(plugin_id).map(|p| p.clone())
    }

    /// List all plugins
    pub fn list(&self) -> Vec<PluginInfo> {
        self.plugins
            .iter()
            .map(|p| p.info())
            .collect()
    }

    /// List enabled plugins
    pub fn list_enabled(&self) -> Vec<PluginInfo> {
        self.plugins
            .iter()
            .filter(|p| self.enabled.get(p.info().id.as_str()).map(|e| *e).unwrap_or(false))
            .map(|p| p.info())
            .collect()
    }

    /// Enable a plugin
    pub fn enable(&self, plugin_id: &str) -> bool {
        if self.plugins.contains_key(plugin_id) {
            self.enabled.insert(plugin_id.to_string(), true);
            true
        } else {
            false
        }
    }

    /// Disable a plugin
    pub fn disable(&self, plugin_id: &str) -> bool {
        if self.plugins.contains_key(plugin_id) {
            self.enabled.insert(plugin_id.to_string(), false);
            true
        } else {
            false
        }
    }

    /// Check if plugin is enabled
    pub fn is_enabled(&self, plugin_id: &str) -> bool {
        self.enabled.get(plugin_id).map(|e| *e).unwrap_or(false)
    }

    /// Handle a message through a plugin
    pub async fn handle(&self, message: PluginMessage) -> Result<PluginResponse, PluginError> {
        let plugin_id = &message.plugin_id;
        
        // Check if plugin exists
        let plugin = self.plugins
            .get(plugin_id)
            .ok_or_else(|| PluginError::NotFound(plugin_id.clone()))?;
        
        // Check if plugin is enabled
        if !self.is_enabled(plugin_id) {
            return Err(PluginError::Disabled);
        }
        
        // Handle message
        plugin.handle(message).await
    }

    /// Handle a message through all enabled plugins
    pub async fn broadcast(&self, message: PluginMessage) -> Vec<(String, Result<PluginResponse, PluginError>)> {
        let mut results = Vec::new();
        
        for plugin in self.plugins.iter() {
            if self.is_enabled(&plugin.info().id) {
                let plugin_id = plugin.info().id.clone();
                let mut msg = message.clone();
                msg.plugin_id = plugin_id.clone();
                
                let result = plugin.handle(msg).await;
                results.push((plugin_id, result));
            }
        }
        
        results
    }

    /// Get plugin count
    pub fn len(&self) -> usize {
        self.plugins.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.plugins.is_empty()
    }
}

impl Default for PluginHost {
    fn default() -> Self {
        Self::new()
    }
}
