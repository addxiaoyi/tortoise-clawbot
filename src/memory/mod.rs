//! Memory Module
//! 
//! Three-tier memory system with short-term, medium-term, and long-term memory.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Memory configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryConfig {
    /// Short-term memory configuration
    pub short_term: TierConfig,
    /// Medium-term memory configuration
    pub medium_term: TierConfig,
    /// Long-term memory configuration
    pub long_term: TierConfig,
    /// Vector store configuration
    pub vector_store: VectorStoreConfig,
    /// Importance threshold for promotion to long-term
    pub promotion_threshold: f32,
    /// Auto-forget threshold
    pub forget_threshold: f32,
    /// Check interval in seconds
    pub check_interval_secs: u64,
}

impl Default for MemoryConfig {
    fn default() -> Self {
        Self {
            short_term: TierConfig {
                max_items: 100,
                max_size_mb: 50,
                ttl_seconds: 3600, // 1 hour
            },
            medium_term: TierConfig {
                max_items: 1000,
                max_size_mb: 200,
                ttl_seconds: 86400 * 7, // 1 week
            },
            long_term: TierConfig {
                max_items: 10000,
                max_size_mb: 1000,
                ttl_seconds: 0, // Never expire
            },
            vector_store: VectorStoreConfig::default(),
            promotion_threshold: 0.8,
            forget_threshold: 0.2,
            check_interval_secs: 300, // 5 minutes
        }
    }
}

/// Tier configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TierConfig {
    /// Maximum number of items
    pub max_items: usize,
    /// Maximum size in MB
    pub max_size_mb: u64,
    /// TTL in seconds (0 = no expiry)
    pub ttl_seconds: u64,
}

/// Vector store configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VectorStoreConfig {
    /// Provider type
    pub provider: VectorStoreProvider,
    /// Embedding dimension
    pub embedding_dim: usize,
    /// Index type
    pub index_type: IndexType,
}

impl Default for VectorStoreConfig {
    fn default() -> Self {
        Self {
            provider: VectorStoreProvider::InMemory,
            embedding_dim: 1536,
            index_type: IndexType::HNSW,
        }
    }
}

/// Vector store provider
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum VectorStoreProvider {
    /// In-memory storage (default)
    InMemory,
    /// SQLite with FTS
    SQLite,
    /// Redis
    Redis,
    /// Pinecone
    Pinecone,
    /// Weaviate
    Weaviate,
    /// Qdrant
    Qdrant,
}

/// Index type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum IndexType {
    /// Flat index
    Flat,
    /// HNSW index
    HNSW,
    /// IVF index
    IVF,
}

/// Memory type/tier
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MemoryType {
    /// Short-term memory (STM)
    ShortTerm,
    /// Medium-term memory (MTM)
    MediumTerm,
    /// Long-term memory (LTM)
    LongTerm,
}

/// Memory item
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryItem {
    /// Unique ID
    pub id: String,
    /// Content
    pub content: String,
    /// Memory type
    pub memory_type: MemoryType,
    /// Importance score (0.0 - 1.0)
    pub importance: f32,
    /// Creation timestamp
    pub created_at: i64,
    /// Last access timestamp
    pub last_accessed: i64,
    /// Access count
    pub access_count: u32,
    /// Metadata
    pub metadata: HashMap<String, serde_json::Value>,
    /// Embedding vector (if available)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub embedding: Option<Vec<f32>>,
}

impl MemoryItem {
    /// Create a new memory item
    pub fn new(content: String, memory_type: MemoryType, importance: f32) -> Self {
        let now = chrono::Utc::now().timestamp();
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            content,
            memory_type,
            importance,
            created_at: now,
            last_accessed: now,
            access_count: 0,
            metadata: HashMap::new(),
            embedding: None,
        }
    }

    /// Update access time and count
    pub fn access(&mut self) {
        self.last_accessed = chrono::Utc::now().timestamp();
        self.access_count += 1;
    }
}

/// Memory query
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryQuery {
    /// Query text
    pub query: String,
    /// Memory type filter
    pub memory_type: Option<MemoryType>,
    /// Maximum results
    pub limit: Option<usize>,
    /// Similarity threshold
    pub threshold: Option<f32>,
    /// Include embeddings
    pub include_embeddings: bool,
}

impl Default for MemoryQuery {
    fn default() -> Self {
        Self {
            query: String::new(),
            memory_type: None,
            limit: Some(10),
            threshold: Some(0.7),
            include_embeddings: false,
        }
    }
}

/// Memory statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStats {
    /// Short-term count
    pub short_term_count: usize,
    /// Medium-term count
    pub medium_term_count: usize,
    /// Long-term count
    pub long_term_count: usize,
    /// Total items
    pub total_count: usize,
    /// Total size in bytes
    pub total_size_bytes: u64,
}

/// Search result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchResult {
    /// Item ID
    pub id: String,
    /// Similarity score
    pub score: f32,
    /// Memory item
    pub item: MemoryItem,
}

/// Memory manager
pub struct MemoryManager {
    config: MemoryConfig,
    short_term: Arc<dyn MemoryStore>,
    medium_term: Arc<dyn MemoryStore>,
    long_term: Arc<dyn MemoryStore>,
    vector_store: Arc<dyn VectorStore>,
}

impl MemoryManager {
    /// Create a new memory manager
    pub async fn new(config: &MemoryConfig) -> Result<Self> {
        let short_term = InMemoryStore::new(&config.short_term)?;
        let medium_term = InMemoryStore::new(&config.medium_term)?;
        let long_term = InMemoryStore::new(&config.long_term)?;
        let vector_store = VectorStore::new(&config.vector_store).await?;

        Ok(Self {
            config: config.clone(),
            short_term,
            medium_term,
            long_term,
            vector_store,
        })
    }

    /// Remember new information
    pub async fn remember(&self, content: String, importance: f32) -> Result<String> {
        // Generate embedding
        let embedding = self.vector_store.embed(&content).await?;

        let mut item = MemoryItem::new(content, MemoryType::ShortTerm, importance);
        item.embedding = Some(embedding);

        // Store in short-term
        let id = self.short_term.store(item.clone()).await?;

        // If importance is high, promote to long-term
        if importance > self.config.promotion_threshold {
            self.promote_to_long_term(&id).await?;
        }

        Ok(id)
    }

    /// Recall relevant memories
    pub async fn recall(&self, query: &str) -> Result<Vec<MemoryItem>> {
        let query_embedding = self.vector_store.embed(query).await?;

        let mut results = Vec::new();
        let limit = 15;

        // Query all tiers
        let short_results = self.search_tier(&self.short_term, &query_embedding, limit / 3).await?;
        let medium_results = self.search_tier(&self.medium_term, &query_embedding, limit / 3).await?;
        let long_results = self.search_tier(&self.long_term, &query_embedding, limit / 3).await?;

        results.extend(short_results);
        results.extend(medium_results);
        results.extend(long_results);

        // Sort by importance and score
        results.sort_by(|a, b| {
            let score_a = a.item.importance * a.score;
            let score_b = b.item.importance * b.score;
            score_b.partial_cmp(&score_a).unwrap()
        });

        // Deduplicate and limit
        let mut seen = std::collections::HashSet::new();
        let items: Vec<_> = results
            .into_iter()
            .filter(|r| seen.insert(r.id.clone()))
            .map(|r| r.item)
            .take(limit)
            .collect();

        Ok(items)
    }

    /// Search in a specific tier
    async fn search_tier(
        &self,
        store: &Arc<dyn MemoryStore>,
        query_embedding: &[f32],
        limit: usize,
    ) -> Result<Vec<SearchResult>> {
        // For simplicity, return items without vector search
        // In production, would use actual similarity search
        let items = store.query(MemoryQuery {
            query: String::new(),
            memory_type: None,
            limit: Some(limit),
            threshold: None,
            include_embeddings: true,
        }).await?;

        Ok(items.into_iter().map(|item| SearchResult {
            id: item.id.clone(),
            score: 0.5, // Placeholder
            item,
        }).collect())
    }

    /// Promote memory to long-term
    pub async fn promote_to_long_term(&self, id: &str) -> Result<()> {
        // Find in medium-term or short-term
        if let Some(item) = self.medium_term.get(id).await? {
            self.long_term.store(item).await?;
            self.medium_term.delete(id).await?;
        } else if let Some(item) = self.short_term.get(id).await? {
            self.long_term.store(item).await?;
            self.short_term.delete(id).await?;
        }
        Ok(())
    }

    /// Forget low-importance memories
    pub async fn forget_low_importance(&self) -> Result<usize> {
        let threshold = self.config.forget_threshold;
        let mut count = 0;

        // Check medium-term
        let items = self.medium_term.query(MemoryQuery {
            query: String::new(),
            memory_type: Some(MemoryType::MediumTerm),
            limit: Some(1000),
            threshold: None,
            include_embeddings: false,
        }).await?;

        for item in items {
            if item.importance < threshold {
                self.medium_term.delete(&item.id).await?;
                count += 1;
            }
        }

        Ok(count)
    }

    /// Get memory statistics
    pub async fn stats(&self) -> Result<MemoryStats> {
        let short_count = self.short_term.count().await?;
        let medium_count = self.medium_term.count().await?;
        let long_count = self.long_term.count().await?;

        Ok(MemoryStats {
            short_term_count: short_count,
            medium_term_count: medium_count,
            long_term_count: long_count,
            total_count: short_count + medium_count + long_count,
            total_size_bytes: 0, // Would calculate from actual stored data
        })
    }

    /// Clear all memory
    pub async fn clear(&self) -> Result<()> {
        self.short_term.clear().await?;
        self.medium_term.clear().await?;
        self.long_term.clear().await?;
        Ok(())
    }
}

/// Memory store trait
#[async_trait::async_trait]
pub trait MemoryStore: Send + Sync {
    /// Store a memory item
    async fn store(&self, item: MemoryItem) -> Result<String>;

    /// Get a memory item by ID
    async fn get(&self, id: &str) -> Result<Option<MemoryItem>>;

    /// Query memories
    async fn query(&self, query: MemoryQuery) -> Result<Vec<MemoryItem>>;

    /// Update a memory item
    async fn update(&self, id: &str, item: MemoryItem) -> Result<()>;

    /// Delete a memory item
    async fn delete(&self, id: &str) -> Result<()>;

    /// Count items
    async fn count(&self) -> Result<usize>;

    /// Clear all items
    async fn clear(&self) -> Result<()>;
}

/// Vector store trait
#[async_trait::async_trait]
pub trait VectorStore: Send + Sync {
    /// Generate embedding for text
    async fn embed(&self, text: &str) -> Result<Vec<f32>>;

    /// Generate embeddings for batch
    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>>;

    /// Search for similar vectors
    async fn search(&self, query: &[f32], limit: usize) -> Result<Vec<SearchResult>>;
}

/// In-memory store implementation
pub struct InMemoryStore {
    items: RwLock<HashMap<String, MemoryItem>>,
    config: TierConfig,
}

impl InMemoryStore {
    pub fn new(config: &TierConfig) -> Result<Self> {
        Ok(Self {
            items: RwLock::new(HashMap::new()),
            config: config.clone(),
        })
    }
}

#[async_trait::async_trait]
impl MemoryStore for InMemoryStore {
    async fn store(&self, item: MemoryItem) -> Result<String> {
        let mut items = self.items.write().await;
        let id = item.id.clone();
        items.insert(id.clone(), item);
        
        // Enforce max items
        while items.len() > self.config.max_items {
            if let Some(oldest) = items.iter()
                .min_by_key(|(_, v)| v.created_at)
                .map(|(k, _)| k.clone())
            {
                items.remove(&oldest);
            } else {
                break;
            }
        }
        
        Ok(id)
    }

    async fn get(&self, id: &str) -> Result<Option<MemoryItem>> {
        let items = self.items.read().await;
        Ok(items.get(id).cloned())
    }

    async fn query(&self, query: MemoryQuery) -> Result<Vec<MemoryItem>> {
        let items = self.items.read().await;
        let mut results: Vec<_> = items.values()
            .filter(|item| {
                if let Some(mt) = query.memory_type {
                    if item.memory_type != mt {
                        return false;
                    }
                }
                true
            })
            .cloned()
            .collect();

        // Sort by importance and recency
        results.sort_by(|a, b| {
            let score_a = a.importance + (1.0 / (a.last_accessed as f32 + 1.0));
            let score_b = b.importance + (1.0 / (b.last_accessed as f32 + 1.0));
            score_b.partial_cmp(&score_a).unwrap()
        });

        if let Some(limit) = query.limit {
            results.truncate(limit);
        }

        Ok(results)
    }

    async fn update(&self, id: &str, item: MemoryItem) -> Result<()> {
        let mut items = self.items.write().await;
        if items.contains_key(id) {
            items.insert(id.to_string(), item);
        }
        Ok(())
    }

    async fn delete(&self, id: &str) -> Result<()> {
        let mut items = self.items.write().await;
        items.remove(id);
        Ok(())
    }

    async fn count(&self) -> Result<usize> {
        let items = self.items.read().await;
        Ok(items.len())
    }

    async fn clear(&self) -> Result<()> {
        let mut items = self.items.write().await;
        items.clear();
        Ok(())
    }
}

/// Simple vector store implementation
pub struct VectorStore {
    config: VectorStoreConfig,
}

impl VectorStore {
    pub fn new(config: &VectorStoreConfig) -> Self {
        Self {
            config: config.clone(),
        }
    }
}

#[async_trait::async_trait]
impl VectorStore for VectorStore {
    async fn embed(&self, text: &str) -> Result<Vec<f32>> {
        // Simplified embedding - in production would call actual embedding API
        let dim = self.config.embedding_dim;
        let mut embedding = vec![0.0; dim];
        
        // Simple hash-based "embedding"
        for (i, byte) in text.bytes().enumerate() {
            embedding[i % dim] += (byte as f32) / 255.0;
        }
        
        // Normalize
        let sum: f32 = embedding.iter().map(|x| x * x).sum::<f32>().sqrt();
        if sum > 0.0 {
            for v in &mut embedding {
                *v /= sum;
            }
        }
        
        Ok(embedding)
    }

    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>> {
        let mut results = Vec::new();
        for text in texts {
            results.push(self.embed(text).await?);
        }
        Ok(results)
    }

    async fn search(&self, query: &[f32], limit: usize) -> Result<Vec<SearchResult>> {
        // Placeholder - would perform actual vector search
        Ok(vec![])
    }
}
