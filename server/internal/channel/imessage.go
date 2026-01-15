package channel

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"
)

// ============ iMessage Channel (BlueBubbles Integration) ============

// iMessageChannel iMessage 渠道 (通过 BlueBubbles API)
type iMessageChannel struct {
	config       *iMessageConfig
	aiEngine    *ai.Engine
	running     bool
	mu          sync.RWMutex
	httpClient  *http.Client
	bbServer    string // BlueBubbles Server URL
	httpServer  *http.Server
	messageQueue chan *iMessage
}

// iMessageConfig iMessage 配置
type iMessageConfig struct {
	// BlueBubbles Server 配置
	ServerURL   string // BlueBubbles 服务器 URL (如 https://your-server:1234)
	Password    string // 服务器密码
	
	// 认证
	ServerPassword string
	OCSPPassword  string
	
	// 消息配置
	AutoReply     bool
	ReplyPrefix   string // 自动回复前缀
	MentionOnly   bool   // 只响应@消息
	
	// 通知配置
	NotifyOnMention bool
	NotifyAll      bool
	
	// 过滤配置
	AllowedChats   []string // 允许的聊天 ID
	BlockedChats   []string // 屏蔽的聊天 ID
	
	// 群组配置
	GroupChatEnabled bool
	DMOnly          bool
}

// iMessage iMessage 消息
type iMessage struct {
	ID           string             `json:"id"`
	GUID         string             `json:"guid"`
	Text         string             `json:"text"`
	HandleID     string             `json:"handle_id"`
	SenderName   string             `json:"sender_name"`
	Sender       iMessageSender     `json:"sender"`
	Chat         iMessageChat       `json:"chat"`
	IsFromMe     bool               `json:"is_from_me"`
	IsGroup      bool               `json:"is_group"`
	Timestamp    time.Time          `json:"timestamp"`
	DateDelivered *time.Time       `json:"date_delivered,omitempty"`
	DateRead     *time.Time         `json:"date_read,omitempty"`
	Attachments  []iMessageAttachment `json:"attachments,omitempty"`
	ReplyToGUID  string             `json:"reply_to_guid,omitempty"`
	EffectID     string             `json:"effect_id,omitempty"`
	Subject      string             `json:"subject,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// iMessageSender 发送者
type iMessageSender struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
}

// iMessageChat 聊天
type iMessageChat struct {
	GUID           string `json:"guid"`
	DisplayName    string `json:"display_name"`
	IsGroup        bool   `json:"is_group"`
	Participants    []iMessageParticipant `json:"participants,omitempty"`
}

// iMessageParticipant 聊天参与者
type iMessageParticipant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Admin bool   `json:"admin,omitempty"`
}

// iMessageAttachment 附件
type iMessageAttachment struct {
	GUID         string `json:"guid"`
	HTMLContent  string `json:"html_content,omitempty"`
	Text         string `json:"text,omitempty"`
	MimeType     string `json:"mime_type"`
	FileName     string `json:"file_name"`
	FileSize     int    `json:"file_size"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	TotalChunks  int    `json:"total_chunks,omitempty"`
	BytesTotal   int    `json:"bytes_total,omitempty"`
	TransferName string `json:"transfer_name,omitempty"`
}

// BlueBubblesMessage BlueBubbles API 消息
type BlueBubblesMessage struct {
	Method  string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// BlueBubblesResponse BlueBubbles API 响应
type BlueBubblesResponse struct {
	Status      int             `json:"status"`
	Message     string          `json:"message"`
	Response    json.RawMessage `json:"response"`
}

// NewiMessageChannel 创建 iMessage 渠道
func NewiMessageChannel(config *iMessageConfig) *iMessageChannel {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	return &iMessageChannel{
		config:      config,
		httpClient:  &http.Client{Transport: tr, Timeout: 30 * time.Second},
		bbServer:    strings.TrimSuffix(config.ServerURL, "/"),
		messageQueue: make(chan *iMessage, 100),
	}
}

// SetAIEngine 设置 AI 引擎
func (c *iMessageChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 iMessage 渠道
func (c *iMessageChannel) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// 启动消息处理器
	go c.processMessages()
	
	// 启动 HTTP 服务器 (用于接收 BlueBubbles Webhook)
	go c.startHTTPServer()
	
	// 测试连接
	if err := c.testConnection(); err != nil {
		log.Printf("⚠️ BlueBubbles 连接测试失败: %v", err)
	}

	log.Printf("✅ iMessage 渠道已启动 (BlueBubbles: %s)", c.bbServer)
	return nil
}

// Stop 停止 iMessage 渠道
func (c *iMessageChannel) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return
	}
	c.running = false
	
	if c.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.httpServer.Shutdown(ctx)
	}
	
	close(c.messageQueue)
	log.Printf("🛑 iMessage 渠道已停止")
}

// startHTTPServer 启动 HTTP 服务器
func (c *iMessageChannel) startHTTPServer() {
	mux := http.NewServeMux()
	
	// Webhook 端点
	mux.HandleFunc("/webhook", c.handleWebhook)
	mux.HandleFunc("/webhook/new", c.handleNewMessage)
	mux.HandleFunc("/webhook/read", c.handleReadReceipt)
	
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","channel":"imessage"}`))
	})
	
	// 状态检查
	mux.HandleFunc("/status", c.handleStatus)
	
	c.httpServer = &http.Server{
		Addr:         ":9092",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ iMessage HTTP 服务器错误: %v", err)
	}
}

// handleWebhook 处理 BlueBubbles Webhook
func (c *iMessageChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	
	var msg BlueBubblesMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("❌ iMessage Webhook 解析失败: %v", err)
		return
	}

	// 处理不同类型的消息
	switch msg.Method {
	case "new-message":
		c.handleNewMessageEvent(msg.Params)
	case "read-receipt":
		c.handleReadReceiptEvent(msg.Params)
	case "typing":
		c.handleTypingEvent(msg.Params)
	default:
		log.Printf("📱 iMessage 未知事件: %s", msg.Method)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}

// handleNewMessage 处理新消息 (兼容格式)
func (c *iMessageChannel) handleNewMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	
	var msg iMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("❌ iMessage 解析失败: %v", err)
		return
	}

	c.messageQueue <- &msg
	w.WriteHeader(http.StatusOK)
}

// handleReadReceipt 处理已读回执
func (c *iMessageChannel) handleReadReceipt(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleStatus 处理状态请求
func (c *iMessageChannel) handleStatus(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()
	
	status := map[string]interface{}{
		"running": running,
		"server": c.bbServer,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleNewMessageEvent 处理新消息事件
func (c *iMessageChannel) handleNewMessageEvent(params map[string]interface{}) {
	// 解析消息
	data, _ := json.Marshal(params)
	var msg iMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("❌ iMessage 解析失败: %v", err)
		return
	}

	// 检查是否应该处理
	if !c.shouldProcessMessage(&msg) {
		return
	}

	c.messageQueue <- &msg
}

// handleReadReceiptEvent 处理已读事件
func (c *iMessageChannel) handleReadReceiptEvent(params map[string]interface{}) {
	log.Printf("📱 iMessage 已读回执: %v", params)
}

// handleTypingEvent 处理正在输入事件
func (c *iMessageChannel) handleTypingEvent(params map[string]interface{}) {
	log.Printf("📱 iMessage 正在输入: %v", params)
}

// shouldProcessMessage 检查是否应该处理消息
func (c *iMessageChannel) shouldProcessMessage(msg *iMessage) bool {
	// 忽略自己发送的消息
	if msg.IsFromMe {
		return false
	}

	// DM only 模式
	if c.config.DMOnly && msg.IsGroup {
		return false
	}

	// 检查允许列表
	if len(c.config.AllowedChats) > 0 {
		found := false
		for _, chatID := range c.config.AllowedChats {
			if chatID == msg.Chat.GUID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查屏蔽列表
	for _, chatID := range c.config.BlockedChats {
		if chatID == msg.Chat.GUID {
			return false
		}
	}

	// 检查群组 @mention
	if msg.IsGroup && c.config.MentionOnly {
		// 检查消息是否包含 @ 机器人
		if !strings.Contains(msg.Text, "@") {
			return false
		}
	}

	return true
}

// processMessages 处理消息队列
func (c *iMessageChannel) processMessages() {
	for msg := range c.messageQueue {
		go c.handleMessage(msg)
	}
}

// handleMessage 处理消息
func (c *iMessageChannel) handleMessage(msg *iMessage) {
	var response string

	if c.aiEngine != nil {
		// 构建上下文
		prompt := msg.Text
		if msg.IsGroup {
			prompt = fmt.Sprintf("[群组 %s] %s: %s", msg.Chat.DisplayName, msg.Sender.Name, msg.Text)
		}

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

	// 添加回复前缀
	if c.config.ReplyPrefix != "" {
		response = c.config.ReplyPrefix + " " + response
	}

	// 发送回复
	c.SendMessage(msg.Chat.GUID, response, msg.GUID)
}

// SendMessage 发送消息
func (c *iMessageChannel) SendMessage(chatGUID, text, replyToGUID string) error {
	// 构建请求
	params := map[string]interface{}{
		"chatGuid": chatGUID,
		"text":     text,
	}

	if replyToGUID != "" {
		params["replyToGuid"] = replyToGUID
	}

	msg := BlueBubblesMessage{
		Method:  "send-message",
		Params: params,
	}

	data, _ := json.Marshal(msg)

	// 发送到 BlueBubbles API
	return c.sendToBlueBubbles(data)
}

// SendAttachment 发送附件
func (c *iMessageChannel) SendAttachment(chatGUID, filePath, mimeType string, caption string) error {
	params := map[string]interface{}{
		"chatGuid":  chatGUID,
		"mimeType":  mimeType,
		"filePath":  filePath,
		"transferName": filePath,
	}

	msg := BlueBubblesMessage{
		Method:  "send-attachment",
		Params: params,
	}

	data, _ := json.Marshal(msg)
	return c.sendToBlueBubbles(data)
}

// SendTypingIndicator 发送正在输入指示
func (c *iMessageChannel) SendTypingIndicator(chatGUID string, isTyping bool) error {
	state := "stopped"
	if isTyping {
		state = "started"
	}

	params := map[string]interface{}{
		"chatGuid":  chatGUID,
		"typingState": state,
	}

	msg := BlueBubblesMessage{
		Method:  "update-typing",
		Params: params,
	}

	data, _ := json.Marshal(msg)
	return c.sendToBlueBubbles(data)
}

// MarkAsRead 标记消息为已读
func (c *iMessageChannel) MarkAsRead(chatGUID, messageGUID string) error {
	params := map[string]interface{}{
		"chatGuid": chatGUID,
		"guid":     messageGUID,
	}

	msg := BlueBubblesMessage{
		Method:  "mark-read",
		Params: params,
	}

	data, _ := json.Marshal(msg)
	return c.sendToBlueBubbles(data)
}

// GetChatInfo 获取聊天信息
func (c *iMessageChannel) GetChatInfo(chatGUID string) (*iMessageChat, error) {
	params := map[string]interface{}{
		"chatGuid": chatGUID,
	}

	msg := BlueBubblesMessage{
		Method:  "get-chat",
		Params: params,
	}

	data, _ := json.Marshal(msg)
	resp, err := c.sendToBlueBubblesWithResponse(data)
	if err != nil {
		return nil, err
	}

	var chat iMessageChat
	if err := json.Unmarshal(resp, &chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

// ListChats 获取聊天列表
func (c *iMessageChannel) ListChats() ([]iMessageChat, error) {
	msg := BlueBubblesMessage{
		Method: "get-chats",
	}

	data, _ := json.Marshal(msg)
	resp, err := c.sendToBlueBubblesWithResponse(data)
	if err != nil {
		return nil, err
	}

	var chats []iMessageChat
	if err := json.Unmarshal(resp, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

// sendToBlueBubbles 发送到 BlueBubbles API
func (c *iMessageChannel) sendToBlueBubbles(data []byte) error {
	_, err := c.sendToBlueBubblesWithResponse(data)
	return err
}

// sendToBlueBubblesWithResponse 发送到 BlueBubbles 并获取响应
func (c *iMessageChannel) sendToBlueBubblesWithResponse(data []byte) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/api/v1/message", c.bbServer)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	// 设置认证头
	auth := base64.StdEncoding.EncodeToString([]byte("api:" + c.config.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BlueBubbles API 错误: %d - %s", resp.StatusCode, string(body))
	}

	var bbResp BlueBubblesResponse
	if err := json.Unmarshal(body, &bbResp); err != nil {
		return nil, err
	}

	return bbResp.Response, nil
}

// testConnection 测试 BlueBubbles 连接
func (c *iMessageChannel) testConnection() error {
	apiURL := fmt.Sprintf("%s/api/v1/server", c.bbServer)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte("api:" + c.config.Password))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("连接失败: HTTP %d", resp.StatusCode)
	}

	log.Printf("✅ BlueBubbles 连接测试成功")
	return nil
}

// GetUnreadChats 获取未读聊天
func (c *iMessageChannel) GetUnreadChats() ([]iMessageChat, error) {
	msg := BlueBubblesMessage{
		Method: "get-unread-chats",
	}

	data, _ := json.Marshal(msg)
	resp, err := c.sendToBlueBubblesWithResponse(data)
	if err != nil {
		return nil, err
	}

	var chats []iMessageChat
	if err := json.Unmarshal(resp, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

// Heartbeat 心跳任务
func (c *iMessageChannel) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for c.mu.RLock(); c.running; c.mu.RUnlock() {
		select {
		case <-ticker.C:
			if err := c.testConnection(); err != nil {
				log.Printf("❌ iMessage 连接丢失: %v", err)
			} else {
				log.Printf("💓 iMessage 心跳")
			}
		}
	}
}
