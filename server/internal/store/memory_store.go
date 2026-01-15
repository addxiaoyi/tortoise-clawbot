package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeWorking  MemoryType = "working"
	MemoryTypeSemantic MemoryType = "semantic"
	MemoryTypeEpisodic MemoryType = "episodic"
)

// Memory 记忆模型
type Memory struct {
	ID         string      `json:"id"`
	Type       MemoryType `json:"type"`
	Content    string      `json:"content"`
	Importance float64    `json:"importance"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	Tags       []string   `json:"tags,omitempty"`
}

// MemoryStore 记忆存储
type MemoryStore struct {
	memories map[string]*Memory
	mu       sync.RWMutex
}

// NewMemoryStore 创建记忆存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		memories: make(map[string]*Memory),
	}
	store.createSampleData()
	return store
}

func (m *MemoryStore) createSampleData() {
	memories := []*Memory{
		{
			ID:         uuid.New().String(),
			Type:       MemoryTypeSemantic,
			Content:    "用户喜欢深色主题界面",
			Importance: 0.8,
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now().Add(-24 * time.Hour),
			Tags:       []string{"偏好", "界面"},
		},
		{
			ID:         uuid.New().String(),
			Type:       MemoryTypeSemantic,
			Content:    "用户主要使用中文交流",
			Importance: 0.9,
			CreatedAt:  time.Now().Add(-48 * time.Hour),
			UpdatedAt:  time.Now().Add(-48 * time.Hour),
			Tags:       []string{"语言", "交流"},
		},
		{
			ID:         uuid.New().String(),
			Type:       MemoryTypeWorking,
			Content:    "当前正在开发 Tortoise Web UI 项目",
			Importance: 0.7,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-2 * time.Hour),
			Tags:       []string{"项目", "开发"},
		},
		{
			ID:         uuid.New().String(),
			Type:       MemoryTypeEpisodic,
			Content:    "用户今天完成了前端组件开发",
			Importance: 0.5,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
			Tags:       []string{"事件", "开发"},
		},
	}

	for _, memory := range memories {
		m.memories[memory.ID] = memory
	}
}

// GetMemories 获取所有记忆
func (m *MemoryStore) GetMemories(memType string) []*Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memories := make([]*Memory, 0)
	for _, memory := range m.memories {
		if memType == "" || memType == "all" || string(memory.Type) == memType {
			memories = append(memories, memory)
		}
	}
	return memories
}

// GetMemory 获取单个记忆
func (m *MemoryStore) GetMemory(id string) (*Memory, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memory, ok := m.memories[id]
	return memory, ok
}

// CreateMemory 创建记忆
func (m *MemoryStore) CreateMemory(memType string, content string, importance float64) *Memory {
	m.mu.Lock()
	defer m.mu.Unlock()

	memory := &Memory{
		ID:         uuid.New().String(),
		Type:       MemoryType(memType),
		Content:    content,
		Importance: importance,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.memories[memory.ID] = memory
	return memory
}

// DeleteMemory 删除记忆
func (m *MemoryStore) DeleteMemory(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.memories[id]; ok {
		delete(m.memories, id)
		return true
	}
	return false
}

// SearchMemories 搜索记忆
func (m *MemoryStore) SearchMemories(query string) []*Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memories := make([]*Memory, 0)
	for _, memory := range m.memories {
		if containsIgnoreCase(memory.Content, query) {
			memories = append(memories, memory)
		}
	}
	return memories
}
	