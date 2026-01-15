package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Session 会话结构
type Session struct {
	ID        string
	UserID    string
	Name      string
	Channel   string
	State     SessionState
	Context   []Message
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	LastSeen  time.Time
	ExpiresAt time.Time
}

// Message 上下文消息
type Message struct {
	Role    string
	Content string
	Time    time.Time
}

// SessionState 会话状态
type SessionState int

const (
	StateActive SessionState = iota
	StateIdle
	StateExpired
	StateClosed
)

func (s SessionState) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateIdle:
		return "idle"
	case StateExpired:
		return "expired"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Config 会话配置
type Config struct {
	MaxSessions    int
	SessionTimeout time.Duration
	ContextLimit   int
	CleanupInterval time.Duration
}

// Manager 会话管理器
type Manager struct {
	config Config

	// 会话存储
	sessions map[string]*Session
	byUser  map[string][]string // userID -> sessionIDs

	// 索引
	indexMu sync.RWMutex
	index   map[string][]string // keyword -> sessionIDs

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 统计
	stats Stats

	// 锁
	mu sync.RWMutex
}

// Stats 会话统计
type Stats struct {
	TotalSessions  atomic.Int64
	ActiveSessions atomic.Int64
	AvgDurationMs  atomic.Int64
}

// NewManager 创建会话管理器
func NewManager(cfg Config) *Manager {
	if cfg.MaxSessions == 0 {
		cfg.MaxSessions = 100000
	}
	if cfg.SessionTimeout == 0 {
		cfg.SessionTimeout = 24 * time.Hour
	}
	if cfg.ContextLimit == 0 {
		cfg.ContextLimit = 100
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:   cfg,
		sessions: make(map[string]*Session),
		byUser:   make(map[string][]string),
		index:    make(map[string][]string),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 启动清理协程
	go m.cleanup()

	return m
}

// cleanup 定期清理过期会话
func (m *Manager) cleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupExpired()
		}
	}
}

// cleanupExpired 清理过期会话
func (m *Manager) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
			delete(m.sessions, id)
			m.stats.ActiveSessions.Add(-1)

			// 从用户索引中移除
			if ids, ok := m.byUser[session.UserID]; ok {
				newIDs := make([]string, 0, len(ids))
				for _, sid := range ids {
					if sid != id {
						newIDs = append(newIDs, sid)
					}
				}
				m.byUser[session.UserID] = newIDs
			}
		}
	}
}

// Create 创建会话
func (m *Manager) Create(userID, name, channel string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查容量
	if len(m.sessions) >= m.config.MaxSessions {
		m.evictOldest()
	}

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Channel:   channel,
		State:     StateActive,
		Context:   make([]Message, 0),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
		ExpiresAt: time.Now().Add(m.config.SessionTimeout),
	}

	m.sessions[session.ID] = session
	m.byUser[userID] = append(m.byUser[userID], session.ID)

	m.stats.TotalSessions.Add(1)
	m.stats.ActiveSessions.Add(1)

	// 更新索引
	m.updateIndex(session)

	return session
}

// Get 获取会话
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if ok {
		session.LastSeen = time.Now()
		session.ExpiresAt = time.Now().Add(m.config.SessionTimeout)
	}
	return session, ok
}

// GetByUser 获取用户的所有会话
func (m *Manager) GetByUser(userID string) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := m.byUser[userID]
	sessions := make([]*Session, 0, len(ids))

	for _, id := range ids {
		if session, ok := m.sessions[id]; ok {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// Update 更新会话
func (m *Manager) Update(id string, update func(*Session)) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return false
	}

	update(session)
	session.UpdatedAt = time.Now()
	session.LastSeen = time.Now()
	session.ExpiresAt = time.Now().Add(m.config.SessionTimeout)

	m.updateIndex(session)
	return true
}

// Delete 删除会话
func (m *Manager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return false
	}

	delete(m.sessions, id)

	// 从用户索引中移除
	if ids, ok := m.byUser[session.UserID]; ok {
		newIDs := make([]string, 0, len(ids))
		for _, sid := range ids {
			if sid != id {
				newIDs = append(newIDs, sid)
			}
		}
		m.byUser[session.UserID] = newIDs
	}

	m.stats.ActiveSessions.Add(-1)
	return true
}

// AppendContext 添加上下文消息
func (m *Manager) AppendContext(id string, role, content string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return false
	}

	// 限制上下文长度
	if len(session.Context) >= m.config.ContextLimit {
		session.Context = session.Context[1:]
	}

	session.Context = append(session.Context, Message{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})

	session.UpdatedAt = time.Now()
	return true
}

// Search 搜索会话
func (m *Manager) Search(keyword string) []*Session {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()

	ids := m.index[keyword]
	sessions := make([]*Session, 0, len(ids))

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, id := range ids {
		if session, ok := m.sessions[id]; ok {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// updateIndex 更新索引
func (m *Manager) updateIndex(session *Session) {
	m.indexMu.Lock()
	defer m.indexMu.Unlock()

	// 简单关键词索引
	keywords := extractKeywords(session.Name)
	for _, kw := range keywords {
		m.index[kw] = append(m.index[kw], session.ID)
	}
}

// extractKeywords 提取关键词
func extractKeywords(text string) []string {
	// 简单实现，实际应该使用分词
	return []string{text}
}

// evictOldest 驱逐最老的会话
func (m *Manager) evictOldest() {
	var oldest *Session
	for _, session := range m.sessions {
		if oldest == nil || session.LastSeen.Before(oldest.LastSeen) {
			oldest = session
		}
	}
	if oldest != nil {
		delete(m.sessions, oldest.ID)
		m.stats.ActiveSessions.Add(-1)
	}
}

// Stats 获取统计
func (m *Manager) Stats() (s Stats) {
	s.TotalSessions.Store(m.stats.TotalSessions.Load())
	s.ActiveSessions.Store(m.stats.ActiveSessions.Load())
	s.AvgDurationMs.Store(m.stats.AvgDurationMs.Load())
	return
}

// List 列出所有会话
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Close 关闭管理器
func (m *Manager) Close() {
	m.cancel()
}
