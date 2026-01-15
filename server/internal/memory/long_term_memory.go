package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============ Long-Term Memory System ============

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"   //情景记忆
	MemoryTypeSemantic   MemoryType = "semantic"    //语义记忆
	MemoryTypeProcedural MemoryType = "procedural"  //程序性记忆
	MemoryTypeEmotional  MemoryType = "emotional"   //情感记忆
	MemoryTypeWorking    MemoryType = "working"     //工作记忆
)

// MemoryEntry 记忆条目
type MemoryEntry struct {
	ID          string                 `json:"id"`
	Type       MemoryType            `json:"type"`
	Content    string                 `json:"content"`
	Embedding  []float64             `json:"embedding,omitempty"`
	Importance float64               `json:"importance"` // 重要性 0-1
	Emotion    string                 `json:"emotion,omitempty"` //情感标签
	Keywords   []string               `json:"keywords,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time             `json:"created_at"`
	AccessedAt time.Time             `json:"accessed_at"`
	AccessCount int                  `json:"access_count"`
	TTL        time.Duration          `json:"ttl,omitempty"` // 生存时间
	IsPinned   bool                  `json:"is_pinned"`     // 是否固定
	ExpiresAt  *time.Time            `json:"expires_at,omitempty"`
}

// MemoryQuery 记忆查询
type MemoryQuery struct {
	Query       string
	Type        MemoryType
	Limit       int
	Offset      int
	Since       *time.Time
	Until       *time.Time
	MinImportance float64
	Keywords    []string
	Emotion     string
}

// MemoryRecallResult 记忆召回结果
type MemoryRecallResult struct {
	Memory *MemoryEntry `json:"memory"`
	Score  float64      `json:"score"` // 相关性分数
}

// LongTermMemory 长期记忆系统
type LongTermMemory struct {
	storage      map[string]*MemoryEntry
	index        *MemoryIndex
	forgetting   *ForgettingMechanism
	config       *MemoryConfig
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	MaxMemories       int
	ForgetThreshold   float64 // 遗忘阈值
	ConsolidationInterval time.Duration
	ImportanceDecay    float64
	MaxEmbeddingSize  int
	RecallTopK        int
}

// MemoryIndex 记忆索引
type MemoryIndex struct {
	byType       map[MemoryType][]string
	byKeyword    map[string][]string
	byEmotion    map[string][]string
	byImportance []string
	byTime       []string
}

// ForgettingMechanism 遗忘机制
type ForgettingMechanism struct {
	baseDecay      float64
	contextDecay   float64
	timeWeight     float64
	importanceWeight float64
}

// NewLongTermMemory 创建长期记忆系统
func NewLongTermMemory(config *MemoryConfig) *LongTermMemory {
	if config == nil {
		config = &MemoryConfig{
			MaxMemories:        10000,
			ForgetThreshold:    0.1,
			ConsolidationInterval: 1 * time.Hour,
			ImportanceDecay:    0.01,
			MaxEmbeddingSize:   1536,
			RecallTopK:         10,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ltm := &LongTermMemory{
		storage:    make(map[string]*MemoryEntry),
		index:      newMemoryIndex(),
		forgetting: newForgettingMechanism(),
		config:     config,
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动遗忘机制
	go ltm.runConsolidation()

	return ltm
}

// newMemoryIndex 创建记忆索引
func newMemoryIndex() *MemoryIndex {
	return &MemoryIndex{
		byType:       make(map[MemoryType][]string),
		byKeyword:    make(map[string][]string),
		byEmotion:    make(map[string][]string),
		byImportance: make([]string, 0),
		byTime:       make([]string, 0),
	}
}

// newForgettingMechanism 创建遗忘机制
func newForgettingMechanism() *ForgettingMechanism {
	return &ForgettingMechanism{
		baseDecay:        0.1,
		contextDecay:     0.05,
		timeWeight:       0.3,
		importanceWeight: 0.5,
	}
}

// Store 存储记忆
func (m *LongTermMemory) Store(entry *MemoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成ID
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.AccessedAt = now
	entry.AccessCount = 0

	// 计算TTL过期时间
	if entry.TTL > 0 {
		exp := now.Add(entry.TTL)
		entry.ExpiresAt = &exp
	}

	// 生成关键词
	if len(entry.Keywords) == 0 {
		entry.Keywords = extractKeywords(entry.Content)
	}

	// 生成嵌入向量 (简化版本，实际应使用AI模型)
	entry.Embedding = generateEmbedding(entry.Content, m.config.MaxEmbeddingSize)

	// 添加到存储
	m.storage[entry.ID] = entry

	// 更新索引
	m.updateIndex(entry)

	// 检查是否需要遗忘
	if len(m.storage) > m.config.MaxMemories {
		m.triggerForgetting()
	}

	return nil
}

// Recall 召回记忆
func (m *LongTermMemory) Recall(query *MemoryQuery) ([]*MemoryRecallResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if query.Limit <= 0 {
		query.Limit = m.config.RecallTopK
	}

	var results []*MemoryRecallResult

	// 如果有文本查询，使用语义搜索
	if query.Query != "" {
		queryEmbedding := generateEmbedding(query.Query, m.config.MaxEmbeddingSize)
		
		for _, entry := range m.storage {
			if m.shouldExclude(entry, query) {
				continue
			}

			// 计算余弦相似度
			score := cosineSimilarity(queryEmbedding, entry.Embedding)
			
			results = append(results, &MemoryRecallResult{
				Memory: entry,
				Score:  score,
			})
		}

		// 按分数排序
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	} else {
		// 使用索引筛选
		entries := m.filterByIndex(query)
		for _, entry := range entries {
			results = append(results, &MemoryRecallResult{
				Memory: entry,
				Score:  entry.Importance,
			})
		}
	}

	// 限制结果数量
	if len(results) > query.Limit {
		results = results[:query.Limit]
	}

	// 更新访问统计
	for _, r := range results {
		r.Memory.AccessCount++
		r.Memory.AccessedAt = time.Now()
	}

	return results, nil
}

// ShouldExclude 检查是否应排除
func (m *LongTermMemory) shouldExclude(entry *MemoryEntry, query *MemoryQuery) bool {
	// 检查过期
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		return true
	}

	// 检查类型
	if query.Type != "" && entry.Type != query.Type {
		return true
	}

	// 检查重要性
	if entry.Importance < query.MinImportance {
		return true
	}

	// 检查情感
	if query.Emotion != "" && entry.Emotion != query.Emotion {
		return true
	}

	// 检查关键词
	if len(query.Keywords) > 0 {
		hasKeyword := false
		for _, kw := range query.Keywords {
			for _, entryKw := range entry.Keywords {
				if strings.Contains(strings.ToLower(entryKw), strings.ToLower(kw)) {
					hasKeyword = true
					break
				}
			}
			if hasKeyword {
				break
			}
		}
		if !hasKeyword {
			return true
		}
	}

	return false
}

// FilterByIndex 通过索引筛选
func (m *LongTermMemory) filterByIndex(query *MemoryQuery) []*MemoryEntry {
	var entries []*MemoryEntry

	// 按类型
	if query.Type != "" {
		if ids, ok := m.index.byType[query.Type]; ok {
			for _, id := range ids {
				if entry, ok := m.storage[id]; ok {
					entries = append(entries, entry)
				}
			}
		}
	} else {
		// 返回所有
		for _, entry := range m.storage {
			entries = append(entries, entry)
		}
	}

	return entries
}

// UpdateIndex 更新索引
func (m *LongTermMemory) updateIndex(entry *MemoryEntry) {
	// 按类型索引
	m.index.byType[entry.Type] = append(m.index.byType[entry.Type], entry.ID)

	// 按关键词索引
	for _, kw := range entry.Keywords {
		kwLower := strings.ToLower(kw)
		m.index.byKeyword[kwLower] = append(m.index.byKeyword[kwLower], entry.ID)
	}

	// 按情感索引
	if entry.Emotion != "" {
		m.index.byEmotion[entry.Emotion] = append(m.index.byEmotion[entry.Emotion], entry.ID)
	}
}

// TriggerForgetting 触发遗忘
func (m *LongTermMemory) triggerForgetting() {
	// 计算所有记忆的遗忘值
	type forgetEntry struct {
		id    string
		value float64
	}

	var forgetList []forgetEntry
	now := time.Now()

	for id, entry := range m.storage {
		if entry.IsPinned {
			continue
		}

		// 计算遗忘值
		forgetValue := m.forgetting.Calculate(entry, now)
		
		forgetList = append(forgetList, forgetEntry{
			id:    id,
			value: forgetValue,
		})
	}

	// 按遗忘值排序
	sort.Slice(forgetList, func(i, j int) bool {
		return forgetList[i].value < forgetList[j].value
	})

	// 删除最低的记忆直到低于限制
	targetCount := m.config.MaxMemories / 2
	deleted := 0

	for _, fe := range forgetList {
		if len(m.storage)-deleted <= targetCount {
			break
		}

		if entry, ok := m.storage[fe.id]; ok {
			// 软删除：设置过期时间而不是立即删除
			exp := now.Add(24 * time.Hour)
			entry.ExpiresAt = &exp
			deleted++
		}
	}
}

// RunConsolidation 运行记忆整合
func (m *LongTermMemory) runConsolidation() {
	ticker := time.NewTicker(m.config.ConsolidationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.consolidate()
		}
	}
}

// Consolidate 记忆整合
func (m *LongTermMemory) consolidate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for id, entry := range m.storage {
		// 删除过期记忆
		if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			delete(m.storage, id)
			continue
		}

		// 重要性衰减
		if !entry.IsPinned {
			entry.Importance *= (1 - m.config.ImportanceDecay)
			if entry.Importance < m.config.ForgetThreshold {
				entry.Importance = m.config.ForgetThreshold
			}
		}

		// 更新最后访问时间
		if entry.AccessCount > 0 {
			accessInterval := now.Sub(entry.AccessedAt)
			// 长时间未访问的记忆衰减更快
			if accessInterval > 24*time.Hour {
				entry.Importance *= 0.9
			}
		}
	}
}

// Calculate 遗忘值计算
func (f *ForgettingMechanism) Calculate(entry *MemoryEntry, now time.Time) float64 {
	// 基础遗忘值
	base := f.baseDecay

	// 时间因素
	timeSinceAccess := now.Sub(entry.AccessedAt)
	timeFactor := math.Min(1.0, timeSinceAccess.Hours()/168.0) // 7天内完全衰减
	timeDecay := f.timeWeight * timeFactor

	// 重要性因素
	importanceFactor := 1.0 - (entry.Importance * f.importanceWeight)

	// 访问频率因素
	accessFactor := 0.0
	if entry.AccessCount > 0 {
		accessFactor = 1.0 / float64(entry.AccessCount+1)
	}

	// 综合遗忘值
	forgetValue := base + timeDecay + importanceFactor*0.3 + accessFactor*0.2

	return forgetValue
}

// Delete 删除记忆
func (m *LongTermMemory) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.storage[id]; !ok {
		return fmt.Errorf("memory not found: %s", id)
	}

	delete(m.storage, id)
	return nil
}

// Pin 固定记忆
func (m *LongTermMemory) Pin(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.storage[id]; ok {
		entry.IsPinned = true
		entry.Importance = 1.0
		entry.ExpiresAt = nil
		return nil
	}

	return fmt.Errorf("memory not found: %s", id)
}

// Unpin 取消固定
func (m *LongTermMemory) Unpin(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.storage[id]; ok {
		entry.IsPinned = false
		return nil
	}

	return fmt.Errorf("memory not found: %s", id)
}

// GetStats 获取记忆统计
func (m *LongTermMemory) GetStats() *MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &MemoryStats{
		TotalMemories: len(m.storage),
		ByType:       make(map[string]int),
	}

	for _, entry := range m.storage {
		stats.ByType[string(entry.Type)]++
		if entry.Importance > 0.8 {
			stats.HighImportance++
		}
		if entry.IsPinned {
			stats.Pinned++
		}
	}

	return stats
}

// MemoryStats 记忆统计
type MemoryStats struct {
	TotalMemories  int            `json:"total_memories"`
	ByType        map[string]int  `json:"by_type"`
	HighImportance int            `json:"high_importance"`
	Pinned        int            `json:"pinned"`
}

// Export 导出记忆
func (m *LongTermMemory) Export() ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]*MemoryEntry, 0, len(m.storage))
	for _, entry := range m.storage {
		entries = append(entries, entry)
	}

	return entries, nil
}

// Import 导入记忆
func (m *LongTermMemory) Import(entries []*MemoryEntry) error {
	for _, entry := range entries {
		if err := m.Store(entry); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭
func (m *LongTermMemory) Close() {
	m.cancel()
}

// ============ Helper Functions ============

// extractKeywords 提取关键词
func extractKeywords(content string) []string {
	// 移除特殊字符
	re := regexp.MustCompile(`[^\w\s]`)
	content = re.ReplaceAllString(content, " ")

	// 分词
	words := strings.Fields(content)

	// 停用词
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
	}

	// 提取有意义的词
	var keywords []string
	seen := make(map[string]bool)
	for _, word := range words {
		word = strings.ToLower(word)
		if len(word) > 2 && !stopWords[word] && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}

	// 限制关键词数量
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}

	return keywords
}

// generateEmbedding 生成嵌入向量 (简化版本)
func generateEmbedding(content string, size int) []float64 {
	embedding := make([]float64, size)
	
	// 简单的基于字符的嵌入生成
	// 实际应该使用 AI 模型如 OpenAI embeddings
	for i, c := range content {
		embedding[i%size] += float64(c) * 0.01
	}

	// L2 归一化
	var norm float64
	for _, v := range embedding {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
