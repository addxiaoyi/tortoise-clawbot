//! 插件系统错误定义

use crate::error::ErrorCode;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("插件不存在: {0}")]
    PluginNotFound(String),
    
    #[error("插件已存在: {0}")]
    PluginAlreadyExists(String),
    
    #[error("插件加载失败: {0}")]
    LoadFailed(String),
    
    #[error("插件执行失败: {0}")]
    ExecutionFailed(String),
    
    #[error("工具不存在: {0}")]
    ToolNotFound(String),
    
    #[error("工具执行失败: {0}")]
    ToolExecutionFailed(String),
    
    #[error("插件未启用")]
    PluginDisabled,
    
    #[error("插件版本不兼容: {0}")]
    VersionMismatch(String),
    
    #[error("权限不足: {0}")]
    PermissionDenied(String),
    
    #[error("沙箱执行失败: {0}")]
    SandboxError(String),
}

impl Error {
    pub fn code(&self) -> i32 {
        match self {
            Error::PluginNotFound(_) => ErrorCode::PluginNotFound as i32,
            Error::LoadFailed(_) => ErrorCode::PluginLoadFailed as i32,
            Error::ToolExecutionFailed(_) => ErrorCode::ToolExecutionFailed as i32,
            _ => 0,
        }
    }
}

impl From<Error> for crate::error::Error {
    fn from(e: Error) -> Self {
        crate::error::Error::Plugin(e)
    }
}
