//! Tortoise Core Tests

#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_session_creation() {
        let manager = SessionManager::new(100).await.unwrap();
        
        let session = manager.create(Some("user123".to_string()), None);
        
        assert!(session.id.starts_with("sess_"));
        assert_eq!(session.user_id, Some("user123".to_string()));
        assert_eq!(session.status, SessionStatus::Active);
    }
    
    #[tokio::test]
    async fn test_session_update() {
        let manager = SessionManager::new(100).await.unwrap();
        let session = manager.create(None, None);
        
        // Update message count
        manager.increment_message_count(&session.id);
        
        let updated = manager.get(&session.id).unwrap();
        assert_eq!(updated.message_count, 1);
    }
    
    #[tokio::test]
    async fn test_session_deletion() {
        let manager = SessionManager::new(100).await.unwrap();
        let session = manager.create(None, None);
        
        assert!(manager.delete(&session.id));
        assert!(manager.get(&session.id).is_none());
    }
    
    #[tokio::test]
    async fn test_memory_store() {
        let store = MemoryStore::new("./test_data").await.unwrap();
        
        let memory = Memory::new(
            "User prefers dark mode".to_string(),
            MemoryType::Semantic,
        ).with_user("user123".to_string());
        
        store.store(memory);
        
        let retrieved = store.get(&store.get("").unwrap().first().unwrap().id).unwrap();
        assert_eq!(retrieved.content, "User prefers dark mode");
    }
    
    #[tokio::test]
    async fn test_memory_search() {
        let store = MemoryStore::new("./test_data").await.unwrap();
        
        store.store(Memory::new(
            "User prefers dark mode".to_string(),
            MemoryType::Semantic,
        ));
        store.store(Memory::new(
            "User name is John".to_string(),
            MemoryType::Semantic,
        ));
        
        let results = store.search("dark mode", 10);
        assert_eq!(results.len(), 1);
        assert!(results[0].content.contains("dark mode"));
    }
    
    #[tokio::test]
    async fn test_tool_registry() {
        let registry = ToolRegistry::new();
        
        // Built-in tools should be registered
        let tools = registry.list();
        assert!(tools.iter().any(|t| t.name == "web_search"));
        assert!(tools.iter().any(|t| t.name == "calculator"));
    }
    
    #[tokio::test]
    async fn test_tool_execution() {
        let registry = ToolRegistry::new();
        
        let call = ToolCall::new(
            "calculator".to_string(),
            serde_json::json!({"expression": "2 + 2"}),
        );
        
        let result = registry.execute(call).await;
        assert!(result.success);
    }
    
    #[tokio::test]
    async fn test_protocol_codec() {
        let codec = ProtocolCodec::new();
        
        let frame = MessageFrame::new(
            0x0003, // Request
            b"Hello, Tortoise!".to_vec(),
        );
        
        let encoded = codec.encode(&frame).unwrap();
        let decoded = codec.decode(encoded).unwrap();
        
        assert_eq!(frame.msg_type, decoded.msg_type);
        assert_eq!(frame.payload, decoded.payload);
    }
    
    #[tokio::test]
    async fn test_protocol_compression() {
        let codec = ProtocolCodec::new();
        
        let data = vec![0u8; 1000]; // Highly compressible
        
        let compressed = codec.compress(&data).unwrap();
        let decompressed = codec.decompress(&compressed).unwrap();
        
        assert_eq!(data, decompressed);
        assert!(compressed.len() < data.len());
    }
    
    #[tokio::test]
    async fn test_circuit_breaker() {
        let mut breaker = CircuitBreaker::new(3, 60);
        
        // Record failures until open
        breaker.record_failure();
        breaker.record_failure();
        assert!(breaker.is_available());
        
        breaker.record_failure();
        assert!(!breaker.is_available());
        
        // Reset to half-open
        breaker.try_reset();
        breaker.record_success();
        assert!(breaker.is_available());
    }
    
    #[tokio::test]
    async fn test_priority_queue() {
        let mut queue = TaskQueue::new();
        
        queue.push(Task::new("low".to_string()).with_priority(Priority::Low));
        queue.push(Task::new("critical".to_string()).with_priority(Priority::Critical));
        queue.push(Task::new("high".to_string()).with_priority(Priority::High));
        
        assert_eq!(queue.pop().unwrap().name, "critical");
        assert_eq!(queue.pop().unwrap().name, "high");
        assert_eq!(queue.pop().unwrap().name, "low");
    }
    
    #[tokio::test]
    async fn test_agent_request_serialization() {
        let request = AgentRequest::new(
            "sess_123".to_string(),
            "Hello".to_string(),
        );
        
        let json = serde_json::to_string(&request).unwrap();
        let parsed: AgentRequest = serde_json::from_str(&json).unwrap();
        
        assert_eq!(request.session_id, parsed.session_id);
        assert_eq!(request.content, parsed.content);
    }
    
    #[tokio::test]
    async fn test_agent_response_serialization() {
        let response = AgentResponse {
            request_id: "req_123".to_string(),
            session_id: "sess_123".to_string(),
            content: "Hello back".to_string(),
            content_type: ContentType::Text,
            tool_calls: vec![],
            metadata: ResponseMetadata {
                model: "gpt-4o".to_string(),
                tokens: TokenUsage {
                    prompt: 10,
                    completion: 20,
                },
                latency_ms: Some(100),
            },
        };
        
        let json = serde_json::to_string(&response).unwrap();
        let parsed: AgentResponse = serde_json::from_str(&json).unwrap();
        
        assert_eq!(response.content, parsed.content);
        assert_eq!(response.metadata.model, parsed.metadata.model);
    }
}

// Benchmark tests
#[cfg(test)]
mod benchmarks {
    use super::*;
    
    #[tokio::test]
    async fn bench_session_creation(b: &mut tokio::test::BENCH) {
        let manager = SessionManager::new(10000).await.unwrap();
        
        b.iter(|| {
            manager.create(None, None);
        });
    }
    
    #[tokio::test]
    async fn bench_memory_search(b: &mut tokio::test::BENCH) {
        let store = MemoryStore::new("./test_data").await.unwrap();
        
        // Add 1000 memories
        for i in 0..1000 {
            store.store(Memory::new(
                format!("Memory {}", i),
                MemoryType::Semantic,
            ));
        }
        
        b.iter(|| {
            store.search("Memory 500", 10);
        });
    }
}
