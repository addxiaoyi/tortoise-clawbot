//! 插件系统模块
//! 
//! 实现安全的热插拔插件系统

use crate::error::{Error, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use async_trait::async_trait;

mod error;
pub use error::Error;

/// 插件状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PluginState {
    Installed,
    Loaded,
    Running,
    Disabled,
    Error,
}

/// 工具定义
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolDefinition {
    pub name: String,
    pub description: String,
    pub parameters: Vec<ToolParameter>,
    pub require_confirmation: bool,
    pub category: String,
}

impl ToolDefinition {
    pub fn new(name: impl Into<String>, description: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            description: description.into(),
            parameters: Vec::new(),
            require_confirmation: false,
            category: "general".to_string(),
        }
    }
}

/// 工具参数
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolParameter {
    pub name: String,
    pub param_type: String,
    pub description: String,
    pub required: bool,
    pub default: Option<String>,
}

/// 插件元信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginInfo {
    pub id: String,
    pub name: String,
    pub version: String,
    pub description: String,
    pub author: String,
    pub license: String,
    pub homepage: Option<String>,
    pub repository: Option<String>,
}

impl PluginInfo {
    pub fn new(name: impl Into<String>, version: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.into(),
            version: version.into(),
            description: String::new(),
            author: String::new(),
            license: "MIT".to_string(),
            homepage: None,
            repository: None,
        }
    }
}

/// 插件接口
#[async_trait]
pub trait Plugin: Send + Sync {
    /// 获取插件信息
    fn info(&self) -> &PluginInfo;
    
    /// 初始化插件
    async fn initialize(&mut self, config: HashMap<String, String>) -> Result<()>;
    
    /// 获取提供的工具列表
    fn tools(&self) -> Vec<ToolDefinition>;
    
    /// 执行工具
    async fn execute(
        &self,
        tool_name: &str,
        arguments: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value>;
    
    /// 关闭插件
    async fn shutdown(&mut self) -> Result<()>;
}

/// 插件实例
pub struct PluginInstance {
    pub info: PluginInfo,
    pub state: PluginState,
    pub tools: Vec<ToolDefinition>,
    pub config: HashMap<String, String>,
    pub installed_at: DateTime<Utc>,
    pub loaded_at: Option<DateTime<Utc>>,
}

impl PluginInstance {
    pub fn new(info: PluginInfo) -> Self {
        Self {
            info,
            state: PluginState::Installed,
            tools: Vec::new(),
            config: HashMap::new(),
            installed_at: Utc::now(),
            loaded_at: None,
        }
    }
    
    pub fn load(&mut self, tools: Vec<ToolDefinition>) {
        self.tools = tools;
        self.state = PluginState::Loaded;
        self.loaded_at = Some(Utc::now());
    }
    
    pub fn enable(&mut self) {
        self.state = PluginState::Running;
    }
    
    pub fn disable(&mut self) {
        self.state = PluginState::Disabled;
    }
}

/// 插件管理器
pub struct PluginManager {
    plugins: Arc<RwLock<HashMap<String, PluginInstance>>>,
    tool_registry: Arc<RwLock<HashMap<String, ToolRegistryEntry>>>,
}

struct ToolRegistryEntry {
    plugin_id: String,
    tool: ToolDefinition,
}

impl PluginManager {
    /// 创建新的插件管理器
    pub fn new() -> Self {
        Self {
            plugins: Arc::new(RwLock::new(HashMap::new())),
            tool_registry: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    /// 注册插件
    pub fn register(&self, info: PluginInfo) -> String {
        let id = info.id.clone();
        let instance = PluginInstance::new(info);
        
        let mut plugins = self.plugins.write();
        plugins.insert(id.clone(), instance);
        
        id
    }
    
    /// 加载插件
    pub fn load(&self, plugin_id: &str, tools: Vec<ToolDefinition>) -> Result<()> {
        let mut plugins = self.plugins.write();
        
        if let Some(plugin) = plugins.get_mut(plugin_id) {
            plugin.load(tools.clone());
            
            // 注册工具
            drop(plugins);
            let mut registry = self.tool_registry.write();
            for tool in tools {
                let key = format!("{}.{}", plugin_id, tool.name);
                registry.insert(key, ToolRegistryEntry {
                    plugin_id: plugin_id.to_string(),
                    tool,
                });
            }
            
            Ok(())
        } else {
            Err(Error::PluginNotFound(plugin_id.to_string()))
        }
    }
    
    /// 启用插件
    pub fn enable(&self, plugin_id: &str) -> Result<()> {
        let mut plugins = self.plugins.write();
        
        if let Some(plugin) = plugins.get_mut(plugin_id) {
            plugin.enable();
            Ok(())
        } else {
            Err(Error::PluginNotFound(plugin_id.to_string()))
        }
    }
    
    /// 禁用插件
    pub fn disable(&self, plugin_id: &str) -> Result<()> {
        let mut plugins = self.plugins.write();
        
        if let Some(plugin) = plugins.get_mut(plugin_id) {
            plugin.disable();
            Ok(())
        } else {
            Err(Error::PluginNotFound(plugin_id.to_string()))
        }
    }
    
    /// 获取插件
    pub fn get(&self, plugin_id: &str) -> Option<PluginInstance> {
        let plugins = self.plugins.read();
        plugins.get(plugin_id).cloned()
    }
    
    /// 列出所有插件
    pub fn list(&self) -> Vec<PluginInstance> {
        let plugins = self.plugins.read();
        plugins.values().cloned().collect()
    }
    
    /// 获取工具
    pub fn get_tool(&self, plugin_id: &str, tool_name: &str) -> Option<ToolDefinition> {
        let key = format!("{}.{}", plugin_id, tool_name);
        let registry = self.tool_registry.read();
        registry.get(&key).map(|e| e.tool.clone())
    }
    
    /// 列出所有工具
    pub fn list_tools(&self) -> Vec<ToolDefinition> {
        let registry = self.tool_registry.read();
        registry.values().map(|e| e.tool.clone()).collect()
    }
    
    /// 获取工具所属插件
    pub fn get_tool_plugin(&self, plugin_id: &str, tool_name: &str) -> Option<String> {
        let key = format!("{}.{}", plugin_id, tool_name);
        let registry = self.tool_registry.read();
        registry.get(&key).map(|e| e.plugin_id.clone())
    }
    
    /// 卸载插件
    pub fn unregister(&self, plugin_id: &str) -> bool {
        // 移除工具
        {
            let mut registry = self.tool_registry.write();
            registry.retain(|_, entry| entry.plugin_id != plugin_id);
        }
        
        // 移除插件
        let mut plugins = self.plugins.write();
        plugins.remove(plugin_id).is_some()
    }
}

impl Default for PluginManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_register_plugin() {
        let manager = PluginManager::new();
        let info = PluginInfo::new("test-plugin", "1.0.0");
        let id = manager.register(info);
        
        assert!(!id.is_empty());
        assert!(manager.get(&id).is_some());
    }
    
    #[test]
    fn test_load_and_enable_plugin() {
        let manager = PluginManager::new();
        let info = PluginInfo::new("test-plugin", "1.0.0");
        let id = manager.register(info);
        
        let tools = vec![
            ToolDefinition::new("hello", "Say hello"),
        ];
        
        manager.load(&id, tools).unwrap();
        manager.enable(&id).unwrap();
        
        let plugin = manager.get(&id).unwrap();
        assert_eq!(plugin.state, PluginState::Running);
        assert_eq!(plugin.tools.len(), 1);
    }
}
