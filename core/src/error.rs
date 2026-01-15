//! Tortoise Core - 错误类型定义

use thiserror::Error;

/// 核心错误类型
#[derive(Error, Debug)]
pub enum Error {
    #[error("会话错误: {0}")]
    Session(#[from] session::Error),
    
    #[error("协议错误: {0}")]
    Protocol(String),
    
    #[error("插件错误: {0}")]
    Plugin(#[from] plugin::Error),
    
    #[error("工具错误: {0}")]
    Tool(String),
    
    #[error("记忆错误: {0}")]
    Memory(#[from] memory::Error),
    
    #[error("渠道错误: {0}")]
    Channel(String),
    
    #[error("AI 引擎错误: {0}")]
    AIEngine(String),
    
    #[error("网络错误: {0}")]
    Network(String),
    
    #[error("配置错误: {0}")]
    Config(String),
    
    #[error("认证错误: {0}")]
    Auth(String),
    
    #[error("权限错误: {0}")]
    Permission(String),
    
    #[error("资源不存在: {0}")]
    NotFound(String),
    
    #[error("资源已存在: {0}")]
    AlreadyExists(String),
    
    #[error("超时: {0}")]
    Timeout(String),
    
    #[error("无效参数: {0}")]
    InvalidArgument(String),
    
    #[error("内部错误: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, Error>;

/// 错误码定义
#[derive(Debug, Clone, Copy)]
pub enum ErrorCode {
    // 会话相关 (1000-1999)
    SessionNotFound = 1001,
    SessionExpired = 1002,
    SessionLimitExceeded = 1003,
    
    // 协议相关 (2000-2999)
    InvalidMessage = 2001,
    InvalidFormat = 2002,
    SerializationFailed = 2003,
    
    // 插件相关 (3000-3999)
    PluginNotFound = 3001,
    PluginLoadFailed = 3002,
    PluginExecutionFailed = 3003,
    PluginTimeout = 3004,
    
    // 工具相关 (4000-4999)
    ToolNotFound = 4001,
    ToolExecutionFailed = 4002,
    ToolTimeout = 4003,
    ToolPermissionDenied = 4004,
    
    // AI 相关 (5000-5999)
    ModelNotAvailable = 5001,
    ModelRateLimited = 5002,
    ModelQuotaExceeded = 5003,
    ModelTimeout = 5004,
    
    // 网络相关 (6000-6999)
    ConnectionFailed = 6001,
    ConnectionClosed = 6002,
    ConnectionTimeout = 6003,
    
    // 认证相关 (7000-7999)
    Unauthorized = 7001,
    TokenExpired = 7002,
    InvalidToken = 7003,
}

impl Error {
    /// 获取错误码
    pub fn code(&self) -> i32 {
        match self {
            Error::Session(e) => e.code(),
            Error::Plugin(e) => e.code(),
            Error::Memory(e) => e.code(),
            _ => 0,
        }
    }
    
    /// 是否可重试
    pub fn is_retryable(&self) -> bool {
        matches!(
            self,
            Error::Network(_) 
            | Error::Timeout(_)
            | Error::AIEngine(s) if s.contains("rate limit")
        )
    }
}
