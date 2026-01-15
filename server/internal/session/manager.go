package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager - 会话管理器
type Manager struct {
	sessions    map[string]*Session
	mu          sync.RWMutex
	maxSessions int
}

// Session - 会话
type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
	Messages  []*Message
}

// Message - 消息
type Message struct {
	ID        string
	SessionID string
	Role      string
	Content   string
	Format    string
	Type      string
	Timestamp time.Time
	Metadata  map[string]string
	ParentID  string
}

// NewManager 创建会话管理器
func NewManager(maxSessions int) *Manager {
	return &Manager{
		sessions:    make(map[string]*Session),
		maxSessions: maxSessions,
	}
}

// Create 创建新会话
func (m *Manager) Create(userID string, metadata map[string]string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理过期会话
	m.cleanupLocked()

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  metadata,
		Messages:  make([]*Message, 0),
	}

	// 如果超出限制，移除最旧的会话
	if len(m.sessions) >= m.maxSessions {
		var oldest *Session
		for _, s := range m.sessions {
			if oldest == nil || s.UpdatedAt.Before(oldest.UpdatedAt) {
				oldest = s
			}
		}
		if oldest != nil {
			delete(m.sessions, oldest.ID)
		}
	}

	m.sessions[session.ID] = session
	return session
}

// Get 获取会话
func (m *Manager) Get(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	session.UpdatedAt = time.Now()
	return session, nil
}

// Delete 删除会话
func (m *Manager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[id]; ok {
		delete(m.sessions, id)
		return true
	}
	return false
}

// List 列出所有会话
func (m *Manager) List(userID string) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, s := range m.sessions {
		if userID == "" || s.UserID == userID {
			sessions = append(sessions, s)
		}
	}
	return sessions
}

// Count 获取会话数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// AddMessage 添加消息到会话
func (s *Session) AddMessage(msg *Message) {
	msg.SessionID = s.ID
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
}

// GetMessages 获取消息
func (s *Session) GetMessages(limit, offset int) []*Message {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	end := offset + limit
	if end > len(s.Messages) {
		end = len(s.Messages)
	}
	if offset > len(s.Messages) {
		return []*Message{}
	}

	return s.Messages[offset:end]
}

// cleanupLocked 清理过期会话（需要持有锁）
func (m *Manager) cleanupLocked() {
	maxIdle := 24 * time.Hour
	now := time.Now()

	for id, s := range m.sessions {
		if now.Sub(s.UpdatedAt) > maxIdle {
			delete(m.sessions, id)
		}
	}
}
