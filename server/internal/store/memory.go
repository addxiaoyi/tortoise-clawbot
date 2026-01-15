package store

import (
	"sync"
	"time"
)

// ============ MemoryStore ============

// MemoryStore 记忆存储
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*Memory
}

// Memory 记忆
type Memory struct {
	ID         string
	Type       string // working | semantic | episodic
	Content    string
	Importance int
	Tags       []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewMemoryStore 创建记忆存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*Memory),
	}
}

// List 列出所有记忆
func (s *MemoryStore) List() []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memories := make([]*Memory, 0, len(s.data))
	for _, m := range s.data {
		memories = append(memories, m)
	}
	return memories
}

// Get 获取记忆
func (s *MemoryStore) Get(id string) *Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

// Create 创建记忆
func (s *MemoryStore) Create(mtype, content string, importance int, tags []string) *Memory {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := &Memory{
		ID:         generateID(),
		Type:       mtype,
		Content:    content,
		Importance: importance,
		Tags:       tags,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	s.data[m.ID] = m
	return m
}

// Update 更新记忆
func (s *MemoryStore) Update(id, content string, importance int, tags []string) *Memory {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.data[id]
	if !ok {
		return nil
	}

	m.Content = content
	m.Importance = importance
	m.Tags = tags
	m.UpdatedAt = time.Now()

	return m
}

// Delete 删除记忆
func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; ok {
		delete(s.data, id)
		return true
	}
	return false
}

// Search 搜索记忆
func (s *MemoryStore) Search(query string) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Memory, 0)
	for _, m := range s.data {
		if contains(m.Content, query) || containsAny(m.Tags, query) {
			results = append(results, m)
		}
	}
	return results
}

// Count 计数
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// ============ SessionStore ============

// SessionStore 会话存储
type SessionStore struct {
	mu   sync.RWMutex
	data map[string]*Session
}

// Session 会话
type Session struct {
	ID           string
	Name         string
	UserID       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MessageCount int
}

// NewSessionStore 创建会话存储
func NewSessionStore() *SessionStore {
	return &SessionStore{
		data: make(map[string]*Session),
	}
}

// List 列出所有会话
func (s *SessionStore) List() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.data))
	for _, session := range s.data {
		sessions = append(sessions, session)
	}
	return sessions
}

// Get 获取会话
func (s *SessionStore) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

// Create 创建会话
func (s *SessionStore) Create(name, userID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:        generateID(),
		Name:      name,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.data[session.ID] = session
	return session
}

// Update 更新会话
func (s *SessionStore) Update(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.data[id]; ok {
		session.UpdatedAt = time.Now()
	}
}

// Delete 删除会话
func (s *SessionStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; ok {
		delete(s.data, id)
		return true
	}
	return false
}

// Count 计数
func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// ============ MessageStore ============

// MessageStore 消息存储
type MessageStore struct {
	mu   sync.RWMutex
	data map[string]*Message
}

// Message 消息
type Message struct {
	ID        string
	SessionID string
	Role      string // user | assistant | system
	Content   string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

// NewMessageStore 创建消息存储
func NewMessageStore() *MessageStore {
	return &MessageStore{
		data: make(map[string]*Message),
	}
}

// List 列出所有消息
func (s *MessageStore) List() []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]*Message, 0, len(s.data))
	for _, m := range s.data {
		messages = append(messages, m)
	}
	return messages
}

// Get 获取消息
func (s *MessageStore) Get(id string) *Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

// GetBySession 获取会话的所有消息
func (s *MessageStore) GetBySession(sessionID string) []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]*Message, 0)
	for _, m := range s.data {
		if m.SessionID == sessionID {
			messages = append(messages, m)
		}
	}
	return messages
}

// Create 创建消息
func (s *MessageStore) Create(msg *Message) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.ID == "" {
		msg.ID = generateID()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	s.data[msg.ID] = msg
	return msg
}

// Delete 删除消息
func (s *MessageStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; ok {
		delete(s.data, id)
		return true
	}
	return false
}

// Count 计数
func (s *MessageStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// ============ PluginStore ============

// PluginStore 插件存储
type PluginStore struct {
	mu   sync.RWMutex
	data map[string]*Plugin
}

// Plugin 插件
type Plugin struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	Enabled     bool
	Tools       []Tool
	Status      string // active | inactive | error
}

// Tool 工具
type Tool struct {
	Name        string
	Description string
	Parameters  []Parameter
}

// Parameter 参数
type Parameter struct {
	Name     string
	Type     string
	Required bool
}

// NewPluginStore 创建插件存储
func NewPluginStore() *PluginStore {
	s := &PluginStore{
		data: make(map[string]*Plugin),
	}
	// 添加默认插件
	s.initDefaultPlugins()
	return s
}

// initDefaultPlugins 初始化默认插件
func (s *PluginStore) initDefaultPlugins() {
	defaultPlugins := []*Plugin{
		{
			ID:          "search",
			Name:        "Web Search",
			Version:     "1.0.0",
			Description: "网页搜索工具",
			Author:      "Tortoise",
			Enabled:     true,
			Status:      "active",
			Tools: []Tool{
				{
					Name:        "search",
					Description: "搜索网页内容",
					Parameters: []Parameter{
						{Name: "query", Type: "string", Required: true},
						{Name: "limit", Type: "number", Required: false},
					},
				},
			},
		},
		{
			ID:          "calculator",
			Name:        "Calculator",
			Version:     "1.0.0",
			Description: "数学计算工具",
			Author:      "Tortoise",
			Enabled:     true,
			Status:      "active",
			Tools: []Tool{
				{
					Name:        "calculate",
					Description: "执行数学计算",
					Parameters: []Parameter{
						{Name: "expression", Type: "string", Required: true},
					},
				},
			},
		},
	}

	for _, p := range defaultPlugins {
		s.data[p.ID] = p
	}
}

// List 列出所有插件
func (s *PluginStore) List() []*Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(s.data))
	for _, p := range s.data {
		plugins = append(plugins, p)
	}
	return plugins
}

// Get 获取插件
func (s *PluginStore) Get(id string) *Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

// Toggle 切换插件状态
func (s *PluginStore) Toggle(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, ok := s.data[id]; ok {
		p.Enabled = !p.Enabled
	}
}

// Count 计数
func (s *PluginStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// ============ Helpers ============

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAny(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
