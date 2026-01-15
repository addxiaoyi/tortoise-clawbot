//! 记忆系统错误定义

use crate::error::ErrorCode;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("记忆不存在: {0}")]
    NotFound(String),
    
    #[error("记忆存储已满")]
    StorageFull,
    
    #[error("无法序列化记忆: {0}")]
    SerializationFailed(String),
    
    #[error("向量化失败: {0}")]
    VectorizationFailed(String),
    
    #[error("存储错误: {0}")]
    StorageError(String),
    
    #[error("检索错误: {0}")]
    RetrievalError(String),
}

impl Error {
    pub fn code(&self) -> i32 {
        match self {
            Error::NotFound(_) => 1001,
            Error::StorageFull => 1002,
            _ => 0,
        }
    }
}

impl From<Error> for crate::error::Error {
    fn from(e: Error) -> Self {
        crate::error::Error::Memory(e)
    }
}
