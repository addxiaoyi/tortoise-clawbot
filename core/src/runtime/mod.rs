//! Agent runtime engine

use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, error, instrument};

use crate::{Error, Result, session::SessionManager, tools::ToolRegistry, memory::MemoryStore};

mod config;
mod executor;

pub use config::RuntimeConfig;
pub use executor::Executor;

/// The main Tortoise runtime
pub struct AgentRuntime {
    config: RuntimeConfig,
    session_manager: Arc<RwLock<SessionManager>>,
    tool_registry: Arc<ToolRegistry>,
    memory_store: Arc<RwLock<MemoryStore>>,
    running: Arc<RwLock<bool>>,
}

impl AgentRuntime {
    /// Create a new runtime instance
    pub async fn new(config: RuntimeConfig) -> Result<Self> {
        info!("Initializing Tortoise runtime v{}", env!("CARGO_PKG_VERSION"));
        
        let session_manager = SessionManager::new(config.max_sessions).await?;
        let tool_registry = ToolRegistry::new();
        let memory_store = MemoryStore::new(&config.data_dir).await?;
        
        Ok(Self {
            config,
            session_manager: Arc::new(RwLock::new(session_manager)),
            tool_registry: Arc::new(tool_registry),
            memory_store: Arc::new(RwLock::new(memory_store)),
            running: Arc::new(RwLock::new(false)),
        }
    }
    
    /// Start the runtime
    #[instrument(skip(self))]
    pub async fn start(&self) -> Result<()> {
        let mut running = self.running.write().await;
        if *running {
            return Err(Error::Internal("Runtime already running".into()));
        }
        *running = true;
        
        info!("Tortoise runtime started");
        Ok(())
    }
    
    /// Stop the runtime
    #[instrument(skip(self))]
    pub async fn stop(&self) -> Result<()> {
        let mut running = self.running.write().await;
        if !*running {
            return Ok(());
        }
        *running = false;
        
        info!("Tortoise runtime stopped");
        Ok(())
    }
    
    /// Check if runtime is running
    pub async fn is_running(&self) -> bool {
        *self.running.read().await
    }
    
    /// Get session manager
    pub fn session