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
)

// ============ Telegram Bot ============

// TelegramChannel Telegram 渠道
type TelegramChannel struct {
	botToken     string
	apiURL       string
	allowedChats map[int64]bool
	aiEngine    *ai.Engine
	offset      int64
	running     bool
	mu          sync.RWMutex
}

// TelegramUpdate Telegram 更新
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
	MessageID int64  `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      *TelegramChat `json:"chat"`
	Text      string `json:"text,omitempty"`
	Date      int64  `json:"date"`
}

// TelegramUser Telegram 用户
type TelegramUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
}

// TelegramChat Telegram 聊天
type TelegramChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
}

// NewTelegramChannel 创建 Telegram 渠道
func NewTelegramChannel(botToken string, allowedChats []int64) *TelegramChannel {
	ch := &TelegramChannel{
		botToken:     botToken,
		apiURL:       "https://api.telegram.org/bot" + botToken,
		allowedChats: make(map[int64]bool),
		offset:       0,
	}

	for _, chatID := range allowedChats {
		ch.allowedChats[chatID] = true
	}

	return ch
}

// SetAIEngine 设置 AI 引擎
func (c *TelegramChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Telegram Bot
func (c *TelegramChannel) Start() error {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	go c.poll()
	log.Println("Telegram Bot started")
	return nil
}

// Stop 停止 Telegram Bot
func (c *TelegramChannel) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	log.Println("Telegram Bot stopped")
}

// poll 轮询获取更新
func (c *TelegramChannel) poll() {
	for {
		c.mu.RLock()
		running := c.running
		c.mu.RUnlock()

		if !running {
			break
		}

		updates, err := c.getUpdates()
		if err != nil {
			log.Printf("Telegram polling error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			c.handleUpdate(update)
			c.offset = update.UpdateID + 1
		}
	}
}

// getUpdates 获取更新
func (c *TelegramChannel) getUpdates() ([]TelegramUpdate, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", c.apiURL, c.offset)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool              `json:"ok"`
		Result []TelegramUpdate `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("Telegram API error")
	}

	return result.Result, nil
}

// handleUpdate 处理更新
func (c *TelegramChannel) handleUpdate(update TelegramUpdate) {
	if update.Message == nil {
		return
	}

	msg := update.Message

	// 检查是否允许的聊天
	if len(c.allowedChats) > 0 && !c.allowedChats[msg.Chat.ID] {
		log.Printf("Message from unauthorized chat: %d", msg.Chat.ID)
		return
	}

	// 处理命令
	if strings.HasPrefix(msg.Text, "/") {
		c.handleCommand(msg)
		return
	}

	// 处理普通消息
	c.handleMessage(msg)
}

// handleCommand 处理命令
func (c *TelegramChannel) handleCommand(msg *TelegramMessage) {
	text := msg.Text

	switch {
	case text == "/start":
		c.sendMessage(msg.Chat.ID, "欢迎使用 Tortoise AI Bot！\n\n发送任意消息与我对话。")
	case text == "/help":
		c.sendMessage(msg.Chat.ID, "可用命令：\n/start - 开始\n/help - 帮助\n/model - 查看当前模型")
	case text == "/model":
		if c.aiEngine != nil {
			c.sendMessage(msg.Chat.ID, "当前使用 GPT-4")
		} else {
			c.sendMessage(msg.Chat.ID, "AI 未配置")
		}
	default:
		c.sendMessage(msg.Chat.ID, "未知命令，请发送普通消息与我对话。")
	}
}

// handleMessage 处理消息
func (c *TelegramChannel) handleMessage(msg *TelegramMessage) {
	// 发送"正在输入"提示
	c.sendChatAction(msg.Chat.ID, "typing")

	var response string

	if c.aiEngine != nil {
		// 使用真实 AI
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   4096,
			Messages: []ai.Message{
				{Role: "user", Content: msg.Text},
			},
		}

		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，AI 服务出错：%v", err)
		} else {
			response = resp.Content
		}
	} else {
		// 模拟响应
		response = fmt.Sprintf("你说了: %s\n\n(AI 未配置，无法回复)", msg.Text)
	}

	c.sendMessage(msg.Chat.ID, response)
}

// sendMessage 发送消息
func (c *TelegramChannel) sendMessage(chatID int64, text string) error {
	url := fmt.Sprintf("%s/sendMessage", c.apiURL)
	
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// sendChatAction 发送聊天动作
func (c *TelegramChannel) sendChatAction(chatID int64, action string) error {
	url := fmt.Sprintf("%s/sendChatAction", c.apiURL)
	
	payload := map[string]interface{}{
		"chat_id": chatID,
		"action":  action,
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetMe 获取 Bot 信息
func (c *TelegramChannel) GetMe() (*TelegramUser, error) {
	url := fmt.Sprintf("%s/getMe", c.apiURL)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool          `json:"ok"`
		Result *TelegramUser `json:"result"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get bot info")
	}

	return result.Result, nil
}
