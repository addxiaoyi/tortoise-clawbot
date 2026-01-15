package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryType - 记忆类型
type MemoryType int

const (
	MemoryTypeUnspecified MemoryType = 0
	MemoryTypeWorking     MemoryType = 1
	MemoryTypeSemantic    MemoryType = 2
	MemoryTypeEpisodic    MemoryType = 3
)

// Memory - 记忆
type Memory struct {
	ID         string
	Type       MemoryType
	Content    string
	Importance float32
	CreatedAt  time.Time
	AccessedAt time.Time
	Metadata   map[string]string
}

// Manager - 记忆管理器
type Manager struct {
	working  []*Memory
	semantic []*Memory
	episodic []*Memory
	mu       sync.RWMutex
	maxSize  int
}

// NewManager 创建记忆管理器
func NewManager() *Manager {
	return &Manager{
		working:  make([]*Memory, 0, 100),
		semantic: make([]*Memory, 0, 10000),
		episodic: make([]*Memory, 0, 5000),
		maxSize:  10000,
	}
}

// Save 保存记忆
func (m *Manager) Save(memType, content string, importance float32, metadata map[string]string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	mem := &Memory{
		ID:         uuid.New().String(),
		Content:    content,
		Importance: importance,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Metadata:   metadata,
	}

	switch memType {
	case "working":
		mem.Type = MemoryTypeWorking
		m.working = append(m.working, mem)
		m.trimMemory(&m.working)
	case "semantic":
		mem.Type = MemoryTypeSemantic
		m.semantic = append(m.semantic, mem)
		m.trimMemory(&m.semantic)
	case "episodic":
		mem.Type = MemoryTypeEpisodic
		m.episodic = append(m.episodic, mem)
		m.trimMemory(&m.episodic)
	}

	return mem.ID
}

// Query 查询记忆
func (m *Manager) Query(query, memType string, limit int) []*Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var memories []*Memory
	switch memType {
	case "working":
		memories = m.working
	case "semantic":
		memories = m.semantic
	case "episodic":
		memories = m.episodic
	default:
		memories = append(memories, m.working...)
		memories = append(memories, m.semantic...)
		memories = append(memories, m.episodic...)
	}

	if limit > 0 && limit < len(memories) {
		return memories[:limit]
	}
	return memories
}

// Get 获取记忆
func (m *Manager) Get(id string) (*Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 搜索所有记忆
	for _, mem := range m.working {
		if mem.ID == id {
			mem.AccessedAt = time.Now()
			return mem, nil
		}
	}
	for _, mem := range m.semantic {
		if mem.ID == id {
			mem.AccessedAt = time.Now()
			return mem, nil
		}
	}
	for _, mem := range m.episodic {
		if mem.ID == id {
			mem.AccessedAt = time.Now()
			return mem, nil
		}
	}

	return nil, ErrMemoryNotFound
}

// Delete 删除记忆
func (m *Manager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 尝试从所有记忆中删除
	if m.deleteFromList(&m.working, id) {
		return true
	}
	if m.deleteFromList(&m.semantic, id) {
		return true
	}
	if m.deleteFromList(&m.episodic, id) {
		return true
	}
	return false
}

// deleteFromList 从列表中删除记忆
func (m *Manager) deleteFromList(memories *[]*Memory, id string) bool {
	for i, mem := range *memories {
		if mem.ID == id {
			*memories = append((*memories)[:i], (*memories)[i+1:]...)
			return true
		}
	}
	return false
}

// trimMemory 裁剪记忆
func (m *Manager) trimMemory(memories *[]*Memory) {
	maxLen := m.maxSize
	if len(*memories) > maxLen {
		*memories = (*memories)[len(*memories)-maxLen:]
	}
}

// Clear 清除记忆
func (m *Manager) Clear(memType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch memType {
	case "working":
		m.working = make([]*Memory, 0, 100)
	case "semantic":
		m.semantic = make([]*Memory, 0, 10000)
	case "episodic":
		m.episodic = make([]*Memory, 0, 5000)
	default:
		m.working = make([]*Memory, 0, 100)
		m.semantic = make([]*Memory, 0, 10000)
		m.episodic = make([]*Memory, 0, 5000)
	}
}
