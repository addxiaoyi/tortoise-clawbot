package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session 会话模型
type Session struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	UserID       string    `json:"userId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MessageCount int       `json:"messageCount"`
	LastMessage  string    `json:"lastMessage,omitempty"`
}

// SessionStore 会话存储
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore 创建会话存储
func NewSessionStore() *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
	}
	// 创建一些示例会话
	store.createSampleData()
	return store
}

func (s *SessionStore) createSampleData() {
	sessions := []*Session{
		{
			ID:           uuid.New().String(),
			Name:         "项目讨论",
			UserID:       "user1",
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			UpdatedAt:    time.Now().Add(-2 * time.Minute),
			MessageCount: 45,
			LastMessage:  "这个功能听起来不错",
		},
		{
			ID:           uuid.New().String(),
			Name:         "代码审查",
			UserID:       "user1",
			CreatedAt:    time.Now().Add(-5 * time.Hour),
			UpdatedAt:    time.Now().Add(-15 * time.Minute),
			MessageCount: 23,
			LastMessage:  "建议添加错误处理",
		},
		{
			ID:           uuid.New().String(),
			Name:         "技术调研",
			UserID:       "user1",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now().Add(-1 * time.Hour),
			MessageCount: 67,
			LastMessage:  "Rust 的性能确实很强",
		},
	}

	for _, session := range sessions {
		s.sessions[session.ID] = session
	}
}

// GetSessions 获取所有会话
func (s *SessionStore) GetSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetSession 获取单个会话
func (s *SessionStore) GetSession(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	return session, ok
}

// CreateSession 创建会话
func (s *SessionStore) CreateSession(name, userID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:           uuid.New().String(),
		Name:         name,
		UserID:       userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		MessageCount: 0,
	}
	s.sessions[session.ID] = session
	return session
}

// DeleteSession 删除会话
func (s *SessionStore) DeleteSession(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[id]; ok {
		delete(s.sessions, id)
		return true
	}
	return false
}

// UpdateSession 更新会话
func (s *SessionStore) UpdateSession(id string, name string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, false
	}

	if name != "" {
		session.Name = name
	}
	session.UpdatedAt = time.Now()
	return session, true
}

// IncrementMessageCount 增加消息计数
func (s *SessionStore) IncrementMessageCount(id string, lastMessage string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[id]; ok {
		session.MessageCount++
		session.LastMessage = lastMessage
		session.UpdatedAt = time.Now()
	}
}
