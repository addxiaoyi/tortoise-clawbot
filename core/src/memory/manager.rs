//! Memory Manager
//! 
//! Orchestrates the three-tier memory system.

use crate::memory::{MemoryItem, MemoryQuery, MemoryStats, MemoryStore, MemoryType, VectorStore};
use anyhow::Result;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Memory importance thresholds
const HIGH_IMPORTANCE_THRESHOLD: f32 = 0.8;
const LOW_IMPORTANCE_THRESHOLD: f32 = 0.3;

/// Memory manager coordinating three memory tiers
pub struct MemoryManager {
    /// Short-term memory store
    short_term: Arc<dyn MemoryStore>,
    /// Medium-term memory store
    medium_term: Arc<dyn MemoryStore>,
    /// Long-term memory store (vector-based)
    long_term: Arc<dyn MemoryStore>,
    /// Vector store for semantic search
    vector_store: Arc<dyn VectorStore>,
    /// Access statistics
    stats: RwLock<MemoryAccessStats>,
}

/// Memory access statistics
#[derive(Debug, Clone)]
pub struct MemoryAccessStats {
    pub hits: u64,
    pub misses: u64,
    pub promotions: u64,
    pub demotions: u64,
}

impl Default for MemoryAccessStats {
    fn default() -> Self {
        Self {
            hits: 0,
            misses: 0,
            promotions: 0,
            demotions: 0,
        }
    }
}

impl MemoryManager {
    /// Create a new memory manager
    pub fn new(
        short_term: Arc<dyn MemoryStore>,
        medium_term: Arc<dyn MemoryStore>,
        long_term: Arc<dyn MemoryStore>,
        vector_store: Arc<dyn VectorStore>,
    ) -> Self {
        Self {
            short_term,
            medium_term,
            long_term,
            vector_store,
            stats: RwLock::new(Default::default()),
        }
    }

    /// Store a new memory item
    pub async fn remember(&self, content: String, importance: f32) -> Result<String> {
        // Generate embedding for the content
        let embedding = self.vector_store.embed(&content).await?;

        let item = MemoryItem {
            id: uuid::Uuid::new_v4().to_string(),
            content,
            memory_type: MemoryType::ShortTerm,
            importance,
            created_at: chrono::Utc::now().timestamp(),
            last_accessed: chrono::Utc::now().timestamp(),
            access_count: 0,
            metadata: serde_json::json!({ "embedding": embedding }),
        };

        let id = self.short_term.store(item.clone()).await?;

        // Handle importance-based promotion
        if importance >= HIGH_IMPORTANCE_THRESHOLD {
            self.promote_to_long_term(&id).await?;
        }

        Ok(id)
    }

    /// Recall memories relevant to a query
    pub async fn recall(&self, query: &str) -> Result<Vec<MemoryItem>> {
        // Generate embedding for the query
        let query_embedding = self.vector_store.embed(query).await?;

        let mut results = Vec::new();

        // Search long-term memory first (most important)
        let long_term_results = self.long_term.retrieve(MemoryQuery {
            query: query.to_string(),
            memory_type: Some(MemoryType::LongTerm),
            limit: Some(10),
            threshold: Some(0.7),
            embedding: Some(query_embedding.clone()),
        }).await?;
        
        results.extend(long_term_results);

        // Then search medium-term memory
        let medium_term_results = self.medium_term.retrieve(MemoryQuery {
            query: query.to_string(),
            memory_type: Some(MemoryType::MediumTerm),
            limit: Some(5),
            threshold: Some(0.7),
            embedding: Some(query_embedding.clone()),
        }).await?;
        
        results.extend(medium_term_results);

        // Finally search short-term memory
        let short_term_results = self.short_term.retrieve(MemoryQuery {
            query: query.to_string(),
            memory_type: Some(MemoryType::ShortTerm),
            limit: Some(3),
            threshold: Some(0.6),
            embedding: Some(query_embedding),
        }).await?;
        
        results.extend(short_term_results);

        // Sort by importance and deduplicate
        results.sort_by(|a, b| b.importance.partial_cmp(&a.importance).unwrap());
        results.dedup_by(|a, b| a.id == b.id);

        // Update access statistics
        {
            let mut stats = self.stats.write().await;
            if results.is_empty() {
                stats.misses += 1;
            } else {
                stats.hits += 1;
            }
        }

        // Update access times
        for item in &results {
            self.update_access_time(&item.id).await?;
        }

        Ok(results)
    }

    /// Promote a memory item to a higher tier
    pub async fn promote(&self, id: &str) -> Result<()> {
        // Try to find in short-term first
        if let Ok(Some(item)) = self.short_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(item) = item.first() {
                return self.promote_item(item).await;
            }
        }

        // Try medium-term
        if let Ok(Some(item)) = self.medium_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(item) = item.first() {
                return self.promote_item(item).await;
            }
        }

        Ok(())
    }

    async fn promote_item(&self, item: &MemoryItem) -> Result<()> {
        match item.memory_type {
            MemoryType::ShortTerm => {
                self.promote_to_medium_term(&item.id).await?;
            }
            MemoryType::MediumTerm => {
                self.promote_to_long_term(&item.id).await?;
            }
            MemoryType::LongTerm => {
                // Already at top tier, update importance
                let mut updated = item.clone();
                updated.importance = (updated.importance + 0.1).min(1.0);
                self.long_term.update(&item.id, updated).await?;
            }
        }
        Ok(())
    }

    async fn promote_to_medium_term(&self, id: &str) -> Result<()> {
        // Find and remove from short-term
        if let Ok(Some(mut items)) = self.short_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(mut item) = items.pop() {
                item.memory_type = MemoryType::MediumTerm;
                self.medium_term.store(item.clone()).await?;
                self.short_term.delete(id).await?;

                let mut stats = self.stats.write().await;
                stats.promotions += 1;
            }
        }
        Ok(())
    }

    async fn promote_to_long_term(&self, id: &str) -> Result<()> {
        // Find and remove from current tier
        if let Ok(Some(mut items)) = self.medium_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(mut item) = items.pop() {
                item.memory_type = MemoryType::LongTerm;
                
                // Store in long-term with vector embedding
                let embedding = self.vector_store.embed(&item.content).await?;
                item.metadata = serde_json::json!({ "embedding": embedding });
                
                self.long_term.store(item.clone()).await?;
                self.medium_term.delete(id).await?;

                let mut stats = self.stats.write().await;
                stats.promotions += 1;
            }
        }
        Ok(())
    }

    /// Demote low-importance memories
    pub async fn forget_low_importance(&self, threshold: f32) -> Result<usize> {
        let mut count = 0;

        // Check medium-term
        let items = self.medium_term.retrieve(MemoryQuery {
            query: String::new(),
            memory_type: Some(MemoryType::MediumTerm),
            limit: Some(1000),
            ..Default::default()
        }).await?;

        for item in items {
            if item.importance < threshold {
                self.medium_term.delete(&item.id).await?;
                count += 1;
            }
        }

        // Check long-term for very low importance
        let items = self.long_term.retrieve(MemoryQuery {
            query: String::new(),
            memory_type: Some(MemoryType::LongTerm),
            limit: Some(1000),
            ..Default::default()
        }).await?;

        for item in items {
            if item.importance < threshold / 2.0 {
                self.long_term.delete(&item.id).await?;
                count += 1;

                let mut stats = self.stats.write().await;
                stats.demotions += 1;
            }
        }

        Ok(count)
    }

    /// Update memory item access time
    pub async fn update_access_time(&self, id: &str) -> Result<()> {
        // Try all tiers
        if let Ok(Some(mut items)) = self.short_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(mut item) = items.pop() {
                item.last_accessed = chrono::Utc::now().timestamp();
                item.access_count += 1;
                self.short_term.update(id, item).await?;
                return Ok(());
            }
        }

        if let Ok(Some(mut items)) = self.medium_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(mut item) = items.pop() {
                item.last_accessed = chrono::Utc::now().timestamp();
                item.access_count += 1;
                self.medium_term.update(id, item).await?;
                return Ok(());
            }
        }

        if let Ok(Some(mut items)) = self.long_term.retrieve(MemoryQuery {
            query: id.to_string(),
            limit: Some(1),
            ..Default::default()
        }).await {
            if let Some(mut item) = items.pop() {
                item.last_accessed = chrono::Utc::now().timestamp();
                item.access_count += 1;
                self.long_term.update(id, item).await?;
                return Ok(());
            }
        }

        Ok(())
    }

    /// Get memory statistics
    pub async fn get_stats(&self) -> Result<MemoryStats> {
        let short_stats = self.short_term.stats().await?;
        let medium_stats = self.medium_term.stats().await?;
        let long_stats = self.long_term.stats().await?;

        Ok(MemoryStats {
            short_term_count: short_stats.count,
            medium_term_count: medium_stats.count,
            long_term_count: long_stats.count,
            total_size_bytes: short_stats.size_bytes 
                + medium_stats.size_bytes 
                + long_stats.size_bytes,
        })
    }

    /// Get access statistics
    pub async fn get_access_stats(&self) -> MemoryAccessStats {
        self.stats.read().await.clone()
    }

    /// Consolidate short-term to medium-term
    pub async fn consolidate_short_term(&self, batch_size: usize) -> Result<usize> {
        let items = self.short_term.retrieve(MemoryQuery {
            query: String::new(),
            memory_type: Some(MemoryType::ShortTerm),
            limit: Some(batch_size),
            ..Default::default()
        }).await?;

        let mut count = 0;
        for item in items {
            // Check if item should be promoted
            if item.access_count > 2 || item.importance > 0.5 {
                if item.importance >= HIGH_IMPORTANCE_THRESHOLD {
                    self.promote_to_long_term(&item.id).await?;
                } else {
                    self.promote_to_medium_term(&item.id).await?;
                }
            } else {
                count += 1;
            }
        }

        Ok(count)
    }
}
