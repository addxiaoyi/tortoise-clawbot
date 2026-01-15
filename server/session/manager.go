// Session management

package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Session status
type SessionStatus string

const (
	StatusActive    SessionStatus = "active"
	StatusIdle      SessionStatus = "idle"
	StatusArchived  SessionStatus = "archived"
)

// Session represents a client session
type Session struct {
	ID            string
	UserID        string
	Conn          *websocket.Conn
	Status        SessionStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastActiveAt  time.Time
	MessageCount  int
	Config        SessionConfig
	Metadata      SessionMetadata
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Model        string  `json:"model"`
	Temperature  float32 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
}

// SessionMetadata holds additional session data
type SessionMetadata struct {
	Channel string                 `json:"channel,omitempty"`
	Tags    []string              `json:"tags,omitempty"`
	Custom  map[string]interface{} `json:"custom,omitempty"`
}

// Manager manages sessions
type Manager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	maxSessions int
}

// NewManager creates a new session manager
func NewManager(maxSessions int) *Manager {
	return &Manager{
		sessions:    make(map[string]*Session),
		maxSessions: maxSessions,
	}
}

// NewSession creates a new session
func (m *Manager) NewSession(conn *websocket.Conn) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:           "sess_" + uuid.New().String()[:8],
		Conn:         conn,
		Status:       StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		MessageCount: 0,
		Config: SessionConfig{
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Metadata: SessionMetadata{
			Custom: make(map[string]interface{}),
		},
	}

	m.sessions[session.ID] = session
	return session
}

// Get retrieves a session by ID
func (m *Manager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// Remove removes a session
func (m *Manager) Remove(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
	return true
}

// RemoveSession removes a session (alias for Remove)
func (m *Manager) RemoveSession(id string) bool {
	return m.Remove(id)
}

// Touch updates session's last active time
func (m *Manager) Touch(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[id]; ok {
		session.LastActiveAt = time.Now()
		return true
	}
	return false
}

// IncrementMessageCount increments the message count
func (m *Manager) IncrementMessageCount(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[id]; ok {
		session.MessageCount++
		session.LastActiveAt = time.Now()
		session.UpdatedAt = time.Now()
		return true
	}
	return false
}

// List returns all sessions
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// Stats returns session statistics
func (m *Manager) Stats() SessionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active, idle int
	for _, s := range m.sessions {
		switch s.Status {
		case StatusActive:
			active++
		case StatusIdle:
			idle++
		}
	}

	return SessionStats{
		Total:  len(m.sessions),
		Active: active,
		Idle:   idle,
	}
}

// SessionStats holds session statistics
type SessionStats struct {
	Total  int
	Active int
	Idle   int
}
