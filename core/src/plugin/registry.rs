//! Plugin Registry
//! 
//! Registry for discovering and managing plugins.

use crate::plugin::{PluginMetadata, PluginType};
use anyhow::Result;
use std::collections::HashMap;

/// Plugin registry
pub struct PluginRegistry {
    /// Available plugins
    plugins: HashMap<String, PluginMetadata>,
}

impl PluginRegistry {
    /// Create a new registry
    pub fn new() -> Self {
        Self {
            plugins: HashMap::new(),
        }
    }

    /// Register a plugin
    pub fn register(&mut self, metadata: PluginMetadata) {
        self.plugins.insert(metadata.id.clone(), metadata);
    }

    /// Unregister a plugin
    pub fn unregister(&mut self, id: &str) -> Option<PluginMetadata> {
        self.plugins.remove(id)
    }

    /// Get plugin metadata
    pub fn get(&self, id: &str) -> Option<&PluginMetadata> {
        self.plugins.get(id)
    }

    /// List all plugins
    pub fn list(&self) -> Vec<&PluginMetadata> {
        self.plugins.values().collect()
    }

    /// Search plugins by keyword
    pub fn search(&self, keyword: &str) -> Vec<&PluginMetadata> {
        let keyword_lower = keyword.to_lowercase();
        self.plugins.values()
            .filter(|p| {
                p.name.to_lowercase().contains(&keyword_lower)
                    || p.description.to_lowercase().contains(&keyword_lower)
                    || p.keywords.iter().any(|k| k.to_lowercase().contains(&keyword_lower))
            })
            .collect()
    }

    /// List plugins by type
    pub fn list_by_type(&self, plugin_type: PluginType) -> Vec<&PluginMetadata> {
        // This would need a type field in PluginMetadata
        self.plugins.values().collect()
    }

    /// Get plugin count
    pub fn count(&self) -> usize {
        self.plugins.len()
    }
}

impl Default for PluginRegistry {
    fn default() -> Self {
        Self::new()
    }
}
