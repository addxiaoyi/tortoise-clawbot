package session

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PiMessage represents a message in Pi session
type PiMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Tokens      int                    `json:"tokens,omitempty"`
	Latency     int64                  `json:"latency,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Attachments []Attachment           `json:"attachments,omitempty"`
}

// Attachment represents file attachments
type Attachment struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int    `json:"size"`
	Content  string `json:"content,omitempty"`
}

// PiSession represents a Pi session
type PiSession struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Messages     []PiMessage        `json:"messages"`
	Metadata     SessionMetadata    `json:"metadata"`
	Tree         *MessageTree       `json:"tree,omitempty"`
	Compressions []CompressionRecord `json:"compressions,omitempty"`
}

// SessionMetadata session metadata
type SessionMetadata struct {
	Model         string   `json:"model"`
	Provider      string   `json:"provider"`
	TotalTokens   int      `json:"total_tokens"`
	TotalMessages int      `json:"total_messages"`
	TotalLatency  int64    `json:"total_latency"`
	UserID        string   `json:"user_id,omitempty"`
	ChannelID     string   `json:"channel_id,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// MessageTree threaded message tree
type MessageTree struct {
	RootID string            `json:"root_id"`
	Nodes  map[string]*TreeNode `json:"nodes"`
}

// TreeNode message in tree
type TreeNode struct {
	Message  *PiMessage `json:"message"`
	Children []string   `json:"children"`
	Depth    int        `json:"depth"`
}

// CompressionRecord compression history
type CompressionRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string   `json:"method"`
	Before    int      `json:"before"`
	After     int      `json:"after"`
	Ratio     float64  `json:"ratio"`
}

// PiSessionManager manages Pi sessions
type PiSessionManager struct {
	sessions    map[string]*PiSession
	mu          sync.RWMutex
	storagePath string
	maxMessages int
	compressAge time.Duration
}

// NewPiSessionManager creates a new Pi session manager
func NewPiSessionManager(storagePath string) *PiSessionManager {
	return &PiSessionManager{
		sessions:    make(map[string]*PiSession),
		storagePath: storagePath,
		maxMessages: 1000,
		compressAge: 24 * time.Hour,
	}
}

// CreateSession creates a new Pi session
func (m *PiSessionManager) CreateSession(title, model, provider string) *PiSession {
	session := &PiSession{
		ID:        uuid.New().String(),
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []PiMessage{},
		Metadata: SessionMetadata{
			Model:    model,
			Provider: provider,
		},
		Tree: &MessageTree{
			Nodes: make(map[string]*TreeNode),
		},
	}

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session
}

// AddMessage adds a message to session
func (m *PiSessionManager) AddMessage(sessionID string, msg PiMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	session.Messages = append(session.Messages, msg)
	session.Metadata.TotalMessages++
	session.UpdatedAt = time.Now()

	// Update tree
	if session.Tree != nil {
		session.Tree.Nodes[msg.ID] = &TreeNode{
			Message:  &msg,
			Children: []string{},
			Depth:    0,
		}
		if msg.ParentID != "" && session.Tree.Nodes[msg.ParentID] != nil {
			session.Tree.Nodes[msg.ParentID].Children = append(
				session.Tree.Nodes[msg.ParentID].Children, msg.ID,
			)
			session.Tree.Nodes[msg.ID].Depth = session.Tree.Nodes[msg.ParentID].Depth + 1
		}
	}

	return nil
}

// GetMessages retrieves messages with pagination
func (m *PiSessionManager) GetMessages(sessionID string, limit, offset int) ([]PiMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	messages := session.Messages
	if offset >= len(messages) {
		return []PiMessage{}, nil
	}

	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}

	return messages[offset:end], nil
}

// ExportJSON exports session to JSON
func (m *PiSessionManager) ExportJSON(sessionID string) ([]byte, error) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return json.MarshalIndent(session, "", "  ")
}

// ExportJSONL exports session to JSONL format
func (m *PiSessionManager) ExportJSONL(sessionID string) ([]byte, error) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var buf bytes.Buffer
	for _, msg := range session.Messages {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

// ExportMarkdown exports session to Markdown
func (m *PiSessionManager) ExportMarkdown(sessionID string) ([]byte, error) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString(fmt.Sprintf("# %s\n\n", session.Title))
	buf.WriteString(fmt.Sprintf("**Created:** %s\n", session.CreatedAt.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("**Model:** %s (%s)\n\n", session.Metadata.Model, session.Metadata.Provider))
	buf.WriteString("---\n\n")

	// Messages
	for i, msg := range session.Messages {
		role := strings.ToUpper(msg.Role)
		if role == "" {
			role = strings.ToUpper(msg.Type)
		}

		buf.WriteString(fmt.Sprintf("## [%d] %s\n\n", i+1, role))
		buf.WriteString(fmt.Sprintf("**Time:** %s\n", msg.Timestamp.Format(time.RFC3339)))

		if msg.Model != "" {
			buf.WriteString(fmt.Sprintf("**Model:** %s\n", msg.Model))
		}
		if msg.Tokens > 0 {
			buf.WriteString(fmt.Sprintf("**Tokens:** %d\n", msg.Tokens))
		}
		if msg.Latency > 0 {
			buf.WriteString(fmt.Sprintf("**Latency:** %dms\n", msg.Latency))
		}

		buf.WriteString("\n")
		buf.WriteString(msg.Content)
		buf.WriteString("\n\n---\n\n")
	}

	return buf.Bytes(), nil
}

// CompressSession compresses old messages
func (m *PiSessionManager) CompressSession(sessionID string, method string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if len(session.Messages) < 10 {
		return fmt.Errorf("not enough messages to compress")
	}

	before := len(session.Messages)

	// Strategy: keep last N messages and summarize older ones
	keepCount := len(session.Messages) / 2
	if keepCount < 10 {
		keepCount = 10
	}

	kept := session.Messages[len(session.Messages)-keepCount:]
	compressed := session.Messages[:len(session.Messages)-keepCount]

	// Generate summary
	summary := generateSessionSummary(compressed)

	// Replace old messages with summary
	session.Messages = append([]PiMessage{
		{
			ID:        uuid.New().String(),
			Type:      "system",
			Role:      "system",
			Content:   fmt.Sprintf("[Previous %d messages summarized]: %s", len(compressed), summary),
			Timestamp: time.Now(),
		},
	}, kept...)

	// Record compression
	record := CompressionRecord{
		Timestamp: time.Now(),
		Method:    method,
		Before:    before,
		After:     len(session.Messages),
		Ratio:     float64(len(session.Messages)) / float64(before),
	}
	session.Compressions = append(session.Compressions, record)

	return nil
}

// generateSessionSummary generates a summary of messages
func generateSessionSummary(messages []PiMessage) string {
	var userMsgs, assistantMsgs int
	var topics []string

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userMsgs++
		case "assistant":
			assistantMsgs++
		}

		// Extract potential topics (simplified)
		words := strings.Fields(msg.Content)
		if len(words) > 0 {
			for _, w := range words[:min(5, len(words))] {
				if len(w) > 5 {
					topics = append(topics, w)
				}
			}
		}
	}

	return fmt.Sprintf("Chat about %d user messages and %d assistant responses", userMsgs, assistantMsgs)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SaveSession saves session to disk
func (m *PiSessionManager) SaveSession(sessionID string) error {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	filename := filepath.Join(m.storagePath, sessionID+".json")
	return os.WriteFile(filename, data, 0644)
}

// LoadSession loads session from disk
func (m *PiSessionManager) LoadSession(sessionID string) error {
	filename := filepath.Join(m.storagePath, sessionID+".json")

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var session PiSession
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	m.mu.Lock()
	m.sessions[sessionID] = &session
	m.mu.Unlock()

	return nil
}

// DeleteSession deletes a session
func (m *PiSessionManager) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(m.sessions, sessionID)

	// Also delete file
	filename := filepath.Join(m.storagePath, sessionID+".json")
	os.Remove(filename)

	return nil
}

// ListSessions lists all sessions
func (m *PiSessionManager) ListSessions() []*PiSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*PiSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}
