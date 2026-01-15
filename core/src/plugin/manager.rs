//! Plugin Manager
//! 
//! Manages plugin lifecycle and execution.

use crate::plugin::{PluginState, State, Sandbox};
use anyhow::{anyhow, Result};
use std::collections::HashMap;
use std::path::Path;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Plugin metadata
#[derive(Debug, Clone)]
pub struct PluginMetadata {
    /// Plugin ID
    pub id: String,
    /// Plugin name
    pub name: String,
    /// Version
    pub version: String,
    /// Description
    pub description: String,
    /// Author
    pub author: String,
    /// License
    pub license: String,
    /// Repository URL
    pub repository: Option<String>,
    /// Keywords
    pub keywords: Vec<String>,
    /// Dependencies
    pub dependencies: HashMap<String, String>,
}

/// Plugin type
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PluginType {
    Channel,
    Skill,
    Tool,
    Integration,
    Custom,
}

/// Plugin manager
pub struct PluginManager {
    /// Loaded plugins
    plugins: RwLock<HashMap<String, PluginInstance>>,
    /// Plugin directory
    plugin_dir: String,
    /// Sandbox enabled
    sandbox_enabled: bool,
}

/// Plugin instance
struct PluginInstance {
    /// Metadata
    metadata: PluginMetadata,
    /// Plugin type
    plugin_type: PluginType,
    /// State
    state: State,
    /// Sandbox
    sandbox: Option<Box<dyn Sandbox>>,
    /// Loaded at
    loaded_at: i64,
}

impl PluginManager {
    /// Create a new plugin manager
    pub fn new(plugin_dir: impl Into<String>, sandbox_enabled: bool) -> Self {
        Self {
            plugins: RwLock::new(HashMap::new()),
            plugin_dir: plugin_dir.into(),
            sandbox_enabled,
        }
    }

    /// Load a plugin
    pub async fn load_plugin(&self, path: &Path) -> Result<String> {
        // Load metadata
        let metadata = self.load_metadata(path)?;

        // Determine plugin type
        let plugin_type = self.detect_plugin_type(path)?;

        // Create sandbox if enabled
        let sandbox: Option<Box<dyn Sandbox>> = if self.sandbox_enabled {
            Some(Box::new(Sandbox::new(128 * 1024 * 1024))?) // 128MB limit
        } else {
            None
        };

        let instance = PluginInstance {
            metadata: metadata.clone(),
            plugin_type,
            state: State::Loaded,
            sandbox,
            loaded_at: chrono::Utc::now().timestamp(),
        };

        let mut plugins = self.plugins.write().await;
        plugins.insert(metadata.id.clone(), instance);

        tracing::info!("Loaded plugin: {} v{}", metadata.name, metadata.version);
        Ok(metadata.id)
    }

    /// Load plugin metadata
    fn load_metadata(&self, path: &Path) -> Result<PluginMetadata> {
        let meta_path = path.join("plugin.json");
        let content = std::fs::read_to_string(meta_path)?;
        let metadata: PluginMetadata = serde_json::from_str(&content)?;
        Ok(metadata)
    }

    /// Detect plugin type from path
    fn detect_plugin_type(&self, path: &Path) -> Result<PluginType> {
        let path_str = path.to_string_lossy();
        
        if path_str.contains("channel") {
            Ok(PluginType::Channel)
        } else if path_str.contains("skill") {
            Ok(PluginType::Skill)
        } else if path_str.contains("tool") {
            Ok(PluginType::Tool)
        } else if path_str.contains("integration") {
            Ok(PluginType::Integration)
        } else {
            Ok(PluginType::Custom)
        }
    }

    /// Enable a plugin
    pub async fn enable_plugin(&self, id: &str) -> Result<()> {
        let mut plugins = self.plugins.write().await;
        
        if let Some(plugin) = plugins.get_mut(id) {
            if plugin.state == State::Loaded || plugin.state == State::Stopped {
                plugin.state = State::Running;
                tracing::info!("Enabled plugin: {}", id);
                Ok(())
            } else {
                Err(anyhow!("Plugin {} is not in a state that can be enabled", id))
            }
        } else {
            Err(anyhow!("Plugin not found: {}", id))
        }
    }

    /// Disable a plugin
    pub async fn disable_plugin(&self, id: &str) -> Result<()> {
        let mut plugins = self.plugins.write().await;
        
        if let Some(plugin) = plugins.get_mut(id) {
            if plugin.state == State::Running {
                plugin.state = State::Stopped;
                tracing::info!("Disabled plugin: {}", id);
                Ok(())
            } else {
                Err(anyhow!("Plugin {} is not running", id))
            }
        } else {
            Err(anyhow!("Plugin not found: {}", id))
        }
    }

    /// Unload a plugin
    pub async fn unload_plugin(&self, id: &str) -> Result<()> {
        let mut plugins = self.plugins.write().await;
        
        if let Some(plugin) = plugins.remove(id) {
            tracing::info!("Unloaded plugin: {}", id);
            Ok(())
        } else {
            Err(anyhow!("Plugin not found: {}", id))
        }
    }

    /// Get plugin state
    pub async fn get_plugin_state(&self, id: &str) -> Result<PluginState> {
        let plugins = self.plugins.read().await;
        
        if let Some(plugin) = plugins.get(id) {
            Ok(PluginState {
                id: plugin.metadata.id.clone(),
                name: plugin.metadata.name.clone(),
                version: plugin.metadata.version.clone(),
                state: plugin.state,
            })
        } else {
            Err(anyhow!("Plugin not found: {}", id))
        }
    }

    /// List all plugins
    pub async fn list_plugins(&self) -> Vec<PluginState> {
        let plugins = self.plugins.read().await;
        plugins.values()
            .map(|p| PluginState {
                id: p.metadata.id.clone(),
                name: p.metadata.name.clone(),
                version: p.metadata.version.clone(),
                state: p.state,
            })
            .collect()
    }

    /// List plugins by type
    pub async fn list_plugins_by_type(&self, plugin_type: PluginType) -> Vec<PluginState> {
        let plugins = self.plugins.read().await;
        plugins.values()
            .filter(|p| p.plugin_type == plugin_type)
            .map(|p| PluginState {
                id: p.metadata.id.clone(),
                name: p.metadata.name.clone(),
                version: p.metadata.version.clone(),
                state: p.state,
            })
            .collect()
    }

    /// Get running plugin count
    pub async fn running_count(&self) -> usize {
        let plugins = self.plugins.read().await;
        plugins.values()
            .filter(|p| p.state == State::Running)
            .count()
    }

    /// Execute plugin method
    pub async fn call_plugin(&self, id: &str, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        let plugins = self.plugins.read().await;
        
        if let Some(plugin) = plugins.get(id) {
            if plugin.state != State::Running {
                return Err(anyhow!("Plugin {} is not running", id));
            }

            // In a full implementation, this would execute the method
            // For now, return a placeholder
            Ok(serde_json::json!({
                "success": true,
                "result": format!("Called {}.{} with {:?}", id, method, args)
            }))
        } else {
            Err(anyhow!("Plugin not found: {}", id))
        }
    }
}
