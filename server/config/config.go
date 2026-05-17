package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	AI       AIConfig
	Channels ChannelsConfig
	Plugins  PluginsConfig
	Matrix   MatrixConfig
	Email    EmailConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Address   string
	SecretKey string
	Debug     bool
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// AIConfig AI服务配置
type AIConfig struct {
	Providers []AIProviderConfig
}

// AIProviderConfig AI提供商配置
type AIProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
	Models  []string
}

// ChannelsConfig 渠道配置
type ChannelsConfig struct {
	Channels []ChannelConfig
}

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Type    string
	Token   string
	Webhook string
}

// PluginsConfig 插件配置
type PluginsConfig struct {
	Directory string
	Registry  string
}

// MatrixConfig Matrix 配置
type MatrixConfig struct {
	Enabled     bool
	Homeserver  string
	Username    string
	Password    string
	DeviceID    string
	RoomID      string
}

// EmailConfig Email 配置
type EmailConfig struct {
	Enabled       bool
	DisplayName   string
	Username      string
	Password      string
	SMTPHost      string
	SMTPPort      int
	SMTPTLS       bool
	SMTPEnabled   bool
	IMAPHost      string
	IMAPPort      int
	IMAPEnabled   bool
	PollInterval  int
}

// Load 加载配置
func Load() *Config {
	// 设置默认值
	viper.SetDefault("app.address", ":8080")
	viper.SetDefault("app.debug", false)
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/tortoise.db")

	// 从环境变量读取
	viper.AutomaticEnv()

	// 读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		// 配置文件不存在，使用默认值
	}

	cfg := &Config{
		App: AppConfig{
			Address:   viper.GetString("app.address"),
			SecretKey: viper.GetString("app.secret_key"),
			Debug:     viper.GetBool("app.debug"),
		},
		Database: DatabaseConfig{
			Type:    viper.GetString("database.type"),
			Host:    viper.GetString("database.host"),
			Port:    viper.GetInt("database.port"),
			User:    viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			Database: viper.GetString("database.database"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		AI: AIConfig{
			Providers: []AIProviderConfig{
				{
					Name:    "openai",
					BaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
					APIKey:  getEnv("OPENAI_API_KEY", ""),
					Models:  []string{"gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"},
				},
				{
					Name:    "anthropic",
					BaseURL: getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
					APIKey:  getEnv("ANTHROPIC_API_KEY", ""),
					Models:  []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
				},
			},
		},
		Channels: ChannelsConfig{
			Channels: []ChannelConfig{
				{
					Type:  "telegram",
					Token: getEnv("TELEGRAM_BOT_TOKEN", ""),
				},
				{
					Type:  "discord",
					Token: getEnv("DISCORD_BOT_TOKEN", ""),
				},
			},
		},
		Plugins: PluginsConfig{
			Directory: getEnv("PLUGINS_DIR", "./plugins"),
			Registry:  getEnv("PLUGINS_REGISTRY", "https://plugins.tortoise.ai"),
		},
		Matrix: MatrixConfig{
			Enabled:     viper.GetBool("matrix.enabled"),
			Homeserver:  getEnv("MATRIX_HOMESERVER", "https://matrix.org"),
			Username:    getEnv("MATRIX_USERNAME", ""),
			Password:    getEnv("MATRIX_PASSWORD", ""),
			DeviceID:    getEnv("MATRIX_DEVICE_ID", "TortoiseBot"),
			RoomID:      viper.GetString("matrix.room_id"),
		},
		Email: EmailConfig{
			Enabled:      viper.GetBool("email.enabled"),
			DisplayName:  getEnv("EMAIL_DISPLAY_NAME", "Tortoise Bot"),
			Username:     getEnv("EMAIL_USERNAME", ""),
			Password:     getEnv("EMAIL_PASSWORD", ""),
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     viper.GetInt("smtp.port"),
			SMTPTLS:      viper.GetBool("smtp.tls"),
			SMTPEnabled:  viper.GetBool("smtp.enabled"),
			IMAPHost:     getEnv("IMAP_HOST", ""),
			IMAPPort:     viper.GetInt("imap.port"),
			IMAPEnabled:  viper.GetBool("imap.enabled"),
			PollInterval: viper.GetInt("email.poll_interval"),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize        int
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}
