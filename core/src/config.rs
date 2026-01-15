//! 配置模块

use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// 配置初始化
pub fn init_config() {
    // 加载配置文件
    let config_paths = vec![
        PathBuf::from("~/.tortoise/config.toml"),
        PathBuf::from("./config.toml"),
        PathBuf::from("/etc/tortoise/config.toml"),
    ];
    
    for path in config_paths {
        if path.exists() {
            tracing::info!("Found config at: {:?}", path);
            // 加载配置
            break;
        }
    }
}

/// 全局配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GlobalConfig {
    pub version: String,
    pub data_dir: PathBuf,
    pub log_level: String,
    pub agent: AgentConfig,
    pub memory: MemoryConfig,
    pub plugin: PluginConfig,
    pub security: SecurityConfig,
    pub network: NetworkConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentConfig {
    pub id: String,
    pub name: String,
    pub model: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryConfig {
    pub storage_path: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PluginConfig {
    pub sandbox_enabled: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityConfig {
    pub encryption_enabled: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NetworkConfig {
    pub enabled: bool,
}

impl Default for GlobalConfig {
    fn default() -> Self {
        Self {
            version: env!("CARGO_PKG_VERSION").to_string(),
            data_dir: PathBuf::from("~/.tortoise"),
            log_level: "info".to_string(),
            agent: AgentConfig {
                id: "default".to_string(),
                name: "Tortoise".to_string(),
                model: "gpt-4".to_string(),
            },
            memory: MemoryConfig {
                storage_path: "~/.tortoise/memory".to_string(),
            },
            plugin: PluginConfig {
                sandbox_enabled: true,
            },
            security: SecurityConfig {
                encryption_enabled: true,
            },
            network: NetworkConfig {
                enabled: true,
            },
        }
    }
}
