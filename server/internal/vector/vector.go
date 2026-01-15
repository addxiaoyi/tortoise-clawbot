package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// ============ VectorStore 向量存储接口 ============

// VectorStore 向量存储接口
type VectorStore interface {
	// 存储向量
	Upsert(id string, vector []float32, metadata map[string]interface{}) error
	
	// 搜索最近邻
	Search(query []float32, topK int, filters map[string]interface{}) ([]SearchResult, error)
	
	// 删除
	Delete(id string) error
	
	// 批量删除
	DeleteMany(ids []string) error
	
	// 获取
	Get(id string) (*VectorResult, error)
	
	// 获取统计
	Stats() (*StoreStats, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       string
	Score    float32
	Metadata map[string]interface{}
}

// VectorResult 向量结果
type VectorResult struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
}

// StoreStats 存储统计
type StoreStats struct {
	TotalVectors int
	Dimension   int
}

// ============ MemoryVectorStore 内存向量存储 ============

// MemoryVectorStore 内存实现的向量存储
type MemoryVectorStore struct {
	mu       sync.RWMutex
	vectors  map[string]*VectorEntry
	dimension int
}

// VectorEntry 向量条目
type VectorEntry struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
}

// NewMemoryVectorStore 创建内存向量存储
func NewMemoryVectorStore(dimension int) *MemoryVectorStore {
	return &MemoryVectorStore{
		vectors:  make(map[string]*VectorEntry),
		dimension: dimension,
	}
}

// Upsert 存储向量
func (s *MemoryVectorStore) Upsert(id string, vector []float32, metadata map[string]interface{}) error {
	if len(vector) != s.dimension {
		return fmt.Errorf("vector dimension mismatch: got %d, expected %d", len(vector), s.dimension)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[id] = &VectorEntry{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}

	return nil
}

// Search 搜索最近邻
func (s *MemoryVectorStore) Search(query []float32, topK int, filters map[string]interface{}) ([]SearchResult, error) {
	if len(query) != s.dimension {
		return nil, fmt.Errorf("query dimension mismatch: got %d, expected %d", len(query), s.dimension)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]SearchResult, 0, len(s.vectors))

	for _, entry := range s.vectors {
		// 应用过滤器
		if filters != nil && !s.matchFilters(entry.Metadata, filters) {
			continue
		}

		// 计算余弦相似度
		score := cosineSimilarity(query, entry.Vector)

		results = append(results, SearchResult{
			ID:       entry.ID,
			Score:    score,
			Metadata: entry.Metadata,
		})
	}

	// 按相似度排序
	quickSort(results, 0, len(results)-1)

	// 返回 topK
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK], nil
}

// matchFilters 检查是否匹配过滤器
func (s *MemoryVectorStore) matchFilters(metadata, filters map[string]interface{}) bool {
	for key, value := range filters {
		if metaValue, ok := metadata[key]; ok {
			if metaValue != value {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Delete 删除
func (s *MemoryVectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.vectors, id)
	return nil
}

// DeleteMany 批量删除
func (s *MemoryVectorStore) DeleteMany(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.vectors, id)
	}
	return nil
}

// Get 获取
func (s *MemoryVectorStore) Get(id string) (*VectorResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.vectors[id]
	if !ok {
		return nil, fmt.Errorf("vector not found: %s", id)
	}

	return &VectorResult{
		ID:       entry.ID,
		Vector:   entry.Vector,
		Metadata: entry.Metadata,
	}, nil
}

// Stats 获取统计
func (s *MemoryVectorStore) Stats() (*StoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &StoreStats{
		TotalVectors: len(s.vectors),
		Dimension:   s.dimension,
	}, nil
}

// ============ Embedding 生成器 ============

// Embedder 向量嵌入生成器
type Embedder interface {
	// 生成嵌入向量
	Embed(ctx context.Context, text string) ([]float32, error)
	
	// 批量生成嵌入向量
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	
	// 获取向量维度
	Dimension() int
}

// SimpleEmbedder 简单嵌入实现 (用于演示，实际应使用 OpenAI/Cohere 等)
type SimpleEmbedder struct {
	dimension int
}

// NewSimpleEmbedder 创建简单嵌入器
func NewSimpleEmbedder(dimension int) *SimpleEmbedder {
	return &SimpleEmbedder{dimension: dimension}
}

// Embed 生成嵌入
func (e *SimpleEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// 简化的词袋嵌入实现
	// 实际应使用 TF-IDF 或预训练模型
	vector := make([]float32, e.dimension)
	
	words := strings.Fields(strings.ToLower(text))
	hash := 0
	for i, word := range words {
		hash = 0
		for _, c := range word {
			hash = hash*31 + int(c)
		}
		
		idx := hash % e.dimension
		vector[idx] += 1.0
		
		// 分散权重
		if i > 0 && i < len(words)-1 {
			idx2 := (hash / 100) % e.dimension
			vector[idx2] += 0.5
		}
	}
	
	// 归一化
	normalize(vector)
	
	return vector, nil
}

// EmbedBatch 批量生成嵌入
func (e *SimpleEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := e.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		vectors[i] = vec
	}
	return vectors, nil
}

// Dimension 获取维度
func (e *SimpleEmbedder) Dimension() int {
	return e.dimension
}

// ============ SemanticMemory 语义记忆 ============

// SemanticMemory 语义记忆存储
type SemanticMemory struct {
	store   VectorStore
	embedder Embedder
	mu      sync.RWMutex
}

// NewSemanticMemory 创建语义记忆
func NewSemanticMemory(embedder Embedder) *SemanticMemory {
	return &SemanticMemory{
		store:   NewMemoryVectorStore(embedder.Dimension()),
		embedder: embedder,
	}
}

// Add 添加记忆
func (m *SemanticMemory) Add(ctx context.Context, content string, metadata map[string]interface{}) (string, error) {
	vector, err := m.embedder.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("failed to embed: %w", err)
	}

	id := uuid.New().String()
	
	// 添加内容到 metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["content"] = content

	if err := m.store.Upsert(id, vector, metadata); err != nil {
		return "", fmt.Errorf("failed to store: %w", err)
	}

	return id, nil
}

// Search 搜索记忆
func (m *SemanticMemory) Search(ctx context.Context, query string, topK int, filters map[string]interface{}) ([]SearchResult, error) {
	vector, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	return m.store.Search(vector, topK, filters)
}

// Delete 删除记忆
func (m *SemanticMemory) Delete(id string) error {
	return m.store.Delete(id)
}

// Stats 获取统计
func (m *SemanticMemory) Stats() (*StoreStats, error) {
	return m.store.Stats()
}

// ============ Helper Functions ============

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// normalize 归一化向量
func normalize(v []float32) {
	var norm float64
	for _, f := range v {
		norm += float64(f) * float64(f)
	}
	norm = math.Sqrt(norm)
	
	if norm > 0 {
		for i := range v {
			v[i] = float32(float64(v[i]) / norm)
		}
	}
}

// quickSort 快速排序
func quickSort(results []SearchResult, left, right int) {
	if left >= right {
		return
	}

	pivot := results[(left+right)/2].Score
	i := left
	j := right

	for i < j {
		for results[i].Score > pivot {
			i++
		}
		for results[j].Score < pivot {
			j--
		}
		if i <= j {
			results[i], results[j] = results[j], results[i]
			i++
			j--
		}
	}

	quickSort(results, left, j)
	quickSort(results, i, right)
}

// ============ JSON 序列化支持 ============

// MarshalJSON 序列化
func (s *SearchResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":       s.ID,
		"score":    s.Score,
		"metadata": s.Metadata,
	})
}

// MarshalJSON 序列化
func (s *StoreStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"total_vectors": s.TotalVectors,
		"dimension":    s.Dimension,
	})
}
