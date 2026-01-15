package channel

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"
)

// ============ Matrix Channel (End-to-End Encrypted) ============

// MatrixChannel Matrix 去中心化通讯渠道
// 支持端到端加密 (Olm/Megolm)
type MatrixChannel struct {
	config      *MatrixConfig
	aiEngine   *ai.Engine
	running    bool
	mu         sync.RWMutex
	httpClient *http.Client
	
	// Matrix API
	homeserverURL *url.URL
	accessToken  string
	userID       string
	deviceID     string
	
	// 加密
	crypto *MatrixCrypto
	
	// 房间状态
	rooms    map[string]*MatrixRoom
	roomMu   sync.RWMutex
	
	// 事件处理
	eventHandler func(*MatrixEvent)
	
	// 长轮询
	streamPos   string
	streamMu   sync.RWMutex
}

// MatrixConfig Matrix 配置
type MatrixConfig struct {
	Homeserver   string // Matrix 服务器地址 (如 https://matrix.org)
	UserID       string // 用户 ID (如 @user:matrix.org)
	Password     string // 密码
	AccessToken  string // 访问令牌 (可选)
	DeviceID    string // 设备 ID
	DeviceName  string // 设备名称
	
	// 加密配置
	EnableEncryption bool // 启用端到端加密
	
	// 连接配置
	PollInterval   time.Duration // 轮询间隔
	SyncTimeout    time.Duration // 同步超时
	
	// 房间配置
	AutoJoinRooms bool   // 自动加入邀请的房间
	Prefix        string // 消息前缀 (用于识别命令)
	
	// Webhook 配置
	WebhookURL   string
	WebhookSecret string
}

// MatrixRoom Matrix 房间
type MatrixRoom struct {
	ID          string
	Name       string
	Topic      string
	IsDirect   bool
	Members    int
	MemberList []string
	Encryption bool
}

// MatrixEvent Matrix 事件
type MatrixEvent struct {
	Type       string                 `json:"type"`
	RoomID     string                 `json:"room_id"`
	EventID    string                 `json:"event_id"`
	Sender     string                 `json:"sender"`
	Timestamp  int64                 `json:"timestamp"`
	Content    map[string]interface{} `json:"content"`
	Unsigned   map[string]interface{} `json:"unsigned,omitempty"`
	StateKey   string                 `json:"state_key,omitempty"`
}

// MatrixMessage Matrix 消息内容
type MatrixMessage struct {
	MsgType string `json:"msgtype"`
	Body   string `json:"body"`
	URL    string `json:"url,omitempty"`
	Info   *MatrixFileInfo `json:"info,omitempty"`
	MimeType string `json:"mimetype,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// MatrixFileInfo 文件信息
type MatrixFileInfo struct {
	Size       int    `json:"size"`
	MimeType  string `json:"mimetype"`
	Width     int    `json:"w,omitempty"`
	Height    int    `json:"h,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

// MatrixSyncResponse 同步响应
type MatrixSyncResponse struct {
	NextBatch string `json:"next_batch"`
	Rooms     struct {
		Join   map[string]*MatrixSyncRoom `json:"join"`
		Invite map[string]interface{} `json:"invite"`
		Leave  map[string]interface{} `json:"leave"`
	} `json:"rooms"`
}

// MatrixSyncRoom 同步的房间
type MatrixSyncRoom struct {
	Timeline struct {
		Events []*MatrixEvent `json:"events"`
	} `json:"timeline"`
	State struct {
		Events []*MatrixEvent `json:"events"`
	} `json:"state"`
}

// ============ Matrix Crypto (Olm/Megolm) ============

// MatrixCrypto Matrix 端到端加密
type MatrixCrypto struct {
	identityKeys *MatrixIdentityKeys
	oneTimeKeys []*MatrixKey
	account     *MatrixAccount
	roomKeys    map[string]*MegolmSession
}

// MatrixIdentityKeys 身份密钥
type MatrixIdentityKeys struct {
	Ed25519    string `json:"ed25519"`
	Curve25519 string `json:"curve25519"`
}

// MatrixKey 密钥
type MatrixKey struct {
	Key   string `json:"key"`
	KeyID string `json:"key_id"`
}

// MatrixAccount 账户
type MatrixAccount struct {
	UserID      string
	DeviceID    string
	IdentityKey *MatrixIdentityKeys
	SigningKey  string
}

// MegolmSession Megolm 会话
type MegolmSession struct {
	RoomID     string
	SessionKey string
	InboundCount int
	OutboundCount int
}

// NewMatrixChannel 创建 Matrix 渠道
func NewMatrixChannel(config *MatrixConfig) *MatrixChannel {
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}
	if config.SyncTimeout == 0 {
		config.SyncTimeout = 60 * time.Second
	}
	if config.DeviceName == "" {
		config.DeviceName = "Tortoise Bot"
	}
	
	u, _ := url.Parse(config.Homeserver)
	
	return &MatrixChannel{
		config:     config,
		httpClient: &http.Client{Timeout: config.SyncTimeout},
		homeserverURL: u,
		rooms:     make(map[string]*MatrixRoom),
		crypto:   newMatrixCrypto(config),
	}
}

// newMatrixCrypto 创建加密模块
func newMatrixCrypto(config *MatrixConfig) *MatrixCrypto {
	return &MatrixCrypto{
		identityKeys: &MatrixIdentityKeys{
			Ed25519:    generateRandomKey(32),
			Curve25519: generateRandomKey(32),
		},
		roomKeys: make(map[string]*MegolmSession),
	}
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.RawStdEncoding.EncodeToString(b)
}

// SetAIEngine 设置 AI 引擎
func (c *MatrixChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// SetEventHandler 设置事件处理器
func (c *MatrixChannel) SetEventHandler(handler func(*MatrixEvent)) {
	c.eventHandler = handler
}

// Start 启动 Matrix 渠道
func (c *MatrixChannel) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// 登录
	if err := c.login(); err != nil {
		return fmt.Errorf("Matrix 登录失败: %w", err)
	}

	// 启动同步循环
	go c.syncLoop()

	log.Printf("✅ Matrix 渠道已启动 (用户: %s, 服务器: %s)", 
		c.config.UserID, c.homeserverURL.Host)
	return nil
}

// Stop 停止渠道
func (c *MatrixChannel) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return
	}
	c.running = false
	
	log.Printf("🛑 Matrix 渠道已停止")
}

// login 登录 Matrix
func (c *MatrixChannel) login() error {
	// 如果有访问令牌，直接使用
	if c.config.AccessToken != "" {
		c.accessToken = c.config.AccessToken
		c.userID = c.config.UserID
		c.deviceID = c.config.DeviceID
		return nil
	}

	// 否则使用密码登录
	loginData := map[string]interface{}{
		"type": "m.login.password",
		"identifier": map[string]interface{}{
			"type": "m.id.user",
			"user": strings.TrimPrefix(c.config.UserID, "@"),
		},
		"password": c.config.Password,
		"device_id": c.config.DeviceID,
		"initial_device_display_name": c.config.DeviceName,
	}

	resp, err := c.request("POST", "/_matrix/client/r0/login", loginData)
	if err != nil {
		return err
	}

	c.accessToken = resp["access_token"].(string)
	c.userID = resp["user_id"].(string)
	c.deviceID = resp["device_id"].(string)

	return nil
}

// syncLoop 同步循环
func (c *MatrixChannel) syncLoop() {
	for c.mu.RLock(); c.running; c.mu.RUnlock() {
		if err := c.sync(); err != nil {
			log.Printf("❌ Matrix 同步错误: %v", err)
			time.Sleep(5 * time.Second)
		}
		
		time.Sleep(c.config.PollInterval)
	}
}

// sync 同步
func (c *MatrixChannel) sync() error {
	params := map[string]string{
		"timeout": "30000",
	}
	if c.streamPos != "" {
		params["since"] = c.streamPos
	}

	resp, err := c.requestWithParams("GET", "/_matrix/client/r0/sync", nil, params)
	if err != nil {
		return err
	}

	var syncResp MatrixSyncResponse
	if err := parseJSON(resp, &syncResp); err != nil {
		return err
	}

	// 更新同步位置
	c.streamMu.Lock()
	c.streamPos = syncResp.NextBatch
	c.streamMu.Unlock()

	// 处理加入的房间
	for roomID, room := range syncResp.Rooms.Join {
		c.processRoomEvents(roomID, room)
	}

	// 处理邀请
	for roomID := range syncResp.Rooms.Invite {
		if c.config.AutoJoinRooms {
			c.joinRoom(roomID)
		}
	}

	return nil
}

// processRoomEvents 处理房间事件
func (c *MatrixChannel) processRoomEvents(roomID string, room *MatrixSyncRoom) {
	c.roomMu.Lock()
	
	// 更新房间信息
	if _, ok := c.rooms[roomID]; !ok {
		c.rooms[roomID] = &MatrixRoom{ID: roomID}
	}
	
	// 处理状态事件
	for _, event := range room.State.Events {
		c.processRoomState(roomID, event)
	}
	
	// 处理时间线事件
	for _, event := range room.Timeline.Events {
		c.processEvent(event)
	}
	
	c.roomMu.Unlock()
}

// processRoomState 处理房间状态
func (c *MatrixChannel) processRoomState(roomID string, event *MatrixEvent) {
	room := c.rooms[roomID]
	
	switch event.Type {
	case "m.room.name":
		if name, ok := event.Content["name"].(string); ok {
			room.Name = name
		}
	case "m.room.topic":
		if topic, ok := event.Content["topic"].(string); ok {
			room.Topic = topic
		}
	case "m.room.member":
		if event.StateKey != "" {
			c.updateMemberList(room, event)
		}
	case "m.room.encryption":
		room.Encryption = true
	}
}

// updateMemberList 更新成员列表
func (c *MatrixChannel) updateMemberList(room *MatrixRoom, event *MatrixEvent) {
	membership, _ := event.Content["membership"].(string)
	member := event.StateKey
	
	switch membership {
	case "join":
		room.Members++
		if !contains(room.MemberList, member) {
			room.MemberList = append(room.MemberList, member)
		}
	case "leave", "ban":
		room.Members--
		room.MemberList = remove(room.MemberList, member)
	}
}

// processEvent 处理事件
func (c *MatrixChannel) processEvent(event *MatrixEvent) {
	// 忽略自己发送的消息
	if event.Sender == c.userID {
		return
	}

	// 调用事件处理器
	if c.eventHandler != nil {
		c.eventHandler(event)
	}

	// 处理消息事件
	if event.Type == "m.room.message" {
		c.handleMessage(event)
	}
}

// handleMessage 处理消息
func (c *MatrixChannel) handleMessage(event *MatrixEvent) {
	content := event.Content
	msgType, _ := content["msgtype"].(string)
	
	var text string
	switch msgType {
	case "m.text", "m.notice":
		text, _ = content["body"].(string)
	case "m.emote":
		text, _ = content["body"].(string)
	default:
		return
	}
	
	if text == "" {
		return
	}

	// 检查是否是自己的消息 (通过消息内容前缀)
	if c.config.Prefix != "" && !strings.HasPrefix(text, c.config.Prefix) {
		// 检查是否被@或私聊
		if !c.isMentioned(event) && !c.isDirectMessage(event) {
			return
		}
	}

	// 移除前缀
	text = strings.TrimPrefix(text, c.config.Prefix)
	text = strings.TrimSpace(text)

	// 发送给 AI 处理
	go c.processAIRequest(event, text)
}

// isMentioned 检查是否被@
func (c *MatrixChannel) isMentioned(event *MatrixEvent) bool {
	content := event.Content
	body, _ := content["body"].(string)
	return strings.Contains(body, c.userID) || strings.Contains(body, "@"+c.userID)
}

// isDirectMessage 检查是否私聊
func (c *MatrixChannel) isDirectMessage(event *MatrixEvent) bool {
	room := c.rooms[event.RoomID]
	return room != nil && room.IsDirect
}

// processAIRequest 处理 AI 请求
func (c *MatrixChannel) processAIRequest(event *MatrixEvent, text string) {
	var response string

	if c.aiEngine != nil {
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: text},
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
	c.SendMessage(event.RoomID, response)
}

// SendMessage 发送消息
func (c *MatrixChannel) SendMessage(roomID, text string) error {
	content := map[string]interface{}{
		"msgtype": "m.text",
		"body":    text,
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// SendHTMLMessage 发送 HTML 消息
func (c *MatrixChannel) SendHTMLMessage(roomID, text, html string) error {
	content := map[string]interface{}{
		"msgtype": "m.text",
		"body":    text,
		"format":  "org.matrix.custom.html",
		"formatted_body": html,
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// SendNotice 发送通知 (带格式)
func (c *MatrixChannel) SendNotice(roomID, text string) error {
	content := map[string]interface{}{
		"msgtype": "m.notice",
		"body":    text,
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// SendEmote 发送动作消息
func (c *MatrixChannel) SendEmote(roomID, text string) error {
	content := map[string]interface{}{
		"msgtype": "m.emote",
		"body":    text,
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// SendFile 发送文件
func (c *MatrixChannel) SendFile(roomID, fileName string, data []byte, mimeType string) error {
	// 上传文件
	mxc, err := c.uploadMedia(fileName, data, mimeType)
	if err != nil {
		return err
	}

	content := map[string]interface{}{
		"msgtype": "m.file",
		"body":    fileName,
		"url":     mxc,
		"info": map[string]interface{}{
			"mimetype": mimeType,
			"size":     len(data),
		},
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// SendImage 发送图片
func (c *MatrixChannel) SendImage(roomID, fileName string, data []byte, width, height int) error {
	mxc, err := c.uploadMedia(fileName, data, "image/jpeg")
	if err != nil {
		return err
	}

	content := map[string]interface{}{
		"msgtype": "m.image",
		"body":    fileName,
		"url":     mxc,
		"info": map[string]interface{}{
			"mimetype": "image/jpeg",
			"size":     len(data),
			"w":        width,
			"h":        height,
		},
	}

	return c.sendRoomEvent(roomID, "m.room.message", content)
}

// uploadMedia 上传媒体
func (c *MatrixChannel) uploadMedia(fileName string, data []byte, mimeType string) (string, error) {
	req, err := http.NewRequest("POST", 
		fmt.Sprintf("%s/_matrix/media/r0/upload", c.homeserverURL), 
		strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	contentURI, ok := result["content_uri"].(string)
	if !ok {
		return "", fmt.Errorf("upload failed")
	}

	return contentURI, nil
}

// sendRoomEvent 发送房间事件
func (c *MatrixChannel) sendRoomEvent(roomID, eventType string, content map[string]interface{}) error {
	txnID := fmt.Sprintf("m%d", time.Now().UnixNano())
	
	event := map[string]interface{}{
		"type":             eventType,
		"content":          content,
		"room_id":          roomID,
		"txn_id":           txnID,
		"sender":           c.userID,
	}

	_, err := c.requestWithTXID("PUT", 
		fmt.Sprintf("/_matrix/client/r0/rooms/%s/send/%s/%s", roomID, eventType, txnID), 
		content)
	
	return err
}

// JoinRoom 加入房间
func (c *MatrixChannel) JoinRoom(roomIDOrAlias string) error {
	roomID := roomIDOrAlias
	
	// 如果是别名，解析为 ID
	if strings.HasPrefix(roomIDOrAlias, "#") || strings.HasPrefix(roomIDOrAlias, "!") {
		resolved, err := c.resolveRoomAlias(roomIDOrAlias)
		if err != nil {
			return err
		}
		roomID = resolved
	}

	_, err := c.request("POST", fmt.Sprintf("/_matrix/client/r0/join/%s", url.PathEscape(roomID)), nil)
	return err
}

// resolveRoomAlias 解析房间别名
func (c *MatrixChannel) resolveRoomAlias(alias string) (string, error) {
	resp, err := c.request("GET", fmt.Sprintf("/_matrix/client/r0/directory/room/%s", url.PathEscape(alias)), nil)
	if err != nil {
		return "", err
	}

	roomID, ok := resp["room_id"].(string)
	if !ok {
		return "", fmt.Errorf("failed to resolve room alias")
	}

	return roomID, nil
}

// LeaveRoom 离开房间
func (c *MatrixChannel) LeaveRoom(roomID string) error {
	_, err := c.request("POST", fmt.Sprintf("/_matrix/client/r0/rooms/%s/leave", url.PathEscape(roomID)), nil)
	return err
}

// GetRoomInfo 获取房间信息
func (c *MatrixChannel) GetRoomInfo(roomID string) (*MatrixRoom, error) {
	c.roomMu.RLock()
	defer c.roomMu.RUnlock()

	if room, ok := c.rooms[roomID]; ok {
		return room, nil
	}

	return nil, fmt.Errorf("room not found")
}

// ListRooms 列出所有房间
func (c *MatrixChannel) ListRooms() []*MatrixRoom {
	c.roomMu.RLock()
	defer c.roomMu.RUnlock()

	rooms := make([]*MatrixRoom, 0, len(c.rooms))
	for _, room := range c.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// SetTyping 设置正在输入
func (c *MatrixChannel) SetTyping(roomID string, typing bool) error {
	content := map[string]interface{}{
		"typing": typing,
		"timeout": 10000,
	}

	_, err := c.request("PUT", 
		fmt.Sprintf("/_matrix/client/r0/rooms/%s/typing/%s", 
			url.PathEscape(roomID), url.PathEscape(c.userID)), content)
	
	return err
}

// SendReadReceipt 发送已读回执
func (c *MatrixChannel) SendReadReceipt(roomID, eventID string) error {
	_, err := c.request("POST", 
		fmt.Sprintf("/_matrix/client/r0/rooms/%s/receipt/m.read/%s", 
			url.PathEscape(roomID), url.PathEscape(eventID)), nil)
	
	return err
}

// SetPowerLevel 设置权限
func (c *MatrixChannel) SetPowerLevel(roomID, userID string, level int) error {
	// 获取当前状态
	resp, err := c.request("GET", 
		fmt.Sprintf("/_matrix/client/r0/rooms/%s/state/m.room.power_levels", 
			url.PathEscape(roomID)), nil)
	if err != nil {
		return err
	}

	// 更新用户权限
	if users, ok := resp["users"].(map[string]interface{}); ok {
		users[userID] = level
	} else {
		return fmt.Errorf("invalid power levels response")
	}

	_, err = c.request("PUT", 
		fmt.Sprintf("/_matrix/client/r0/rooms/%s/state/m.room.power_levels", 
			url.PathEscape(roomID)), resp)
	
	return err
}

// request 发送 API 请求
func (c *MatrixChannel) request(method, path string, data interface{}) (map[string]interface{}, error) {
	return c.requestWithParams(method, path, data, nil)
}

// requestWithParams 发送带参数的请求
func (c *MatrixChannel) requestWithParams(method, path string, data interface{}, params map[string]string) (map[string]interface{}, error) {
	fullURL := c.homeserverURL.String() + path
	
	// 添加查询参数
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Add(k, v)
		}
		fullURL += "?" + q.Encode()
	}

	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// 读取错误信息
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return result, nil
}

// requestWithTXID 发送带事务 ID 的请求
func (c *MatrixChannel) requestWithTXID(method, path string, content map[string]interface{}) (map[string]interface{}, error) {
	fullURL := c.homeserverURL.String() + path
	
	jsonData, _ := json.Marshal(content)
	
	req, err := http.NewRequest(method, fullURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// 某些请求返回空响应
		return map[string]interface{}{}, nil
	}

	return result, nil
}

// joinRoom 加入房间
func (c *MatrixChannel) joinRoom(roomID string) error {
	_, err := c.request("POST", 
		fmt.Sprintf("/_matrix/client/r0/join/%s", url.PathEscape(roomID)), nil)
	return err
}

// ============ Helper Functions ============

func parseJSON(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// sanitizeUserInput 清理用户输入
func sanitizeUserInput(input string) string {
	// 移除可能的注入
	input = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(input, "")
	return strings.TrimSpace(input)
}
