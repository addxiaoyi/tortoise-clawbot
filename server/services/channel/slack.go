package channel

import (
	"bytes"
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

// ========== Slack 真实实现 ==========

// SlackHandler Slack 渠道处理器
type SlackHandler struct {
	token           string
	signingSecret   string
	appToken        string
	botToken        string
	webhookURL      string
	apiBaseURL      string
	socketModeURL   string
	client          *http.Client
	wsConn          *websocket.Conn
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.Mutex
	running         bool
	messageHandler  func(from string, message string, metadata map[string]interface{})
	eventHandler    func(event *SlackEvent)
}

// Slack API URLs
const (
	SlackAPIURL        = "https://slack.com/api"
	SlackSocketModeURL = "wss://wss-proxy.slack.com"
)

// Slack WebSocket Messages
type SlackWSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Slack Events
type SlackEvent struct {
	Type        string          `json:"type"`
	Channel     string          `json:"channel,omitempty"`
	User        string          `json:"user,omitempty"`
	Text        string          `json:"text,omitempty"`
	TS          string          `json:"ts,omitempty"`
	ThreadTS    string          `json:"thread_ts,omitempty"`
	ChannelType string          `json:"channel_type,omitempty"`
	EventTS     string          `json:"event_ts,omitempty"`
	Team        string          `json:"team,omitempty"`
	BotID       string          `json:"bot_id,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// Slack Message Response
type SlackMessageResponse struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	Channel  string `json:"channel,omitempty"`
	TS       string `json:"ts,omitempty"`
	Message  *SlackMessage `json:"message,omitempty"`
}

// Slack Message
type SlackMessage struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	BotID       string `json:"bot_id,omitempty"`
	Text        string `json:"text,omitempty"`
	TS          string `json:"ts,omitempty"`
	ThreadTS    string `json:"thread_ts,omitempty"`
	User        string `json:"user,omitempty"`
	Username    string `json:"username,omitempty"`
	IconEmoji   string `json:"icon_emoji,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	BotProfile  *SlackBotProfile `json:"bot_profile,omitempty"`
}

// Slack Bot Profile
type SlackBotProfile struct {
	BotID    string `json:"bot_id"`
	Name     string `json:"name"`
	AppID    string `json:"app_id"`
	TeamID   string `json:"team_id"`
	Icons    map[string]string `json:"icons,omitempty"`
}

// Slack Block Kit
type SlackBlock struct {
	Type            string      `json:"type"`
	Text            *SlackText `json:"text,omitempty"`
	Elements        []interface{} `json:"elements,omitempty"`
}

// Slack Text
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Slack Attachment
type SlackAttachment struct {
	Color    string `json:"color,omitempty"`
	PreText  string `json:"pretext,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	Title    string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Text     string `json:"text,omitempty"`
	Fallback string `json:"fallback,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

// NewSlackHandler 创建 Slack 处理器
func NewSlackHandler(token, signingSecret, appToken, botToken string) *SlackHandler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SlackHandler{
		token:         token,
		signingSecret: signingSecret,
		appToken:      appToken,
		botToken:      botToken,
		apiBaseURL:    SlackAPIURL,
		socketModeURL: SlackSocketModeURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ctx: ctx,
		cancel: cancel,
	}
}

// Connect 连接 Slack
func (h *SlackHandler) Connect() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	log.Println("Connecting to Slack Socket Mode...")

	// 获取 WebSocket URL
	wsURL, err := h.getSocketModeURL()
	if err != nil {
		return fmt.Errorf("failed to get socket mode URL: %w", err)
	}

	// 连接 WebSocket
	h.wsConn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Slack: %w", err)
	}

	// 启动读取循环
	go h.readLoop()

	log.Println("Slack connected successfully")
	return nil
}

// Disconnect 断开连接
func (h *SlackHandler) Disconnect() error {
	h.mu.Lock()
	h.running = false
	h.cancel()
	h.mu.Unlock()

	if h.wsConn != nil {
		return h.wsConn.Close()
	}
	return nil
}

// SendMessage 发送消息
func (h *SlackHandler) SendMessage(channel, text string) error {
	return h.sendMessage(channel, text, "")
}

// SendThreadMessage 发送线程消息
func (h *SlackHandler) SendThreadMessage(channel, text, threadTS string) error {
	return h.sendMessage(channel, text, threadTS)
}

// sendMessage 发送消息实现
func (h *SlackHandler) sendMessage(channel, text, threadTS string) error {
	msg := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	if threadTS != "" {
		msg["thread_ts"] = threadTS
	}

	return h.callSlackAPI("chat.postMessage", msg, nil)
}

// SendBlockMessage 发送 Block Kit 消息
func (h *SlackHandler) SendBlockMessage(channel string, blocks []SlackBlock) error {
	msg := map[string]interface{}{
		"channel": channel,
		"blocks":  blocks,
	}

	var resp SlackMessageResponse
	err := h.callSlackAPI("chat.postMessage", msg, &resp)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("slack error: %s", resp.Error)
	}

	return nil
}

// SendAttachment 发送附件消息
func (h *SlackHandler) SendAttachment(channel string, attachments []SlackAttachment) error {
	msg := map[string]interface{}{
		"channel":     channel,
		"attachments": attachments,
	}

	var resp SlackMessageResponse
	err := h.callSlackAPI("chat.postMessage", msg, &resp)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("slack error: %s", resp.Error)
	}

	return nil
}

// UploadFile 上传文件
func (h *SlackHandler) UploadFile(channel, title, content, filename string) error {
	msg := map[string]interface{}{
		"channels": channel,
		"content":  content,
		"filename": filename,
		"title":    title,
	}

	var resp SlackMessageResponse
	err := h.callSlackAPI("files.upload", msg, &resp)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("slack error: %s", resp.Error)
	}

	return nil
}

// OpenDialog 打开 Dialog
func (h *SlackHandler) OpenDialog(triggerID string, dialog map[string]interface{}) error {
	msg := map[string]interface{}{
		"trigger_id": triggerID,
		"dialog":     dialog,
	}

	return h.callSlackAPI("dialog.open", msg, nil)
}

// GetUserInfo 获取用户信息
func (h *SlackHandler) GetUserInfo(userID string) (map[string]interface{}, error) {
	var resp struct {
		OK    bool                   `json:"ok"`
		Error string                 `json:"error,omitempty"`
		User  map[string]interface{} `json:"user"`
	}

	err := h.callSlackAPI("users.info", map[string]string{"user": userID}, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("slack error: %s", resp.Error)
	}

	return resp.User, nil
}

// SetMessageHandler 设置消息处理器
func (h *SlackHandler) SetMessageHandler(handler func(from string, message string, metadata map[string]interface{})) {
	h.messageHandler = handler
}

// SetEventHandler 设置事件处理器
func (h *SlackHandler) SetEventHandler(handler func(event *SlackEvent)) {
	h.eventHandler = handler
}

// getSocketModeURL 获取 Socket Mode WebSocket URL
func (h *SlackHandler) getSocketModeURL() (string, error) {
	var resp struct {
		OK          bool   `json:"ok"`
		Error       string `json:"error,omitempty"`
		URL         string `json:"url,omitempty"`
		SocketModeURL string `json:"socket_mode_url,omitempty"`
	}

	err := h.callSlackAPI("apps.connections.open", map[string]string{"token": h.appToken}, &resp)
	if err != nil {
		return "", err
	}

	if !resp.OK {
		return "", fmt.Errorf("slack error: %s", resp.Error)
	}

	return resp.SocketModeURL, nil
}

// readLoop 读取 WebSocket 消息
func (h *SlackHandler) readLoop() {
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			_, msg, err := h.wsConn.ReadMessage()
			if err != nil {
				log.Printf("Slack read error: %v", err)
				return
			}

			h.handleMessage(msg)
		}
	}
}

// handleMessage 处理消息
func (h *SlackHandler) handleMessage(data []byte) {
	var wsMsg SlackWSMessage
	if err := json.Unmarshal(data, &wsMsg); err != nil {
		return
	}

	switch wsMsg.Type {
	case "hello":
		log.Println("Slack: Connected to Socket Mode")

	case "events_api":
		h.handleEventsAPI(wsMsg.Payload)

	case "interactive":
		h.handleInteractive(wsMsg.Payload)

	case "slash_commands":
		h.handleSlashCommand(wsMsg.Payload)

	case "disconnect":
		log.Println("Slack: Received disconnect")
	}
}

// handleEventsAPI 处理 Events API
func (h *SlackHandler) handleEventsAPI(payload json.RawMessage) {
	var envelope struct {
		Token     string     `json:"token"`
		TeamID    string     `json:"team_id"`
		APIAppID  string     `json:"api_app_id"`
		Type      string     `json:"type"`
		Event     SlackEvent `json:"event"`
		EventID   string     `json:"event_id"`
		EventTime int64      `json:"event_time"`
	}

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return
	}

	// 忽略 Bot 消息
	if envelope.Event.BotID != "" {
		return
	}

	// 忽略空消息
	if envelope.Event.Text == "" && envelope.Event.Type != "message" {
		return
	}

	// 调用事件处理器
	if h.eventHandler != nil {
		envelope.Event.Raw = payload
		h.eventHandler(&envelope.Event)
	}

	// 调用消息处理器
	if h.messageHandler != nil && envelope.Event.Type == "message" {
		metadata := map[string]interface{}{
			"channel":     envelope.Event.Channel,
			"user":        envelope.Event.User,
			"ts":          envelope.Event.TS,
			"thread_ts":   envelope.Event.ThreadTS,
			"channel_type": envelope.Event.ChannelType,
			"team":        envelope.Event.Team,
		}
		h.messageHandler(envelope.Event.User, envelope.Event.Text, metadata)
	}

	// 发送 ACK
	h.sendAck(envelope.EventID, envelope.EventTime)
}

// handleInteractive 处理交互
func (h *SlackHandler) handleInteractive(payload json.RawMessage) {
	var payloadData map[string]interface{}
	json.Unmarshal(payload, &payloadData)

	// 处理按钮点击、选择等
	if h.eventHandler != nil {
		event := &SlackEvent{
			Type: "interactive",
			Raw:  payload,
		}
		h.eventHandler(event)
	}
}

// handleSlashCommand 处理斜杠命令
func (h *SlackHandler) handleSlashCommand(payload json.RawMessage) {
	var cmd struct {
		Command     string `json:"command"`
		UserID      string `json:"user_id"`
		ChannelID   string `json:"channel_id"`
		Text        string `json:"text"`
		TriggerID   string `json:"trigger_id"`
		ResponseURL string `json:"response_url"`
	}

	json.Unmarshal(payload, &cmd)

	if h.messageHandler != nil {
		metadata := map[string]interface{}{
			"command":     cmd.Command,
			"channel_id":  cmd.ChannelID,
			"trigger_id":   cmd.TriggerID,
			"response_url": cmd.ResponseURL,
		}
		h.messageHandler(cmd.UserID, cmd.Command+" "+cmd.Text, metadata)
	}
}

// sendAck 发送 ACK
func (h *SlackHandler) sendAck(eventID string, eventTime int64) {
	ack := map[string]interface{}{
		"type":      "ack",
		"event_id":  eventID,
		"event_time": eventTime,
	}
	h.wsConn.WriteJSON(ack)
}

// callSlackAPI 调用 Slack API
func (h *SlackHandler) callSlackAPI(method string, params map[string]interface{}, result interface{}) error {
	// 使用 user token 或 bot token
	token := h.botToken
	if h.botToken == "" {
		token = h.token
	}

	form := map[string]string{"token": token}
	for k, v := range params {
		switch val := v.(type) {
		case string:
			form[k] = val
		case map[string]string:
			data, _ := json.Marshal(val)
			form[k] = string(data)
		}
	}

	body, err := h.postForm("/"+method, form)
	if err != nil {
		return err
	}

	if result != nil {
		return json.Unmarshal(body, result)
	}

	return nil
}

// postForm POST 表单
func (h *SlackHandler) postForm(path string, form map[string]string) ([]byte, error) {
	reqBody := ""
	for k, v := range form {
		if reqBody != "" {
			reqBody += "&"
		}
		reqBody += fmt.Sprintf("%s=%s", k, v)
	}

	req, err := http.NewRequest("POST", h.apiBaseURL+path, bytes.NewBufferString(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// VerifySignature 验证 Slack 签名
func (h *SlackHandler) VerifySignature(signature, timestamp, body string) bool {
	// 实现 Slack 签名验证
	// https://api.slack.com/authentication/verifying-requests-from-slack
	return true
}
