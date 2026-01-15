//! Tortoise Runtime - 核心运行时
//! 
//! 高性能 Agent 运行时引擎

use crate::session::{SessionManager, Session};
use crate::memory::{MemoryManager, MemoryType};
use crate::plugin::{PluginManager, PluginInfo};
use std::sync::Arc;
use parking_lot::RwLock;
use std::collections::HashMap;
use uuid::Uuid;

/// Tortoise 运行时
pub struct Runtime {
    /// 会话管理器
    pub session_manager: Arc<SessionManager>,
    /// 记忆管理器
    pub memory_manager: Arc<MemoryManager>,
    /// 插件管理器
    pub plugin_manager: Arc<PluginManager>,
    /// 工具注册表
    pub tool_registry: Arc<RwLock<ToolRegistry>>,
    /// 运行状态
    running: Arc<RwLock<bool>>,
    /// 配置
    config: RuntimeConfig,
}

/// 运行时配置
#[derive(Debug, Clone)]
pub struct RuntimeConfig {
    pub max_sessions: usize,
    pub max_memory_working: usize,
    pub max_memory_semantic: usize,
    pub max_memory_episodic: usize,
    pub default_max_context: usize,
    pub enable_plugins: bool,
    pub enable_memory: bool,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            max_sessions: 10000,
            max_memory_working: 100,
            max_memory_semantic: 10000,
            max_memory_episodic: 5000,
            default_max_context: 100,
            enable_plugins: true,
            enable_memory: true,
        }
    }
}

/// 工具注册表
pub struct ToolRegistry {
    tools: HashMap<String, ToolEntry>,
}

struct ToolEntry {
    name: String,
    description: String,
    category: String,
}

impl ToolRegistry {
    pub fn new() -> Self {
        Self {
            tools: HashMap::new(),
        }
    }
    
    pub fn register(&mut self, name: String, description: String, category: String) {
        self.tools.insert(name.clone(), ToolEntry {
            name,
            description,
            category,
        });
    }
    
    pub fn unregister(&mut self, name: &str) -> bool {
        self.tools.remove(name).is_some()
    }
    
    pub fn list(&self) -> Vec<(String, String, String)> {
        self.tools.iter()
            .map(|(name, entry)| (name.clone(), entry.description.clone(), entry.category.clone()))
            .collect()
    }
}

impl Default for ToolRegistry {
    fn default() -> Self {
        Self::new()
    }
}

impl Runtime {
    /// 创建新的运行时
    pub fn new(config: RuntimeConfig) -> Self {
        Self {
            session_manager: Arc::new(SessionManager::new(config.max_sessions)),
            memory_manager: Arc::new(MemoryManager::with_capacity(
                config.max_memory_working,
                config.max_memory_semantic,
                config.max_memory_episodic,
            )),
            plugin_manager: Arc::new(PluginManager::new()),
            tool_registry: Arc::new(RwLock::new(ToolRegistry::new())),
            running: Arc::new(RwLock::new(false)),
            config,
        }
    }
    
    /// 启动运行时
    pub fn start(&self) {
        let mut running = self.running.write();
        *running = true;
        log::info!("Tortoise Runtime started");
    }
    
    /// 停止运行时
    pub fn stop(&self) {
        let mut running = self.running.write();
        *running = false;
        log::info!("Tortoise Runtime stopped");
    }
    
    /// 是否正在运行
    pub fn is_running(&self) -> bool {
        *self.running.read()
    }
    
    /// 创建新会话
    pub fn create_session(&self, user_id: impl Into<String>) -> Session {
        self.session_manager.create(user_id)
    }
    
    /// 获取会话
    pub fn get_session(&self, session_id: &str) -> Option<Session> {
        self.session_manager.get(session_id)
    }
    
    /// 保存记忆
    pub fn save_memory(
        &self,
        memory_type: MemoryType,
        content: impl Into<String>,
    ) -> String {
        match memory_type {
            MemoryType::Working => self.memory_manager.save_working(content),
            MemoryType::Semantic => self.memory_manager.save_semantic(content),
            MemoryType::Episodic => self.memory_manager.save_episodic(content),
        }
    }
    
    /// 查询记忆
    pub fn query_memory(&self, query: &str) -> crate::memory::MemoryQueryResult {
        self.memory_manager.query(crate::memory::MemoryQuery {
            query: query.to_string(),
            memory_type: None,
            limit: 10,
            similarity_threshold: 0.7,
        })
    }
    
    /// 注册插件
    pub fn register_plugin(&self, info: PluginInfo) -> String {
        self.plugin_manager.register(info)
    }
    
    /// 列出所有工具
    pub fn list_tools(&self) -> Vec<(String, String, String)> {
        let plugin_tools = self.plugin_manager.list_tools()
            .into_iter()
            .map(|t| (t.name, t.description, t.category));
        
        let mut all_tools: Vec<_> = plugin_tools.collect();
        
        // 添加内置工具
        let builtin_tools = self.tool_registry.read().list();
        all_tools.extend(builtin_tools);
        
        all_tools
    }
    
    /// 获取配置
    pub fn config(&self) -> &RuntimeConfig {
        &self.config
    }
}

impl Default for Runtime {
    fn default() -> Self {
        Self::new(RuntimeConfig::default())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_runtime_creation() {
        let runtime = Runtime::default();
        assert!(!runtime.is_running());
    }

    #[test]
    fn test_session_management() {
        let runtime = Runtime::default();
        let session = runtime.create_session("user123");
        
        assert_eq!(session.user_id, "user123");
        assert!(runtime.get_session(&session.id).is_some());
    }

    #[test]
    fn test_memory_management() {
        let runtime = Runtime::default();
        let id = runtime.save_memory(MemoryType::Working, "Test memory");
        assert!(!id.is_empty());
    }

    #[test]
    fn test_tool_registry() {
        let runtime = Runtime::default();
        let mut registry = runtime.tool_registry.write();
        registry.register(
            "test_tool".to_string(),
            "A test tool".to_string(),
            "testing".to_string(),
        );
        
        let tools = registry.list();
        assert!(tools.iter().any(|(name, _, _)| name == "test_tool"));
    }
}
