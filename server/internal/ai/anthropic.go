package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ============ Anthropic Provider ============

// AnthropicProvider Anthropic Claude 提供商
type AnthropicProvider struct {
	config *ProviderConfig
	client *http.Client
}

// Anthropic 请求结构
type anthropicRequest struct {
	Model       string   `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string   `json:"system,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens  int      `json:"max_tokens"`
	Stream      bool    `json:"stream,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Anthropic 响应结构
type anthropicResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Role     string `json:"role"`
	Content  []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"content"`
	Model       string `json:"model"`
	StopReason string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Anthropic 流式响应
type anthropicStreamResponse struct {
	Type string `json:"type"`
	Index int `json:"index,omitempty"`
	ContentBlock *struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type         string `json:"type"`
		Text         string `json:"text"`
		StopReason   string `json:"stop_reason,omitempty"`
	} `json:"delta,omitempty"`
	MessageStop *struct {
		Index int `json:"index"`
	} `json:"message_stop,omitempty"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewAnthropicProvider 创建 Anthropic Provider
func NewAnthropicProvider(config *ProviderConfig) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &AnthropicProvider{
		config: config,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Name 返回提供商名称
func (p *AnthropicProvider) Name() string {
	return "Anthropic"
}

// ID 返回提供商 ID
func (p *AnthropicProvider) ID() string {
	return "anthropic"
}

// IsEnabled 是否启用
func (p *AnthropicProvider) IsEnabled() bool {
	return p.config.Enabled
}

// Chat 发送聊天请求
func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
	}

	// 分离 system message
	var systemPrompt string
	messages := make([]anthropicMessage, 0, len(req.Messages))
	
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			// Anthropic 使用 user/assistant
			role := msg.Role
			if role == "assistant" {
				role = "assistant"
			} else {
				role = "user"
			}
			messages = append(messages, anthropicMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	// Anthropic 要求 max_tokens
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicRequest{
		Model:       model,
		Messages:    messages,
		System:      systemPrompt,
		Temperature: req.Temperature,
		MaxTokens:   maxTokens,
	}

	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/v1/messages"
	if !strings.Contains(p.config.BaseURL, "/v1") {
		url = "https://api.anthropic.com/v1/messages"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置 Anthropic 特定的请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("Anthropic error: %s - %s", anthropicResp.Error.Type, anthropicResp.Error.Message)
	}

	// 提取文本内容
	content := ""
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &ChatResponse{
		Model:        anthropicResp.Model,
		Content:      content,
		FinishReason: anthropicResp.StopReason,
		Tokens:       anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
	}, nil
}

// ChatStream 发送流式聊天请求
func (p *AnthropicProvider) ChatStream(ctx context.Context, req *ChatRequest, callback func(*StreamingChunk)) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
	}

	var systemPrompt string
	messages := make([]anthropicMessage, 0, len(req.Messages))
	
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			role := msg.Role
			if role != "assistant" {
				role = "user"
			}
			messages = append(messages, anthropicMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicRequest{
		Model:       model,
		Messages:    messages,
		System:      systemPrompt,
		Temperature: req.Temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}

	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/v1/messages"
	if !strings.Contains(p.config.BaseURL, "/v1") {
		url = "https://api.anthropic.com/v1/messages"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, string(body))
	}

	// 处理 SSE 流
	reader := resp.Body
	buf := make([]byte, 0, 4096)
	lineBuf := make([]byte, 0, 4096)
	totalTokens := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tmp := make([]byte, 1024)
		n, err := reader.Read(tmp)
		if err != nil {
			break
		}
		buf = append(buf, tmp[:n]...)

		for {
			// 查找换行
			idx := bytes.IndexByte(buf, '\n')
			if idx < 0 {
				break
			}

			line := buf[:idx]
			buf = buf[idx+1:]

			// 跳过空行和非 data 行
			if len(line) == 0 || !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := line[6:] // 去掉 "data: " 前缀
			if bytes.Equal(data, []byte("[DONE]")) {
				callback(&StreamingChunk{Done: true, TotalTokens: totalTokens})
				return nil
			}

			var chunk anthropicStreamResponse
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			if chunk.Error != nil {
				return fmt.Errorf("Anthropic stream error: %s", chunk.Error.Message)
			}

			content := ""
			done := false

			if chunk.Delta != nil {
				content = chunk.Delta.Text
				if chunk.Delta.StopReason != "" {
					done = true
				}
			}

			if chunk.Usage != nil {
				totalTokens = chunk.Usage.InputTokens + chunk.Usage.OutputTokens
			}

			callback(&StreamingChunk{
				Content:    content,
				Done:       done,
				TotalTokens: totalTokens,
			})

			if done {
				return nil
			}
		}

		_ = lineBuf
	}

	callback(&StreamingChunk{Done: true, TotalTokens: totalTokens})
	return nil
}

// GetModels 获取可用模型列表
func (p *AnthropicProvider) GetModels() []ModelInfo {
	return []ModelInfo{
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "anthropic"},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "anthropic"},
		{ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet", Provider: "anthropic"},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Provider: "anthropic"},
	}
}
