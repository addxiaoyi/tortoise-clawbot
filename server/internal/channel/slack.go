package channel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/slack-go/slack"

	"tortoise-server/internal/ai"
)

// ============ Slack Channel ============

// SlackChannel Slack 渠道
type SlackChannel struct {
	botToken      string
	signingSecret string
	webhookURL   string
	apiClient    *slack.Client
	aiEngine     *ai.Engine
	running      bool
	mu           sync.RWMutex
}

// SlackEvent Slack 事件
type SlackEvent struct {
	Token          string          `json:"token"`
	TeamID         string          `json:"team_id"`
	APIAppID       string          `json:"api_app_id"`
	Type           string          `json:"type"`
	Event          json.RawMessage `json:"event"`
	Challenge      string          `json:"challenge,omitempty"`
	EventID        string          `json:"event_id,omitempty"`
	EventTime      int             `json:"event_time,omitempty"`
	AuthedUsers    []string        `json:"authed_users,omitempty"`
	AuthedTeams    []string        `json:"authed_teams,omitempty"`
	Authorizations []SlackAuth     `json:"authorizations,omitempty"`
}

// SlackAuth Slack 授权
type SlackAuth struct {
	TeamID      string `json:"team_id"`
	UserID      string `json:"user_id"`
	AppID       string `json:"app_id"`
	Token       string `json:"token"`
	Scope       string `json:"scope"`
	ActorID     string `json:"actor_id"`
	AuthedAppID string `json:"authed_app_id"`
}

// SlackMessageEvent 消息事件
type SlackMessageEvent struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"`
	User        string `json:"user"`
	Text        string `json:"text"`
	Ts          string `json:"ts"`
	ThreadTs    string `json:"thread_ts,omitempty"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
}

// NewSlackChannel 创建 Slack 渠道
func NewSlackChannel(botToken, signingSecret, webhookURL string) *SlackChannel {
	return &SlackChannel{
		botToken:      botToken,
		signingSecret: signingSecret,
		webhookURL:    webhookURL,
		apiClient:     slack.New(botToken),
	}
}

// SetAIEngine 设置 AI 引擎
func (c *SlackChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Slack Bot
func (c *SlackChannel) Start() error {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	log.Println("✅ Slack Bot 已启动")
	return nil
}

// Stop 停止 Slack Bot
func (c *SlackChannel) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	log.Println("🛑 Slack Bot 已停止")
}

// VerifySignature 验证 Slack 签名
func (c *SlackChannel) VerifySignature(signature, timestamp, body string) bool {
	if c.signingSecret == "" {
		return true
	}

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	mac := hmac.New(sha256.New, []byte(c.signingSecret))
	mac.Write([]byte(baseString))
	expectedSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// HandleEvent 处理 Slack 事件
func (c *SlackChannel) HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "static/slack.html")
		return
	}

	// URL 验证请求
	if r.URL.Query().Get("type") == "url_verification" {
		var req struct {
			Challenge string `json:"challenge"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(req.Challenge))
		return
	}

	// 验证签名
	signature := r.Header.Get("X-Slack-Signature")
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	body, _ := io.ReadAll(r.Body)

	if !c.VerifySignature(signature, timestamp, string(body)) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 处理事件
	go c.processEvent(event)

	w.WriteHeader(http.StatusOK)
}

// processEvent 处理事件
func (c *SlackChannel) processEvent(event SlackEvent) {
	if event.Type != "event_callback" {
		return
	}

	var msgEvent SlackMessageEvent
	if err := json.Unmarshal(event.Event, &msgEvent); err != nil {
		return
	}

	// 忽略机器人消息
	if msgEvent.User == "" || strings.HasPrefix(msgEvent.Text, "<@U") {
		return
	}

	// 处理消息
	go c.handleMessage(&msgEvent)
}

// handleMessage 处理消息
func (c *SlackChannel) handleMessage(event *SlackMessageEvent) {
	var response string

	if c.aiEngine != nil {
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
			Messages: []ai.Message{
				{Role: "user", Content: event.Text},
			},
		}

		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，AI 服务出错：%v", err)
		} else {
			response = resp.Content
		}
	} else {
		response = "AI 未配置"
	}

	// 回复消息
	c.replyMessage(event.Channel, event.ThreadTs, response)
}

// replyMessage 回复消息
func (c *SlackChannel) replyMessage(channel, threadTs, text string) {
	_, _, err := c.apiClient.PostMessage(
		channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTs),
	)
	if err != nil {
		log.Printf("Slack 回复失败: %v", err)
	}
}

// SendWebhook 通过 Webhook 发送消息
func (c *SlackChannel) SendWebhook(channel, text string) error {
	if c.webhookURL == "" {
		return fmt.Errorf("webhook URL 未配置")
	}

	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(c.webhookURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
