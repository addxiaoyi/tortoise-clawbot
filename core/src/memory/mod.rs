//! Memory system module - Three-tier memory architecture

use std::sync::Arc;
use tokio::sync::RwLock;
use dashmap::DashMap;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::BinaryHeap;
use std::cmp::Ordering;

/// Memory type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum MemoryType {
    Working,    // Short-term memory
    Semantic,   // Long-term with semantic search
    Episodic,   // Experience-based memory
}

/// Memory entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Memory {
    pub id: String,
    pub content: String,
    pub memory_type: MemoryType,
    pub session_id: Option<String>,
    pub user_id: Option<String>,
    pub importance: f32,
    pub tags: Vec<String>,
    pub created_at: DateTime<Utc>,
    pub accessed_at: DateTime<Utc>,
    pub access_count: usize,
    pub vector: Option<Vec<f32>>,
    pub metadata: serde_json::Value,
}

impl Memory {
    pub fn new(content: String, memory_type: MemoryType) -> Self {
        let now = Utc::now();
        Self {
            id: format!("mem_{}", Uuid::new_v4()),
            content,
            memory_type,
            session_id: None,
            user_id: None,
            importance: 0.5,
            tags: Vec::new(),
            created_at: now,
            accessed_at: now,
            access_count: 0,
            vector: None,
            metadata: serde_json::Value::Null,
        }
    }

    pub fn with_session(mut self, session_id: String) -> Self {
        self.session_id = Some(session_id);
        self
    }

    pub fn with_user(mut self, user_id: String) -> Self {
        self.user_id = Some(user_id);
        self
    }

    pub fn with_importance(mut self, importance: f32) -> Self {
        self.importance = importance;
        self
    }

    pub fn with_tags(mut self, tags: Vec<String>) -> Self {
        self.tags = tags;
        self
    }

    pub fn touch(&mut self) {
        self.accessed_at = Utc::now();
        self.access_count += 1;
    }
}

/// Memory store
pub struct MemoryStore {
    working: Arc<DashMap<String, Memory>>,
    semantic: Arc<DashMap<String, Memory>>,
    episodic: Arc<DashMap<String, Memory>>,
    working_capacity: usize,
    vector_dimension: usize,
}

impl MemoryStore {
    /// Create a new memory store
    pub async fn new(data_dir: &str) -> Result<Self, super::Error> {
        Ok(Self {
            working: Arc::new(DashMap::new()),
            semantic: Arc::new(DashMap::new()),
            episodic: Arc::new(DashMap::new()),
            working_capacity: 100,
            vector_dimension: 1536,
        })
    }

    /// Store a memory entry
    pub fn store(&self, memory: Memory) {
        match memory.memory_type {
            MemoryType::Working => {
                self.working.insert(memory.id.clone(), memory);
                self.trim_working();
            }
            MemoryType::Semantic => {
                self.semantic.insert(memory.id.clone(), memory);
            }
            MemoryType::Episodic => {
                self.episodic.insert(memory.id.clone(), memory);
            }
        }
    }

    /// Retrieve a memory by ID
    pub fn get(&self, id: &str) -> Option<Memory> {
        if let Some(m) = self.working.get(id) {
            return Some(m.clone());
        }
        if let Some(m) = self.semantic.get(id) {
            return Some(m.clone());
        }
        if let Some(m) = self.episodic.get(id) {
            return Some(m.clone());
        }
        None
    }

    /// Search memories by content (simple keyword match)
    pub fn search(&self, query: &str, limit: usize) -> Vec<Memory> {
        let query_lower = query.to_lowercase();
        let mut results: Vec<Memory> = Vec::new();

        // Search in semantic memory (main search target)
        for entry in self.semantic.iter() {
            if entry.content.to_lowercase().contains(&query_lower) {
                results.push(entry.clone());
            }
        }

        // Also search in episodic
        for entry in self.episodic.iter() {
            if entry.content.to_lowercase().contains(&query_lower) {
                results.push(entry.clone());
            }
        }

        // Sort by relevance (importance * recency)
        results.sort_by(|a, b| {
            let score_a = a.importance * recency_score(&a.accessed_at);
            let score_b = b.importance * recency_score(&b.accessed_at);
            score_b.partial_cmp(&score_a).unwrap_or(Ordering::Equal)
        });

        results.truncate(limit);
        results
    }

    /// Search by vector similarity (semantic search)
    pub fn search_vector(&self, query: &[f32], limit: usize) -> Vec<(Memory, f32)> {
        let mut results: Vec<(Memory, f32)> = Vec::new();

        for entry in self.semantic.iter() {
            if let Some(vector) = &entry.vector {
                let similarity = cosine_similarity(query, vector);
                results.push((entry.clone(), similarity));
            }
        }

        // Sort by similarity
        results.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap_or(Ordering::Equal));
        results.truncate(limit);
        results
    }

    /// Get memories for a session
    pub fn get_session_memories(&self, session_id: &str) -> Vec<Memory> {
        let mut results = Vec::new();

        for entry in self.working.iter() {
            if entry.session_id.as_deref() == Some(session_id) {
                results.push(entry.clone());
            }
        }

        for entry in self.semantic.iter() {
            if entry.session_id.as_deref() == Some(session_id) {
                results.push(entry.clone());
            }
        }

        results
    }

    /// Get memories for a user
    pub fn get_user_memories(&self, user_id: &str) -> Vec<Memory> {
        let mut results = Vec::new();

        for entry in self.semantic.iter() {
            if entry.user_id.as_deref() == Some(user_id) {
                results.push(entry.clone());
            }
        }

        for entry in self.episodic.iter() {
            if entry.user_id.as_deref() == Some(user_id) {
                results.push(entry.clone());
            }
        }

        results
    }

    /// Delete a memory
    pub fn delete(&self, id: &str) -> bool {
        if self.working.remove(id).is_some() {
            return true;
        }
        if self.semantic.remove(id).is_some() {
            return true;
        }
        if self.episodic.remove(id).is_some() {
            return true;
        }
        false
    }

    /// Clear a session's memories
    pub fn clear_session(&self, session_id: &str) {
        self.working.retain(|_, m| m.session_id.as_deref() != Some(session_id));
        self.semantic.retain(|_, m| m.session_id.as_deref() != Some(session_id));
    }

    /// Trim working memory to capacity
    fn trim_working(&self) {
        while self.working.len() > self.working_capacity {
            if let Some(oldest) = self.find_oldest_working() {
                self.working.remove(&oldest);
            } else {
                break;
            }
        }
    }

    /// Find oldest entry in working memory
    fn find_oldest_working(&self) -> Option<String> {
        let mut oldest: Option<String> = None;
        let mut oldest_time: Option<DateTime<Utc>> = None;

        for entry in self.working.iter() {
            if let Some(time) = oldest_time {
                if entry.accessed_at < time {
                    oldest_time = Some(entry.accessed_at);
                    oldest = Some(entry.id.clone());
                }
            } else {
                oldest_time = Some(entry.accessed_at);
                oldest = Some(entry.id.clone());
            }
        }

        oldest
    }

    /// Consolidate working memory to semantic
    pub fn consolidate(&self) {
        for entry in self.working.iter() {
            if entry.access_count > 3 || entry.importance > 0.7 {
                // Important or frequently accessed - move to semantic
                let mut memory = entry.clone();
                memory.memory_type = MemoryType::Semantic;
                self.semantic.insert(memory.id.clone(), memory);
            }
        }
        self.working.clear();
    }

    /// Get memory statistics
    pub fn stats(&self) -> MemoryStats {
        MemoryStats {
            working_count: self.working.len(),
            semantic_count: self.semantic.len(),
            episodic_count: self.episodic.len(),
            working_capacity: self.working_capacity,
        }
    }
}

/// Memory statistics
#[derive(Debug, Clone)]
pub struct MemoryStats {
    pub working_count: usize,
    pub semantic_count: usize,
    pub episodic_count: usize,
    pub working_capacity: usize,
}

/// Calculate recency score (more recent = higher score)
fn recency_score(time: &DateTime<Utc>) -> f32 {
    let age_secs = (Utc::now() - *time).num_seconds() as f32;
    let decay = (-age_secs / (7 * 24 * 3600) as f32).exp();
    0.5 + 0.5 * decay
}

/// Cosine similarity between two vectors
fn cosine_similarity(a: &[f32], b: &[f32]) -> f32 {
    if a.len() != b.len() {
        return 0.0;
    }

    let dot_product: f32 = a.iter().zip(b.iter()).map(|(x, y)| x * y).sum();
    let magnitude_a: f32 = a.iter().map(|x| x * x).sum::<f32>().sqrt();
    let magnitude_b: f32 = b.iter().map(|x| x * x).sum::<f32>().sqrt();

    if magnitude_a == 0.0 || magnitude_b == 0.0 {
        return 0.0;
    }

    dot_product / (magnitude_a * magnitude_b)
}
