//! 流式响应
//!
//! 处理代理响应的流式传输

use anyhow::Result;
use futures::Stream;
use std::pin::Pin;
use std::sync::Arc;
use tokio::sync::mpsc;
use tokio::task::JoinHandle;

use super::engine::AgentEvent;

/// 流式响应
pub struct StreamingResponse {
    pub events: mpsc::Receiver<AgentEvent>,
    task: Option<JoinHandle<()>>,
}

impl StreamingResponse {
    /// 创建新的流式响应
    pub fn new(rx: mpsc::Receiver<AgentEvent>) -> Self {
        Self {
            events: rx,
            task: None,
        }
    }

    /// 创建带后台任务的流式响应
    pub fn with_task(
        rx: mpsc::Receiver<AgentEvent>,
        task: JoinHandle<()>,
    ) -> Self {
        Self {
            events: rx,
            task: Some(task),
        }
    }

    /// 等待下一个事件
    pub async fn next(&mut self) -> Option<AgentEvent> {
        self.events.recv().await
    }

    /// 收集所有事件
    pub async fn collect(mut self) -> Vec<AgentEvent> {
        let mut events = Vec::new();
        while let Some(event) = self.events.recv().await {
            events.push(event);
        }
        events
    }

    /// 获取最终内容
    pub async fn final_content(&mut self) -> String {
        let mut content = String::new();
        while let Some(event) = self.events.recv().await {
            if let AgentEvent::Generation { text } = event {
                content.push_str(&text);
            }
        }
        content
    }
}

impl Stream for StreamingResponse {
    type Item = AgentEvent;

    fn poll_next(
        mut self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Option<Self::Item>> {
        use mpsc::Receiver;
        // 安全地委托给内部 receiver
        Pin::new(&mut self.events).poll_next(cx)
    }
}

/// 事件接收器
pub struct EventSink {
    sender: mpsc::Sender<AgentEvent>,
}

impl EventSink {
    /// 创建新的事件接收器
    pub fn new(sender: mpsc::Sender<AgentEvent>) -> Self {
        Self { sender }
    }

    /// 发送事件
    pub async fn send(&self, event: AgentEvent) {
        if let Err(e) = self.sender.send(event).await {
            tracing::warn!("Failed to send event: {:?}", e);
        }
    }

    /// 尝试发送事件
    pub fn try_send(&self, event: AgentEvent) -> bool {
        self.sender.try_send(event).is_ok()
    }
}

impl Clone for EventSink {
    fn clone(&self) -> Self {
        Self {
            sender: self.sender.clone(),
        }
    }
}

/// 创建流式响应通道
pub fn create_event_channel(buffer: usize) -> (EventSink, mpsc::Receiver<AgentEvent>) {
    let (tx, rx) = mpsc::channel(buffer);
    (EventSink::new(tx), rx)
}

/// 流式处理器
pub struct StreamingHandler {
    sink: EventSink,
}

impl StreamingHandler {
    /// 创建新的流式处理器
    pub fn new(sink: EventSink) -> Self {
        Self { sink }
    }

    /// 处理内容块
    pub async fn handle_content(&self, content: String) {
        self.sink.send(AgentEvent::Generation { content }).await;
    }

    /// 处理工具调用
    pub async fn handle_tool_call(&self, call: super::engine::ToolCall) {
        self.sink.send(AgentEvent::ToolCall { call }).await;
    }

    /// 处理完成
    pub async fn handle_complete(&self, content: String) {
        self.sink.send(AgentEvent::ResponseComplete {
            content,
            tool_results: vec![],
        }).await;
    }

    /// 处理错误
    pub async fn handle_error(&self, code: String, message: String) {
        self.sink.send(AgentEvent::Error { code, message }).await;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_event_channel() {
        let (sink, mut rx) = create_event_channel(10);
        
        sink.send(AgentEvent::ThinkingStarted {
            mode: super::super::thinking::ThinkMode::Balanced,
        }).await;
        
        let event = rx.recv().await.unwrap();
        match event {
            AgentEvent::ThinkingStarted { mode } => {
                assert_eq!(mode, super::super::thinking::ThinkMode::Balanced);
            }
            _ => panic!("Unexpected event type"),
        }
    }

    #[tokio::test]
    async fn test_streaming_response() {
        let (sink, rx) = create_event_channel(10);
        
        sink.send(AgentEvent::Generation {
            content: "Hello".to_string(),
        }).await;
        sink.send(AgentEvent::Generation {
            content: " World".to_string(),
        }).await;
        drop(sink);
        
        let mut response = StreamingResponse::new(rx);
        let content = response.final_content().await;
        
        assert_eq!(content, "Hello World");
    }
}
