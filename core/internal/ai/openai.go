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

	"github.com/google/uuid"
)

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
}

// OpenAIProvider OpenAI 提供商实现
type OpenAIProvider struct {
	config  OpenAIConfig
	client  *http.Client
	latency time.Duration
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	// 优先使用环境变量中的 API Key
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && cfg.APIKey == "" {
		cfg.APIKey = apiKey
	}

	return &OpenAIProvider{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Chat 实现真实的 OpenAI API 调用
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	start := time.Now()
	p.latency = start.Sub(start) // 重置延迟计算

	// 构建请求
	model := p.config.Model
	if req.Model != "" {
		model = req.Model
	}

	openaiReq := OpenAIChatRequest{
		Model: model,
		Messages: req.Messages,
	}

	// 设置可选参数
	if req.Temperature > 0 {
		openaiReq.Temperature = req.Temperature
	}
	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = req.MaxTokens
	}
	if req.Stream {
		openaiReq.Stream = true
	}

	// 序列化请求
	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))

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
		var errResp OpenAIErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("OpenAI API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var openaiResp OpenAIChatResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 更新延迟
	p.latency = time.Since(start)

	// 构建响应
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := ""
	finishReason := ""
	if openaiResp.Choices[0].Message.Content != "" {
		content = openaiResp.Choices[0].Message.Content
		finishReason = openaiResp.Choices[0].FinishReason
	} else if openaiResp.Choices[0].Delta.Content != "" {
		content = openaiResp.Choices[0].Delta.Content
		finishReason = openaiResp.Choices[0].FinishReason
	}

	return &ChatResponse{
		ID:            openaiResp.ID,
		Model:         openaiResp.Model,
		Content:       content,
		FinishReason:  finishReason,
		Usage: Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// Embeddings 生成嵌入向量
func (p *OpenAIProvider) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	model := "text-embedding-ada-002"
	if req.Model != "" {
		model = req.Model
	}

	embedReq := OpenAIEmbeddingsRequest{
		Model: model,
		Input: req.Input,
	}

	jsonData, err := json.Marshal(embedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI embeddings API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var embedResp OpenAIEmbeddingsResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &EmbeddingsResponse{
		Embedding: embedResp.Data[0].Embedding,
	}, nil
}

// Latency 返回延迟
func (p *OpenAIProvider) Latency() time.Duration {
	return p.latency
}

// Cost 返回每 1K token 的成本
func (p *OpenAIProvider) Cost() float64 {
	switch p.config.Model {
	case "gpt-4":
		return 0.03 // $0.03/1K input, $0.06/1K output
	case "gpt-4-turbo":
		return 0.01
	case "gpt-3.5-turbo":
		return 0.0005
	default:
		return 0.0005
	}
}

// Type 返回提供商类型
func (p *OpenAIProvider) Type() ProviderType {
	return ProviderOpenAI
}

// GetAPIKey 获取 API Key
func (p *OpenAIProvider) GetAPIKey() string {
	return p.config.APIKey
}

// SetAPIKey 设置 API Key
func (p *OpenAIProvider) SetAPIKey(apiKey string) {
	p.config.APIKey = apiKey
}

// UpdateConfig 更新配置
func (p *OpenAIProvider) UpdateConfig(cfg OpenAIConfig) {
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

// ==================== OpenAI API 类型定义 ====================

// OpenAIChatRequest OpenAI 聊天请求
type OpenAIChatRequest struct {
	Model       string      `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32     `json:"temperature,omitempty"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
}

// OpenAIChatResponse OpenAI 聊天响应
type OpenAIChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message,omitempty"`
	Delta        Delta       `json:"delta,omitempty"`
	FinishReason string      `json:"finish_reason"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Delta 增量消息
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// OpenAIEmbeddingsRequest OpenAI 嵌入请求
type OpenAIEmbeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// OpenAIEmbeddingsResponse OpenAI 嵌入响应
type OpenAIEmbeddingsResponse struct {
	Object string    `json:"object"`
	Data   []Data    `json:"data"`
	Model  string    `json:"model"`
	Usage  EmbedUsage `json:"usage"`
}

// Data 数据
type Data struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbedUsage 使用量
type EmbedUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// OpenAIErrorResponse OpenAI 错误响应
type OpenAIErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
}

// InitOpenAIProviders 初始化 OpenAI 提供商
func InitOpenAIProviders(providers []string, apiKeys map[string]string) map[ProviderType]Provider {
	result := make(map[ProviderType]Provider)

	for _, p := range providers {
		switch p {
		case "openai":
			cfg := OpenAIConfig{
				APIKey:  apiKeys["openai"],
				Model:   "gpt-4",
				Timeout: 60 * time.Second,
			}
			if key := os.Getenv("OPENAI_API_KEY"); key != "" {
				cfg.APIKey = key
			}
			result[ProviderOpenAI] = NewOpenAIProvider(cfg)
		}
	}

	return result
}
