package ai

import (
	"context"
	"fmt"
)

// ============ 消息结构 ============

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 消息内容
	Name    string `json:"name,omitempty"` // 可选的名字
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"` // 0-2 之间，越高越有创意
	MaxTokens   int       `json:"max_tokens"` // 最大 token 数
	Stream      bool      `json:"stream"`     // 是否流式响应
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Model      string   `json:"model"`
	Content    string   `json:"content"`
	FinishReason string `json:"finish_reason"` // stop, length, content_filter
	Tokens     int      `json:"tokens"`       // 使用的 token 数
	Error      string   `json:"error,omitempty"`
}

// StreamingChunk 流式响应块
type StreamingChunk struct {
	Content    string `json:"content"`
	Done       bool   `json:"done"`
	TotalTokens int   `json:"total_tokens,omitempty"`
}

// ============ Provider 接口 ============

// Provider AI 提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() string
	// ID 返回提供商 ID
	ID() string
	// IsEnabled 是否启用
	IsEnabled() bool
	// Chat 发送聊天请求
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	// ChatStream 发送流式聊天请求
	ChatStream(ctx context.Context, req *ChatRequest, callback func(*StreamingChunk)) error
	// GetModels 获取可用模型列表
	GetModels() []ModelInfo
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// ============ Provider 工厂 ============

// ProviderConfig 提供商配置
type ProviderConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

// NewProvider 创建 AI Provider
func NewProvider(config *ProviderConfig) (Provider, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", config.ID)
	}

	switch config.ID {
	case "openai":
		return NewOpenAIProvider(config)
	case "anthropic":
		return NewAnthropicProvider(config)
	case "ollama":
		return NewOllamaProvider(config)
	default:
		return nil, fmt.Errorf("unknown provider: %s", config.ID)
	}
}

// NewProviders 创建多个 Provider
func NewProviders(configs []ProviderConfig) ([]Provider, error) {
	providers := make([]Provider, 0, len(configs))
	
	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}
		
		p, err := NewProvider(&cfg)
		if err != nil {
			// 记录错误但继续
			fmt.Printf("Warning: failed to create provider %s: %v\n", cfg.ID, err)
			continue
		}
		providers = append(providers, p)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no AI providers available")
	}

	return providers, nil
}
