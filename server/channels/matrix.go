package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tortoise/config"
)

// MatrixChannel Matrix 渠道实现
type MatrixChannel struct {
	config     config.MatrixConfig
	client     *MatrixClient
	userID     string
	deviceID   string
	accessToken string
	rooms      map[string]*MatrixRoom
	handlers   map[string]MessageHandler
	mu         sync.RWMutex
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// MatrixRoom Matrix 房间
type MatrixRoom struct {
	RoomID    string
	Name      string
	Topic     string
	Members   []string
	CreatedAt time.Time
}

// MatrixClient Matrix API 客户端
type MatrixClient struct {
	homeserver string
	client     *MatrixHTTPClient
}

// MatrixMessage Matrix 消息
type MatrixMessage struct {
	EventID   string `json:"event_id"`
	RoomID    string `json:"room_id"`
	Sender    string `json:"sender"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Content   MatrixMessageContent `json:"content"`
}

// MatrixMessageContent 消息内容
type MatrixMessageContent struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// NewMatrixChannel 创建 Matrix 渠道
func NewMatrixChannel(cfg config.MatrixConfig) *MatrixChannel {
	ctx, cancel := context.WithCancel(context.Background())
	return &MatrixChannel{
		config:   cfg,
		rooms:    make(map[string]*MatrixRoom),
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Connect 连接到 Matrix 服务器
func (m *MatrixChannel) Connect() error {
	m.client = &MatrixClient{
		homeserver: m.config.Homeserver,
		client:     NewMatrixHTTPClient(m.config.Homeserver),
	}

	// 登录获取 access token
	loginReq := MatrixLoginRequest{
		Type:     "m.login.password",
		User:     m.config.Username,
		Password: m.config.Password,
		DeviceID: m.config.DeviceID,
	}

	var loginResp MatrixLoginResponse
	if err := m.client.Login(&loginReq, &loginResp); err != nil {
		return fmt.Errorf("matrix login failed: %w", err)
	}

	m.accessToken = loginResp.AccessToken
	m.userID = loginResp.UserID
	m.deviceID = loginResp.DeviceID

	// 同步房间列表
	if err := m.syncRooms(); err != nil {
		return fmt.Errorf("sync rooms failed: %w", err)
	}

	m.running = true
	return nil
}

// syncRooms 同步房间列表
func (m *MatrixChannel) syncRooms() error {
	var resp MatrixSyncResponse
	if err := m.client.Sync(m.accessToken, "", "", &resp); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for roomID, joinRoom := range resp.Rooms.Join {
		room := &MatrixRoom{
			RoomID:  roomID,
			Members: joinRoom.State.Members,
		}
		m.rooms[roomID] = room
	}

	return nil
}

// Start 开始监听消息
func (m *MatrixChannel) Start() error {
	if !m.running {
		if err := m.Connect(); err != nil {
			return err
		}
	}

	go m.messageLoop()
	return nil
}

// messageLoop 消息循环
func (m *MatrixChannel) messageLoop() {
	nextBatch := ""

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		var resp MatrixSyncResponse
		if err := m.client.Sync(m.accessToken, nextBatch, "5000ms", &resp); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// 处理新消息
		for roomID, joinRoom := range resp.Rooms.Join {
			for _, event := range joinRoom.Timeline.Events {
				if event.Type == "m.room.message" {
					m.handleMessage(roomID, event)
				}
			}
		}

		nextBatch = resp.NextBatch
	}
}

// handleMessage 处理消息
func (m *MatrixChannel) handleMessage(roomID string, event MatrixEvent) {
	msg := Message{
		ID:        event.EventID,
		Channel:   "matrix",
		ChannelID: roomID,
		Sender:    event.Sender,
		Content:   event.Content.Body,
		Type:      event.Content.MsgType,
		Timestamp: time.UnixMilli(event.Timestamp),
		Raw:       event,
	}

	m.mu.RLock()
	handler := m.handlers[roomID]
	m.mu.RUnlock()

	if handler != nil {
		go handler(msg)
	}
}

// SendMessage 发送消息
func (m *MatrixChannel) SendMessage(roomID, content string) error {
	sendReq := MatrixSendMessageRequest{
		MsgType: "m.text",
		Body:    content,
	}

	var eventID string
	return m.client.SendMessage(m.accessToken, roomID, &sendReq, &eventID)
}

// SendHTMLMessage 发送 HTML 消息
func (m *MatrixChannel) SendHTMLMessage(roomID, body, html string) error {
	content := map[string]interface{}{
		"msgtype": "m.text",
		"body":    body,
		"format":  "org.matrix.custom.html",
		"formatted_body": html,
	}

	var eventID string
	return m.client.SendMessage(m.accessToken, roomID, content, &eventID)
}

// SendFile 发送文件
func (m *MatrixChannel) SendFile(roomID, filename string, data []byte, contentType string) error {
	// 上传文件
	var uploadResp MatrixUploadResponse
	if err := m.client.UploadMedia(m.accessToken, data, contentType, filename, &uploadResp); err != nil {
		return fmt.Errorf("upload media failed: %w", err)
	}

	// 发送文件消息
	content := map[string]interface{}{
		"msgtype": "m.file",
		"body":    filename,
		"info": map[string]interface{}{
			"mimetype": contentType,
			"size":     len(data),
		},
		"url": uploadResp.ContentURI,
	}

	var eventID string
	return m.client.SendMessage(m.accessToken, roomID, content, &eventID)
}

// JoinRoom 加入房间
func (m *MatrixChannel) JoinRoom(roomIDOrAlias string) error {
	var roomID string
	if err := m.client.JoinRoom(m.accessToken, roomIDOrAlias, &roomID); err != nil {
		return fmt.Errorf("join room failed: %w", err)
	}

	m.mu.Lock()
	m.rooms[roomID] = &MatrixRoom{RoomID: roomID}
	m.mu.Unlock()

	return nil
}

// LeaveRoom 离开房间
func (m *MatrixChannel) LeaveRoom(roomID string) error {
	return m.client.LeaveRoom(m.accessToken, roomID)
}

// SetHandler 设置消息处理器
func (m *MatrixChannel) SetHandler(roomID string, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[roomID] = handler
}

// GetRooms 获取房间列表
func (m *MatrixChannel) GetRooms() []*MatrixRoom {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]*MatrixRoom, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// Stop 停止渠道
func (m *MatrixChannel) Stop() error {
	m.cancel()
	m.running = false
	return nil
}

// GetUserID 获取用户 ID
func (m *MatrixChannel) GetUserID() string {
	return m.userID
}

// Matrix API 请求/响应结构

type MatrixLoginRequest struct {
	Type     string `json:"type"`
	User     string `json:"identifier"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type MatrixLoginResponse struct {
	AccessToken  string `json:"access_token"`
	DeviceID     string `json:"device_id"`
	UserID       string `json:"user_id"`
	HomeServer   string `json:"home_server"`
	RefreshToken string `json:"refresh_token"`
}

type MatrixSyncResponse struct {
	NextBatch string         `json:"next_batch"`
	Rooms     MatrixRooms    `json:"rooms"`
	Presence  MatrixPresence `json:"presence"`
}

type MatrixRooms struct {
	Join  map[string]MatrixJoinRoom `json:"join"`
	Invite map[string]MatrixJoinRoom `json:"invite"`
	Leave map[string]struct{}       `json:"leave"`
}

type MatrixJoinRoom struct {
	State       MatrixRoomState `json:"state"`
	Timeline    MatrixTimeline  `json:"timeline"`
	Ephemeral   MatrixTimeline   `json:"ephemeral"`
	AccountData struct{}         `json:"account_data"`
}

type MatrixRoomState struct {
	Members []string    `json:"members"`
	Events  []MatrixEvent `json:"events"`
}

type MatrixTimeline struct {
	Events    []MatrixEvent `json:"events"`
	Limited   bool         `json:"limited"`
	PrevBatch string       `json:"prev_batch"`
}

type MatrixEvent struct {
	EventID   string                 `json:"event_id"`
	RoomID    string                 `json:"room_id"`
	Sender    string                 `json:"sender"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"origin_server_ts"`
	Content   MatrixMessageContent   `json:"content"`
}

type MatrixPresence struct {
	Events []struct{} `json:"events"`
}

type MatrixSendMessageRequest struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

type MatrixUploadResponse struct {
	ContentURI string `json:"content_uri"`
}

// MatrixHTTPClient Matrix HTTP 客户端
type MatrixHTTPClient struct {
	baseURL string
}

func NewMatrixHTTPClient(homeserver string) *MatrixHTTPClient {
	return &MatrixHTTPClient{
		baseURL: fmt.Sprintf("%s/_matrix/client/r0", homeserver),
	}
}

func (c *MatrixHTTPClient) Login(req *MatrixLoginRequest, resp *MatrixLoginResponse) error {
	return c.post("/login", req, resp)
}

func (c *MatrixHTTPClient) Sync(token, since, timeout string, resp *MatrixSyncResponse) error {
	url := fmt.Sprintf("/sync?filter={\"room\":{\"timeline\":{\"limit\":20}}}&timeout=%s", timeout)
	if since != "" {
		url += "&since=" + since
	}
	return c.getWithToken(url, token, resp)
}

func (c *MatrixHTTPClient) SendMessage(token, roomID string, content interface{}, eventID *string) error {
	txnID := fmt.Sprintf("t%d", time.Now().UnixNano())
	url := fmt.Sprintf("/rooms/%s/send/m.room.message/%s", roomID, txnID)
	return c.putWithToken(url, token, content, eventID)
}

func (c *MatrixHTTPClient) JoinRoom(token, roomIDOrAlias string, roomID *string) error {
	url := fmt.Sprintf("/join/%s", roomIDOrAlias)
	return c.postWithToken(url, token, struct{}{}, roomID)
}

func (c *MatrixHTTPClient) LeaveRoom(token, roomID string) error {
	url := fmt.Sprintf("/rooms/%s/leave", roomID)
	return c.postWithToken(url, token, struct{}{}, nil)
}

func (c *MatrixHTTPClient) UploadMedia(token string, data []byte, contentType, filename string, resp *MatrixUploadResponse) error {
	return c.uploadWithToken("/upload", token, data, contentType, filename, resp)
}

// 辅助方法
func (c *MatrixHTTPClient) get(url string, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpGet(fullURL)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, result)
}

func (c *MatrixHTTPClient) getWithToken(url, token string, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpGetWithAuth(fullURL, token)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, result)
}

func (c *MatrixHTTPClient) post(url string, req, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpPost(fullURL, req)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, result)
}

func (c *MatrixHTTPClient) postWithToken(url, token string, req, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpPostWithAuth(fullURL, token, req)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(resp, result)
	}
	return nil
}

func (c *MatrixHTTPClient) putWithToken(url, token string, req, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpPutWithAuth(fullURL, token, req)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(resp, result)
	}
	return nil
}

func (c *MatrixHTTPClient) uploadWithToken(url, token string, data []byte, contentType, filename string, result interface{}) error {
	fullURL := c.baseURL + url
	resp, err := httpUpload(fullURL, token, data, contentType, filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, result)
}

// HTTP 请求辅助函数
func httpGet(url string) ([]byte, error) {
	return doHTTPRequest("GET", url, nil, nil)
}

func httpGetWithAuth(url, token string) ([]byte, error) {
	return doHTTPRequest("GET", url, nil, map[string]string{"Authorization": "Bearer " + token})
}

func httpPost(url string, body interface{}) ([]byte, error) {
	return doHTTPRequest("POST", url, body, nil)
}

func httpPostWithAuth(url, token string, body interface{}) ([]byte, error) {
	return doHTTPRequest("POST", url, body, map[string]string{"Authorization": "Bearer " + token})
}

func httpPutWithAuth(url, token string, body interface{}) ([]byte, error) {
	return doHTTPRequest("PUT", url, body, map[string]string{"Authorization": "Bearer " + token})
}

func httpUpload(url, token string, data []byte, contentType, filename string) ([]byte, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  contentType,
		"Content-Length": fmt.Sprintf("%d", len(data)),
		"X-Filename":     filename,
	}
	return doHTTPRequest("PUT", url, data, headers)
}

func doHTTPRequest(method, url string, body interface{}, headers map[string]string) ([]byte, error) {
	// TODO: 实现实际的 HTTP 请求
	return nil, fmt.Errorf("not implemented")
}
