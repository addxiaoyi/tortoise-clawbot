//! Memory Types
//! 
//! Core types for the memory system.

use serde::{Deserialize, Serialize};

/// Memory item
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryItem {
    /// Unique ID
    pub id: String,
    /// Content
    pub content: String,
    /// Memory tier
    pub memory_type: MemoryType,
    /// Importance score (0.0 - 1.0)
    pub importance: f32,
    /// Creation timestamp
    pub created_at: i64,
    /// Last access timestamp
    pub last_accessed: i64,
    /// Access count
    pub access_count: u32,
    /// Additional metadata
    pub metadata: serde_json::Value,
}

/// Memory tier
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MemoryType {
    /// Short-term memory (working memory)
    ShortTerm,
    /// Medium-term memory (recent memories)
    MediumTerm,
    /// Long-term memory (persistent, vectorized)
    LongTerm,
}

/// Memory query
#[derive(Debug, Clone, Default)]
pub struct MemoryQuery {
    /// Query string for semantic search
    pub query: String,
    /// Filter by memory type
    pub memory_type: Option<MemoryType>,
    /// Maximum results to return
    pub limit: Option<usize>,
    /// Minimum similarity threshold
    pub threshold: Option<f32>,
    /// Pre-computed embedding vector
    pub embedding: Option<Vec<f32>>,
}

impl MemoryQuery {
    /// Create a new query
    pub fn new(query: impl Into<String>) -> Self {
        Self {
            query: query.into(),
            ..Default::default()
        }
    }

    /// Set memory type filter
    pub fn with_type(mut self, memory_type: MemoryType) -> Self {
        self.memory_type = Some(memory_type);
        self
    }

    /// Set limit
    pub fn with_limit(mut self, limit: usize) -> Self {
        self.limit = Some(limit);
        self
    }

    /// Set threshold
    pub fn with_threshold(mut self, threshold: f32) -> Self {
        self.threshold = Some(threshold);
        self
    }
}

/// Memory statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStats {
    /// Short-term memory count
    pub short_term_count: usize,
    /// Medium-term memory count
    pub medium_term_count: usize,
    /// Long-term memory count
    pub long_term_count: usize,
    /// Total size in bytes
    pub total_size_bytes: u64,
}

impl Default for MemoryStats {
    fn default() -> Self {
        Self {
            short_term_count: 0,
            medium_term_count: 0,
            long_term_count: 0,
            total_size_bytes: 0,
        }
    }
}

/// Memory store trait
#[async_trait::async_trait]
pub trait MemoryStore: Send + Sync {
    /// Store a memory item
    async fn store(&self, item: MemoryItem) -> Result<String>;

    /// Retrieve memory items matching a query
    async fn retrieve(&self, query: MemoryQuery) -> Result<Option<Vec<MemoryItem>>>;

    /// Update a memory item
    async fn update(&self, id: &str, item: MemoryItem) -> Result<()>;

    /// Delete a memory item
    async fn delete(&self, id: &str) -> Result<()>;

    /// Get store statistics
    async fn stats(&self) -> Result<MemoryStoreStats>;
}

/// Memory store statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStoreStats {
    /// Item count
    pub count: usize,
    /// Size in bytes
    pub size_bytes: u64,
}

/// Vector store trait
#[async_trait::async_trait]
pub trait VectorStore: Send + Sync {
    /// Generate embedding for text
    async fn embed(&self, text: &str) -> Result<Vec<f32>>;

    /// Generate embeddings for multiple texts
    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>>;

    /// Search for similar vectors
    async fn search(&self, query: &[f32], limit: usize) -> Result<Vec<VectorSearchResult>>;
}

/// Vector search result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VectorSearchResult {
    /// Item ID
    pub id: String,
    /// Similarity score
    pub score: f32,
    /// Content
    pub content: String,
}

/// In-memory vector store implementation
pub struct InMemoryVectorStore {
    vectors: std::sync::RwLock<std::collections::HashMap<String, Vec<f32>>>,
    dimension: usize,
}

impl InMemoryVectorStore {
    /// Create a new in-memory vector store
    pub fn new(dimension: usize) -> Self {
        Self {
            vectors: std::sync::RwLock::new(std::collections::HashMap::new()),
            dimension,
        }
    }

    /// Add a vector
    pub fn add(&self, id: &str, vector: Vec<f32>) {
        let mut vectors = self.vectors.write().unwrap();
        vectors.insert(id.to_string(), vector);
    }

    /// Compute cosine similarity
    fn cosine_similarity(a: &[f32], b: &[f32]) -> f32 {
        if a.len() != b.len() {
            return 0.0;
        }

        let dot_product: f32 = a.iter().zip(b.iter()).map(|(x, y)| x * y).sum();
        let norm_a: f32 = a.iter().map(|x| x * x).sum::<f32>().sqrt();
        let norm_b: f32 = b.iter().map(|x| x * x).sum::<f32>().sqrt();

        if norm_a == 0.0 || norm_b == 0.0 {
            return 0.0;
        }

        dot_product / (norm_a * norm_b)
    }
}

#[async_trait::async_trait]
impl VectorStore for InMemoryVectorStore {
    async fn embed(&self, text: &str) -> Result<Vec<f32>> {
        // Simple embedding using hash-based approach
        // In production, use a real embedding model
        use sha2::{Sha256, Digest};
        
        let mut hasher = Sha256::new();
        hasher.update(text.as_bytes());
        let hash = hasher.finalize();
        
        let mut vector = vec![0.0f32; self.dimension];
        for (i, byte) in hash.iter().enumerate() {
            let idx = (i * 4) % self.dimension;
            if idx + 3 < self.dimension {
                let val = *byte as f32 / 255.0;
                vector[idx] = val;
                vector[idx + 1] = val * 0.5;
                vector[idx + 2] = val * 0.25;
                vector[idx + 3] = (1.0 - val) * 0.5;
            }
        }
        
        Ok(vector)
    }

    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>> {
        let mut results = Vec::with_capacity(texts.len());
        for text in texts {
            results.push(self.embed(text).await?);
        }
        Ok(results)
    }

    async fn search(&self, query: &[f32], limit: usize) -> Result<Vec<VectorSearchResult>> {
        let vectors = self.vectors.read().unwrap();
        
        let mut results: Vec<_> = vectors.iter()
            .map(|(id, vector)| {
                let score = Self::cosine_similarity(query, vector);
                (id.clone(), score)
            })
            .collect();

        results.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap());
        results.truncate(limit);

        Ok(results.into_iter()
            .map(|(id, score)| VectorSearchResult {
                id,
                score,
                content: String::new(),
            })
            .collect())
    }
}
