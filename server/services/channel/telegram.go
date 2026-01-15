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

// ========== Telegram 真实实现 ==========

// TelegramHandler Telegram 渠道处理器
type TelegramHandler struct {
	token       string
	webhookURL  string
	apiBaseURL  string
	client      *http.Client
	updates     chan *TelegramUpdate
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	running     bool
	messageHandler func(from string, message string, metadata map[string]interface{})
}

// TelegramUpdate Telegram 更新
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message *TelegramMessage `json:"message,omitempty"`
	EditedMessage *TelegramMessage `json:"edited_message,omitempty"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query,omitempty"`
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
	MessageID int64 `json:"message_id"`
	From     *TelegramUser `json:"from,omitempty"`
	Chat     *TelegramChat `json:"chat"`
	Text     string `json:"text,omitempty"`
	Photo    []TelegramPhoto `json:"photo,omitempty"`
	Document *TelegramDocument `json:"document,omitempty"`
	Venue    *TelegramVenue `json:"venue,omitempty"`
	Location *TelegramLocation `json:"location,omitempty"`
	Date     int64 `json:"date"`
}

// TelegramUser Telegram 用户
type TelegramUser struct {
	ID           int64 `json:"id"`
	IsBot        bool `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// TelegramChat Telegram 聊天
type TelegramChat struct {
	ID        int64 `json:"id"`
	Type      string `json:"type"` // private, group, supergroup, channel
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// TelegramPhoto Telegram 照片
type TelegramPhoto struct {
	FileID   string `json:"file_id"`
	UniqueID string `json:"file_unique_id"`
	Width    int `json:"width"`
	Height   int `json:"height"`
	FileSize int `json:"file_size,omitempty"`
}

// TelegramDocument Telegram 文档
type TelegramDocument struct {
	FileID   string `json:"file_id"`
	UniqueID string `json:"file_unique_id"`
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int `json:"file_size,omitempty"`
}

// TelegramVenue Telegram 地点
type TelegramVenue struct {
	Location *TelegramLocation `json:"location"`
	Title    string `json:"title"`
	Address  string `json:"address"`
}

// TelegramLocation Telegram 位置
type TelegramLocation struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
}

// TelegramCallbackQuery Telegram 回调查询
type TelegramCallbackQuery struct {
	ID   string `json:"id"`
	From *TelegramUser `json:"from"`
	Data string `json:"data,omitempty"`
}

// TelegramResponse Telegram API 响应
type TelegramResponse struct {
	OK          bool        `json:"ok"`
	ErrorCode   int         `json:"error_code,omitempty"`
	Description string      `json:"description,omitempty"`
	Result     interface{} `json:"result"`
}

// newTelegramHandler 创建 Telegram 处理器
func newTelegramHandler(token string) *TelegramHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &TelegramHandler{
		token:      token,
		apiBaseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		updates: make(chan *TelegramUpdate, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Connect 连接 Telegram Bot
func (h *TelegramHandler) Connect() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	// 获取机器人信息
	if err := h.getMe(); err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	// 启动轮询
	go h.pollUpdates()

	log.Printf("Telegram bot connected successfully")
	return nil
}

// Disconnect 断开连接
func (h *TelegramHandler) Disconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.cancel()
	h.running = false
	close(h.updates)

	log.Printf("Telegram bot disconnected")
	return nil
}

// SendMessage 发送消息
func (h *TelegramHandler) SendMessage(chatID string, text string) error {
	return h.sendMessage(chatID, text, nil)
}

// SendMessageWithKeyboard 发送带键盘的消息
func (h *TelegramHandler) SendMessageWithKeyboard(chatID, text string, keyboard *TelegramInlineKeyboardMarkup) error {
	return h.sendMessage(chatID, text, keyboard)
}

func (h *TelegramHandler) sendMessage(chatID, text string, keyboard *TelegramInlineKeyboardMarkup) error {
	reqBody := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"parse_mode": "Markdown",
	}

	if keyboard != nil {
		reqBody["reply_markup"] = keyboard
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := h.post("/sendMessage", body)
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}

	return nil
}

// SendPhoto 发送照片
func (h *TelegramHandler) SendPhoto(chatID, photoURL, caption string) error {
	reqBody := map[string]interface{}{
		"chat_id": chatID,
		"photo":   photoURL,
		"caption": caption,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := h.post("/sendPhoto", body)
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}

	return nil
}

// SendDocument 发送文档
func (h *TelegramHandler) SendDocument(chatID, documentURL, caption string) error {
	reqBody := map[string]interface{}{
		"chat_id":  chatID,
		"document": documentURL,
		"caption":  caption,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := h.post("/sendDocument", body)
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}

	return nil
}

// SendVenue 发送位置
func (h *TelegramHandler) SendVenue(chatID string, lat, lon float64, title, address string) error {
	reqBody := map[string]interface{}{
		"chat_id":  chatID,
		"latitude":  lat,
		"longitude": lon,
		"title":    title,
		"address":  address,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := h.post("/sendVenue", body)
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}

	return nil
}

// AnswerCallbackQuery 回答回调查询
func (h *TelegramHandler) AnswerCallbackQuery(queryID, text string) error {
	reqBody := map[string]interface{}{
		"callback_query_id": queryID,
		"text":             text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	_, err = h.post("/answerCallbackQuery", body)
	return err
}

// SetWebhook 设置 Webhook
func (h *TelegramHandler) SetWebhook(url string) error {
	reqBody := map[string]interface{}{
		"url": url,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := h.post("/setWebhook", body)
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}

	return nil
}

// DeleteWebhook 删除 Webhook
func (h *TelegramHandler) DeleteWebhook() error {
	_, err := h.get("/deleteWebhook")
	return err
}

// OnMessage 设置消息处理函数
func (h *TelegramHandler) OnMessage(handler func(from string, message string, metadata map[string]interface{})) {
	h.messageHandler = handler
}

// HandleWebhook 处理 Webhook
func (h *TelegramHandler) HandleWebhook(update *TelegramUpdate) {
	h.processUpdate(update)
}

// ========== 私有方法 ==========

func (h *TelegramHandler) getMe() error {
	resp, err := h.get("/getMe")
	if err != nil {
		return err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("getMe error: %s", result.Description)
	}

	var user struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
	}
	if data, err := json.Marshal(result.Result); err == nil {
		json.Unmarshal(data, &user)
		log.Printf("Bot info: @%s (%s)", user.Username, user.FirstName)
	}

	return nil
}

func (h *TelegramHandler) pollUpdates() {
	offset := int64(0)

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			updates, err := h.getUpdates(offset)
			if err != nil {
				log.Printf("Error getting updates: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				h.processUpdate(update)
				offset = update.UpdateID + 1
			}
		}
	}
}

func (h *TelegramHandler) getUpdates(offset int64) ([]*TelegramUpdate, error) {
	url := fmt.Sprintf("/getUpdates?offset=%d&timeout=30&limit=100", offset)
	resp, err := h.get(url)
	if err != nil {
		return nil, err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("getUpdates error: %s", result.Description)
	}

	updates := make([]*TelegramUpdate, 0)
	if data, err := json.Marshal(result.Result); err == nil {
		json.Unmarshal(data, &updates)
	}

	return updates, nil
}

func (h *TelegramHandler) processUpdate(update *TelegramUpdate) {
	if update.Message != nil {
		metadata := map[string]interface{}{
			"message_id":   update.Message.MessageID,
			"chat_id":     update.Message.Chat.ID,
			"chat_type":   update.Message.Chat.Type,
			"username":    update.Message.From.Username,
			"first_name":  update.Message.From.FirstName,
			"timestamp":   update.Message.Date,
		}

		// 处理不同类型的消息
		if update.Message.Text != "" {
			if h.messageHandler != nil {
				h.messageHandler(fmt.Sprintf("%d", update.Message.Chat.ID), update.Message.Text, metadata)
			}
		} else if len(update.Message.Photo) > 0 {
			if h.messageHandler != nil {
				h.messageHandler(fmt.Sprintf("%d", update.Message.Chat.ID), "[Photo]", metadata)
			}
		} else if update.Message.Document != nil {
			if h.messageHandler != nil {
				h.messageHandler(fmt.Sprintf("%d", update.Message.Chat.ID), "[Document]", metadata)
			}
		} else if update.Message.Location != nil {
			if h.messageHandler != nil {
				h.messageHandler(fmt.Sprintf("%d", update.Message.Chat.ID), 
					fmt.Sprintf("[Location: %.4f, %.4f]", update.Message.Location.Latitude, update.Message.Location.Longitude), 
					metadata)
			}
		}
	}

	if update.CallbackQuery != nil {
		if h.messageHandler != nil {
			metadata := map[string]interface{}{
				"query_id": update.CallbackQuery.ID,
				"data":     update.CallbackQuery.Data,
			}
			h.messageHandler(fmt.Sprintf("%d", update.CallbackQuery.From.ID), update.CallbackQuery.Data, metadata)
		}
	}
}

func (h *TelegramHandler) get(path string) ([]byte, error) {
	url := h.apiBaseURL + path
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (h *TelegramHandler) post(path string, body []byte) ([]byte, error) {
	url := h.apiBaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// TelegramInlineKeyboardMarkup 内联键盘
type TelegramInlineKeyboardMarkup struct {
	InlineKeyboard [][]TelegramInlineKeyboardButton `json:"inline_keyboard"`
}

// TelegramInlineKeyboardButton 内联键盘按钮
type TelegramInlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// NewInlineKeyboard 创建内联键盘
func NewInlineKeyboard() *TelegramInlineKeyboardMarkup {
	return &TelegramInlineKeyboardMarkup{
		InlineKeyboard: make([][]TelegramInlineKeyboardButton, 0),
	}
}

// AddRow 添加一行
func (k *TelegramInlineKeyboardMarkup) AddRow(buttons ...TelegramInlineKeyboardButton) *TelegramInlineKeyboardMarkup {
	k.InlineKeyboard = append(k.InlineKeyboard, buttons)
	return k
}

// NewButton 创建按钮
func NewButton(text, callbackData string) TelegramInlineKeyboardButton {
	return TelegramInlineKeyboardButton{
		Text:         text,
		CallbackData: callbackData,
	}
}

// NewURLButton 创建 URL 按钮
func NewURLButton(text, url string) TelegramInlineKeyboardButton {
	return TelegramInlineKeyboardButton{
		Text: text,
		URL:  url,
	}
}
