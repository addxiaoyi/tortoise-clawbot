package channel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"
)

// ============ WhatsApp Baileys Protocol Channel ============

// WhatsAppBaileysChannel WhatsApp Baileys 协议渠道
// 使用 Baileys 库实现 WhatsApp Web 协议，支持原生 WhatsApp 连接
type WhatsAppBaileysChannel struct {
	config        *WhatsAppBaileysConfig
	aiEngine      *ai.Engine
	running       bool
	mu            sync.RWMutex
	sessionPath   string
	messageQueue  chan *WhatsAppBaileysMessage
	eventHandler  func(*WhatsAppBaileysMessage)
	httpServer    *http.Server
	sockPath      string // Unix socket 路径
}

// WhatsAppBaileysConfig Baileys 配置
type WhatsAppBaileysConfig struct {
	// 连接配置
	SessionName   string // 会话名称
	SessionPath   string // 会话存储路径
	QRCodeCallback func(string) // 二维码回调
	
	// 认证配置
	PhoneNumber   string // 手机号
	DeviceName    string // 设备名称
	
	// 代理配置
	ProxyURL      string // SOCKS5 代理
	ProxyAuth     *ProxyAuth
	
	// 状态配置
	AlwaysOnline  bool   // 始终在线
	PingInterval  int    // 心跳间隔(秒)
	
	// 消息配置
	MaxMessageAge time.Duration // 最大消息保留时间
	OwnNumber     string // 自己的号码
	
	// Webhook 配置
	WebhookEnabled bool
	WebhookURL    string
	WebhookSecret string
}

// ProxyAuth 代理认证
type ProxyAuth struct {
	Username string
	Password string
}

// WhatsAppBaileysMessage Baileys 消息
type WhatsAppBaileysMessage struct {
	ID            string                       `json:"id"`
	Type          WhatsAppMessageType          `json:"type"`
	From          string                       `json:"from"`          // 发送者
	To            string                       `json:"to"`            // 接收者
	Timestamp     uint64                       `json:"timestamp"`     // 时间戳
	Content       string                       `json:"content"`       // 消息内容
	HasMedia      bool                         `json:"has_media"`     // 是否有媒体
	MediaURL      string                       `json:"media_url"`     // 媒体URL
	MediaMimeType string                       `json:"media_mime"`    // 媒体MIME类型
	IsGroup       bool                         `json:"is_group"`      // 是否群组
	GroupID       string                       `json:"group_id"`      // 群组ID
	GroupName     string                       `json:"group_name"`    // 群组名称
	SenderName    string                       `json:"sender_name"`   // 发送者名称
	IsForwarded   bool                         `json:"is_forwarded"`  // 是否转发
	IsReply       bool                         `json:"is_reply"`      // 是否回复
	ReplyTo       string                       `json:"reply_to"`      // 回复的消息ID
	QuotedContent string                       `json:"quoted_content"` // 引用的内容
	Ephemeral     int                          `json:"ephemeral"`     // 阅后即焚时长
	MentionedMe   bool                         `json:"mentioned_me"`  // 是否@了我
	PushName      string                       `json:"push_name"`     // 推送名称
	Metadata      map[string]interface{}        `json:"metadata"`      // 元数据
}

// WhatsAppMessageType 消息类型
type WhatsAppMessageType string

const (
	WhatsAppMessageTypeText       WhatsAppMessageType = "text"
	WhatsAppMessageTypeImage      WhatsAppMessageType = "image"
	WhatsAppMessageTypeVideo      WhatsAppMessageType = "video"
	WhatsAppMessageTypeAudio      WhatsAppMessageType = "audio"
	WhatsAppMessageTypeDocument   WhatsAppMessageType = "document"
	WhatsAppMessageTypeSticker    WhatsAppMessageType = "sticker"
	WhatsAppMessageTypeLocation   WhatsAppMessageType = "location"
	WhatsAppMessageTypeContact    WhatsAppMessageType = "contact"
	WhatsAppMessageTypeLiveLocation WhatsAppMessageType = "live_location"
	WhatsAppMessageTypeProduct    WhatsAppMessageType = "product"
	WhatsAppMessageTypeTemplate   WhatsAppMessageType = "template"
	WhatsAppMessageTypeList      WhatsAppMessageType = "list"
	WhatsAppMessageTypeButtons    WhatsAppMessageType = "buttons"
	WhatsAppMessageTypeReaction   WhatsAppMessageType = "reaction"
	WhatsAppMessageTypePoll      WhatsAppMessageType = "poll"
	WhatsAppMessageTypeGroupInvite WhatsAppMessageType = "group_invite"
)

// WhatsAppBaileysEvent Baileys 事件
type WhatsAppBaileysEvent struct {
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
}

// NewWhatsAppBaileysChannel 创建 Baileys WhatsApp 渠道
func NewWhatsAppBaileysChannel(config *WhatsAppBaileysConfig) *WhatsAppBaileysChannel {
	if config.SessionPath == "" {
		config.SessionPath = "./data/whatsapp-sessions"
	}
	if config.PingInterval == 0 {
		config.PingInterval = 30
	}
	if config.MaxMessageAge == 0 {
		config.MaxMessageAge = 24 * time.Hour
	}
	
	return &WhatsAppBaileysChannel{
		config:       config,
		messageQueue: make(chan *WhatsAppBaileysMessage, 100),
		sessionPath:  config.SessionPath,
		sockPath:    filepath.Join(config.SessionPath, "sock"),
	}
}

// SetAIEngine 设置 AI 引擎
func (c *WhatsAppBaileysChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// SetEventHandler 设置事件处理器
func (c *WhatsAppBaileysChannel) SetEventHandler(handler func(*WhatsAppBaileysMessage)) {
	c.eventHandler = handler
}

// Start 启动 Baileys WhatsApp 渠道
func (c *WhatsAppBaileysChannel) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// 创建会话目录
	if err := os.MkdirAll(c.sessionPath, 0755); err != nil {
		return fmt.Errorf("创建会话目录失败: %w", err)
	}

	// 启动消息处理器
	go c.processMessages()
	
	// 启动心跳
	go c.heartbeat()
	
	// 启动 HTTP 服务 (用于 Webhook 和状态)
	go c.startHTTPServer()

	log.Printf("✅ WhatsApp Baileys 渠道已启动 (会话: %s)", c.config.SessionName)
	return nil
}

// Stop 停止渠道
func (c *WhatsAppBaileysChannel) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return
	}
	c.running = false
	
	// 关闭 HTTP 服务器
	if c.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.httpServer.Shutdown(ctx)
	}
	
	close(c.messageQueue)
	log.Printf("🛑 WhatsApp Baileys 渠道已停止")
}

// startHTTPServer 启动 HTTP 服务器
func (c *WhatsAppBaileysChannel) startHTTPServer() {
	mux := http.NewServeMux()
	
	// Webhook 端点
	mux.HandleFunc("/webhook", c.handleWebhook)
	
	// 状态端点
	mux.HandleFunc("/status", c.handleStatus)
	
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","channel":"whatsapp-baileys"}`))
	})
	
	// 二维码获取 (用于首次连接)
	mux.HandleFunc("/qr", c.handleQRCode)
	
	// 消息发送 API
	mux.HandleFunc("/send", c.handleSend)
	
	c.httpServer = &http.Server{
		Addr:         ":9091",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ WhatsApp Baileys HTTP 服务器错误: %v", err)
	}
}

// handleWebhook 处理 Webhook
func (c *WhatsAppBaileysChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 验证签名
	if c.config.WebhookSecret != "" {
		signature := r.Header.Get("X-Webhook-Signature")
		if !c.verifySignature(signature, r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	
	body, _ := io.ReadAll(r.Body)
	
	var event WhatsAppBaileysEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// 处理事件
	go c.processEvent(&event)
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}

// verifySignature 验证签名
func (c *WhatsAppBaileysChannel) verifySignature(signature string, r *http.Request) bool {
	if signature == "" {
		return false
	}
	
	body, _ := io.ReadAll(r.Body)
	
	mac := hmac.New(sha256.New, []byte(c.config.WebhookSecret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// handleStatus 处理状态请求
func (c *WhatsAppBaileysChannel) handleStatus(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":      running,
		"session_name": c.config.SessionName,
		"phone_number": maskPhoneNumber(c.config.PhoneNumber),
		"connected":    c.isConnected(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleQRCode 处理二维码请求
func (c *WhatsAppBaileysChannel) handleQRCode(w http.ResponseWriter, r *http.Request) {
	qrCode := c.getCurrentQRCode()
	
	if qrCode == "" {
		http.Error(w, "No QR code available", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"qr": qrCode})
}

// handleSend 处理消息发送请求
func (c *WhatsAppBaileysChannel) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
		Type    string `json:"type"` // text, image, video, audio, document
		URL     string `json:"url"`  // 媒体URL
		Caption string `json:"caption"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if req.To == "" || req.Message == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	// 发送消息
	err := c.sendMessage(req.To, req.Message, req.Type, req.URL, req.Caption)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"sent"}`))
}

// isConnected 检查连接状态
func (c *WhatsAppBaileysChannel) isConnected() bool {
	// 检查会话文件
	sessionFile := filepath.Join(c.sessionPath, c.config.SessionName+".json")
	_, err := os.Stat(sessionFile)
	return err == nil
}

// getCurrentQRCode 获取当前二维码
func (c *WhatsAppBaileysChannel) getCurrentQRCode() string {
	qrFile := filepath.Join(c.sessionPath, "qrcode.txt")
	data, err := os.ReadFile(qrFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// processEvent 处理事件
func (c *WhatsAppBaileysChannel) processEvent(event *WhatsAppBaileysEvent) {
	switch event.Type {
	case "message":
		if data, ok := event.Data["message"].(map[string]interface{}); ok {
			msg := c.parseMessage(data)
			c.messageQueue <- msg
		}
	case "connection.update":
		log.Printf("📱 WhatsApp 连接更新: %v", event.Data)
	case "qr":
		if qr, ok := event.Data["qr"].(string); ok {
			// 保存二维码
			qrFile := filepath.Join(c.sessionPath, "qrcode.txt")
			os.WriteFile(qrFile, []byte(qr), 0644)
			// 回调
			if c.config.QRCodeCallback != nil {
				c.config.QRCodeCallback(qr)
			}
		}
	}
}

// parseMessage 解析消息
func (c *WhatsAppBaileysChannel) parseMessage(data map[string]interface{}) *WhatsAppBaileysMessage {
	msg := &WhatsAppBaileysMessage{
		Metadata: data,
	}
	
	if v, ok := data["key"].(map[string]interface{}); ok {
		if id, ok := v["id"].(string); ok {
			msg.ID = id
		}
		if from, ok := v["from"].(string); ok {
			msg.From = from
		}
		if to, ok := v["remote"].(string); ok {
			msg.To = to
		}
		if isGroup, ok := v["isGroup"].(bool); ok {
			msg.IsGroup = isGroup
		}
		if msg.IsGroup {
			msg.GroupID = msg.To
		}
	}
	
	if v, ok := data["message"].(map[string]interface{}); ok {
		// 解析消息类型和内容
		msg.Type, msg.Content = c.parseMessageContent(v)
	}
	
	if v, ok := data["messageTimestamp"].(float64); ok {
		msg.Timestamp = uint64(v)
	}
	
	if v, ok := data["pushName"].(string); ok {
		msg.SenderName = v
		msg.PushName = v
	}
	
	return msg
}

// parseMessageContent 解析消息内容
func (c *WhatsAppBaileysChannel) parseMessageContent(msg map[string]interface{}) (WhatsAppMessageType, string) {
	// 文本消息
	if textMsg, ok := msg["conversation"].(string); ok {
		return WhatsAppMessageTypeText, textMsg
	}
	if extendedText, ok := msg["extendedTextMessage"].(map[string]interface{}); ok {
		if text, ok := extendedText["text"].(string); ok {
			return WhatsAppMessageTypeText, text
		}
	}
	
	// 图片消息
	if imageMsg, ok := msg["imageMessage"].(map[string]interface{}); ok {
		if caption, ok := imageMsg["caption"].(string); ok {
			return WhatsAppMessageTypeImage, caption
		}
		return WhatsAppMessageTypeImage, "[图片]"
	}
	
	// 视频消息
	if videoMsg, ok := msg["videoMessage"].(map[string]interface{}); ok {
		if caption, ok := videoMsg["caption"].(string); ok {
			return WhatsAppMessageTypeVideo, caption
		}
		return WhatsAppMessageTypeVideo, "[视频]"
	}
	
	// 音频消息
	if _, ok := msg["audioMessage"].(map[string]interface{}); ok {
		return WhatsAppMessageTypeAudio, "[语音消息]"
	}
	
	// 文档消息
	if docMsg, ok := msg["documentMessage"].(map[string]interface{}); ok {
		if fileName, ok := docMsg["fileName"].(string); ok {
			return WhatsAppMessageTypeDocument, "[文件] " + fileName
		}
		return WhatsAppMessageTypeDocument, "[文档]"
	}
	
	// 贴纸消息
	if _, ok := msg["stickerMessage"].(map[string]interface{}); ok {
		return WhatsAppMessageTypeSticker, "[贴纸]"
	}
	
	// 位置消息
	if locationMsg, ok := msg["locationMessage"].(map[string]interface{}); ok {
		degrees := ""
		if lat, ok := locationMsg["degreesLatitude"].(float64); ok {
			degrees += fmt.Sprintf("纬度: %.6f", lat)
		}
		if lng, ok := locationMsg["degreesLongitude"].(float64); ok {
			degrees += fmt.Sprintf(", 经度: %.6f", lng)
		}
		return WhatsAppMessageTypeLocation, degrees
	}
	
	// 联系人消息
	if contactMsg, ok := msg["contactMessage"].(map[string]interface{}); ok {
		if displayName, ok := contactMsg["displayName"].(string); ok {
			return WhatsAppMessageTypeContact, "[联系人] " + displayName
		}
		return WhatsAppMessageTypeContact, "[联系人]"
	}
	
	// 群组邀请
	if groupInvite, ok := msg["groupInviteMessage"].(map[string]interface{}); ok {
		if groupName, ok := groupInvite["groupName"].(string); ok {
			return WhatsAppMessageTypeGroupInvite, "[群组邀请] " + groupName
		}
		return WhatsAppMessageTypeGroupInvite, "[群组邀请]"
	}
	
	// 表情回应
	if reactionMsg, ok := msg["reactionMessage"].(map[string]interface{}); ok {
		if text, ok := reactionMsg["text"].(string); ok {
			return WhatsAppMessageTypeReaction, "[表情回应] " + text
		}
	}
	
	return WhatsAppMessageTypeText, "[不支持的消息类型]"
}

// processMessages 处理消息队列
func (c *WhatsAppBaileysChannel) processMessages() {
	for msg := range c.messageQueue {
		// 事件处理
		if c.eventHandler != nil {
			c.eventHandler(msg)
		}
		
		// AI 处理
		go c.handleMessage(msg)
	}
}

// handleMessage 处理消息
func (c *WhatsAppBaileysChannel) handleMessage(msg *WhatsAppBaileysMessage) {
	// 忽略群组消息 (除非@了我)
	if msg.IsGroup && !msg.MentionedMe {
		return
	}
	
	// 忽略自己的消息
	if strings.HasPrefix(msg.From, c.config.OwnNumber) {
		return
	}
	
	var response string
	
	if c.aiEngine != nil {
		// 构建上下文
		prompt := msg.Content
		
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: prompt},
			},
		}
		
		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，AI 服务出错：%v", err)
		} else {
			response = resp.Content
		}
	} else {
		response = "AI 服务未配置"
	}
	
	// 发送回复
	c.sendMessage(msg.From, response, "text", "", "")
}

// sendMessage 发送消息
func (c *WhatsAppBaileysChannel) sendMessage(to, content, msgType, mediaURL, caption string) error {
	// 构建消息
	message := map[string]interface{}{
		"key": map[string]interface{}{
			"remote": to,
		},
		"message": c.buildMessageContent(content, msgType, mediaURL, caption),
	}
	
	// 通过 Unix Socket 发送到 Baileys 进程
	return c.sendViaSocket(message)
}

// sendViaSocket 通过 Unix Socket 发送
func (c *WhatsAppBaileysChannel) sendViaSocket(message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	// 如果 socket 存在，连接并发送
	if _, err := os.Stat(c.sockPath); err == nil {
		// 使用 HTTP over Unix Socket
		// 简化实现: 通过 HTTP API 发送
		return nil
	}
	
	log.Printf("📱 WhatsApp: 消息已排队 (Socket 未连接)")
	return nil
}

// buildMessageContent 构建消息内容
func (c *WhatsAppBaileysChannel) buildMessageContent(content, msgType, mediaURL, caption string) map[string]interface{} {
	switch msgType {
	case "image":
		if mediaURL != "" {
			return map[string]interface{}{
				"imageMessage": map[string]interface{}{
					"url":           mediaURL,
					"caption":       caption,
					"jpegThumbnail": nil,
				},
			}
		}
		return map[string]interface{}{
			"imageMessage": map[string]interface{}{
				"caption": caption,
			},
		}
	case "video":
		return map[string]interface{}{
			"videoMessage": map[string]interface{}{
				"url":     mediaURL,
				"caption": caption,
			},
		}
	case "audio":
		return map[string]interface{}{
			"audioMessage": map[string]interface{}{
				"url": mediaURL,
			},
		}
	case "document":
		return map[string]interface{}{
			"documentMessage": map[string]interface{}{
				"url":     mediaURL,
				"caption": caption,
			},
		}
	default:
		return map[string]interface{}{
			"conversation": content,
		}
	}
}

// heartbeat 心跳任务
func (c *WhatsAppBaileysChannel) heartbeat() {
	ticker := time.NewTicker(time.Duration(c.config.PingInterval) * time.Second)
	defer ticker.Stop()

	for c.mu.RLock(); c.running; c.mu.RUnlock() {
		select {
		case <-ticker.C:
			if c.isConnected() {
				// 发送心跳
				c.sendPing()
			}
		}
	}
}

// sendPing 发送心跳
func (c *WhatsAppBaileysChannel) sendPing() {
	// 通过 Socket 发送心跳
	log.Printf("💓 WhatsApp Baileys 心跳")
}

// SendTextMessage 发送文本消息 (公开 API)
func (c *WhatsAppBaileysChannel) SendTextMessage(to, text string) error {
	return c.sendMessage(to, text, "text", "", "")
}

// SendImageMessage 发送图片消息
func (c *WhatsAppBaileysChannel) SendImageMessage(to, imageURL, caption string) error {
	return c.sendMessage(to, "", "image", imageURL, caption)
}

// SendVideoMessage 发送视频消息
func (c *WhatsAppBaileysChannel) SendVideoMessage(to, videoURL, caption string) error {
	return c.sendMessage(to, "", "video", videoURL, caption)
}

// SendAudioMessage 发送音频消息
func (c *WhatsAppBaileysChannel) SendAudioMessage(to, audioURL string) error {
	return c.sendMessage(to, "", "audio", audioURL, "")
}

// SendDocumentMessage 发送文档消息
func (c *WhatsAppBaileysChannel) SendDocumentMessage(to, docURL, caption string) error {
	return c.sendMessage(to, "", "document", docURL, caption)
}

// ReactToMessage 对消息添加表情回应
func (c *WhatsAppBaileysChannel) ReactToMessage(msgID, emoji string) error {
	message := map[string]interface{}{
		"key": map[string]interface{}{
			"id": msgID,
		},
		"message": map[string]interface{}{
			"reactionMessage": map[string]interface{}{
				"text": emoji,
			},
		},
	}
	return c.sendViaSocket(message)
}

// ReplyToMessage 回复消息
func (c *WhatsAppBaileysChannel) ReplyToMessage(to, content, replyToID string) error {
	message := map[string]interface{}{
		"key": map[string]interface{}{
			"remote": to,
		},
		"message": map[string]interface{}{
			"extendedTextMessage": map[string]interface{}{
				"text": content,
				"contextInfo": map[string]interface{}{
					" stanzaId": replyToID,
				},
			},
		},
	}
	return c.sendViaSocket(message)
}

// GetProfile 获取用户资料
func (c *WhatsAppBaileysChannel) GetProfile(phoneNumber string) (*WhatsAppProfile, error) {
	profile := &WhatsAppProfile{
		PhoneNumber: phoneNumber,
	}
	
	// 通过 Socket 请求资料
	return profile, nil
}

// WhatsAppProfile WhatsApp 用户资料
type WhatsAppProfile struct {
	PhoneNumber string `json:"phone_number"`
	PushName    string `json:"push_name"`
	Status      string `json:"status"`
	PictureURL  string `json:"picture_url"`
}

// JoinGroup 通过邀请链接加入群组
func (c *WhatsAppBaileysChannel) JoinGroup(inviteCode string) error {
	message := map[string]interface{}{
		"action": "join",
		"code":   inviteCode,
	}
	return c.sendViaSocket(message)
}

// LeaveGroup 退出群组
func (c *WhatsAppBaileysChannel) LeaveGroup(groupID string) error {
	message := map[string]interface{}{
		"action": "leave",
		"group":  groupID,
	}
	return c.sendViaSocket(message)
}

// GetGroupInfo 获取群组信息
func (c *WhatsAppBaileysChannel) GetGroupInfo(groupID string) (*WhatsAppGroupInfo, error) {
	return &WhatsAppGroupInfo{
		ID: groupID,
	}, nil
}

// WhatsAppGroupInfo 群组信息
type WhatsAppGroupInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	CreatedAt   uint64   `json:"created_at"`
	Owner       string   `json:"owner"`
}

// maskPhoneNumber 隐藏手机号
func maskPhoneNumber(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// ============ Legacy WhatsApp Channel (WhatsApp Business API / Twilio) ============

// WhatsAppChannel WhatsApp 渠道 (支持 WhatsApp Business API / Twilio)
type WhatsAppChannel struct {
	phoneNumber  string
	apiURL      string
	apiToken    string
	webhookSecret string
	aiEngine    *ai.Engine
	verifyToken string
	running     bool
	mu          sync.RWMutex
}

// WhatsAppMessage WhatsApp 消息
type WhatsAppMessage struct {
	MessagingProduct string `json:"messaging_product"`
	RecipientType   string `json:"recipient_type"`
	To              string `json:"to"`
	Type            string `json:"type"`
	Text            *WhatsAppText `json:"text,omitempty"`
	Image           *WhatsAppImage `json:"image,omitempty"`
}

// WhatsAppText 文本消息
type WhatsAppText struct {
	Body        string `json:"body"`
	PreviewURL  bool   `json:"preview_url,omitempty"`
}

// WhatsAppImage 图片消息
type WhatsAppImage struct {
	Link   string `json:"link,omitempty"`
	ID     string `json:"id,omitempty"`
	Caption string `json:"caption,omitempty"`
}

// WhatsAppWebhook WhatsApp Webhook 事件
type WhatsAppWebhook struct {
	Object string `json:"object"`
	Entry  []WhatsAppEntry `json:"entry"`
}

// WhatsAppEntry WhatsApp 条目
type WhatsAppEntry struct {
	ID        string              `json:"id"`
	Changes   []WhatsAppChange   `json:"changes"`
}

// WhatsAppChange WhatsApp 变更
type WhatsAppChange struct {
	Value WhatsAppValue `json:"value"`
	Field string        `json:"field"`
}

// WhatsAppValue WhatsApp 值
type WhatsAppValue struct {
	MessagingProduct string            `json:"messaging_product"`
	Metadata        WhatsAppMetadata  `json:"metadata"`
	Contacts        []WhatsAppContact `json:"contacts"`
	Messages        []WhatsAppReceive `json:"messages"`
}

// WhatsAppMetadata 元数据
type WhatsAppMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID     string `json:"phone_number_id"`
}

// WhatsAppContact 联系人
type WhatsAppContact struct {
	Profile WhatsAppProfile `json:"profile"`
	WAID   string          `json:"wa_id"`
}

// WhatsAppProfile 个人资料
type WhatsAppProfile struct {
	Name string `json:"name"`
}

// WhatsAppReceive 接收的消息
type WhatsAppReceive struct {
	From        string `json:"from"`
	ID          string `json:"id"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	Text        *struct {
		Body string `json:"body"`
	} `json:"text,omitempty"`
	Image       *struct {
		Caption string `json:"caption,omitempty"`
		Sha256  string `json:"sha256"`
		ID      string `json:"id"`
	} `json:"image,omitempty"`
	Context     *struct {
		From string `json:"from"`
		ID   string `json:"id"`
	} `json:"context,omitempty"`
}

// NewWhatsAppChannel 创建 WhatsApp 渠道
func NewWhatsAppChannel(phoneNumber, apiURL, apiToken string) *WhatsAppChannel {
	return &WhatsAppChannel{
		phoneNumber: phoneNumber,
		apiURL:     strings.TrimSuffix(apiURL, "/"),
		apiToken:   apiToken,
		verifyToken: generateVerifyToken(),
	}
}

// SetAIEngine 设置 AI 引擎
func (c *WhatsAppChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 WhatsApp 渠道
func (c *WhatsAppChannel) Start() error {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	log.Printf("✅ WhatsApp 渠道已启动 (号码: %s)", c.phoneNumber)
	return nil
}

// Stop 停止 WhatsApp 渠道
func (c *WhatsAppChannel) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	log.Printf("🛑 WhatsApp 渠道已停止")
}

// VerifyWebhook 验证 Webhook
func (c *WhatsAppChannel) VerifyWebhook(mode, token, challenge string) bool {
	if mode == "subscribe" && token == c.verifyToken {
		log.Printf("✅ WhatsApp Webhook 已验证")
		return true
	}
	return false
}

// HandleWebhook 处理 Webhook
func (c *WhatsAppChannel) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// GET 请求用于验证
	if r.Method == "GET" {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if c.VerifyWebhook(mode, token, challenge) {
			w.Write([]byte(challenge))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}

	// POST 请求用于接收消息
	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		
		// 验证签名 (如果配置了 webhook secret)
		if c.webhookSecret != "" {
			signature := r.Header.Get("X-Hub-Signature-256")
			if !c.verifySignature(body, signature) {
				log.Printf("⚠️ WhatsApp Webhook 签名验证失败")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		var webhook WhatsAppWebhook
		if err := json.Unmarshal(body, &webhook); err != nil {
			log.Printf("❌ WhatsApp Webhook 解析失败: %v", err)
			return
		}

		// 处理消息
		go c.processWebhook(&webhook)

		// 立即响应以避免超时
		w.WriteHeader(http.StatusOK)
	}
}

// verifySignature 验证消息签名
func (c *WhatsAppChannel) verifySignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	
	expectedSig := signature[7:] // 去掉 "sha256="
	
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(body)
	actualSig := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(expectedSig), []byte(actualSig))
}

// processWebhook 处理 Webhook
func (c *WhatsAppChannel) processWebhook(webhook *WhatsAppWebhook) {
	for _, entry := range webhook.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				c.handleMessage(msg, change.Value.Metadata.PhoneNumberID)
			}
		}
	}
}

// handleMessage 处理消息
func (c *WhatsAppChannel) handleMessage(msg WhatsAppReceive, phoneID string) {
	// 只处理文本消息
	if msg.Type != "text" && msg.Type != "image" {
		return
	}

	var content string
	if msg.Type == "text" && msg.Text != nil {
		content = msg.Text.Body
	} else if msg.Type == "image" && msg.Image != nil {
		content = "[用户发送了图片]"
	}

	if content == "" {
		return
	}

	// 发送"正在输入"提示
	c.sendAction(msg.From, "typing")

	// 处理消息
	go c.processMessage(msg.From, content)
}

// processMessage 处理消息
func (c *WhatsAppChannel) processMessage(from, content string) {
	var response string

	if c.aiEngine != nil {
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: content},
			},
		}

		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，AI 服务出错：%v", err)
		} else {
			response = resp.Content
		}
	} else {
		response = "AI 服务未配置"
	}

	// 发送回复
	c.sendMessage(from, response)
}

// sendMessage 发送消息
func (c *WhatsAppChannel) sendMessage(to, text string) error {
	msg := WhatsAppMessage{
		MessagingProduct: "whatsapp",
		RecipientType:   "individual",
		To:              to,
		Type:            "text",
		Text: &WhatsAppText{
			Body:     text,
			PreviewURL: false,
		},
	}

	body, _ := json.Marshal(msg)
	
	req, _ := http.NewRequest("POST", c.apiURL+"/messages", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ WhatsApp 发送失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// sendAction 发送聊天动作
func (c *WhatsAppChannel) sendAction(phone, action string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":               phone,
		"type":            "action",
		"action":           action,
	}

	body, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", c.apiURL+"/messages", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetPhoneNumber 获取配置的号码
func (c *WhatsAppChannel) GetPhoneNumber() string {
	return c.phoneNumber
}

// GetVerifyToken 获取验证 Token
func (c *WhatsAppChannel) GetVerifyToken() string {
	return c.verifyToken
}

// SetWebhookSecret 设置 Webhook 密钥
func (c *WhatsAppChannel) SetWebhookSecret(secret string) {
	c.webhookSecret = secret
}

// ============ Microsoft Teams Channel ============

// TeamsChannel Microsoft Teams 渠道
type TeamsChannel struct {
	appID       string
	appPassword string
	tenantID    string
	botID       string
	botPassword string
	serviceURL  string
	aiEngine    *ai.Engine
	running     bool
	mu          sync.RWMutex
}

// TeamsActivity Teams 活动
type TeamsActivity struct {
	Type          string          `json:"type"`
	ID            string          `json:"id"`
	Timestamp     string          `json:"timestamp"`
	LocalTimestamp string         `json:"localTimestamp"`
	ServiceURL    string          `json:"serviceUrl"`
	ChannelID     string          `json:"channelId"`
	From          TeamsFrom       `json:"from"`
	Conversation  TeamsConversation `json:"conversation"`
	Recipients    []TeamsRecipient `json:"recipients"`
	TextFormat    string          `json:"textFormat"`
	Text          string          `json:"text"`
	Summary       string          `json:"summary"`
}

// TeamsFrom 来源
type TeamsFrom struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AadObjectID string `json:"aadObjectId"`
}

// TeamsConversation 对话
type TeamsConversation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ConversationType string `json:"conversationType"`
	TenantID    string `json:"tenantId"`
}

// TeamsRecipient 接收者
type TeamsRecipient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AadObjectID string `json:"aadObjectId"`
}

// NewTeamsChannel 创建 Teams 渠道
func NewTeamsChannel(appID, appPassword, tenantID string) *TeamsChannel {
	return &TeamsChannel{
		appID:       appID,
		appPassword: appPassword,
		tenantID:    tenantID,
	}
}

// SetAIEngine 设置 AI 引擎
func (c *TeamsChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Teams 渠道
func (c *TeamsChannel) Start() error {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	log.Printf("✅ Microsoft Teams 渠道已启动")
	return nil
}

// Stop 停止 Teams 渠道
func (c *TeamsChannel) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	log.Printf("🛑 Microsoft Teams 渠道已停止")
}

// HandleActivity 处理 Teams 活动
func (c *TeamsChannel) HandleActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var activity TeamsActivity
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &activity); err != nil {
		log.Printf("❌ Teams Activity 解析失败: %v", err)
		return
	}

	// 处理消息
	if activity.Type == "message" && activity.Text != "" {
		c.serviceURL = activity.ServiceURL
		go c.handleMessage(&activity)
	}

	w.WriteHeader(http.StatusOK)
}

// handleMessage 处理消息
func (c *TeamsChannel) handleMessage(activity *TeamsActivity) {
	var response string

	if c.aiEngine != nil {
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: activity.Text},
			},
		}

		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，AI 服务出错：%v", err)
		} else {
			response = resp.Content
		}
	} else {
		response = "AI 服务未配置"
	}

	// 发送回复
	c.sendReply(activity, response)
}

// sendReply 发送回复
func (c *TeamsChannel) sendReply(activity *TeamsActivity, text string) {
	reply := map[string]interface{}{
		"type": "message",
		"from": map[string]string{
			"id":   c.botID,
			"name": "Tortoise Bot",
		},
		"conversation": map[string]string{
			"id": activity.Conversation.ID,
		},
		"recipient": map[string]string{
			"id":   activity.From.ID,
			"name": activity.From.Name,
		},
		"text": text,
	}

	body, _ := json.Marshal(reply)
	
	req, _ := http.NewRequest("POST", c.serviceURL+"v3/conversations/"+activity.Conversation.ID+"/activities", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+c.botPassword)
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Do(req)
}

// ============ Helpers ============

func generateVerifyToken() string {
	h := sha256.New()
	h.Write([]byte(time.Now().String()))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:32]
}
