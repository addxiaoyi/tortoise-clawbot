package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// Config 应用配置
type Config struct {
	// 服务器
	Server struct {
		Port     int    `json:"port"`
		Host     string `json:"host"`
		LogLevel string `json:"log_level"`
	} `json:"server"`

	// AI 配置
	AI struct {
		Providers    []AIProvider `json:"providers"`
		DefaultModel string       `json:"default_model"`
		Routing      string       `json:"routing"` // "latency" | "load" | "cost"
	} `json:"ai"`

	// 渠道配置
	Channels struct {
		Telegram  ChannelConfig  `json:"telegram"`
		Discord   DiscordConfig   `json:"discord"`
		Slack     SlackConfig    `json:"slack"`
		WhatsApp  WhatsAppConfig `json:"whatsapp"`
		Teams     TeamsConfig    `json:"teams"`
	} `json:"channels"`

	// 设备发现
	Discovery struct {
		Enabled            bool `json:"enabled"`
		MDNS              bool `json:"mdns"`
		UPnP              bool `json:"upnp"`
		SSDP              bool `json:"ssdp"`
		Port              int  `json:"port"`
		AdvertiseInterval  int  `json:"advertise_interval"`
	} `json:"discovery"`

	// 数据库
	Database struct {
		Type   string         `json:"type"` // "memory" | "sqlite" | "redis"
		SQLite SQLiteConfig   `json:"sqlite"`
		Redis  RedisConfig    `json:"redis"`
	} `json:"database"`

	// 安全
	Security struct {
		APIKeyRequired bool        `json:"api_key_required"`
		APIKeys       []APIKey    `json:"api_keys"`
		JWTSecret     string      `json:"jwt_secret"`
		RateLimit     RateLimit   `json:"rate_limit"`
		CORS          CORSConfig  `json:"cors"`
	} `json:"security"`

	// 高级设置
	Advanced struct {
		MaxSessions        int    `json:"max_sessions"`
		SessionTimeout     int    `json:"session_timeout"`
		MessageBufferSize  int    `json:"message_buffer_size"`
		WorkerPoolSize     int    `json:"worker_pool_size"`
		MemoryPoolSize     int    `json:"memory_pool_size"`
		LogLevel           string `json:"log_level"`
		EnableMetrics      bool   `json:"enable_metrics"`
		EnableTracing      bool   `json:"enable_tracing"`
		OTELEndpoint       string `json:"otel_endpoint"`
	} `json:"advanced"`
}

// AIProvider AI 提供商配置
type AIProvider struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
}

// ChannelConfig 渠道基础配置
type ChannelConfig struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
}

// DiscordConfig Discord 配置
type DiscordConfig struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	GuildID string `json:"guild_id"`
}

// SlackConfig Slack 配置
type SlackConfig struct {
	Enabled       bool   `json:"enabled"`
	BotToken      string `json:"bot_token"`
	SigningSecret string `json:"signing_secret"`
	WebhookURL    string `json:"webhook_url"`
}

// WhatsAppConfig WhatsApp 配置
type WhatsAppConfig struct {
	Enabled     bool   `json:"enabled"`
	PhoneNumber string `json:"phone_number"`
	APIURL      string `json:"api_url"`
	APIToken    string `json:"api_token"`
}

// TeamsConfig Teams 配置
type TeamsConfig struct {
	Enabled     bool   `json:"enabled"`
	AppID       string `json:"app_id"`
	AppPassword string `json:"app_password"`
	TenantID    string `json:"tenant_id"`
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	Path string `json:"path"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	URL      string `json:"url"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// APIKey API Key 配置
type APIKey struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// RateLimit 限流配置
type RateLimit struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// Manager 配置管理器
type Manager struct {
	config *Config
	mu     sync.RWMutex
	path   string
}

// NewManager 创建配置管理器
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		path: configPath,
	}

	// 尝试加载现有配置
	if err := m.Load(); err != nil {
		// 如果文件不存在，使用默认配置
		m.config = DefaultConfig()
		if err := m.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// 从环境变量覆盖配置
	m.loadFromEnv()

	return m, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	cfg := &Config{}

	// 服务器默认值
	cfg.Server.Port = 18792
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.LogLevel = "info"

	// AI 默认值
	cfg.AI.DefaultModel = "gpt-4"
	cfg.AI.Routing = "latency"
	cfg.AI.Providers = []AIProvider{
		{
			ID:      "openai",
			Name:    "OpenAI",
			Enabled: false,
			Model:   "gpt-4",
			BaseURL: "https://api.openai.com/v1",
		},
		{
			ID:      "anthropic",
			Name:    "Anthropic",
			Enabled: false,
			Model:   "claude-3-sonnet-20240229",
			BaseURL: "https://api.anthropic.com",
		},
		{
			ID:      "ollama",
			Name:    "Ollama (本地)",
			Enabled: false,
			Model:   "llama2",
			BaseURL: "http://localhost:11434",
		},
	}

	// 渠道默认值
	cfg.Channels.Telegram.Enabled = false
	cfg.Channels.Discord.Enabled = false
	cfg.Channels.Slack.Enabled = false
	cfg.Channels.WhatsApp.Enabled = false
	cfg.Channels.Teams.Enabled = false

	// 设备发现默认值
	cfg.Discovery.Enabled = true
	cfg.Discovery.MDNS = true
	cfg.Discovery.UPnP = true
	cfg.Discovery.SSDP = false
	cfg.Discovery.Port = 18792
	cfg.Discovery.AdvertiseInterval = 30

	// 数据库默认值
	cfg.Database.Type = "memory"
	cfg.Database.SQLite.Path = "./data/tortoise.db"
	cfg.Database.Redis.URL = "redis://localhost:6379"
	cfg.Database.Redis.DB = 0

	// 安全默认值
	cfg.Security.APIKeyRequired = false
	cfg.Security.APIKeys = []APIKey{}
	cfg.Security.JWTSecret = ""
	cfg.Security.RateLimit.Enabled = true
	cfg.Security.RateLimit.RequestsPerMinute = 60
	cfg.Security.CORS.Enabled = true
	cfg.Security.CORS.AllowedOrigins = []string{"*"}

	// 高级设置默认值
	cfg.Advanced.MaxSessions = 100000
	cfg.Advanced.SessionTimeout = 86400
	cfg.Advanced.MessageBufferSize = 10000
	cfg.Advanced.WorkerPoolSize = 100
	cfg.Advanced.MemoryPoolSize = 1024
	cfg.Advanced.LogLevel = "info"
	cfg.Advanced.EnableMetrics = true
	cfg.Advanced.EnableTracing = false

	return cfg
}

// Load 加载配置
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.config = cfg
	return nil
}

// Save 保存配置
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Get 获取配置副本
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回深拷贝
	cfg := *m.config
	return &cfg
}

// Update 更新配置
func (m *Manager) Update(cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg
	return m.Save()
}

// GetAIProviders 获取启用的 AI 提供商
func (m *Manager) GetAIProviders() []AIProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabled := make([]AIProvider, 0)
	for _, p := range m.config.AI.Providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// SetAIProviderAPIKey 设置 AI 提供商的 API Key
func (m *Manager) SetAIProviderAPIKey(providerID, apiKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.AI.Providers {
		if m.config.AI.Providers[i].ID == providerID {
			m.config.AI.Providers[i].APIKey = apiKey
			return m.Save()
		}
	}
	return fmt.Errorf("provider not found: %s", providerID)
}

// AddAPIKey 添加 API Key
func (m *Manager) AddAPIKey(name string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := &APIKey{
		ID:        uuid.New().String(),
		Key:       generateAPIKey(),
		Name:      name,
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	m.config.Security.APIKeys = append(m.config.Security.APIKeys, *key)
	if err := m.Save(); err != nil {
		return nil, err
	}

	return key, nil
}

// RemoveAPIKey 移除 API Key
func (m *Manager) RemoveAPIKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, k := range m.config.Security.APIKeys {
		if k.ID == id {
			m.config.Security.APIKeys = append(
				m.config.Security.APIKeys[:i],
				m.config.Security.APIKeys[i+1:]...,
			)
			return m.Save()
		}
	}
	return fmt.Errorf("API key not found: %s", id)
}

// loadFromEnv 从环境变量加载配置
func (m *Manager) loadFromEnv() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// OpenAI
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		for i := range m.config.AI.Providers {
			if m.config.AI.Providers[i].ID == "openai" {
				m.config.AI.Providers[i].APIKey = key
				m.config.AI.Providers[i].Enabled = true
				break
			}
		}
	}

	// Anthropic
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		for i := range m.config.AI.Providers {
			if m.config.AI.Providers[i].ID == "anthropic" {
				m.config.AI.Providers[i].APIKey = key
				m.config.AI.Providers[i].Enabled = true
				break
			}
		}
	}

	// Telegram
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		m.config.Channels.Telegram.Token = token
		m.config.Channels.Telegram.Enabled = true
	}

	// Discord
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		m.config.Channels.Discord.Token = token
		m.config.Channels.Discord.Enabled = true
	}

	// Slack
	if token := os.Getenv("SLACK_BOT_TOKEN"); token != "" {
		m.config.Channels.Slack.BotToken = token
		m.config.Channels.Slack.Enabled = true
	}

	// Redis
	if url := os.Getenv("REDIS_URL"); url != "" {
		m.config.Database.Redis.URL = url
		m.config.Database.Type = "redis"
	}

	// JWT Secret
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		m.config.Security.JWTSecret = secret
	}
}

// generateAPIKey 生成随机 API Key
func generateAPIKey() string {
	return fmt.Sprintf("thp_%s", uuid.New().String())
}

// GetJWTSecret 获取 JWT Secret
func (m *Manager) GetJWTSecret() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret := m.config.Security.JWTSecret
	if secret == "" {
		secret = generateAPIKey()
		m.config.Security.JWTSecret = secret
	}
	return secret
}
