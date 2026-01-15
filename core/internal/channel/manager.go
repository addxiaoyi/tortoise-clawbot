package channel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ChannelType 渠道类型
type ChannelType string

const (
	ChannelTelegram  ChannelType = "telegram"
	ChannelDiscord  ChannelType = "discord"
	ChannelSlack    ChannelType = "slack"
	ChannelTeams    ChannelType = "teams"
	ChannelWhatsApp ChannelType = "whatsapp"
	ChannelWeb      ChannelType = "web"
	ChannelAPI      ChannelType = "api"
)

// Message 消息结构
type Message struct {
	ID         string
	Channel    ChannelType
	From      string
	To        string
	Content   string
	Type      string
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// Config 渠道配置
type Config struct {
	BufferSize    int
	Workers       int
	RetryAttempts int
	RetryDelay   time.Duration
}

// Handler 消息处理函数
type Handler func(msg *Message) error

// Manager 消息渠道管理器
type Manager struct {
	config Config

	// 渠道注册
	channels map[ChannelType]*Channel

	// 消息缓冲
	messageQueue chan *Message
	incoming    chan *Message

	// 处理器
	handlers map[string]Handler

	// 统计
	stats Stats

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	running atomic.Bool
}

// Stats 渠道统计
type Stats struct {
	MessagesReceived  atomic.Int64
	MessagesSent     atomic.Int64
	MessagesDropped  atomic.Int64
	ChannelsActive   atomic.Int64
	AvgLatencyUs    atomic.Int64
}

// Channel 渠道接口
type Channel interface {
	Connect() error
	Disconnect() error
	Send(msg *Message) error
	IsConnected() bool
	Type() ChannelType
}

// NewManager 创建渠道管理器
func NewManager(cfg Config) *Manager {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 10000
	}
	if cfg.Workers == 0 {
		cfg.Workers = 100
	}
	if cfg.RetryAttempts == 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 100 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:      cfg,
		channels:    make(map[ChannelType]*Channel),
		messageQueue: make(chan *Message, cfg.BufferSize),
		incoming:    make(chan *Message, cfg.BufferSize),
		handlers:   make(map[string]Handler),
		ctx:        ctx,
		cancel:      cancel,
	}

	m.running.Store(true)

	// 启动工作协程
	m.startWorkers()

	return m
}

// startWorkers 启动工作协程
func (m *Manager) startWorkers() {
	for i := 0; i < m.config.Workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// worker 消息处理工作协程
func (m *Manager) worker(id int) {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case msg := <-m.incoming:
			if msg == nil {
				continue
			}

			start := time.Now()

			// 调用处理器
			if handler, ok := m.handlers[msg.Type]; ok {
				if err := handler(msg); err != nil {
					// 重试逻辑
					m.retry(msg, handler)
				}
			}

			// 更新统计
			m.stats.MessagesReceived.Add(1)
			m.stats.AvgLatencyUs.Store(time.Since(start).Microseconds())
		}
	}
}

// retry 重试机制
func (m *Manager) retry(msg *Message, handler Handler) {
	for attempt := 1; attempt <= m.config.RetryAttempts; attempt++ {
		time.Sleep(m.config.RetryDelay)

		if err := handler(msg); err == nil {
			return
		}
	}
	m.stats.MessagesDropped.Add(1)
}

// RegisterChannel 注册渠道
func (m *Manager) RegisterChannel(ch Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channels[ch.Type()] = &ch
	m.stats.ChannelsActive.Add(1)
}

// UnregisterChannel 注销渠道
func (m *Manager) UnregisterChannel(chType ChannelType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, ok := m.channels[chType]; ok {
		(*ch).Disconnect()
		delete(m.channels, chType)
		m.stats.ChannelsActive.Add(-1)
	}
}

// RegisterHandler 注册消息处理器
func (m *Manager) RegisterHandler(msgType string, handler Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[msgType] = handler
}

// Send 发送消息
func (m *Manager) Send(chType ChannelType, msg *Message) error {
	m.mu.RLock()
	ch, ok := m.channels[chType]
	m.mu.RUnlock()

	if !ok {
		return ErrChannelNotFound
	}

	if !(*ch).IsConnected() {
		if err := (*ch).Connect(); err != nil {
			return err
		}
	}

	start := time.Now()
	err := (*ch).Send(msg)
	if err == nil {
		m.stats.MessagesSent.Add(1)
	}
	m.stats.AvgLatencyUs.Store(time.Since(start).Microseconds())

	return err
}

// Queue 队列消息
func (m *Manager) Queue(msg *Message) {
	select {
	case m.incoming <- msg:
	default:
		m.stats.MessagesDropped.Add(1)
	}
}

// QueueBatch 批量队列
func (m *Manager) QueueBatch(msgs []*Message) {
	for _, msg := range msgs {
		m.Queue(msg)
	}
}

// Broadcast 广播到所有渠道
func (m *Manager) Broadcast(msg *Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.channels {
		go func(channel *Channel) {
			if (*channel).IsConnected() {
				(*channel).Send(msg)
			}
		}(channel)
	}
}

// GetChannels 获取所有渠道
func (m *Manager) GetChannels() []ChannelType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]ChannelType, 0, len(m.channels))
	for chType := range m.channels {
		types = append(types, chType)
	}
	return types
}

// Stats 获取统计
func (m *Manager) Stats() Stats {
	return Stats{
		MessagesReceived: m.stats.MessagesReceived,
		MessagesSent:     m.stats.MessagesSent,
		MessagesDropped:  m.stats.MessagesDropped,
		ChannelsActive:  m.stats.ChannelsActive,
		AvgLatencyUs:   m.stats.AvgLatencyUs,
	}
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.running.Store(false)
	m.cancel()
	m.wg.Wait()

	// 断开所有渠道
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ch := range m.channels {
		(*ch).Disconnect()
	}
}

// Errors
var (
	ErrChannelNotFound = &ChannelError{Code: "CHANNEL_NOT_FOUND", Message: "渠道未找到"}
)

// ChannelError 渠道错误
type ChannelError struct {
	Code    string
	Message string
}

func (e *ChannelError) Error() string {
	return e.Code + ": " + e.Message
}

// ==================== Telegram 渠道实现 ====================

type TelegramChannel struct {
	botToken string
	chatIDs  []string
	connected bool
	mu       sync.RWMutex
}

func NewTelegramChannel(token string, chats []string) *TelegramChannel {
	return &TelegramChannel{
		botToken: token,
		chatIDs:  chats,
	}
}

func (c *TelegramChannel) Connect() error {
	// 连接 Telegram Bot API
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	return nil
}

func (c *TelegramChannel) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return nil
}

func (c *TelegramChannel) Send(msg *Message) error {
	// 发送消息到 Telegram
	return nil
}

func (c *TelegramChannel) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *TelegramChannel) Type() ChannelType {
	return ChannelTelegram
}

// ==================== Discord 渠道实现 ====================

type DiscordChannel struct {
	token    string
	guildID string
	channels map[string]string
	connected bool
	mu       sync.RWMutex
}

func NewDiscordChannel(token, guildID string) *DiscordChannel {
	return &DiscordChannel{
		token:    token,
		guildID: guildID,
		channels: make(map[string]string),
	}
}

func (c *DiscordChannel) Connect() error {
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	return nil
}

func (c *DiscordChannel) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return nil
}

func (c *DiscordChannel) Send(msg *Message) error {
	return nil
}

func (c *DiscordChannel) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *DiscordChannel) Type() ChannelType {
	return ChannelDiscord
}

// ==================== Slack 渠道实现 ====================

type SlackChannel struct {
	token   string
	webhook string
	connected bool
	mu       sync.RWMutex
}

func NewSlackChannel(token string) *SlackChannel {
	return &SlackChannel{
		token: token,
	}
}

func (c *SlackChannel) Connect() error {
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	return nil
}

func (c *SlackChannel) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return nil
}

func (c *SlackChannel) Send(msg *Message) error {
	return nil
}

func (c *SlackChannel) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *SlackChannel) Type() ChannelType {
	return ChannelSlack
}

// CreateMessage 创建消息
func CreateMessage(chType ChannelType, from, to, content string) *Message {
	return &Message{
		ID:         uuid.New().String(),
		Channel:    chType,
		From:       from,
		To:         to,
		Content:    content,
		Timestamp:  time.Now(),
	}
}
