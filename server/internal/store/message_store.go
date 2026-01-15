package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Message 消息模型
type Message struct {
	ID         string                 `json:"id"`
	SessionID  string                 `json:"sessionId"`
	Role       string                 `json:"role"` // user, assistant, system
	Content    string                 `json:"content"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MessageStore 消息存储
type MessageStore struct {
	messages map[string]*Message
	mu       sync.RWMutex
}

// NewMessageStore 创建消息存储
func NewMessageStore() *MessageStore {
	store := &MessageStore{
		messages: make(map[string]*Message),
	}
	store.createSampleData()
	return store
}

func (m *MessageStore) createSampleData() {
	// 创建一些示例消息
	sessions := NewSessionStore()
	sessionList := sessions.GetSessions()

	if len(sessionList) > 0 {
		sessionID := sessionList[0].ID

		messages := []*Message{
			{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "user",
				Content:   "你好，Tortoise！帮我介绍一下你自己",
				Timestamp: time.Now().Add(-1 * time.Hour),
			},
			{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "assistant",
				Content:   "你好！我是 Tortoise，一个高性能的 AI 代理框架。我基于 OpenClaw 的生态优势，融合了 Hermes 的智能能力，可以帮助你完成各种任务。",
				Timestamp: time.Now().Add(-59 * time.Minute),
			},
			{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "user",
				Content:   "你支持哪些功能？",
				Timestamp: time.Now().Add(-58 * time.Minute),
			},
			{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Role:      "assistant",
				Content:   "我支持多种功能，包括：\n\n• 多渠道消息处理\n• 智能记忆系统\n• 插件扩展系统\n• 多模型路由\n• 实时流式响应\n• WebSocket 通信\n\n需要了解更多吗？",
				Timestamp: time.Now().Add(-57 * time.Minute),
			},
		}

		for _, msg := range messages {
			m.messages[msg.ID] = msg
		}
	}
}

// GetMessages 获取会话的所有消息
func (m *MessageStore) GetMessages(sessionID string, limit int) []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]*Message, 0)
	for _, msg := range m.messages {
		if msg.SessionID == sessionID {
			messages = append(messages, msg)
		}
	}

	// 按时间排序
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].Timestamp.After(messages[j].Timestamp) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}

	// 限制数量
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	return messages
}

// CreateMessage 创建消息
func (m *MessageStore) CreateMessage(sessionID, role, content string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := &Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	m.messages[msg.ID] = msg
	return msg
}

// DeleteMessages 删除会话的所有消息
func (m *MessageStore) DeleteMessages(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, msg := range m.messages {
		if msg.SessionID == sessionID {
			delete(m.messages, id)
		}
	}
}
