//! AI Provider Trait Definitions
//! 
//! 定义 AI Provider 接口

use async_trait::async_trait;
use futures_util::Stream;
use std::pin::Pin;

use super::{AIError, ChatRequest, ChatResponse};

/// AI Provider 接口
#[async_trait]
pub trait AIProvider: Send + Sync {
    /// 获取提供商名称
    fn name(&self) -> super::ModelProvider;
    
    /// 获取支持的模型列表
    fn supported_models(&self) -> Vec<String>;
    
    /// 聊天完成
    async fn chat(&self, request: ChatRequest) -> Result<ChatResponse, AIError>;
    
    /// 流式聊天完成
    async fn chat_stream(
        &self,
        request: ChatRequest,
    ) -> Result<Pin<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>>, AIError>;
    
    /// 检查健康状态
    async fn health_check(&self) -> bool;
}

pub type ChatStream = Pin<Box<dyn Stream<Item = Result<ChatResponse, AIError>> + Send>>;
