package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// DiscordConfig Discord 配置
type DiscordConfig struct {
	BotToken    string
	GuildID     string // 可选的服务器 ID
	Intents     int    // Gateway Intents
	ProxyURL    string // 可选的代理 URL
}

// DiscordChannel Discord 渠道实现
type DiscordChannel struct {
	config    DiscordConfig
	client    *http.Client
	connected bool
	mu        sync.RWMutex

	// WebSocket 连接
	gateway   string
	ws        *websocket.Conn
	wsMu      sync.Mutex
	
	// 消息处理
	messageHandler func(msg *DiscordMessage) error
	
	// 会话信息
	sessionID  string
	lastSeq    int
	
	// 上下文控制
	ctx    context.Context
	cancel context.CancelFunc
	
	// 心跳
	heartbeatInterval time.Duration
	heartbeatAck      chan bool
}

// NewDiscordChannel 创建 Discord 渠道
func NewDiscordChannel(token, guildID string) *DiscordChannel {
	cfg := DiscordConfig{
		BotToken: token,
		GuildID:  guildID,
		Intents:  // 默认 Intents
			GatewayIntentGuildMessages |
			GatewayIntentDirectMessages |
			GatewayIntentMessageContent,
	}

	// 优先使用环境变量
	if botToken := os.Getenv("DISCORD_BOT_TOKEN"); botToken != "" {
		cfg.BotToken = botToken
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DiscordChannel{
		config:       cfg,
		client:       &http.Client{Timeout: 30 * time.Second},
		ctx:          ctx,
		cancel:       cancel,
		heartbeatAck: make(chan bool, 1),
	}
}

// Connect 连接到 Discord Gateway
func (c *DiscordChannel) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.BotToken == "" {
		return fmt.Errorf("Discord bot token is required")
	}

	// 获取 Gateway URL
	if err := c.getGateway(); err != nil {
		c.connected = false
		return fmt.Errorf("failed to get Discord gateway: %w", err)
	}

	// 验证 token
	if err := c.verifyToken(); err != nil {
		c.connected = false
		return fmt.Errorf("failed to verify Discord bot token: %w", err)
	}

	// 建立 WebSocket 连接
	if err := c.connectWebSocket(); err != nil {
		c.connected = false
		return fmt.Errorf("failed to connect to Discord gateway: %w", err)
	}

	c.connected = true
	
	// 启动消息处理协程
	go c.handleEvents()
	
	return nil
}

// getGateway 获取 Gateway URL
func (c *DiscordChannel) getGateway() error {
	resp, err := c.callDiscordAPI("GET", "/gateway/bot", nil)
	if err != nil {
		return err
	}

	var gatewayResp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(resp, &gatewayResp); err != nil {
		return fmt.Errorf("failed to parse gateway response: %w", err)
	}

	c.gateway = gatewayResp.URL + "?v=10&encoding=json"
	return nil
}

// verifyToken 验证 Bot Token
func (c *DiscordChannel) verifyToken() error {
	_, err := c.callDiscordAPI("GET", "/users/@me", nil)
	return err
}

// connectWebSocket 建立 WebSocket 连接
func (c *DiscordChannel) connectWebSocket() error {
	// 解析 URL
	u := c.gateway

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return fmt.Errorf("failed to dial gateway: %w", err)
	}
	c.ws = ws

	// 等待 Hello 事件
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read hello: %w", err)
	}

	var hello struct {
		Op   int `json:"op"`
		D    struct {
			HeartbeatInterval int `json:"heartbeat_interval"`
		} `json:"d"`
	}
	if err := json.Unmarshal(msg, &hello); err != nil {
		return fmt.Errorf("failed to parse hello: %w", err)
	}

	if hello.Op != 10 { // Hello opcode
		return fmt.Errorf("expected hello, got opcode %d", hello.Op)
	}

	c.heartbeatInterval = time.Duration(hello.D.HeartbeatInterval) * time.Millisecond

	// 发送 Identify
	identify := map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token": c.config.BotToken,
			"intents": c.config.Intents,
			"properties": map[string]interface{}{
				"os":      "windows",
				"browser": "tortoise",
				"device":  "tortoise",
			},
		},
	}
	if err := ws.WriteJSON(identify); err != nil {
		return fmt.Errorf("failed to send identify: %w", err)
	}

	// 启动心跳
	go c.startHeartbeat()

	return nil
}

// startHeartbeat 启动心跳
func (c *DiscordChannel) startHeartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心跳
func (c *DiscordChannel) sendHeartbeat() {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	if c.ws == nil {
		return
	}

	heartbeat := map[string]interface{}{
		"op": 1,
		"d":  c.lastSeq,
	}

	if err := c.ws.WriteJSON(heartbeat); err != nil {
		fmt.Printf("Failed to send heartbeat: %v\n", err)
	}
}

// handleEvents 处理事件
func (c *DiscordChannel) handleEvents() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, msg, err := c.ws.ReadMessage()
			if err != nil {
				if c.ctx.Err() != nil {
					return
				}
				fmt.Printf("Discord read error: %v\n", err)
				time.Sleep(5 * time.Second)
				// 尝试重连
				c.reconnect()
				return
			}

			c.processMessage(msg)
		}
	}
}

// processMessage 处理消息
func (c *DiscordChannel) processMessage(data []byte) {
	var msg DiscordWSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// 更新序列号
	if msg.S != nil {
		c.lastSeq = *msg.S
	}

	switch msg.Op {
	case 0: // Dispatch
		switch msg.T {
		case "MESSAGE_CREATE":
			c.handleMessageCreate(msg.D)
		case "READY":
			c.handleReady(msg.D)
		}
	case 1: // Heartbeat
		c.sendHeartbeat()
	case 7: // Reconnect
		c.reconnect()
	case 11: // Heartbeat ACK
		select {
		case c.heartbeatAck <- true:
		default:
		}
	}
}

// handleMessageCreate 处理消息创建
func (c *DiscordChannel) handleMessageCreate(data json.RawMessage) {
	var msg DiscordMessageData
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// 忽略 Bot 消息
	if msg.Author.Bot {
		return
	}

	// 忽略空消息
	if msg.Content == "" && len(msg.Attachments) == 0 {
		return
	}

	dmMsg := &DiscordMessage{
		ID:            msg.ID,
		ChannelID:     msg.ChannelID,
		GuildID:       msg.GuildID,
		AuthorID:      msg.Author.ID,
		AuthorName:    msg.Author.Username,
		Content:       msg.Content,
		Timestamp:     msg.Timestamp,
		ChannelType:   msg.ChannelType,
	}

	// 调用处理器
	if c.messageHandler != nil {
		c.messageHandler(dmMsg)
	}

	// 转换为通用消息格式
	message := &Message{
		ID:        uuid.New().String(),
		Channel:   ChannelDiscord,
		From:      msg.Author.ID,
		To:        msg.ChannelID,
		Content:   msg.Content,
		Type:      "discord",
		Timestamp: msg.Timestamp,
		Metadata: map[string]interface{}{
			"channel_id":  msg.ChannelID,
			"guild_id":    msg.GuildID,
			"author_name": msg.Author.Username,
			"channel_type": msg.ChannelType,
		},
	}
	QueueMessage(message)
}

// handleReady 处理就绪事件
func (c *DiscordChannel) handleReady(data json.RawMessage) {
	var ready struct {
		SessionID string `json:"session_id"`
	}
	json.Unmarshal(data, &ready)
	c.sessionID = ready.SessionID
}

// reconnect 重连
func (c *DiscordChannel) reconnect() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	// 关闭旧连接
	if c.ws != nil {
		c.ws.Close()
	}

	// 重试连接
	for i := 0; i < 5; i++ {
		if err := c.connectWebSocket(); err == nil {
			c.mu.Lock()
			c.connected = true
			c.mu.Unlock()
			return
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
}

// Disconnect 断开连接
func (c *DiscordChannel) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()
	c.connected = false

	if c.ws != nil {
		c.ws.Close()
		c.ws = nil
	}

	return nil
}

// Send 发送消息
func (c *DiscordChannel) Send(msg *Message) error {
	if !c.IsConnected() {
		return fmt.Errorf("Discord channel not connected")
	}

	// 获取目标 Channel ID
	channelID := msg.To
	if channelID == "" {
		if msg.Metadata != nil {
			if chID, ok := msg.Metadata["channel_id"]; ok {
				channelID = chID.(string)
			}
		}
	}

	if channelID == "" {
		return fmt.Errorf("channel_id is required")
	}

	params := map[string]interface{}{
		"content": msg.Content,
	}

	_, err := c.callDiscordAPI("POST", fmt.Sprintf("/channels/%s/messages", channelID), params)
	return err
}

// SendEmbed 发送嵌入消息
func (c *DiscordChannel) SendEmbed(channelID string, embed DiscordEmbed) error {
	params := map[string]interface{}{
		"embeds": []DiscordEmbed{embed},
	}

	_, err := c.callDiscordAPI("POST", fmt.Sprintf("/channels/%s/messages", channelID), params)
	return err
}

// IsConnected 检查是否已连接
func (c *DiscordChannel) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Type 返回渠道类型
func (c *DiscordChannel) Type() ChannelType {
	return ChannelDiscord
}

// SetMessageHandler 设置消息处理器
func (c *DiscordChannel) SetMessageHandler(handler func(msg *DiscordMessage) error) {
	c.messageHandler = handler
}

// callDiscordAPI 调用 Discord REST API
func (c *DiscordChannel) callDiscordAPI(method, endpoint string, params map[string]interface{}) ([]byte, error) {
	url := fmt.Sprintf("https://discord.com/api/v10%s", endpoint)

	var body io.Reader
	if params != nil {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.config.BotToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DiscordBot (tortoise, 1.0.0)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode >= 400 {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Discord API error %d: %s", resp.StatusCode, errResp.Message)
	}

	return respBody, nil
}

// UpdateConfig 更新配置
func (c *DiscordChannel) UpdateConfig(cfg DiscordConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cfg.BotToken != "" {
		c.config.BotToken = cfg.BotToken
	}
	if cfg.GuildID != "" {
		c.config.GuildID = cfg.GuildID
	}
}

// ==================== Discord API 类型定义 ====================

// DiscordMessage Discord 消息
type DiscordMessage struct {
	ID          string
	ChannelID   string
	GuildID     string
	AuthorID    string
	AuthorName  string
	Content     string
	Timestamp   time.Time
	ChannelType int
}

// DiscordWSMessage WebSocket 消息
type DiscordWSMessage struct {
	Op   int             `json:"op"`
	T    string           `json:"t"`
	S    *int            `json:"s"`
	D    json.RawMessage  `json:"d"`
}

// DiscordMessageData 消息数据
type DiscordMessageData struct {
	ID             string           `json:"id"`
	ChannelID      string           `json:"channel_id"`
	GuildID        string           `json:"guild_id,omitempty"`
	Content        string           `json:"content"`
	Timestamp      time.Time        `json:"timestamp"`
	ChannelType    int              `json:"channel_type"`
	Author         DiscordUser      `json:"author"`
	Attachments    []DiscordAttachment `json:"attachments"`
}

// DiscordUser Discord 用户
type DiscordUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Bot       bool   `json:"bot"`
}

// DiscordAttachment 附件
type DiscordAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

// DiscordEmbed 嵌入消息
type DiscordEmbed struct {
	Title       string                  `json:"title,omitempty"`
	Description string                  `json:"description,omitempty"`
	URL         string                  `json:"url,omitempty"`
	Color       int                     `json:"color,omitempty"`
	Author      DiscordEmbedAuthor      `json:"author,omitempty"`
	Fields      []DiscordEmbedField     `json:"fields,omitempty"`
	Footer      DiscordEmbedFooter      `json:"footer,omitempty"`
	Timestamp   string                  `json:"timestamp,omitempty"`
}

// DiscordEmbedAuthor 嵌入作者
type DiscordEmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedField 嵌入字段
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedFooter 嵌入页脚
type DiscordEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// ==================== Gateway Intents ====================

const (
	GatewayIntentGuildMessages              = 1 << 9
	GatewayIntentDirectMessages             = 1 << 12
	GatewayIntentMessageContent             = 1 << 23
)
