//! 会话错误定义

use crate::error::ErrorCode;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("会话不存在: {0}")]
    NotFound(String),
    
    #[error("会话已过期")]
    Expired,
    
    #[error("会话数量超限")]
    LimitExceeded,
    
    #[error("会话已关闭")]
    SessionClosed,
    
    #[error("会话已暂停")]
    SessionPaused,
    
    #[error("上下文超限")]
    ContextOverflow,
    
    #[error("无效的操作: {0}")]
    InvalidOperation(String),
}

impl Error {
    pub fn code(&self) -> i32 {
        match self {
            Error::NotFound(_) => ErrorCode::SessionNotFound as i32,
            Error::Expired => ErrorCode::SessionExpired as i32,
            Error::LimitExceeded => ErrorCode::SessionLimitExceeded as i32,
            _ => 0,
        }
    }
}

impl From<Error> for crate::error::Error {
    fn from(e: Error) -> Self {
        crate::error::Error::Session(e)
    }
}
