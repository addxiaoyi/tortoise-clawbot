//! Agent Module Tests

use tortoise_core::agent::*;
use std::sync::Arc;

#[tokio::test]
async fn test_agent_config_default() {
    let config = AgentConfig::default();
    assert_eq!(config.name, "Tortoise");
    assert_eq!(config.temperature, 0.7);
}

#[tokio::test]
async fn test_think_mode_timeout() {
    assert_eq!(ThinkMode::Fast.timeout_ms(), 100);
    assert_eq!(ThinkMode::Balanced.timeout_ms(), 500);
    assert_eq!(ThinkMode::Deep.timeout_ms(), 2000);
    assert_eq!(ThinkMode::Research.timeout_ms(), 5000);
    assert_eq!(ThinkMode::Creative.timeout_ms(), 10000);
}

#[tokio::test]
async fn test_think_mode_temperature() {
    assert_eq!(ThinkMode::Fast.default_temperature(), 0.0);
    assert_eq!(ThinkMode::Balanced.default_temperature(), 0.5);
    assert_eq!(ThinkMode::Deep.default_temperature(), 0.7);
    assert_eq!(ThinkMode::Creative.default_temperature(), 1.0);
}

#[tokio::test]
async fn test_message_builder() {
    let msg = MessageBuilder::user("Hello, Tortoise!")
        .sender_id("user-123")
        .channel("discord")
        .build();
    
    assert_eq!(msg.role, MessageRole::User);
    assert_eq!(msg.content, "Hello, Tortoise!");
    assert_eq!(msg.metadata.sender_id, Some("user-123".to_string()));
    assert_eq!(msg.metadata.channel, Some("discord".to_string()));
}

#[tokio::test]
async fn test_message_builder_system() {
    let msg = MessageBuilder::system("You are a helpful assistant").build();
    assert_eq!(msg.role, MessageRole::System);
}

#[tokio::test]
async fn test_message_builder_assistant() {
    let msg = MessageBuilder::assistant("How can I help you?").build();
    assert_eq!(msg.role, MessageRole::Assistant);
}
