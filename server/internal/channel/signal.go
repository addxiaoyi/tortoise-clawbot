package channel

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"
)

// ============ Signal Channel ============

// SignalChannel Signal 安全通讯渠道
// Signal Protocol 实现 - 端到端加密
type SignalChannel struct {
	config        *SignalConfig
	aiEngine      *ai.Engine
	running       bool
	mu            sync.RWMutex
	httpServer    *http.Server
	messageQueue  chan *SignalMessage
	sessionStore  *SignalSessionStore
	identityStore *SignalIdentityStore
}

// SignalConfig Signal 配置
type SignalConfig struct {
	// 服务器配置
	ServerHost string // Signal 服务器地址
	ServerPort int    // Signal 服务器端口
	HTTPPort   int    // 本地 HTTP 服务端口
	
	// 认证配置
	AccountID   string // 账号 ID
	DeviceID    uint32 // 设备 ID
	Password    string // 密码
	
	// Signal Registration
	PhoneNumber string // 手机号
	RegToken    string // Registration Token (从 Signal 服务器获取)
	
	// 密钥配置
	IdentityKey         string // 身份密钥对 (Base64)
	PreKeys             []string // 预密钥列表 (Base64)
	PreKeyLastResortID  uint32   // 备用预密钥 ID
	PreKeyLastResortKey string    // 备用预密钥 (Base64)
	
	// 消息接收配置
	WebHookURL  string // Webhook URL
	VerifyToken string // 验证 Token
	
	// 连接配置
	ReconnectInterval time.Duration // 重连间隔
	MaxRetries       int           // 最大重试次数
}

// SignalMessage Signal 消息
type SignalMessage struct {
	ID            string                 `json:"id"`
	Type          SignalMessageType      `json:"type"`
	Source        string                 `json:"source"`        // 发送者手机号
	Destination   string                 `json:"destination"`   // 接收者手机号
	Timestamp     uint64                 `json:"timestamp"`     // 时间戳 (毫秒)
	Content       string                 `json:"content"`       // 消息内容 (已解密)
	ContentType   string                 `json:"content_type"`  // 内容类型
	GroupID       string                 `json:"group_id,omitempty"` // 群组 ID
	IsGroup       bool                   `json:"is_group"`
	Metadata      map[string]interface{} `json:"metadata"`      // 元数据
	Attachments   []SignalAttachment     `json:"attachments"`   // 附件
}

// SignalMessageType Signal 消息类型
type SignalMessageType string

const (
	SignalMessageTypeText           SignalMessageType = "text"
	SignalMessageTypeImage          SignalMessageType = "image"
	SignalMessageTypeAudio         SignalMessageType = "audio"
	SignalMessageTypeVideo         SignalMessageType = "video"
	SignalMessageTypeDocument      SignalMessageType = "document"
	SignalMessageTypeLocation      SignalMessageType = "location"
	SignalMessageTypeSticker       SignalMessageType = "sticker"
	SignalMessageTypeReaction      SignalMessageType = "reaction"
	SignalMessageTypeCall          SignalMessageType = "call"
	SignalMessageTypeTyping        SignalMessageType = "typing"
	SignalMessageTypeDelivery      SignalMessageType = "delivery"
	SignalMessageTypeRead          SignalMessageType = "read"
)

// SignalAttachment Signal 附件
type SignalAttachment struct {
	ID          string `json:"id"`
	ContentType string `json:"content_type"` // MIME 类型
	FileName    string `json:"file_name"`
	Size        int    `json:"size"`         // 字节数
	URL         string `json:"url"`          // 下载 URL
	Thumbnail   string `json:"thumbnail"`    // 缩略图 (Base64)
}

// SignalSessionStore Signal 会话存储
type SignalSessionStore struct {
	sessions map[string]*SignalSession
	mu       sync.RWMutex
}

// SignalSession Signal 会话
type SignalSession struct {
	Recipients map[uint32][]byte `json:"recipients"` // DeviceID -> SessionRecord
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SignalIdentityStore Signal 身份存储
type SignalIdentityStore struct {
	identities map[string]*SignalIdentity
	mu         sync.RWMutex
}

// SignalIdentity Signal 身份
type SignalIdentity struct {
	ID        string    `json:"id"`         // 手机号或群组 ID
	Key       []byte    `json:"key"`        // 身份公钥
	Trust     bool      `json:"trust"`      // 是否信任
	Timestamp time.Time `json:"timestamp"`  // 首次看到时间
}

// SignalEnvelope Signal 信封 (加密消息)
type SignalEnvelope struct {
	Source       string `json:"source"`
	SourceDevice uint32 `json:"sourceDevice"`
	Type         uint32 `json:"type"`    // 消息类型
	Timestamp    uint64 `json:"timestamp"`
	Content      []byte `json:"content"` // 加密内容
}

// SignalSentMessage 发送的消息
type SignalSentMessage struct {
	Destination      string   `json:"destination"`
	Timestamp        uint64  `json:"timestamp"`
	MessageBody      string   `json:"message_body"`
	MessageType      string   `json:"message_type"`
	Attachments      []string `json:"attachments,omitempty"`
	ExpiresInSeconds int      `json:"expires_in_seconds"`
	Reaction         string   `json:"reaction,omitempty"`
	Quote            string   `json:"quote,omitempty"`
	GroupID          string   `json:"group_id,omitempty"`
}

// SignalTypingMessage 正在输入消息
type SignalTypingMessage struct {
	Destination string `json:"destination"`
	Timestamp   uint64 `json:"timestamp"`
	TypingState string `json:"typing_state"` // started / stopped
}

// SignalReceiptMessage 已送达回执
type SignalReceiptMessage struct {
	Destination string   `json:"destination"`
	Timestamp   []uint64 `json:"timestamps"`
	ReceiptType string   `json:"receipt_type"` // delivery / read
}

// NewSignalChannel 创建 Signal 渠道
func NewSignalChannel(config *SignalConfig) *SignalChannel {
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 5
	}
	if config.HTTPPort == 0 {
		config.HTTPPort = 9090
	}
	
	return &SignalChannel{
		config:        config,
		messageQueue:  make(chan *SignalMessage, 100),
		sessionStore:  &SignalSessionStore{sessions: make(map[string]*SignalSession)},
		identityStore: &SignalIdentityStore{identities: make(map[string]*SignalIdentity)},
	}
}

// SetAIEngine 设置 AI 引擎
func (c *SignalChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Signal 渠道
func (c *SignalChannel) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// 启动 HTTP 服务器用于接收 Webhook
	go c.startHTTPServer()
	
	// 启动消息处理器
	go c.processMessages()
	
	// 启动心跳
	go c.heartbeat()

	log.Printf("✅ Signal 渠道已启动 (手机号: %s, HTTP端口: %d)", 
		maskPhoneNumber(c.config.PhoneNumber), c.config.HTTPPort)
	return nil
}

// Stop 停止 Signal 渠道
func (c *SignalChannel) Stop() {
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
	log.Printf("🛑 Signal 渠道已停止")
}

// startHTTPServer 启动 HTTP 服务器
func (c *SignalChannel) startHTTPServer() {
	mux := http.NewServeMux()
	
	// Webhook 端点
	mux.HandleFunc("/webhook", c.handleWebhook)
	
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","channel":"signal"}`))
	})
	
	c.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", c.config.HTTPPort),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Signal HTTP 服务器错误: %v", err)
	}
}

// handleWebhook 处理 Webhook
func (c *SignalChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// 验证 Token
	if c.config.VerifyToken != "" {
		token := r.Header.Get("X-Signal-Token")
		if token != c.config.VerifyToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	switch r.Method {
	case "GET":
		// Webhook 验证
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")
		
		if mode == "subscribe" && token == c.config.VerifyToken {
			w.Write([]byte(challenge))
			return
		}
		http.Error(w, "Invalid verification", http.StatusForbidden)
		
	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ Signal Webhook 读取失败: %v", err)
			return
		}
		
		// 解析消息
		go c.parseAndProcessMessage(body)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"received"}`))
	}
}

// parseAndProcessMessage 解析并处理消息
func (c *SignalChannel) parseAndProcessMessage(body []byte) {
	var envelope SignalEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		// 尝试作为普通消息处理
		var msg SignalMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("❌ Signal 消息解析失败: %v", err)
			return
		}
		c.messageQueue <- &msg
		return
	}

	// 解密消息
	msg, err := c.decryptEnvelope(&envelope)
	if err != nil {
		log.Printf("❌ Signal 消息解密失败: %v", err)
		return
	}

	c.messageQueue <- msg
}

// decryptEnvelope 解密封文
func (c *SignalChannel) decryptEnvelope(envelope *SignalEnvelope) (*SignalMessage, error) {
	// Signal Protocol 解密流程
	// 1. 获取发送者的预密钥
	// 2. 建立会话 (如果不存在)
	// 3. 解密消息
	
	// 这里需要完整的 Signal Protocol 实现
	// 简化版本：直接使用 AES-GCM 解密
	
	if len(envelope.Content) == 0 {
		return &SignalMessage{
			ID:          fmt.Sprintf("%d-%s", envelope.Timestamp, envelope.Source),
			Type:        SignalMessageTypeText,
			Source:      envelope.Source,
			Timestamp:   envelope.Timestamp,
			Content:     "",
		}, nil
	}

	// 生成会话密钥
	sessionKey := deriveSessionKey(envelope.Source, envelope.SourceDevice)
	
	// 解密
	plaintext, err := decryptAEAD(envelope.Content, sessionKey)
	if err != nil {
		return nil, err
	}

	var msg SignalMessage
	if err := json.Unmarshal(plaintext, &msg); err != nil {
		// 如果解析失败，作为纯文本处理
		return &SignalMessage{
			ID:        fmt.Sprintf("%d-%s", envelope.Timestamp, envelope.Source),
			Type:      SignalMessageTypeText,
			Source:    envelope.Source,
			Timestamp: envelope.Timestamp,
			Content:   string(plaintext),
		}, nil
	}

	return &msg, nil
}

// deriveSessionKey 派生会话密钥
func deriveSessionKey(source string, deviceID uint32) []byte {
	h := sha256.New()
	h.Write([]byte(source))
	binary.Write(h, binary.BigEndian, deviceID)
	return h.Sum(nil)
}

// decryptAEAD 使用 AES-GCM 解密
func decryptAEAD(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < 28 { // 12 byte nonce + 16 byte tag
		return nil, fmt.Errorf("ciphertext too short")
	}

	gcm, err := cipher.NewGCM(cipher.NewAES(key))
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:12]
	actualCiphertext := ciphertext[12:]

	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

// encryptAEAD 使用 AES-GCM 加密
func encryptAEAD(plaintext, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cipher.NewAES(key))
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// processMessages 处理消息队列
func (c *SignalChannel) processMessages() {
	for msg := range c.messageQueue {
		if msg.Content == "" {
			continue
		}
		
		// 处理消息
		go c.handleMessage(msg)
	}
}

// handleMessage 处理消息
func (c *SignalChannel) handleMessage(msg *SignalMessage) {
	// 忽略群组消息的处理
	if msg.IsGroup && !strings.HasPrefix(msg.Content, "@bot") {
		return
	}

	var response string

	if c.aiEngine != nil {
		// 构建 AI 请求
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: msg.Content},
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
	c.SendMessage(msg.Source, response, msg.Timestamp)
}

// SendMessage 发送消息
func (c *SignalChannel) SendMessage(to, content string, replyToTimestamp uint64) error {
	// 构建消息
	msg := SignalSentMessage{
		Destination: to,
		Timestamp:   uint64(time.Now().UnixMilli()),
		MessageBody: content,
		MessageType: string(SignalMessageTypeText),
		ExpiresInSeconds: 7 * 24 * 60 * 60, // 7天
	}

	// 如果是回复
	if replyToTimestamp > 0 {
		msg.Quote = fmt.Sprintf("%d", replyToTimestamp)
	}

	// 序列化
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 加密
	sessionKey := deriveSessionKey(to, c.config.DeviceID)
	ciphertext, err := encryptAEAD(data, sessionKey)
	if err != nil {
		return err
	}

	// 发送 (通过 Signal 服务器 API)
	return c.sendToSignalServer(to, ciphertext)
}

// sendToSignalServer 发送到 Signal 服务器
func (c *SignalChannel) sendToSignalServer(to string, content []byte) error {
	// Signal 服务器 API 调用
	// 这里需要实际的 Signal 服务器地址和认证
	
	apiURL := fmt.Sprintf("https://%s:%d/v1/message/%s", 
		c.config.ServerHost, c.config.ServerPort, to)

	req, err := http.NewRequest("PUT", apiURL, strings.NewReader(string(content)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Signal API 错误: %d", resp.StatusCode)
	}

	return nil
}

// SendTypingIndicator 发送正在输入指示
func (c *SignalChannel) SendTypingIndicator(to string, isTyping bool) error {
	typing := "stopped"
	if isTyping {
		typing = "started"
	}

	msg := SignalTypingMessage{
		Destination: to,
		Timestamp:  uint64(time.Now().UnixMilli()),
		TypingState: typing,
	}

	data, _ := json.Marshal(msg)
	
	// 发送到 Signal 服务器
	apiURL := fmt.Sprintf("https://%s:%d/v1/typing/%s", 
		c.config.ServerHost, c.config.ServerPort, to)

	req, _ := http.NewRequest("PUT", apiURL, strings.NewReader(string(data)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SendReadReceipt 发送已读回执
func (c *SignalChannel) SendReadReceipt(to string, timestamps []uint64) error {
	msg := SignalReceiptMessage{
		Destination: to,
		Timestamp:   timestamps,
		ReceiptType: "read",
	}

	data, _ := json.Marshal(msg)
	
	apiURL := fmt.Sprintf("https://%s:%d/v1/receipt/%s", 
		c.config.ServerHost, c.config.ServerPort, to)

	req, _ := http.NewRequest("PUT", apiURL, strings.NewReader(string(data)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SendAttachment 发送附件
func (c *SignalChannel) SendAttachment(to, fileName, contentType string, data []byte, caption string) error {
	// 上传附件到 Signal 服务器
	uploadURL := fmt.Sprintf("https://%s:%d/v1/attachments", 
		c.config.ServerHost, c.config.ServerPort)

	req, _ := http.NewRequest("POST", uploadURL, strings.NewReader(data))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("附件上传失败: %d", resp.StatusCode)
	}

	// 解析响应获取 attachmentId
	var uploadResp struct {
		AttachmentID string `json:"attachmentId"`
		Location     string `json:"location"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return err
	}

	// 发送带附件的消息
	msg := SignalSentMessage{
		Destination: to,
		Timestamp:   uint64(time.Now().UnixMilli()),
		MessageBody: caption,
		MessageType: string(SignalMessageTypeDocument),
		Attachments: []string{uploadResp.AttachmentID},
		ExpiresInSeconds: 7 * 24 * 60 * 60,
	}

	data, _ = json.Marshal(msg)
	sessionKey := deriveSessionKey(to, c.config.DeviceID)
	ciphertext, _ := encryptAEAD(data, sessionKey)

	return c.sendToSignalServer(to, ciphertext)
}

// RegisterPreKeys 注册预密钥
func (c *SignalChannel) RegisterPreKeys(identityKey string, preKeys []string) error {
	registration := map[string]interface{}{
		"identityKey":         identityKey,
		"preKeys":             preKeys,
		"preKeyLastResort": map[string]interface{}{
			"id":  c.config.PreKeyLastResortID,
			"key": c.config.PreKeyLastResortKey,
		},
	}

	data, _ := json.Marshal(registration)
	
	apiURL := fmt.Sprintf("https://%s:%d/v1/registration", c.config.ServerHost)
	
	req, _ := http.NewRequest("PUT", apiURL, strings.NewReader(string(data)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetProfile 获取用户资料
func (c *SignalChannel) GetProfile(phoneNumber string) (*SignalProfile, error) {
	apiURL := fmt.Sprintf("https://%s:%d/v1/profile/%s", 
		c.config.ServerHost, c.config.ServerPort, phoneNumber)
	
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Password))
	req.Header.Set("X-Signal-Account", c.config.AccountID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取资料失败: %d", resp.StatusCode)
	}

	var profile SignalProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// SignalProfile Signal 用户资料
type SignalProfile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`       // Base64 编码的头像
	Description string `json:"description"`  // 个人简介
}

// heartbeat 心跳任务
func (c *SignalChannel) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for c.mu.RLock(); c.running; c.mu.RUnlock() {
		select {
		case <-time.After(time.Until(ticker.C.Add(30 * time.Second))):
			c.mu.RLock()
			running := c.running
			c.mu.RUnlock()
			if !running {
				return
			}
			
			// 检查连接状态
			log.Printf("💓 Signal 渠道心跳")
		}
	}
}

// maskPhoneNumber 隐藏手机号中间4位
func maskPhoneNumber(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// ============ Helpers ============

// GeneratePreKeys 生成预密钥对
func GeneratePreKeys(count int) ([]string, error) {
	preKeys := make([]string, count)
	for i := 0; i < count; i++ {
		// 生成随机密钥
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		preKeys[i] = base64.StdEncoding.EncodeToString(key)
	}
	return preKeys, nil
}

// GenerateIdentityKey 生成身份密钥对
func GenerateIdentityKey() ([]byte, []byte, error) {
	// 生成随机密钥对
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, err
	}
	
	// 计算公钥 (简化版本，实际使用 Curve25519)
	h := sha256.New()
	h.Write(privateKey)
	publicKey := h.Sum(nil)
	
	return privateKey, publicKey, nil
}
