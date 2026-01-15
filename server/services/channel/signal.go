package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ========== Signal 渠道实现 ==========

// SignalHandler Signal 渠道处理器
type SignalHandler struct {
	phoneNumber string
	deviceName  string
	apiURL      string
	authToken   string
	client      *http.Client
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	running     bool
	messageHandler func(from string, message string, metadata map[string]interface{})
}

// SignalAccount Signal 账户
type SignalAccount struct {
	Number     string `json:"number"`
	DeviceName string `json:"device_name"`
	Status     string `json:"status"`
}

// SignalMessage Signal 消息
type SignalMessage struct {
	ID           string            `json:"id"`
	Source       string           `json:"source"`
	Destination  string           `json:"destination"`
	Type         string           `json:"type"` // text, image, etc
	Content      string           `json:"content"`
	Timestamp    int64            `json:"timestamp"`
	Receipt      bool             `json:"receipt"`
	Metadata     map[string]any   `json:"metadata,omitempty"`
}

// SignalMessageRequest 发送消息请求
type SignalMessageRequest struct {
	Message  string   `json:"message"`
	Number  string   `json:"number"`
	Recipients []string `json:"recipients"`
}

// NewSignalHandler 创建 Signal 处理器
func NewSignalHandler(phoneNumber, deviceName, apiURL, authToken string) *SignalHandler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SignalHandler{
		phoneNumber: phoneNumber,
		deviceName:  deviceName,
		apiURL:     apiURL,
		authToken:  authToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect 连接 Signal
func (h *SignalHandler) Connect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return nil
	}

	// 验证账户
	account, err := h.getAccount()
	if err != nil {
		return fmt.Errorf("failed to verify Signal account: %w", err)
	}

	if account.Status != "active" {
		return fmt.Errorf("Signal account not active: %s", account.Status)
	}

	// 启动轮询
	go h.pollMessages()

	h.running = true
	log.Printf("Signal handler connected: %s", h.phoneNumber)
	return nil
}

// Disconnect 断开连接
func (h *SignalHandler) Disconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.cancel()
	h.running = false
	log.Printf("Signal handler disconnected")
	return nil
}

// SendMessage 发送消息
func (h *SignalHandler) SendMessage(to string, message string) error {
	req := SignalMessageRequest{
		Message:    message,
		Number:     h.phoneNumber,
		Recipients: []string{to},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", h.apiURL+"/send", bytes.NewReader(data))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.authToken)

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Signal send failed: %d", resp.StatusCode)
	}

	return nil
}

// SendImage 发送图片
func (h *SignalHandler) SendImage(to string, imageURL string, caption string) error {
	// 实现图片发送
	return nil
}

// OnMessage 设置消息处理器
func (h *SignalHandler) OnMessage(handler func(from string, message string, metadata map[string]interface{})) {
	h.messageHandler = handler
}

// getAccount 获取账户信息
func (h *SignalHandler) getAccount() (*SignalAccount, error) {
	httpReq, err := http.NewRequest("GET", h.apiURL+"/account", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+h.authToken)

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var account SignalAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

// pollMessages 轮询消息
func (h *SignalHandler) pollMessages() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastTimestamp int64

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			messages, err := h.fetchMessages(lastTimestamp)
			if err != nil {
				log.Printf("Signal poll error: %v", err)
				continue
			}

			for _, msg := range messages {
				if msg.Timestamp > lastTimestamp {
					lastTimestamp = msg.Timestamp
				}

				if h.messageHandler != nil && msg.Source != h.phoneNumber {
					metadata := map[string]interface{}{
						"source":      msg.Source,
						"destination": msg.Destination,
						"type":        msg.Type,
						"id":          msg.ID,
					}
					h.messageHandler(msg.Source, msg.Content, metadata)
				}
			}
		}
	}
}

// fetchMessages 获取消息
func (h *SignalHandler) fetchMessages(since int64) ([]SignalMessage, error) {
	url := fmt.Sprintf("%s/messages?since=%d", h.apiURL, since)
	
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+h.authToken)

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var messages []SignalMessage
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, err
	}

	return messages, nil
}
