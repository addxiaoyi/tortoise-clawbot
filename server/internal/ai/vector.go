package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
)

// VectorStore 向量存储接口
type VectorStore interface {
	// 添加向量
	Add(ctx context.Context, id string, text string, vector []float32) error
	// 搜索相似向量
	Search(ctx context.Context, vector []float32, topK int) ([]SearchResult, error)
	// 删除向量
	Delete(ctx context.Context, id string) error
	// 获取向量
	Get(ctx context.Context, id string) (string, []float32, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	ID     string  `json:"id"`
	Text   string  `json:"text"`
	Score  float32 `json:"score"` // 相似度分数
}

// ============ 简单内存向量存储 ============

// SimpleVectorStore 简单内存向量存储
type SimpleVectorStore struct {
	vectors map[string]*VectorEntry
	mu      sync.RWMutex
}

type VectorEntry struct {
	ID      string
	Text    string
	Vector  []float32
}

// NewSimpleVectorStore 创建简单向量存储
func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		vectors: make(map[string]*VectorEntry),
	}
}

// Add 添加向量
func (s *SimpleVectorStore) Add(ctx context.Context, id string, text string, vector []float32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[id] = &VectorEntry{
		ID:     id,
		Text:   text,
		Vector: vector,
	}
	return nil
}

// Search 搜索相似向量
func (s *SimpleVectorStore) Search(ctx context.Context, vector []float32, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]SearchResult, 0)

	for _, entry := range s.vectors {
		score := cosineSimilarity(vector, entry.Vector)
		results = append(results, SearchResult{
			ID:    entry.ID,
			Text:  entry.Text,
			Score: score,
		})
	}

	// 按相似度排序
	sortResults(results)

	// 返回 topK
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Delete 删除向量
func (s *SimpleVectorStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vectors, id)
	return nil
}

// Get 获取向量
func (s *SimpleVectorStore) Get(ctx context.Context, id string) (string, []float32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.vectors[id]
	if !ok {
		return "", nil, fmt.Errorf("vector not found: %s", id)
	}
	return entry.Text, entry.Vector, nil
}

// ============ 向量化工具 ============

// TextToVector 将文本转换为向量 (简化版，使用词频)
func TextToVector(text string, dimensions int) []float32 {
	// 简单实现：基于词频的向量
	words := strings.Fields(strings.ToLower(text))
	vector := make([]float32, dimensions)

	// 简单的词频统计
	for i, word := range words {
		hash := simpleHash(word)
		idx := hash % dimensions
		vector[idx] += 1.0
		if i > 0 {
			// 添加一些 n-gram 特征
			prevIdx := simpleHash(words[i-1]) % dimensions
			vector[(idx+prevIdx)%dimensions] += 0.5
		}
	}

	// 归一化
	normalize(vector)
	return vector
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// normalize 归一化向量
func normalize(v []float32) {
	var norm float32
	for _, f := range v {
		norm += f * f
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
}

// simpleHash 简单哈希
func simpleHash(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// sortResults 按分数排序
func sortResults(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// ============ Semantic Memory ============

// SemanticMemory 语义记忆
type SemanticMemory struct {
	store      VectorStore
	dimensions int
}

// NewSemanticMemory 创建语义记忆
func NewSemanticMemory(dimensions int) *SemanticMemory {
	return &SemanticMemory{
		store:      NewSimpleVectorStore(),
		dimensions: dimensions,
	}
}

// Store 存储记忆
func (m *SemanticMemory) Store(id, text string) error {
	vector := TextToVector(text, m.dimensions)
	return m.store.Add(context.Background(), id, text, vector)
}

// Search 搜索记忆
func (m *SemanticMemory) Search(query string, topK int) ([]SearchResult, error) {
	vector := TextToVector(query, m.dimensions)
	return m.store.Search(context.Background(), vector, topK)
}

// Delete 删除记忆
func (m *SemanticMemory) Delete(id string) error {
	return m.store.Delete(context.Background(), id)
}

// ============ Pinecone 集成 (可选) ============

// PineconeConfig Pinecone 配置
type PineconeConfig struct {
	APIKey   string
	Environment string
	Index     string
}

// PineconeStore Pinecone 向量存储
type PineconeStore struct {
	apiKey     string
	environment string
	index      string
	baseURL    string
}

// NewPineconeStore 创建 Pinecone 存储
func NewPineconeStore(config *PineconeConfig) (*PineconeStore, error) {
	return &PineconeStore{
		apiKey:     config.APIKey,
		environment: config.Environment,
		index:      config.Index,
		baseURL:    fmt.Sprintf("https://%s-%s.svc.%s.pinecone.io", 
			config.Index, "xxx", config.Environment),
	}, nil
}

// Add 添加向量到 Pinecone
func (s *PineconeStore) Add(ctx context.Context, id string, text string, vector []float32) error {
	// 简化实现，实际需要调用 Pinecone API
	type upsertRequest struct {
		Vectors []struct {
			ID     string    `json:"id"`
			Values []float32 `json:"values"`
			Metadata map[string]string `json:"metadata"`
		} `json:"vectors"`
	}

	req := upsertRequest{
		Vectors: []struct {
			ID     string    `json:"id"`
			Values []float32 `json:"values"`
			Metadata map[string]string `json:"metadata"`
		}{
			{
				ID:     id,
				Values: vector,
				Metadata: map[string]string{"text": text},
			},
		},
	}

	data, _ := json.Marshal(req)
	// 实际需要发送 HTTP 请求到 Pinecone
	_ = data
	return nil
}

// Search 搜索
func (s *PineconeStore) Search(ctx context.Context, vector []float32, topK int) ([]SearchResult, error) {
	// 简化实现
	return nil, fmt.Errorf("Pinecone search not implemented")
}

// Delete 删除
func (s *PineconeStore) Delete(ctx context.Context, id string) error {
	return nil
}

// Get 获取
func (s *PineconeStore) Get(ctx context.Context, id string) (string, []float32, error) {
	return "", nil, fmt.Errorf("Pinecone get not implemented")
}
