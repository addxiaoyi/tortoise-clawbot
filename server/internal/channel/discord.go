package channel

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"

	"github.com/gorilla/websocket"
)

// ============ Discord Bot ============

// DiscordChannel Discord 渠道
type DiscordChannel struct {
	botToken  string
	apiURL    string
	gatewayURL string
	aiEngine  *ai.Engine
	ws        *websocket.Conn
	sessionID string
	sequence  int
	running   bool
	mu        sync.RWMutex
}

// Discord Gateway Events

type GatewayHello struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type GatewayIdentify struct {
	Token      string            `json:"token"`
	Properties *GatewayProperties `json:"properties"`
	Intents    int              `json:"intents"`
}

type GatewayProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

type GatewayMessage struct {
	Op   int             `json:"op"`
	Data json.RawMessage `json:"d"`
	S    *int            `json:"s"`
	T    string          `json:"t"`
}

// Discord Message
type DiscordMessage struct {
	ID        string          `json:"id"`
	Content   string          `json:"content"`
	ChannelID string          `json:"channel_id"`
	Author    *DiscordUser    `json:"author"`
	GuildID   string          `json:"guild_id,omitempty"`
	Mentions  []*DiscordUser   `json:"mentions,omitempty"`
}

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Bot           bool   `json:"bot,omitempty"`
}

// NewDiscordChannel 创建 Discord 渠道
func NewDiscordChannel(botToken string) *DiscordChannel {
	return &DiscordChannel{
		botToken:   botToken,
		apiURL:     "https://discord.com/api/v10",
		sequence:   0,
	}
}

// SetAIEngine 设置 AI 引擎
func (c *DiscordChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Discord Bot
func (c *DiscordChannel) Start() error {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	// 获取 Gateway URL
	gatewayURL, err := c.getGateway()
	if err != nil {
		return fmt.Errorf("failed to get gateway: %w", err)
	}
	c.gatewayURL = gatewayURL

	// 连接 Gateway
	go c.connect()
	log.Println("Discord Bot started")
	return nil
}

// Stop 停止 Discord Bot
func (c *DiscordChannel) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	
	if c.ws != nil {
		c.ws.Close()
	}
	log.Println("Discord Bot stopped")
}

// getGateway 获取 Gateway URL
func (c *DiscordChannel) getGateway() (string, error) {
	url := c.apiURL + "/gateway"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// connect 连接 Gateway
func (c *DiscordChannel) connect() {
	for {
		c.mu.RLock()
		running := c.running
		c.mu.RUnlock()

		if !running {
			break
		}

		ws, _, err := websocket.DefaultDialer.Dial(c.gatewayURL+"/?v=10&encoding=json", nil)
		if err != nil {
			log.Printf("Discord WebSocket error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		c.ws = ws
		c.handleGateway()
	}
}

// handleGateway 处理 Gateway 消息
func (c *DiscordChannel) handleGateway() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("Discord read error: %v", err)
			break
		}

		var gatewayMsg GatewayMessage
		if err := json.Unmarshal(msg, &gatewayMsg); err != nil {
			continue
		}

		// 更新 sequence
		if gatewayMsg.S != nil {
			c.sequence = *gatewayMsg.S
		}

		switch gatewayMsg.Op {
		case 10: // Hello
			c.handleHello(gatewayMsg.Data)
		case 0: // Dispatch
			c.handleDispatch(gatewayMsg.T, gatewayMsg.Data)
		case 11: // Heartbeat ACK
			// 心跳确认
		}
	}
}

// handleHello 处理 Hello 事件
func (c *DiscordChannel) handleHello(data json.RawMessage) {
	var hello GatewayHello
	json.Unmarshal(data, &hello)

	// 发送 Identify
	identify := GatewayIdentify{
		Token: c.botToken,
		Properties: &GatewayProperties{
			OS:      "linux",
			Browser: "tortoise",
			Device:  "tortoise",
		},
		Intents: 1 << 9, // GUILD_MESSAGES intent
	}

	c.sendJSON(2, identify)

	// 启动心跳
	go c.heartbeat(hello.HeartbeatInterval)
}

// heartbeat 发送心跳
func (c *DiscordChannel) heartbeat(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			seq := c.sequence
			c.mu.RUnlock()
			c.sendJSON(1, seq)
		}
	}
}

// handleDispatch 处理 Dispatch 事件
func (c *DiscordChannel) handleDispatch(t string, data json.RawMessage) {
	switch t {
	case "MESSAGE_CREATE":
		c.handleMessageCreate(data)
	}
}

// handleMessageCreate 处理消息创建
func (c *DiscordChannel) handleMessageCreate(data json.RawMessage) {
	var msg DiscordMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// 忽略 Bot 消息
	if msg.Author != nil && msg.Author.Bot {
		return
	}

	// 检查是否 @了 Bot (简化处理，这里检查是否以 <@开头)
	if !strings.HasPrefix(msg.Content, "<@") {
		// 可以处理非 @ 消息
		return
	}

	// 提取消息内容 (简化处理)
	content := strings.TrimSpace(msg.Content)
	content = strings.TrimPrefix(content, "<@!")
	// 移除 @mentions
	content = strings.TrimPrefix(content, "<@")
	if idx := strings.Index(content, ">"); idx != -1 {
		content = content[idx+1:]
	}
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	// 发送"正在思考"反应
	c.addReaction(msg.ChannelID, msg.ID, "🤔")

	// 处理消息
	go c.processMessage(msg.ChannelID, msg.ID, content)
}

// processMessage 处理消息
func (c *DiscordChannel) processMessage(channelID, messageID, content string) {
	var response string

	if c.aiEngine != nil {
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
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
		response = fmt.Sprintf("你说了: %s\n\n(AI 未配置)", content)
	}

	// 回复消息
	c.replyMessage(channelID, messageID, response)
}

// replyMessage 回复消息
func (c *DiscordChannel) replyMessage(channelID, messageID, content string) {
	url := fmt.Sprintf("%s/channels/%s/messages", c.apiURL, channelID)

	payload := map[string]interface{}{
		"content":   content,
		"message_reference": map[string]string{
			"message_id": messageID,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Discord reply error: %v", err)
		return
	}
	defer resp.Body.Close()
}

// addReaction 添加反应
func (c *DiscordChannel) addReaction(channelID, messageID, emoji string) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", 
		c.apiURL, channelID, messageID, emoji)

	req, _ := http.NewRequest("PUT", url, nil)
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// sendJSON 发送 JSON 消息
func (c *DiscordChannel) sendJSON(op int, data interface{}) {
	payload := map[string]interface{}{
		"op": op,
		"d":  data,
	}

	msg, _ := json.Marshal(payload)
	c.ws.WriteMessage(websocket.TextMessage, msg)
}

// GetBotInfo 获取 Bot 信息
func (c *DiscordChannel) GetBotInfo() (*DiscordUser, error) {
	url := c.apiURL + "/users/@me"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var user DiscordUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
