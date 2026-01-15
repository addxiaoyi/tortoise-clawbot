//! Plugin Module
//! 
//! Plugin system with WASM sandbox support.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Plugin configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginConfig {
    /// Plugin directory
    pub plugin_dir: String,
    /// WASM sandbox memory limit (MB)
    pub wasm_memory_limit_mb: u32,
    /// Whether sandbox is enabled
    pub sandbox_enabled: bool,
    /// Auto-load plugins on startup
    pub auto_load: bool,
    /// Plugin dependencies
    pub dependencies: HashMap<String, String>,
}

impl Default for PluginConfig {
    fn default() -> Self {
        Self {
            plugin_dir: "plugins".to_string(),
            wasm_memory_limit_mb: 128,
            sandbox_enabled: true,
            auto_load: true,
            dependencies: HashMap::new(),
        }
    }
}

/// Plugin metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
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
    /// Plugin type
    pub plugin_type: PluginType,
}

/// Plugin type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PluginType {
    Channel,
    Skill,
    Tool,
    Integration,
    Custom,
}

/// Plugin state
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PluginState {
    Loaded,
    Running,
    Stopped,
    Error(String),
}

/// Plugin instance
pub struct Plugin {
    metadata: PluginMetadata,
    state: RwLock<PluginState>,
    runtime: Box<dyn PluginRuntime>,
}

impl Plugin {
    /// Create a new plugin
    pub fn new(metadata: PluginMetadata, runtime: Box<dyn PluginRuntime>) -> Self {
        Self {
            metadata,
            state: RwLock::new(PluginState::Loaded),
            runtime,
        }
    }

    /// Start the plugin
    pub async fn start(&self) -> Result<()> {
        self.runtime.start().await?;
        *self.state.write().await = PluginState::Running;
        Ok(())
    }

    /// Stop the plugin
    pub async fn stop(&self) -> Result<()> {
        self.runtime.stop().await?;
        *self.state.write().await = PluginState::Stopped;
        Ok(())
    }

    /// Call a plugin method
    pub async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        self.runtime.call(method, args).await
    }

    /// Get plugin state
    pub async fn state(&self) -> PluginState {
        self.state.read().await.clone()
    }

    /// Get plugin metadata
    pub fn metadata(&self) -> &PluginMetadata {
        &self.metadata
    }
}

/// Plugin runtime trait
#[async_trait::async_trait]
pub trait PluginRuntime: Send + Sync {
    /// Start the runtime
    async fn start(&self) -> Result<()>;

    /// Stop the runtime
    async fn stop(&self) -> Result<()>;

    /// Call a method
    async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value>;

    /// Get resource usage
    fn resource_usage(&self) -> ResourceUsage;
}

/// Resource usage statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResourceUsage {
    /// Memory in MB
    pub memory_mb: f64,
    /// CPU percentage
    pub cpu_percent: f32,
    /// Uptime in seconds
    pub uptime_seconds: u64,
}

/// Plugin manager
pub struct PluginManager {
    plugins: RwLock<HashMap<String, Arc<Plugin>>>,
    config: PluginConfig,
}

impl PluginManager {
    /// Create a new plugin manager
    pub fn new(config: &PluginConfig) -> Result<Self> {
        Ok(Self {
            plugins: RwLock::new(HashMap::new()),
            config: config.clone(),
        })
    }

    /// Load a plugin from path
    pub async fn load_plugin(&self, path: &str) -> Result<String> {
        // Read metadata
        let metadata = self.read_metadata(path)?;
        let id = metadata.id.clone();

        // Create runtime based on file extension
        let runtime: Box<dyn PluginRuntime> = if path.ends_with(".wasm") {
            Box::new(WasmRuntime::new(path, self.config.wasm_memory_limit_mb)?)
        } else if path.ends_with(".so") || path.ends_with(".dll") {
            Box::new(DynamicRuntime::load(path)?)
        } else {
            anyhow::bail!("Unsupported plugin format: {}", path);
        };

        let plugin = Arc::new(Plugin::new(metadata, runtime));
        
        let mut plugins = self.plugins.write().await;
        plugins.insert(id.clone(), plugin);

        Ok(id)
    }

    /// Read plugin metadata
    fn read_metadata(&self, path: &str) -> Result<PluginMetadata> {
        let meta_path = format!("{}/plugin.json", path);
        let content = std::fs::read_to_string(meta_path)?;
        let metadata: PluginMetadata = serde_json::from_str(&content)?;
        Ok(metadata)
    }

    /// Enable a plugin
    pub async fn enable_plugin(&self, id: &str) -> Result<()> {
        let plugins = self.plugins.read().await;
        let plugin = plugins.get(id)
            .ok_or_else(|| anyhow::anyhow!("Plugin not found: {}", id))?;
        
        plugin.start().await?;
        Ok(())
    }

    /// Disable a plugin
    pub async fn disable_plugin(&self, id: &str) -> Result<()> {
        let plugins = self.plugins.read().await;
        let plugin = plugins.get(id)
            .ok_or_else(|| anyhow::anyhow!("Plugin not found: {}", id))?;
        
        plugin.stop().await?;
        Ok(())
    }

    /// List all plugins
    pub async fn list_plugins(&self) -> Vec<PluginMetadata> {
        let plugins = self.plugins.read().await;
        plugins.values()
            .map(|p| p.metadata().clone())
            .collect()
    }

    /// Get a plugin by ID
    pub async fn get_plugin(&self, id: &str) -> Option<Arc<Plugin>> {
        let plugins = self.plugins.read().await;
        plugins.get(id).cloned()
    }

    /// Call a plugin method
    pub async fn call_plugin(&self, id: &str, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        let plugins = self.plugins.read().await;
        let plugin = plugins.get(id)
            .ok_or_else(|| anyhow::anyhow!("Plugin not found: {}", id))?;
        
        plugin.call(method, args).await
    }

    /// Unload a plugin
    pub async fn unload_plugin(&self, id: &str) -> Result<()> {
        let mut plugins = self.plugins.write().await;
        plugins.remove(id);
        Ok(())
    }
}

/// WASM runtime implementation
pub struct WasmRuntime {
    path: String,
    memory_limit_mb: u32,
    // In production, would hold wasmer instance
}

impl WasmRuntime {
    pub fn new(path: &str, memory_limit_mb: u32) -> Result<Self> {
        Ok(Self {
            path: path.to_string(),
            memory_limit_mb,
        })
    }
}

#[async_trait::async_trait]
impl PluginRuntime for WasmRuntime {
    async fn start(&self) -> Result<()> {
        tracing::info!("Starting WASM plugin: {}", self.path);
        Ok(())
    }

    async fn stop(&self) -> Result<()> {
        tracing::info!("Stopping WASM plugin: {}", self.path);
        Ok(())
    }

    async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        tracing::debug!("Calling WASM method: {}", method);
        // In production, would invoke WASM function
        Ok(serde_json::json!({
            "success": true,
            "method": method,
            "args": args
        }))
    }

    fn resource_usage(&self) -> ResourceUsage {
        ResourceUsage {
            memory_mb: self.memory_limit_mb as f64 * 0.5, // Estimate
            cpu_percent: 0.0,
            uptime_seconds: 0,
        }
    }
}

/// Dynamic library runtime implementation
pub struct DynamicRuntime {
    path: String,
}

impl DynamicRuntime {
    pub fn load(path: &str) -> Result<Self> {
        Ok(Self {
            path: path.to_string(),
        })
    }
}

#[async_trait::async_trait]
impl PluginRuntime for DynamicRuntime {
    async fn start(&self) -> Result<()> {
        tracing::info!("Starting dynamic plugin: {}", self.path);
        Ok(())
    }

    async fn stop(&self) -> Result<()> {
        tracing::info!("Stopping dynamic plugin: {}", self.path);
        Ok(())
    }

    async fn call(&self, method: &str, args: serde_json::Value) -> Result<serde_json::Value> {
        tracing::debug!("Calling dynamic method: {}", method);
        // In production, would call native function via FFI
        Ok(serde_json::json!({
            "success": true,
            "method": method,
            "args": args
        }))
    }

    fn resource_usage(&self) -> ResourceUsage {
        ResourceUsage {
            memory_mb: 10.0, // Estimate
            cpu_percent: 0.0,
            uptime_seconds: 0,
        }
    }
}
