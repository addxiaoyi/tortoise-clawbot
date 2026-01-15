package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ============ Webhook 系统 ============

// EventType 事件类型
type EventType string

const (
	EventMessage        EventType = "message"         // 收到消息
	EventMessageSent   EventType = "message.sent"    // 发送消息
	EventSessionCreate EventType = "session.create"   // 创建会话
	EventSessionDelete EventType = "session.delete"   // 删除会话
	EventMemoryCreate EventType = "memory.create"     // 创建记忆
	EventMemoryUpdate EventType = "memory.update"    // 更新记忆
	EventMemoryDelete EventType = "memory.delete"    // 删除记忆
	EventPluginEnable  EventType = "plugin.enable"    // 启用插件
	EventPluginDisable EventType = "plugin.disable"   // 禁用插件
	EventAIRequest    EventType = "ai.request"       // AI 请求
	EventAIResponse   EventType = "ai.response"      // AI 响应
	EventError        EventType = "error"            // 错误事件
	EventChannel      EventType = "channel"          // 渠道事件
)

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Webhook Webhook 配置
type Webhook struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Events     []EventType       `json:"events"`
	Secret     string            `json:"secret,omitempty"`
	Enabled    bool              `json:"enabled"`
	RetryCount int               `json:"retry_count"`
	Timeout    int               `json:"timeout"` // 秒
	Headers    map[string]string `json:"headers,omitempty"`
}

// Manager Webhook 管理器
type Manager struct {
	mu       sync.RWMutex
	webhooks map[string]*Webhook
	client   *http.Client
}

// NewManager 创建 Webhook 管理器
func NewManager() *Manager {
	return &Manager{
		webhooks: make(map[string]*Webhook),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register 注册 Webhook
func (m *Manager) Register(webhook *Webhook) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if webhook.ID == "" {
		return fmt.Errorf("webhook ID is required")
	}
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if len(webhook.Events) == 0 {
		return fmt.Errorf("at least one event type is required")
	}

	if webhook.RetryCount == 0 {
		webhook.RetryCount = 3
	}
	if webhook.Timeout == 0 {
		webhook.Timeout = 30
	}

	m.webhooks[webhook.ID] = webhook
	log.Printf("✅ Webhook 已注册: %s (%s)", webhook.Name, webhook.URL)

	return nil
}

// Unregister 取消注册
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.webhooks[id]; !exists {
		return fmt.Errorf("webhook not found: %s", id)
	}

	delete(m.webhooks, id)
	log.Printf("🗑️ Webhook 已取消注册: %s", id)
	return nil
}

// Get 获取 Webhook
func (m *Manager) Get(id string) *Webhook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.webhooks[id]
}

// List 列出所有 Webhook
func (m *Manager) List() []*Webhook {
	m.mu.RLock()
	defer m.mu.RUnlock()

	webhooks := make([]*Webhook, 0, len(m.webhooks))
	for _, w := range m.webhooks {
		webhooks = append(webhooks, w)
	}
	return webhooks
}

// Enable 启用 Webhook
func (m *Manager) Enable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	w, exists := m.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook not found: %s", id)
	}

	w.Enabled = true
	log.Printf("✅ Webhook 已启用: %s", id)
	return nil
}

// Disable 禁用 Webhook
func (m *Manager) Disable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	w, exists := m.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook not found: %s", id)
	}

	w.Enabled = false
	log.Printf("⏸️ Webhook 已禁用: %s", id)
	return nil
}

// Emit 触发事件
func (m *Manager) Emit(ctx context.Context, eventType EventType, data map[string]interface{}) {
	event := &Event{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, webhook := range m.webhooks {
		if !webhook.Enabled {
			continue
		}

		// 检查是否订阅此事件
		subscribed := false
		for _, et := range webhook.Events {
			if et == eventType || et == "*" {
				subscribed = true
				break
			}
		}
		if !subscribed {
			continue
		}

		// 异步发送
		go m.sendEvent(webhook, event)
	}
}

// sendEvent 发送事件
func (m *Manager) sendEvent(webhook *Webhook, event *Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ Webhook %s: 序列化失败: %v", webhook.ID, err)
		return
	}

	// 重试逻辑
	var lastErr error
	for attempt := 0; attempt <= webhook.RetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
			log.Printf("🔄 Webhook %s: 重试 %d/%d", webhook.ID, attempt, webhook.RetryCount)
		}

		err := m.doRequest(webhook, payload)
		if err == nil {
			log.Printf("✅ Webhook %s: 事件已发送 (类型: %s)", webhook.ID, event.Type)
			return
		}

		lastErr = err
		log.Printf("⚠️ Webhook %s: 发送失败 (尝试 %d): %v", webhook.ID, attempt+1, err)
	}

	log.Printf("❌ Webhook %s: 发送失败，已达到最大重试次数: %v", webhook.ID, lastErr)
}

// doRequest 执行 HTTP 请求
func (m *Manager) doRequest(webhook *Webhook, payload []byte) error {
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Tortoise-Webhook/1.0")
	if len(webhook.Events) > 0 {
		req.Header.Set("X-Tortoise-Event", string(webhook.Events[0]))
	}

	// 添加自定义头
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// 添加签名
	if webhook.Secret != "" {
		signature := generateSignature(payload, webhook.Secret)
		req.Header.Set("X-Tortoise-Signature", signature)
	}

	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(webhook.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateSignature 生成签名
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名
func VerifySignature(payload []byte, signature, secret string) bool {
	expected := generateSignature(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
