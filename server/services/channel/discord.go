package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ========== Discord 真实实现 ==========

// DiscordHandler Discord 渠道处理器
type DiscordHandler struct {
	token       string
	intents    int
	gatewayURL  string
	sessionID   string
	sequence    int
	wsConn      *websocket.Conn
	client      *http.Client
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	running     bool
	messageHandler func(from string, message string, metadata map[string]interface{})
}

// Discord  intents
const (
	DiscordIntentGuilds                 = 1 << 0
	DiscordIntentGuildMembers          = 1 << 1
	DiscordIntentGuildBans             = 1 << 2
	DiscordIntentGuildEmojisAndStickers = 1 << 3
	DiscordIntentGuildIntegrations      = 1 << 4
	DiscordIntentGuildWebhooks          = 1 << 5
	DiscordIntentGuildInvites          = 1 << 6
	DiscordIntentGuildVoiceStates      = 1 << 7
	DiscordIntentGuildPresences        = 1 << 8
	DiscordIntentGuildMessages          = 1 << 9
	DiscordIntentGuildMessageReactions = 1 << 10
	DiscordIntentGuildMessageTyping    = 1 << 11
	DiscordIntentDirectMessages        = 1 << 12
	DiscordIntentDirectMessageReactions = 1 << 13
	DiscordIntentDirectMessageTyping    = 1 << 14
	DiscordIntentMessageContent        = 1 << 15
	DiscordIntentGuildScheduledEvents  = 1 << 16

	// 默认 intents (不需要特权)
	DiscordIntentDefault = DiscordIntentGuilds | DiscordIntentGuildMessages | DiscordIntentDirectMessages
)

// GatewayResponse Gateway 响应
type GatewayResponse struct {
	URL string `json:"url"`
}

// GatewayIdentify Gateway 身份验证
type GatewayIdentify struct {
	Token      string         `json:"token"`
	Properties GatewayProperties `json:"properties"`
	Intents   int            `json:"intents"`
}

// GatewayProperties Gateway 属性
type GatewayProperties struct {
	OS              string `json:"os"`
	Browser         string `json:"browser"`
	Device          string `json:"device"`
	Referrer        string `json:"referrer"`
	ReferringDomain string `json:"referring_domain"`
}

// DiscordMessage Discord 消息
type DiscordMessage struct {
	ID              string          `json:"id"`
	ChannelID      string          `json:"channel_id"`
	GuildID        string          `json:"guild_id,omitempty"`
	Content         string          `json:"content"`
	Timestamp       string          `json:"timestamp"`
	EditedTimestamp string          `json:"edited_timestamp,omitempty"`
	Tts             bool            `json:"tts"`
	MentionEveryone bool            `json:"mention_everyone"`
	Author          DiscordUser     `json:"author"`
	Member          *DiscordMember  `json:"member,omitempty"`
	Attachments     []DiscordAttachment `json:"attachments"`
	Embeds          []DiscordEmbed  `json:"embeds"`
	Reactions       []DiscordReaction `json:"reactions"`
	Nonce          string          `json:"nonce,omitempty"`
	Pinned         bool            `json:"pinned"`
	WebhookID      string          `json:"webhook_id,omitempty"`
	Type           int             `json:"type"`
}

// DiscordUser Discord 用户
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name,omitempty"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot,omitempty"`
	System        bool   `json:"system,omitempty"`
	MFAEnabled    bool   `json:"mfa_enabled,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
	Email         string `json:"email,omitempty"`
	Flags         int    `json:"flags,omitempty"`
	PremiumType   int    `json:"premium_type,omitempty"`
	PublicFlags   int    `json:"public_flags,omitempty"`
}

// DiscordMember Discord 成员
type DiscordMember struct {
	Roles  []string          `json:"roles"`
	Deaf   bool             `json:"deaf"`
	Muted  bool             `json:"mute"`
	Nick   string           `json:"nick,omitempty"`
	Avatar string           `json:"avatar,omitempty"`
	User   *DiscordUser     `json:"user,omitempty"`
	Joined string           `json:"joined_at"`
}

// DiscordAttachment Discord 附件
type DiscordAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	ProxyURL    string `json:"proxy_url"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// DiscordEmbed Discord 嵌入
type DiscordEmbed struct {
	Title       string          `json:"title,omitempty"`
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *DiscordFooter  `json:"footer,omitempty"`
	Image       *DiscordImage   `json:"image,omitempty"`
	Thumbnail   *DiscordImage   `json:"thumbnail,omitempty"`
	Video       *DiscordVideo   `json:"video,omitempty"`
	Provider    *DiscordProvider `json:"provider,omitempty"`
	Author      *DiscordAuthor  `json:"author,omitempty"`
	Fields      []DiscordField  `json:"fields,omitempty"`
}

// DiscordFooter Discord 页脚
type DiscordFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// DiscordImage Discord 图片
type DiscordImage struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// DiscordVideo Discord 视频
type DiscordVideo struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// DiscordProvider Discord 提供商
type DiscordProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// DiscordAuthor Discord 作者
type DiscordAuthor struct {
	Name         string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// DiscordField Discord 字段
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordReaction Discord 反应
type DiscordReaction struct {
	Count int           `json:"count"`
	Me    bool         `json:"me"`
	Emoji *DiscordEmoji `json:"emoji"`
}

// DiscordEmoji Discord 表情
type DiscordEmoji struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	User          *DiscordUser `json:"user,omitempty"`
	RequireColons bool   `json:"require_colons,omitempty"`
	Managed       bool   `json:"managed,omitempty"`
	Animated      bool   `json:"animated,omitempty"`
	Available     bool   `json:"available,omitempty"`
}

// DiscordChannel Discord 频道
type DiscordChannel struct {
	ID              string   `json:"id"`
	Type            int      `json:"type"`
	GuildID         string   `json:"guild_id,omitempty"`
	Position        int      `json:"position,omitempty"`
	PermissionOverwrites []DiscordOverwrite `json:"permission_overwrites,omitempty"`
	Name            string   `json:"name,omitempty"`
	Topic           string   `json:"topic,omitempty"`
	NSFW            bool     `json:"nsfw,omitempty"`
	Icon            string   `json:"icon,omitempty"`
	ParentID        string   `json:"parent_id,omitempty"`
}

// DiscordOverwrite Discord 权限覆盖
type DiscordOverwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 = role, 1 = member
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// newDiscordHandler 创建 Discord 处理器
func newDiscordHandler(config map[string]interface{}) *DiscordHandler {
	token := getString(config, "token")
	intents := DiscordIntentDefault

	ctx, cancel := context.WithCancel(context.Background())

	return &DiscordHandler{
		token:      token,
		intents:    intents,
		gatewayURL: "wss://gateway.discord.gg",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect 连接到 Discord Gateway
func (h *DiscordHandler) Connect() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	log.Println("Connecting to Discord Gateway...")

	// 获取 Gateway URL
	resp, err := h.client.Get("https://discord.com/api/v10/gateway")
	if err != nil {
		return fmt.Errorf("failed to get gateway: %w", err)
	}
	defer resp.Body.Close()

	var gatewayResp GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		return fmt.Errorf("failed to decode gateway response: %w", err)
	}
	h.gatewayURL = gatewayResp.URL + "?v=10&encoding=json"

	// 连接 WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(h.gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	h.wsConn = conn

	// 启动读取循环
	go h.readLoop()

	// 发送身份验证
	return h.identify()
}

// Disconnect 断开连接
func (h *DiscordHandler) Disconnect() error {
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()

	h.cancel()

	if h.wsConn != nil {
		return h.wsConn.Close()
	}
	return nil
}

// SendMessage 发送消息
func (h *DiscordHandler) SendMessage(channelID, content string) error {
	// 限制消息长度
	if len(content) > 2000 {
		content = content[:1997] + "..."
	}

	payload := map[string]interface{}{
		"content": content,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID), bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+h.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message: %s", string(body))
	}

	return nil
}

// SendEmbed 发送嵌入消息
func (h *DiscordHandler) SendEmbed(channelID string, embed *DiscordEmbed) error {
	payload := map[string]interface{}{
		"embeds": []interface{}{embed},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID), bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+h.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// OnMessage 设置消息处理函数
func (h *DiscordHandler) OnMessage(handler func(from string, message string, metadata map[string]interface{})) {
	h.messageHandler = handler
}

func (h *DiscordHandler) identify() error {
	identify := GatewayIdentify{
		Token: h.token,
		Properties: GatewayProperties{
			OS:              "linux",
			Browser:         "Tortoise",
			Device:          "Tortoise",
			Referrer:        "",
			ReferringDomain: "",
		},
		Intents: h.intents,
	}

	return h.wsConn.WriteJSON(identify)
}

func (h *DiscordHandler) readLoop() {
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			_, msg, err := h.wsConn.ReadMessage()
			if err != nil {
				log.Printf("Discord read error: %v", err)
				return
			}

			h.handleMessage(msg)
		}
	}
}

func (h *DiscordHandler) handleMessage(data []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	op := int(event["op"].(float64))
	d := event["d"]

	switch op {
	case 10: // Hello
		log.Println("Discord: Received Hello")

	case 11: // Heartbeat ACK
		// log.Println("Discord: Heartbeat ACK")

	case 0: // Dispatch
		s := int(event["s"].(float64))
		h.sequence = s

		t := event["t"].(string)
		switch t {
		case "MESSAGE_CREATE":
			h.handleMessageCreate(d)
		case "MESSAGE_UPDATE":
			h.handleMessageUpdate(d)
		}
	}
}

func (h *DiscordHandler) handleMessageCreate(d interface{}) {
	data, _ := json.Marshal(d)
	var msg DiscordMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// 忽略 Bot 消息
	if msg.Author.Bot {
		return
	}

	// 忽略空消息
	if msg.Content == "" {
		return
	}

	metadata := map[string]interface{}{
		"channel_id": msg.ChannelID,
		"guild_id":   msg.GuildID,
		"author_id":  msg.Author.ID,
		"username":   msg.Author.Username,
	}

	if h.messageHandler != nil {
		h.messageHandler(msg.Author.ID, msg.Content, metadata)
	}
}

func (h *DiscordHandler) handleMessageUpdate(d interface{}) {
	// 处理编辑过的消息
}

// Helper function
func bytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
