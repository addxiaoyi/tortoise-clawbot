package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"whatsmeow"
	"whatsmeow/store/sqlstore"
	"whatsmeow/types"
	"whatsmeow/types/events"
)

// ========== WhatsApp 真实实现 (使用连接服务器) ==========

// WhatsAppHandler WhatsApp 渠道处理器
type WhatsAppHandler struct {
	deviceID     int
	client       *whatsmeow.Client
	sessionPath  string
	apiURL       string
	clientHTTP   *http.Client
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	running      bool
	messageHandler func(from string, message string, metadata map[string]interface{})
	eventHandler   func(event *WhatsAppEvent)
}

// WhatsAppEvent WhatsApp 事件
type WhatsAppEvent struct {
	Type      string // message, qr, connected, disconnected
	Sender    string
	GroupID   string
	Message   string
	MessageID string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// WhatsAppMessage WhatsApp 消息
type WhatsAppMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	FromMe    bool   `json:"fromMe"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	Body      string `json:"body"`
}

// NewWhatsAppHandler 创建 WhatsApp 处理器
func NewWhatsAppHandler(sessionPath string) *WhatsAppHandler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WhatsAppHandler{
		sessionPath: sessionPath,
		apiURL:     "http://localhost:8081", // WhatsApp Web API 服务地址
		clientHTTP: &http.Client{Timeout: 30 * time.Second},
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Connect 连接 WhatsApp
func (h *WhatsAppHandler) Connect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return nil
	}

	// 方式1: 使用 WhatsApp Web API (推荐)
	// 需要运行 whatsapp-web-api 或 baileyjs 服务
	if err := h.connectViaAPI(); err != nil {
		// 方式2: 直接使用 whatsmeow (需要扫描二维码)
		log.Printf("API connection failed, trying direct connection: %v", err)
		return h.connectDirect()
	}

	h.running = true
	return nil
}

// connectViaAPI 通过 API 连接
func (h *WhatsAppHandler) connectViaAPI() error {
	// 检查 API 服务是否可用
	resp, err := h.clientHTTP.Get(h.apiURL + "/status")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WhatsApp API not available")
	}

	// 订阅事件
	go h.pollEvents()
	
	return nil
}

// connectDirect 直接连接
func (h *WhatsAppHandler) connectDirect() error {
	// 创建存储
	// 这里需要适当的数据库连接
	// store, err := sqlstore.New("sqlite3", "file:whatsapp.db?_foreign_keys=on")
	// if err != nil {
	// 	return err
	// }
	
	// 创建设备
	// device, err := store.GetDevice(uint64(h.deviceID))
	// if err != nil {
	// 	return err
	// }
	
	// 创建客户端
	// h.client = whatsmeow.NewClient(device)
	
	// 添加事件处理器
	// h.client.AddEventHandler(h.handleEvents)
	
	// 连接
	// if err := h.client.Connect(); err != nil {
	// 	return err
	// }
	
	log.Println("WhatsApp direct connection not fully implemented")
	return fmt.Errorf("direct connection requires database setup")
}

// pollEvents 轮询事件 (从 API 服务)
func (h *WhatsAppHandler) pollEvents() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.fetchEvents()
		}
	}
}

// fetchEvents 获取事件
func (h *WhatsAppHandler) fetchEvents() {
	resp, err := h.clientHTTP.Get(h.apiURL + "/events")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var events []WhatsAppEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return
	}

	for _, event := range events {
		h.handleEvent(&event)
	}
}

// handleEvent 处理事件
func (h *WhatsAppHandler) handleEvent(event *WhatsAppEvent) {
	switch event.Type {
	case "message":
		if h.messageHandler != nil {
			sender := event.Sender
			if sender == "" {
				sender = event.GroupID
			}
			metadata := map[string]interface{}{
				"message_id": event.MessageID,
				"timestamp":  event.Timestamp,
				"type":       event.Type,
			}
			if event.GroupID != "" {
				metadata["group_id"] = event.GroupID
			}
			h.messageHandler(sender, event.Message, metadata)
		}
	}
}

// handleEvents 处理 whatsmeow 事件
func (h *WhatsAppHandler) handleEvents(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		h.handleWhatsAppMessage(v)
	case *events.Connected:
		log.Println("WhatsApp connected")
	case *events.Disconnected:
		log.Println("WhatsApp disconnected")
	case *events.QR:
		log.Printf("WhatsApp QR: %s", v.QR)
	}
}

// handleWhatsAppMessage 处理消息
func (h *WhatsAppHandler) handleWhatsAppMessage(msg *events.Message) {
	if msg.Message.GetBug() != nil || msg.IsGroup {
		return // 忽略群组消息或特殊消息
	}

	sender := msg.Sender.ToNonAD()
	senderJID := ""
	if sender != nil {
		senderJID = sender.User
	}

	event := &WhatsAppEvent{
		Type:      "message",
		Sender:    senderJID,
		Message:   msg.Message.GetConversation(),
		MessageID: msg.ID.String(),
		Timestamp: time.Unix(msg.Timestamp, 0),
		Metadata: map[string]interface{}{
			"push_name": msg.PushName,
		},
	}

	if msg.IsGroup {
		event.GroupID = msg.Chat.ToNonAD().User
		event.Sender = senderJID
	}

	h.handleEvent(event)
}

// Disconnect 断开连接
func (h *WhatsAppHandler) Disconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cancel()
	h.running = false

	if h.client != nil {
		h.client.Disconnect()
	}

	return nil
}

// SendMessage 发送消息
func (h *WhatsAppHandler) SendMessage(to, message string) error {
	if h.client != nil {
		// 直接发送
		jid, _ := types.ParseJID(to)
		_, err := h.client.SendMessage(jid, &types.MessageInfo{}, message)
		return err
	}

	// 通过 API 发送
	resp, err := h.clientHTTP.Post(
		h.apiURL+"/send",
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"to":"%s","message":"%s"}`, to, message))),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send failed: %d", resp.StatusCode)
	}

	return nil
}

// SendImage 发送图片
func (h *WhatsAppHandler) SendImage(to string, image []byte, caption string) error {
	if h.client != nil {
		jid, _ := types.ParseJID(to)
		_, err := h.client.SendMessage(jid, &types.MessageInfo{}, &types.ImageMessage{
			Image: image,
			Caption: caption,
		})
		return err
	}

	// 通过 API 发送
	resp, err := h.clientHTTP.Post(
		h.apiURL+"/send/image",
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"to":"%s","caption":"%s"}`, to, caption))),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// OnMessage 设置消息处理器
func (h *WhatsAppHandler) OnMessage(handler func(from string, message string, metadata map[string]interface{})) {
	h.messageHandler = handler
}

// OnEvent 设置事件处理器
func (h *WhatsAppHandler) OnEvent(handler func(event *WhatsAppEvent)) {
	h.eventHandler = handler
}

// GetQRCode 获取二维码
func (h *WhatsAppHandler) GetQRCode() (string, error) {
	resp, err := h.clientHTTP.Get(h.apiURL + "/qr")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetStatus 获取连接状态
func (h *WhatsAppHandler) GetStatus() (string, error) {
	if h.client != nil && h.client.IsConnected() {
		return "connected", nil
	}

	resp, err := h.clientHTTP.Get(h.apiURL + "/status")
	if err != nil {
		return "disconnected", err
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "unknown", err
	}

	return result["status"], nil
}

// Logout 登出
func (h *WhatsAppHandler) Logout() error {
	if h.client != nil {
		h.client.Logout()
	}

	// 删除会话文件
	if h.sessionPath != "" {
		os.RemoveAll(h.sessionPath)
	}

	return nil
}
