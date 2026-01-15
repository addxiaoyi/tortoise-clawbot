package memory

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// MemoryType 记忆类型
type MemoryType int

const (
	MemoryTypeWorking MemoryType = iota
	MemoryTypeSemantic
	MemoryTypeEpisodic
)

func (t MemoryType) String() string {
	switch t {
	case MemoryTypeWorking:
		return "working"
	case MemoryTypeSemantic:
		return "semantic"
	case MemoryTypeEpisodic:
		return "episodic"
	default:
		return "unknown"
	}
}

// Memory 记忆单元
type Memory struct {
	ID         string
	Type       MemoryType
	Content    string
	Embedding  []float32 // 向量嵌入
	Importance float32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	AccessCount atomic.Int64
	ExpiresAt  time.Time // 用于 Working Memory
}

// Config 记忆系统配置
type Config struct {
	WorkingCapacity  int
	SemanticCapacity int
	EpisodicCapacity int
	WorkingTTL      time.Duration
}

// System 三层记忆系统
type System struct {
	config Config

	// 三层记忆存储
	working   map[string]*Memory
	semantic  map[string]*Memory
	episodic  map[string]*Memory

	// 索引 (用于快速检索)
	semanticIndex *SemanticIndex
	episodicIndex *EpisodicIndex

	// 锁
	mu sync.RWMutex

	// 统计
	stats Stats
}

// Stats 记忆系统统计
type Stats struct {
	TotalMemories   atomic.Int64
	WorkingCount    atomic.Int64
	SemanticCount   atomic.Int64
	EpisodicCount   atomic.Int64
	SearchLatencyUs atomic.Int64
	IndexUpdates    atomic.Int64
}

// SemanticIndex 语义索引 (简化版向量索引)
type SemanticIndex struct {
	vectors map[string][]float32
	mu     sync.RWMutex
}

// EpisodicIndex 情景索引
type EpisodicIndex struct {
	timeline map[string][]*Memory
	mu      sync.RWMutex
}

// NewSystem 创建记忆系统
func NewSystem(cfg Config) *System {
	if cfg.WorkingCapacity == 0 {
		cfg.WorkingCapacity = 1000
	}
	if cfg.SemanticCapacity == 0 {
		cfg.SemanticCapacity = 100000
	}
	if cfg.EpisodicCapacity == 0 {
		cfg.EpisodicCapacity = 50000
	}
	if cfg.WorkingTTL == 0 {
		cfg.WorkingTTL = 5 * time.Minute
	}

	s := &System{
		config: cfg,
		working:  make(map[string]*Memory),
		semantic: make(map[string]*Memory),
		episodic: make(map[string]*Memory),
		semanticIndex: &SemanticIndex{
			vectors: make(map[string][]float32),
		},
		episodicIndex: &EpisodicIndex{
			timeline: make(map[string][]*Memory),
		},
	}

	// 启动清理协程
	go s.gc()

	return s
}

// gc 垃圾回收 (清理过期记忆)
func (s *System) gc() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupWorking()
	}
}

// cleanupWorking 清理过期的工作记忆
func (s *System) cleanupWorking() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, mem := range s.working {
		if !mem.ExpiresAt.IsZero() && now.After(mem.ExpiresAt) {
			delete(s.working, id)
			s.stats.WorkingCount.Add(-1)
		}
	}
}

// Store 存储记忆
func (s *System) Store(memType MemoryType, content string, embedding []float32) *Memory {
	mem := &Memory{
		ID:         uuid.New().String(),
		Type:       memType,
		Content:    content,
		Embedding:  embedding,
		Importance: 0.5,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch memType {
	case MemoryTypeWorking:
		if len(s.working) >= s.config.WorkingCapacity {
			s.evictWorking()
		}
		mem.ExpiresAt = time.Now().Add(s.config.WorkingTTL)
		s.working[mem.ID] = mem

	case MemoryTypeSemantic:
		if len(s.semantic) >= s.config.SemanticCapacity {
			s.evictSemantic()
		}
		s.semantic[mem.ID] = mem
		s.semanticIndex.vectors[mem.ID] = embedding

	case MemoryTypeEpisodic:
		if len(s.episodic) >= s.config.EpisodicCapacity {
			s.evictEpisodic()
		}
		s.episodic[mem.ID] = mem
		s.episodicIndex.timeline[mem.ID] = append(s.episodicIndex.timeline[mem.ID], mem)
	}

	s.stats.TotalMemories.Add(1)
	return mem
}

// evictWorking 驱逐工作记忆
func (s *System) evictWorking() {
	var oldest *Memory
	for _, mem := range s.working {
		if oldest == nil || mem.CreatedAt.Before(oldest.CreatedAt) {
			oldest = mem
		}
	}
	if oldest != nil {
		delete(s.working, oldest.ID)
		s.stats.WorkingCount.Add(-1)
	}
}

// evictSemantic 驱逐语义记忆
func (s *System) evictSemantic() {
	var lowest *Memory
	for _, mem := range s.semantic {
		if lowest == nil || mem.Importance < lowest.Importance {
			lowest = mem
		}
	}
	if lowest != nil {
		delete(s.semantic, lowest.ID)
		delete(s.semanticIndex.vectors, lowest.ID)
		s.stats.SemanticCount.Add(-1)
	}
}

// evictEpisodic 驱逐情景记忆
func (s *System) evictEpisodic() {
	var oldest *Memory
	for _, mem := range s.episodic {
		if oldest == nil || mem.CreatedAt.Before(oldest.CreatedAt) {
			oldest = mem
		}
	}
	if oldest != nil {
		delete(s.episodic, oldest.ID)
		s.stats.EpisodicCount.Add(-1)
	}
}

// Retrieve 检索记忆
func (s *System) Retrieve(id string) (*Memory, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mem, ok := s.working[id]; ok {
		mem.AccessCount.Add(1)
		return mem, true
	}
	if mem, ok := s.semantic[id]; ok {
		mem.AccessCount.Add(1)
		return mem, true
	}
	if mem, ok := s.episodic[id]; ok {
		mem.AccessCount.Add(1)
		return mem, true
	}
	return nil, false
}

// Search 语义搜索
func (s *System) Search(queryEmbedding []float32, limit int) []*Memory {
	start := time.Now()
	defer func() {
		s.stats.SearchLatencyUs.Store(time.Since(start).Microseconds())
	}()

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Memory, 0, limit)

	// 计算余弦相似度
	for id, mem := range s.semantic {
		if emb, ok := s.semanticIndex.vectors[id]; ok {
			sim := cosineSimilarity(queryEmbedding, emb)
			if sim > 0.7 { // 阈值
				memCopy := *mem
				results = append(results, &memCopy)
			}
		}
	}

	// 按相似度排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if cosineSimilarity(queryEmbedding, s.semanticIndex.vectors[results[j].ID]) >
				cosineSimilarity(queryEmbedding, s.semanticIndex.vectors[results[i].ID]) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32.sqrt(normA) * float32.sqrt(normB))
}

// Delete 删除记忆
func (s *System) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.working[id]; ok {
		delete(s.working, id)
		s.stats.WorkingCount.Add(-1)
		s.stats.TotalMemories.Add(-1)
		return true
	}
	if _, ok := s.semantic[id]; ok {
		delete(s.semantic, id)
		delete(s.semanticIndex.vectors, id)
		s.stats.SemanticCount.Add(-1)
		s.stats.TotalMemories.Add(-1)
		return true
	}
	if _, ok := s.episodic[id]; ok {
		delete(s.episodic, id)
		s.stats.EpisodicCount.Add(-1)
		s.stats.TotalMemories.Add(-1)
		return true
	}
	return false
}

// GetByType 按类型获取
func (s *System) GetByType(memType MemoryType) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var memories []*Memory
	switch memType {
	case MemoryTypeWorking:
		for _, m := range s.working {
			memories = append(memories, m)
		}
	case MemoryTypeSemantic:
		for _, m := range s.semantic {
			memories = append(memories, m)
		}
	case MemoryTypeEpisodic:
		for _, m := range s.episodic {
			memories = append(memories, m)
		}
	}
	return memories
}

// Stats 获取统计
func (s *System) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Stats{
		TotalMemories:   atomic.Int64{int64(len(s.working) + len(s.semantic) + len(s.episodic))},
		WorkingCount:   atomic.Int64{int64(len(s.working))},
		SemanticCount:  atomic.Int64{int64(len(s.semantic))},
		EpisodicCount:  atomic.Int64{int64(len(s.episodic))},
	}
}
