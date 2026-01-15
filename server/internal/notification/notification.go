package notification

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// ============ 通知系统 ============

// Channel 通知渠道
type Channel string

const (
	ChannelEmail Channel = "email" // 邮件
	ChannelSMS   Channel = "sms"   // 短信
	ChannelPush  Channel = "push"  // 推送
	ChannelWeb   Channel = "web"   // WebSocket
)

// Level 通知级别
type Level string

const (
	LevelInfo     Level = "info"     // 信息
	LevelWarning  Level = "warning"  // 警告
	LevelError    Level = "error"    // 错误
	LevelCritical Level = "critical"  // 严重
)

// Notification 通知
type Notification struct {
	ID        string                 `json:"id"`
	Channel   Channel                `json:"channel"`
	Level     Level                  `json:"level"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt string                 `json:"created_at"`
	Read      bool                   `json:"read"`
}

// Subscriber 订阅者
type Subscriber struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Webhook  string    `json:"webhook,omitempty"`
	Channels []Channel `json:"channels"`
	Level    Level     `json:"level"`
}

// Handler 渠道处理器
type Handler interface {
	Send(ctx context.Context, notif *Notification, sub *Subscriber) error
}

// Manager 通知管理器
type Manager struct {
	mu           sync.RWMutex
	subscribers   map[string]*Subscriber
	notifications map[string]*Notification
	handlers     map[Channel]Handler
}

// NewManager 创建通知管理器
func NewManager() *Manager {
	return &Manager{
		subscribers:   make(map[string]*Subscriber),
		notifications: make(map[string]*Notification),
		handlers:     make(map[Channel]Handler),
	}
}

// RegisterHandler 注册渠道处理器
func (m *Manager) RegisterHandler(channel Channel, handler Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = handler
}

// AddSubscriber 添加订阅者
func (m *Manager) AddSubscriber(sub *Subscriber) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.ID == "" {
		return fmt.Errorf("subscriber ID is required")
	}

	m.subscribers[sub.ID] = sub
	log.Printf("✅ 订阅者已添加: %s", sub.Name)
	return nil
}

// RemoveSubscriber 移除订阅者
func (m *Manager) RemoveSubscriber(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscribers[id]; !exists {
		return fmt.Errorf("subscriber not found: %s", id)
	}

	delete(m.subscribers, id)
	return nil
}

// ListSubscribers 列出订阅者
func (m *Manager) ListSubscribers() []*Subscriber {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs := make([]*Subscriber, 0, len(m.subscribers))
	for _, sub := range m.subscribers {
		subs = append(subs, sub)
	}
	return subs
}

// Send 发送通知
func (m *Manager) Send(ctx context.Context, level Level, title, content string) error {
	notif := &Notification{
		Level:   level,
		Title:   title,
		Content: content,
	}

	// 保存通知
	m.mu.Lock()
	m.notifications[notif.ID] = notif
	m.mu.Unlock()

	// 发送到各渠道
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sub := range m.subscribers {
		if !m.shouldSend(sub, level) {
			continue
		}

		for _, channel := range sub.Channels {
			if handler, exists := m.handlers[channel]; exists {
				go func(h Handler, s *Subscriber, n *Notification) {
					if err := h.Send(ctx, n, s); err != nil {
						log.Printf("❌ 通知发送失败: %s -> %s - %v", n.Title, s.Name, err)
					} else {
						log.Printf("✅ 通知已发送: %s -> %s", n.Title, s.Name)
					}
				}(handler, sub, notif)
			}
		}
	}

	return nil
}

// shouldSend 检查是否应该发送
func (m *Manager) shouldSend(sub *Subscriber, level Level) bool {
	levels := map[Level]int{
		LevelInfo:     0,
		LevelWarning:  1,
		LevelError:    2,
		LevelCritical: 3,
	}

	return levels[level] >= levels[sub.Level]
}

// GetNotifications 获取通知列表
func (m *Manager) GetNotifications(limit int) []*Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notifs := make([]*Notification, 0, limit)
	for _, n := range m.notifications {
		notifs = append(notifs, n)
		if len(notifs) >= limit {
			break
		}
	}
	return notifs
}

// MarkAsRead 标记为已读
func (m *Manager) MarkAsRead(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	n, exists := m.notifications[id]
	if !exists {
		return fmt.Errorf("notification not found: %s", id)
	}

	n.Read = true
	return nil
}

// LogHandler 日志处理器
type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Send(ctx context.Context, notif *Notification, sub *Subscriber) error {
	log.Printf("[NOTIFICATION] Level: %s, To: %s, Title: %s, Content: %s",
		notif.Level, sub.Name, notif.Title, notif.Content)
	return nil
}

// RegisterBuiltins 注册内置渠道
func (m *Manager) RegisterBuiltins() {
	m.RegisterHandler(ChannelWeb, NewLogHandler())
	log.Printf("📦 通知渠道已注册")
}
