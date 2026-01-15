package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
)

// ConfigStore 配置存储
type ConfigStore struct {
	mu    sync.RWMutex
	data  *Config
	path  string
}

// Config 应用配置
type Config struct {
	// AI 配置
	AI AIConfig `json:"ai"`

	// 渠道配置
	Channels ChannelsConfig `json:"channels"`

	// 设备发现配置
	Discovery DiscoveryConfig `json:"discovery"`

	// 数据库配置
	Database DatabaseConfig `json:"database"`

	// 安全配置
	Security SecurityConfig `json:"security"`

	// 高级配置
	Advanced AdvancedConfig `json:"advanced"`

	// 内部字段（不返回给前端）
	APIKeys []APIKey `json:"-"`
}

// AIConfig AI 模型配置
type AIConfig struct {
	Providers    []AIProvider `json:"providers"`
	Routing      string       `json:"routing"` // latency | load | cost
	DefaultModel string       `json:"default_model"`
}

// AIProvider AI 提供商
type AIProvider struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

// ChannelsConfig 消息渠道配置
type ChannelsConfig struct {
	Telegram  TelegramConfig  `json:"telegram"`
	Discord   DiscordConfig   `json:"discord"`
	Slack     SlackConfig     `json:"slack"`
	WhatsApp  WhatsAppConfig  `json:"whatsapp"`
	Teams     TeamsConfig     `json:"teams"`
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	Enabled      bool   `json:"enabled"`
	BotToken     string `json:"bot_token"`
	AllowedChats string `json:"allowed_chats"`
}

// DiscordConfig Discord 配置
type DiscordConfig struct {
	Enabled   bool   `json:"enabled"`
	BotToken  string `json:"bot_token"`
	GuildID   string `json:"guild_id"`
}

// SlackConfig Slack 配置
type SlackConfig struct {
	Enabled        bool   `json:"enabled"`
	BotToken       string `json:"bot_token"`
	SigningSecret  string `json:"signing_secret"`
	WebhookURL     string `json:"webhook_url"`
}

// WhatsAppConfig WhatsApp 配置
type WhatsAppConfig struct {
	Enabled    bool   `json:"enabled"`
	PhoneNumber string `json:"phone_number"`
	APIURL     string `json:"api_url"`
	APIToken   string `json:"api_token"`
}

// TeamsConfig Teams 配置
type TeamsConfig struct {
	Enabled     bool   `json:"enabled"`
	AppID       string `json:"app_id"`
	AppPassword string `json:"app_password"`
	TenantID    string `json:"tenant_id"`
}

// DiscoveryConfig 设备发现配置
type DiscoveryConfig struct {
	Enabled            bool   `json:"enabled"`
	mDNS               bool   `json:"mdns"`
	UPnP               bool   `json:"upnp"`
	SSDP               bool   `json:"ssdp"`
	Port               int    `json:"port"`
	AdvertiseInterval  int    `json:"advertise_interval"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type   string          `json:"type"` // memory | sqlite | redis
	SQLite SQLiteConfig     `json:"sqlite"`
	Redis  RedisConfig      `json:"redis"`
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

// SecurityConfig 安全配置
type SecurityConfig struct {
	APIKeyRequired bool          `json:"api_key_required"`
	APIKeys       []APIKey      `json:"api_keys"`
	JWTSecret     string       `json:"jwt_secret"`
	RateLimit     RateLimitConfig `json:"rate_limit"`
	CORS          CORSConfig    `json:"cors"`
}

// APIKey API Key
type APIKey struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled            bool `json:"enabled"`
	RequestsPerMinute  int  `json:"requests_per_minute"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	MaxSessions       int  `json:"max_sessions"`
	SessionTimeout    int  `json:"session_timeout"`
	MessageBufferSize int  `json:"message_buffer_size"`
	WorkerPoolSize    int  `json:"worker_pool_size"`
	MemoryPoolSize    int  `json:"memory_pool_size"`
	LogLevel          string `json:"log_level"`
	EnableMetrics     bool   `json:"enable_metrics"`
	EnableTracing     bool   `json:"enable_tracing"`
}

// NewConfigStore 创建配置存储
func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		data: defaultConfig(),
	}
}

// GetConfig 获取完整配置（敏感字段会被过滤）
func (s *ConfigStore) GetConfig() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 创建副本并过滤敏感字段
	cfg := &Config{
		AI:        s.data.AI,
		Channels:  s.data.Channels,
		Discovery: s.data.Discovery,
		Database:  s.data.Database,
		Security:  s.filterSecurityConfig(),
		Advanced:  s.data.Advanced,
	}

	return cfg
}

// filterSecurityConfig 过滤安全配置中的敏感字段
func (s *ConfigStore) filterSecurityConfig() SecurityConfig {
	sec := s.data.Security
	// 只返回掩码后的 API Keys
	filteredKeys := make([]APIKey, len(sec.APIKeys))
	for i, k := range sec.APIKeys {
		filteredKeys[i] = APIKey{
			ID:        k.ID,
			Key:       maskString(k.Key, 8, "****"),
			Name:      k.Name,
			CreatedAt: k.CreatedAt,
		}
	}
	sec.APIKeys = filteredKeys
	// 不返回 JWT Secret
	sec.JWTSecret = ""
	return sec
}

// UpdateConfig 更新配置
func (s *ConfigStore) UpdateConfig(updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	// 解析更新
	var partial Config
	if err := json.Unmarshal(data, &partial); err != nil {
		return err
	}

	// 合并更新
	if providers, ok := updates["ai"]; ok {
		if aiMap, ok := providers.(map[string]interface{}); ok {
			if prov, ok := aiMap["providers"]; ok {
				s.mergeAIProviders(prov)
			}
			if routing, ok := aiMap["routing"].(string); ok {
				s.data.AI.Routing = routing
			}
			if defaultModel, ok := aiMap["default_model"].(string); ok {
				s.data.AI.DefaultModel = defaultModel
			}
		}
	}

	if channels, ok := updates["channels"]; ok {
		s.mergeChannels(channels)
	}

	if discovery, ok := updates["discovery"]; ok {
		s.mergeDiscovery(discovery)
	}

	if database, ok := updates["database"]; ok {
		s.mergeDatabase(database)
	}

	if security, ok := updates["security"]; ok {
		s.mergeSecurity(security)
	}

	if advanced, ok := updates["advanced"]; ok {
		s.mergeAdvanced(advanced)
	}

	return nil
}

func (s *ConfigStore) mergeAIProviders(providers interface{}) {
	provBytes, _ := json.Marshal(providers)
	var provs []AIProvider
	json.Unmarshal(provBytes, &provs)

	// 更新提供商配置，保留原有 API Key（如果新配置中没有提供）
	for i, newProv := range provs {
		for j, existing := range s.data.AI.Providers {
			if existing.ID == newProv.ID {
				// 如果新配置中没有 API Key，保留原有的
				if newProv.APIKey == "" {
					provs[i].APIKey = existing.APIKey
				}
				// 保留原有的 base_url 如果新配置为空
				if newProv.BaseURL == "" {
					provs[i].BaseURL = existing.BaseURL
				}
				_ = j
				break
			}
		}
	}
	s.data.AI.Providers = provs
}

func (s *ConfigStore) mergeChannels(channels interface{}) {
	chBytes, _ := json.Marshal(channels)
	var ch ChannelsConfig
	json.Unmarshal(chBytes, &ch)

	// 保留原有 Bot Token
	if ch.Telegram.BotToken == "" {
		ch.Telegram.BotToken = s.data.Channels.Telegram.BotToken
	}
	if ch.Discord.BotToken == "" {
		ch.Discord.BotToken = s.data.Channels.Discord.BotToken
	}
	if ch.Slack.BotToken == "" {
		ch.Slack.BotToken = s.data.Channels.Slack.BotToken
	}
	if ch.Slack.SigningSecret == "" {
		ch.Slack.SigningSecret = s.data.Channels.Slack.SigningSecret
	}
	if ch.WhatsApp.APIToken == "" {
		ch.WhatsApp.APIToken = s.data.Channels.WhatsApp.APIToken
	}
	if ch.Teams.AppPassword == "" {
		ch.Teams.AppPassword = s.data.Channels.Teams.AppPassword
	}

	s.data.Channels = ch
}

func (s *ConfigStore) mergeDiscovery(discovery interface{}) {
	discBytes, _ := json.Marshal(discovery)
	var disc DiscoveryConfig
	json.Unmarshal(discBytes, &disc)
	s.data.Discovery = disc
}

func (s *ConfigStore) mergeDatabase(database interface{}) {
	dbBytes, _ := json.Marshal(database)
	var db DatabaseConfig
	json.Unmarshal(dbBytes, &db)
	s.data.Database = db
}

func (s *ConfigStore) mergeSecurity(security interface{}) {
	secBytes, _ := json.Marshal(security)
	var sec SecurityConfig
	json.Unmarshal(secBytes, &sec)

	// 保留原有 JWT Secret
	if sec.JWTSecret == "" {
		sec.JWTSecret = s.data.Security.JWTSecret
	}
	// 保留原有 API Keys
	if len(sec.APIKeys) == 0 {
		sec.APIKeys = s.data.Security.APIKeys
	}

	s.data.Security = sec
}

func (s *ConfigStore) mergeAdvanced(advanced interface{}) {
	advBytes, _ := json.Marshal(advanced)
	var adv AdvancedConfig
	json.Unmarshal(advBytes, &adv)
	s.data.Advanced = adv
}

// GetSecret 获取敏感配置（仅内部使用）
func (s *ConfigStore) GetSecret(providerID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.data.AI.Providers {
		if p.ID == providerID {
			return p.APIKey
		}
	}
	return ""
}

// GetChannelSecret 获取渠道 Token
func (s *ConfigStore) GetChannelSecret(channel string) (token string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch channel {
	case "telegram":
		return s.data.Channels.Telegram.BotToken
	case "discord":
		return s.data.Channels.Discord.BotToken
	case "slack":
		return s.data.Channels.Slack.BotToken
	case "whatsapp":
		return s.data.Channels.WhatsApp.APIToken
	case "teams":
		return s.data.Channels.Teams.AppPassword
	}
	return ""
}

// GenerateAPIKey 生成新的 API Key
func (s *ConfigStore) GenerateAPIKey(name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	key := "sk_" + hex.EncodeToString(bytes)

	s.data.Security.APIKeys = append(s.data.Security.APIKeys, APIKey{
		ID:        generateID(),
		Key:       key,
		Name:      name,
		CreatedAt: "2024-01-15", // TODO: 使用实际时间
	})

	return key, nil
}

// DeleteAPIKey 删除 API Key
func (s *ConfigStore) DeleteAPIKey(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, k := range s.data.Security.APIKeys {
		if k.ID == id {
			s.data.Security.APIKeys = append(s.data.Security.APIKeys[:i], s.data.Security.APIKeys[i+1:]...)
			return true
		}
	}
	return false
}

// GetAPIKey 获取 API Key（仅内部使用）
func (s *ConfigStore) GetAPIKey(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, k := range s.data.Security.APIKeys {
		if k.Key == key {
			return true
		}
	}
	return false
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func maskString(s string, visible int, replacement string) string {
	if len(s) <= visible*2 {
		return replacement
	}
	return s[:visible] + replacement + s[len(s)-visible:]
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		AI: AIConfig{
			Providers: []AIProvider{
				{
					ID:       "openai",
					Name:     "OpenAI",
					Enabled:  true,
					APIKey:   "",
					Model:    "gpt-4",
					BaseURL:  "https://api.openai.com/v1",
				},
				{
					ID:       "anthropic",
					Name:     "Anthropic",
					Enabled:  false,
					APIKey:   "",
					Model:    "claude-3-sonnet-20240229",
					BaseURL:  "https://api.anthropic.com",
				},
				{
					ID:       "ollama",
					Name:     "Ollama (本地)",
					Enabled:  false,
					APIKey:   "",
					Model:    "llama2",
					BaseURL:  "http://localhost:11434",
				},
			},
			Routing:      "latency",
			DefaultModel: "gpt-4",
		},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{
				Enabled:      false,
				BotToken:     "",
				AllowedChats: "",
			},
			Discord: DiscordConfig{
				Enabled:   false,
				BotToken:  "",
				GuildID:   "",
			},
			Slack: SlackConfig{
				Enabled:        false,
				BotToken:       "",
				SigningSecret:  "",
				WebhookURL:     "",
			},
			WhatsApp: WhatsAppConfig{
				Enabled:    false,
				PhoneNumber: "",
				APIURL:     "",
				APIToken:   "",
			},
			Teams: TeamsConfig{
				Enabled:     false,
				AppID:       "",
				AppPassword: "",
				TenantID:    "",
			},
		},
		Discovery: DiscoveryConfig{
			Enabled:           true,
			mDNS:              true,
			UPnP:              true,
			SSDP:              false,
			Port:              18792,
			AdvertiseInterval: 30,
		},
		Database: DatabaseConfig{
			Type: "memory",
			SQLite: SQLiteConfig{
				Path: "./data/tortoise.db",
			},
			Redis: RedisConfig{
				URL:      "redis://localhost:6379",
				Password: "",
				DB:       0,
			},
		},
		Security: SecurityConfig{
			APIKeyRequired: false,
			APIKeys:        []APIKey{},
			JWTSecret:      "",
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute:  60,
			},
			CORS: CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
			},
		},
		Advanced: AdvancedConfig{
			MaxSessions:       100000,
			SessionTimeout:     86400,
			MessageBufferSize:  10000,
			WorkerPoolSize:     10000,
			MemoryPoolSize:     1024,
			LogLevel:           "info",
			EnableMetrics:      true,
			EnableTracing:      false,
		},
	}
}
