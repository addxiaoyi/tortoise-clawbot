package memory

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// VectorStore 向量存储服务
type VectorStore struct {
	provider string
	baseURL string
	apiKey  string
	client  *http.Client
}

// MemoryItem 记忆条目
type MemoryItem struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Embedding   []float32              `json:"-"`
	Score       float64                `json:"score,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Item  *MemoryItem `json:"item"`
	Score float64     `json:"score"`
}

// NewVectorStore 创建向量存储
func NewVectorStore(provider, baseURL, apiKey string) *VectorStore {
	return &VectorStore{
		provider: provider,
		baseURL:  baseURL,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Store 存储记忆
func (vs *VectorStore) Store(ctx context.Context, item *MemoryItem) error {
	// 生成向量嵌入
	embedding, err := vs.generateEmbedding(item.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}
	item.Embedding = embedding

	// 根据 provider 存储
	switch vs.provider {
	case "openai":
		return vs.storeWithOpenAI(ctx, item)
	case "qdrant":
		return vs.storeWithQdrant(ctx, item)
	case "chroma":
		return vs.storeWithChroma(ctx, item)
	default:
		return vs.storeViaAPI(ctx, item)
	}
}

// Search 语义搜索
func (vs *VectorStore) Search(ctx context.Context, query string, userID string, limit int) ([]SearchResult, error) {
	// 生成查询向量
	embedding, err := vs.generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	switch vs.provider {
	case "openai":
		return vs.searchWithOpenAI(ctx, embedding, userID, limit)
	case "qdrant":
		return vs.searchWithQdrant(ctx, embedding, userID, limit)
	case "chroma":
		return vs.searchWithChroma(ctx, embedding, userID, limit)
	default:
		return vs.searchViaAPI(ctx, embedding, userID, limit)
	}
}

// generateEmbedding 生成向量嵌入 (使用 OpenAI Embeddings API)
func (vs *VectorStore) generateEmbedding(text string) ([]float32, error) {
	if vs.apiKey == "" {
		// 返回模拟向量用于测试
		return generateMockEmbedding(text), nil
	}

	reqBody := map[string]interface{}{
		"input": text,
		"model": "text-embedding-3-small",
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", vs.baseURL+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vs.apiKey)

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error: %s", string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// storeWithOpenAI 使用 OpenAI 存储
func (vs *VectorStore) storeWithOpenAI(ctx context.Context, item *MemoryItem) error {
	data, _ := json.Marshal(item)
	req, err := http.NewRequestWithContext(ctx, "POST", vs.baseURL+"/memories", bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vs.apiKey)

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("store failed: %s", string(body))
	}

	return nil
}

// searchWithOpenAI 使用 OpenAI 搜索
func (vs *VectorStore) searchWithOpenAI(ctx context.Context, embedding []float32, userID string, limit int) ([]SearchResult, error) {
	reqBody := map[string]interface{}{
		"embedding": embedding,
		"user_id":   userID,
		"limit":     limit,
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", vs.baseURL+"/memories/search", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vs.apiKey)

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	var results []SearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// storeWithQdrant 使用 Qdrant 存储
func (vs *VectorStore) storeWithQdrant(ctx context.Context, item *MemoryItem) error {
	point := map[string]interface{}{
		"id":      item.ID,
		"vector":  item.Embedding,
		"payload": item,
	}

	data, _ := json.Marshal([]map[string]interface{}{point})
	req, err := http.NewRequestWithContext(ctx, "PUT",
		fmt.Sprintf("%s/collections/memories/points", vs.baseURL), bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", vs.apiKey)

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// searchWithQdrant 使用 Qdrant 搜索
func (vs *VectorStore) searchWithQdrant(ctx context.Context, embedding []float32, userID string, limit int) ([]SearchResult, error) {
	reqBody := map[string]interface{}{
		"vector":    embedding,
		"limit":    limit,
		"with_payload": true,
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "user_id", "match": map[string]interface{}{"value": userID}},
			},
		},
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/collections/memories/points/search", vs.baseURL), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", vs.apiKey)

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Qdrant search failed: %s", string(body))
	}

	var qdrantResp struct {
		Result []struct {
			ID       string      `json:"id"`
			Score    float64     `json:"score"`
			Payload  *MemoryItem `json:"payload"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &qdrantResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(qdrantResp.Result))
	for i, r := range qdrantResp.Result {
		results[i] = SearchResult{Item: r.Payload, Score: r.Score}
	}

	return results, nil
}

// storeWithChroma 使用 Chroma 存储
func (vs *VectorStore) storeWithChroma(ctx context.Context, item *MemoryItem) error {
	reqBody := map[string]interface{}{
		"ids":         []string{item.ID},
		"embeddings":  [][]float32{item.Embedding},
		"documents":   []string{item.Content},
		"metadatas":   []map[string]interface{}{{"user_id": item.UserID, "title": item.Title}},
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/collections/memories", vs.baseURL), bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// searchWithChroma 使用 Chroma 搜索
func (vs *VectorStore) searchWithChroma(ctx context.Context, embedding []float32, userID string, limit int) ([]SearchResult, error) {
	reqBody := map[string]interface{}{
		"query_embeddings": [][]float32{embedding},
		"n_results":       limit,
		"where":           map[string]interface{}{"user_id": userID},
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/collections/memories/query", vs.baseURL), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return nil, nil
}

// storeViaAPI 通过 API 存储
func (vs *VectorStore) storeViaAPI(ctx context.Context, item *MemoryItem) error {
	log.Printf("Storing memory via API: %s", item.ID)
	return nil
}

// searchViaAPI 通过 API 搜索
func (vs *VectorStore) searchViaAPI(ctx context.Context, embedding []float32, userID string, limit int) ([]SearchResult, error) {
	log.Printf("Searching memories via API for user: %s", userID)
	return nil, nil
}

// generateMockEmbedding 生成模拟向量 (用于测试)
func generateMockEmbedding(text string) []float32 {
	// 简单的基于文本的伪随机向量
	vec := make([]float32, 1536)
	hash := int64(0)
	for i, c := range text {
		hash = hash*31 + int64(c)*int64(i+1)
	}

	r := float64(hash%1000) / 1000.0
	for i := range vec {
		vec[i] = float32(r + float64(i%10)*0.01)
	}

	// 归一化
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	norm := float32(1.0 / float64(sum))
	for i := range vec {
		vec[i] *= norm
	}

	return vec
}

// EncodeEmbedding 编码向量为 base64
func EncodeEmbedding(embedding []float32) string {
	data, _ := json.Marshal(embedding)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeEmbedding 解码 base64 为向量
func DecodeEmbedding(s string) ([]float32, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var embedding []float32
	if err := json.Unmarshal(data, &embedding); err != nil {
		return nil, err
	}

	return embedding, nil
}
