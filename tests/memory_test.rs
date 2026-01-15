//! Memory Module Tests

use tortoise_core::memory::*;
use std::sync::Arc;

#[tokio::test]
async fn test_memory_config_default() {
    let config = MemoryConfig::default();
    assert_eq!(config.short_term.max_items, 100);
    assert_eq!(config.medium_term.max_items, 1000);
    assert_eq!(config.long_term.max_items, 10000);
    assert_eq!(config.promotion_threshold, 0.8);
}

#[tokio::test]
async fn test_memory_item_creation() {
    let item = MemoryItem::new(
        "Test content".to_string(),
        MemoryType::ShortTerm,
        0.5
    );
    
    assert_eq!(item.content, "Test content");
    assert_eq!(item.memory_type, MemoryType::ShortTerm);
    assert_eq!(item.importance, 0.5);
    assert!(!item.id.is_empty());
}

#[tokio::test]
async fn test_memory_item_access() {
    let mut item = MemoryItem::new(
        "Test".to_string(),
        MemoryType::ShortTerm,
        0.5
    );
    
    let initial_count = item.access_count;
    item.access();
    assert_eq!(item.access_count, initial_count + 1);
}

#[tokio::test]
async fn test_memory_query_default() {
    let query = MemoryQuery::default();
    assert!(query.query.is_empty());
    assert!(query.memory_type.is_none());
    assert_eq!(query.limit, Some(10));
    assert_eq!(query.threshold, Some(0.7));
}

#[tokio::test]
async fn test_in_memory_store() {
    let config = TierConfig {
        max_items: 10,
        max_size_mb: 50,
        ttl_seconds: 3600,
    };
    
    let store = InMemoryStore::new(&config).unwrap();
    
    // Store an item
    let item = MemoryItem::new(
        "Test memory".to_string(),
        MemoryType::ShortTerm,
        0.7
    );
    
    let id = store.store(item).await.unwrap();
    assert!(!id.is_empty());
    
    // Retrieve the item
    let retrieved = store.get(&id).await.unwrap();
    assert!(retrieved.is_some());
    assert_eq!(retrieved.unwrap().content, "Test memory");
    
    // Count
    let count = store.count().await.unwrap();
    assert_eq!(count, 1);
    
    // Query
    let results = store.query(MemoryQuery::default()).await.unwrap();
    assert_eq!(results.len(), 1);
    
    // Delete
    store.delete(&id).await.unwrap();
    let count_after = store.count().await.unwrap();
    assert_eq!(count_after, 0);
}

#[tokio::test]
async fn test_memory_manager_remember() {
    let config = MemoryConfig::default();
    let manager = MemoryManager::new(&config).await.unwrap();
    
    let id = manager.remember("Important fact".to_string(), 0.9).await.unwrap();
    assert!(!id.is_empty());
}

#[tokio::test]
async fn test_memory_manager_recall() {
    let config = MemoryConfig::default();
    let manager = MemoryManager::new(&config).await.unwrap();
    
    // Remember some items
    manager.remember("Python is a programming language".to_string(), 0.8).await.unwrap();
    manager.remember("Rust is a systems programming language".to_string(), 0.7).await.unwrap();
    
    // Recall
    let memories = manager.recall("programming").await.unwrap();
    assert!(!memories.is_empty());
}

#[tokio::test]
async fn test_memory_stats() {
    let config = MemoryConfig::default();
    let manager = MemoryManager::new(&config).await.unwrap();
    
    let stats = manager.stats().await.unwrap();
    assert_eq!(stats.short_term_count, 0);
    assert_eq!(stats.medium_term_count, 0);
    assert_eq!(stats.long_term_count, 0);
}

#[tokio::test]
async fn test_memory_clear() {
    let config = MemoryConfig::default();
    let manager = MemoryManager::new(&config).await.unwrap();
    
    // Add some memories
    manager.remember("Test 1".to_string(), 0.5).await.unwrap();
    manager.remember("Test 2".to_string(), 0.6).await.unwrap();
    
    // Clear
    manager.clear().await.unwrap();
    
    let stats = manager.stats().await.unwrap();
    assert_eq!(stats.total_count, 0);
}
