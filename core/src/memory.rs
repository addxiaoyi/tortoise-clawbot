//! 记忆系统模块
//! 
//! 实现三层记忆系统: Working Memory, Semantic Memory, Episodic Memory

use crate::error::{Error, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use uuid::Uuid;
use chrono::{DateTime, Utc};

mod error;
pub use error::Error;

/// 记忆类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MemoryType {
    /// 工作记忆 - 短期，保留当前会话信息
    Working,
    /// 语义记忆 - 长期，保留事实和概念
    Semantic,
    /// 情景记忆 - 经验，保留事件和经历
    Episodic,
}

/// 记忆项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Memory {
    pub id: String,
    pub memory_type: MemoryType,
    pub content: String,
    pub importance: f32,
    pub created_at: DateTime<Utc>,
    pub accessed_at: DateTime<Utc>,
    pub metadata: HashMap<String, String>,
    // 向量嵌入 (简化版，实际应使用实际的 embedding)
    pub embedding: Option<Vec<f32>>,
}

impl Memory {
    /// 创建新的记忆
    pub fn new(memory_type: MemoryType, content: impl Into<String>) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            memory_type,
            content: content.into(),
            importance: 0.5,
            created_at: now,
            accessed_at: now,
            metadata: HashMap::new(),
            embedding: None,
        }
    }
    
    /// 创建工作记忆
    pub fn working(content: impl Into<String>) -> Self {
        Self::new(MemoryType::Working, content)
    }
    
    /// 创建语义记忆
    pub fn semantic(content: impl Into<String>) -> Self {
        Self::new(MemoryType::Semantic, content)
    }
    
    /// 创建情景记忆
    pub fn episodic(content: impl Into<String>) -> Self {
        Self::new(MemoryType::Episodic, content)
    }
    
    /// 访问记忆
    pub fn access(&mut self) {
        self.accessed_at = Utc::now();
    }
    
    /// 设置重要性
    pub fn set_importance(&mut self, importance: f32) {
        self.importance = importance.clamp(0.0, 1.0);
    }
}

/// 记忆查询
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryQuery {
    pub query: String,
    pub memory_type: Option<MemoryType>,
    pub limit: usize,
    pub similarity_threshold: f32,
}

impl Default for MemoryQuery {
    fn default() -> Self {
        Self {
            query: String::new(),
            memory_type: None,
            limit: 10,
            similarity_threshold: 0.7,
        }
    }
}

/// 记忆查询结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryQueryResult {
    pub memories: Vec<Memory>,
    pub scores: Vec<f32>,
}

impl MemoryQueryResult {
    pub fn new(memories: Vec<Memory>, scores: Vec<f32>) -> Self {
        Self { memories, scores }
    }
    
    pub fn is_empty(&self) -> bool {
        self.memories.is_empty()
    }
    
    pub fn len(&self) -> usize {
        self.memories.len()
    }
}

/// 记忆存储
pub struct MemoryStore {
    working: Arc<RwLock<Vec<Memory>>>,
    semantic: Arc<RwLock<Vec<Memory>>>,
    episodic: Arc<RwLock<Vec<Memory>>>,
    max_working: usize,
    max_semantic: usize,
    max_episodic: usize,
}

impl MemoryStore {
    /// 创建新的记忆存储
    pub fn new(
        max_working: usize,
        max_semantic: usize,
        max_episodic: usize,
    ) -> Self {
        Self {
            working: Arc::new(RwLock::new(Vec::with_capacity(max_working))),
            semantic: Arc::new(RwLock::new(Vec::with_capacity(max_semantic))),
            episodic: Arc::new(RwLock::new(Vec::with_capacity(max_episodic))),
            max_working,
            max_semantic,
            max_episodic,
        }
    }
    
    /// 获取对应类型的存储
    fn get_store(&self, memory_type: MemoryType) -> Arc<RwLock<Vec<Memory>>> {
        match memory_type {
            MemoryType::Working => self.working.clone(),
            MemoryType::Semantic => self.semantic.clone(),
            MemoryType::Episodic => self.episodic.clone(),
        }
    }
    
    /// 获取最大容量
    fn get_max(&self, memory_type: MemoryType) -> usize {
        match memory_type {
            MemoryType::Working => self.max_working,
            MemoryType::Semantic => self.max_semantic,
            MemoryType::Episodic => self.max_episodic,
        }
    }
    
    /// 保存记忆
    pub fn save(&self, mut memory: Memory) -> String {
        let store = self.get_store(memory.memory_type);
        let max = self.get_max(memory.memory_type);
        let mut store = store.write();
        
        // 生成 ID
        let id = memory.id.clone();
        
        // 如果超出容量，移除最低重要性的记忆
        while store.len() >= max {
            if let Some(pos) = store.iter()
                .position(|m| m.importance <= store.iter()
                    .map(|x| x.importance)
                    .fold(f32::INFINITY, |a, b| a.min(b)))
            {
                store.remove(pos);
            } else {
                break;
            }
        }
        
        // 添加新记忆
        store.push(memory);
        id
    }
    
    /// 获取记忆
    pub fn get(&self, id: &str, memory_type: MemoryType) -> Option<Memory> {
        let store = self.get_store(memory_type);
        let store = store.read();
        
        let mut memory = store.iter().find(|m| m.id == id)?.clone();
        memory.access();
        Some(memory)
    }
    
    /// 删除记忆
    pub fn delete(&self, id: &str, memory_type: MemoryType) -> bool {
        let store = self.get_store(memory_type);
        let mut store = store.write();
        
        if let Some(pos) = store.iter().position(|m| m.id == id) {
            store.remove(pos);
            true
        } else {
            false
        }
    }
    
    /// 查询记忆
    pub fn query(&self, query: &MemoryQuery) -> MemoryQueryResult {
        let memory_type = query.memory_type.unwrap_or(MemoryType::Semantic);
        let store = self.get_store(memory_type);
        let store = store.read();
        
        // 简化版相似度计算（实际应使用向量相似度）
        let results: Vec<(Memory, f32)> = store.iter()
            .filter(|m| {
                query.memory_type.map_or(true, |t| m.memory_type == t)
            })
            .map(|m| {
                // 简化的相似度：基于重要性和最近访问时间
                let importance_score = m.importance;
                let recency_score = {
                    let duration = Utc::now().signed_duration_since(m.accessed_at);
                    let hours = duration.num_hours() as f32;
                    (1.0 / (1.0 + hours / 24.0)) * 0.5
                };
                let score = importance_score * 0.7 + recency_score * 0.3;
                (m.clone(), score)
            })
            .filter(|(_, score)| *score >= query.similarity_threshold)
            .take(query.limit)
            .collect();
        
        let memories: Vec<Memory> = results.iter().map(|(m, _)| m.clone()).collect();
        let scores: Vec<f32> = results.iter().map(|(_, s)| *s).collect();
        
        MemoryQueryResult::new(memories, scores)
    }
    
    /// 列出所有记忆
    pub fn list(&self, memory_type: Option<MemoryType>) -> Vec<Memory> {
        match memory_type {
            Some(t) => {
                let store = self.get_store(t);
                store.read().clone()
            }
            None => {
                let mut all = Vec::new();
                all.extend(self.working.read().clone());
                all.extend(self.semantic.read().clone());
                all.extend(self.episodic.read().clone());
                all
            }
        }
    }
    
    /// 清空指定类型的记忆
    pub fn clear(&self, memory_type: MemoryType) {
        let store = self.get_store(memory_type);
        store.write().clear();
    }
    
    /// 获取记忆数量
    pub fn count(&self, memory_type: Option<MemoryType>) -> usize {
        match memory_type {
            Some(t) => self.get_store(t).read().len(),
            None => {
                self.working.read().len() 
                + self.semantic.read().len() 
                + self.episodic.read().len()
            }
        }
    }
}

impl Default for MemoryStore {
    fn default() -> Self {
        Self::new(100, 10000, 5000)
    }
}

/// 记忆管理器
pub struct MemoryManager {
    store: Arc<MemoryStore>,
}

impl MemoryManager {
    /// 创建新的记忆管理器
    pub fn new() -> Self {
        Self {
            store: Arc::new(MemoryStore::default()),
        }
    }
    
    /// 创建带容量的记忆管理器
    pub fn with_capacity(
        max_working: usize,
        max_semantic: usize,
        max_episodic: usize,
    ) -> Self {
        Self {
            store: Arc::new(MemoryStore::new(max_working, max_semantic, max_episodic)),
        }
    }
    
    /// 保存工作记忆
    pub fn save_working(&self, content: impl Into<String>) -> String {
        let memory = Memory::working(content);
        self.store.save(memory)
    }
    
    /// 保存语义记忆
    pub fn save_semantic(&self, content: impl Into<String>) -> String {
        let memory = Memory::semantic(content);
        self.store.save(memory)
    }
    
    /// 保存情景记忆
    pub fn save_episodic(&self, content: impl Into<String>) -> String {
        let memory = Memory::episodic(content);
        self.store.save(memory)
    }
    
    /// 查询记忆
    pub fn query(&self, query: MemoryQuery) -> MemoryQueryResult {
        self.store.query(&query)
    }
    
    /// 获取记忆
    pub fn get(&self, id: &str, memory_type: MemoryType) -> Option<Memory> {
        self.store.get(id, memory_type)
    }
    
    /// 删除记忆
    pub fn delete(&self, id: &str, memory_type: MemoryType) -> bool {
        self.store.delete(id, memory_type)
    }
    
    /// 清空工作记忆
    pub fn clear_working(&self) {
        self.store.clear(MemoryType::Working);
    }
    
    /// 获取上下文记忆
    pub fn get_context(&self, limit: usize) -> Vec<Memory> {
        let mut working = self.store.list(Some(MemoryType::Working));
        working.extend(self.store.query(&MemoryQuery {
            memory_type: Some(MemoryType::Semantic),
            limit,
            ..Default::default()
        }).memories);
        working
    }
}

impl Default for MemoryManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_save_memory() {
        let manager = MemoryManager::new();
        let id = manager.save_working("Hello, world!");
        
        assert!(!id.is_empty());
    }
    
    #[test]
    fn test_query_memory() {
        let manager = MemoryManager::new();
        manager.save_semantic("The sky is blue");
        manager.save_semantic("The grass is green");
        
        let results = manager.query(MemoryQuery {
            query: "color".to_string(),
            memory_type: Some(MemoryType::Semantic),
            limit: 10,
            similarity_threshold: 0.0,
        });
        
        assert_eq!(results.len(), 2);
    }
    
    #[test]
    fn test_delete_memory() {
        let manager = MemoryManager::new();
        let id = manager.save_working("Test memory");
        
        assert!(manager.delete(&id, MemoryType::Working));
        assert!(manager.get(&id, MemoryType::Working).is_none());
    }
}
