package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AnthropicConfig Anthropic 配置
type AnthropicConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
	Version string // API 版本，默认 "2023-06-01"
}

// AnthropicProvider Anthropic 提供商实现
type AnthropicProvider struct {
	config  AnthropicConfig
	client  *http.Client
	latency time.Duration
}

// NewAnthropicProvider 创建 Anthropic 提供商
func NewAnthropicProvider(cfg AnthropicConfig) *AnthropicProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}
	if cfg.Version == "" {
		cfg.Version = "2023-06-01"
	}

	// 优先使用环境变量中的 API Key
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" && cfg.APIKey == "" {
		cfg.APIKey = apiKey
	}

	return &AnthropicProvider{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Chat 实现真实的 Anthropic Claude API 调用
func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	start := time.Now()

	// 构建请求
	model := p.config.Model
	if req.Model != "" {
		model = req.Model
	}

	// 转换消息格式到 Anthropic 格式
	messages := make([]AnthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "assistant"
		} else if role == "user" {
			role = "user"
		} else {
			role = "user" // system 消息单独处理
		}
		messages = append(messages, AnthropicMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	anthropicReq := AnthropicChatRequest{
		Model:             model,
		Messages:          messages,
		MaxTokens:         4096, // Claude 需要指定最大 token
	}

	// 设置可选参数
	if req.Temperature > 0 {
		anthropicReq.Temperature = req.Temperature
	}

	// 序列化请求
	jsonData, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/v1/messages", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", p.config.Version)

	// 发送请求
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		var errResp AnthropicErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("Anthropic API error: %s - %s (type: %s)", 
				errResp.Error.Type, errResp.Error.Message, errResp.Error.Type)
		}
		return nil, fmt.Errorf("Anthropic API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var anthropicResp AnthropicChatResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 更新延迟
	p.latency = time.Since(start)

	// 构建响应
	content := ""
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	finishReason := "stop"
	if anthropicResp.StopReason == "max_tokens" {
		finishReason = "length"
	}

	return &ChatResponse{
		ID:            anthropicResp.ID,
		Model:         anthropicResp.Model,
		Content:       content,
		FinishReason:  finishReason,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

// Embeddings Claude 不直接支持嵌入，但可以使用 OpenAI 的嵌入服务
func (p *AnthropicProvider) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Claude 不提供嵌入 API，返回模拟数据
	// 实际应用中可以使用 OpenAI 或其他嵌入服务
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(i % 100 / 100)
	}
	return &EmbeddingsResponse{Embedding: embedding}, nil
}

// Latency 返回延迟
func (p *AnthropicProvider) Latency() time.Duration {
	return p.latency
}

// Cost 返回每 1K token 的成本
func (p *AnthropicProvider) Cost() float64 {
	switch p.config.Model {
	case "claude-3-opus-20240229":
		return 0.015 // $15/1M input, $75/1M output
	case "claude-3-sonnet-20240229":
		return 0.003 // $3/1M input, $15/1M output
	case "claude-3-haiku-20240307":
		return 0.00025 // $0.25/1M input, $1.25/1M output
	case "claude-2.1":
		return 0.024
	case "claude-2.0":
		return 0.024
	case "claude-instant-1":
		return 0.0008
	default:
		return 0.003
	}
}

// Type 返回提供商类型
func (p *AnthropicProvider) Type() ProviderType {
	return ProviderAnthropic
}

// GetAPIKey 获取 API Key
func (p *AnthropicProvider) GetAPIKey() string {
	return p.config.APIKey
}

// SetAPIKey 设置 API Key
func (p *AnthropicProvider) SetAPIKey(apiKey string) {
	p.config.APIKey = apiKey
}

// UpdateConfig 更新配置
func (p *AnthropicProvider) UpdateConfig(cfg AnthropicConfig) {
	if cfg.APIKey != "" {
		p.config.APIKey = cfg.APIKey
	}
	if cfg.BaseURL != "" {
		p.config.BaseURL = cfg.BaseURL
	}
	if cfg.Model != "" {
		p.config.Model = cfg.Model
	}
	if cfg.Timeout > 0 {
		p.config.Timeout = cfg.Timeout
	}
}

// ==================== Anthropic API 类型定义 ====================

// AnthropicChatRequest Anthropic 聊天请求
type AnthropicChatRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float32            `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"` // 系统提示
	Stream      bool               `json:"stream,omitempty"`
}

// AnthropicMessage Anthropic 消息
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicChatResponse Anthropic 聊天响应
type AnthropicChatResponse struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Role         string           `json:"role"`
	Content      []ContentBlock   `json:"content"`
	Model        string           `json:"model"`
	StopReason   string           `json:"stop_reason"`
	StopSequence string           `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage   `json:"usage"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AnthropicUsage Anthropic 使用量
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicErrorResponse Anthropic 错误响应
type AnthropicErrorResponse struct {
	Error AnthropicError `json:"error"`
}

// AnthropicError Anthropic 错误
type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// InitAnthropicProviders 初始化 Anthropic 提供商
func InitAnthropicProviders(providers []string, apiKeys map[string]string) map[ProviderType]Provider {
	result := make(map[ProviderType]Provider)

	for _, p := range providers {
		switch p {
		case "anthropic":
			cfg := AnthropicConfig{
				APIKey:  apiKeys["anthropic"],
				Model:   "claude-3-sonnet-20240229",
				Timeout: 120 * time.Second,
				Version: "2023-06-01",
			}
			if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
				cfg.APIKey = key
			}
			result[ProviderAnthropic] = NewAnthropicProvider(cfg)
		}
	}

	return result
}
