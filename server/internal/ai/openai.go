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

// ============ OpenAI Provider ============

// OpenAIProvider OpenAI 提供商
type OpenAIProvider struct {
	config *ProviderConfig
	client *http.Client
}

// OpenAI 请求结构
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// OpenAI 响应结构
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      openAIMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens     int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// 流式响应结构
type openAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int             `json:"index"`
		Delta        openAIMessage   `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *struct {
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// NewOpenAIProvider 创建 OpenAI Provider
func NewOpenAIProvider(config *ProviderConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Name 返回提供商名称
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// ID 返回提供商 ID
func (p *OpenAIProvider) ID() string {
	return "openai"
}

// IsEnabled 是否启用
func (p *OpenAIProvider) IsEnabled() bool {
	return p.config.Enabled
}

// Chat 发送聊天请求
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "gpt-4"
		}
	}

	// 构建请求
	openaiReq := openAIRequest{
		Model:       model,
		Messages:    make([]openAIMessage, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	// 转换消息格式
	for i, msg := range req.Messages {
		openaiReq.Messages[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if msg.Name != "" {
			openaiReq.Messages[i].Name = msg.Name
		}
	}

	// 序列化请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建请求
	url := p.config.BaseURL + "/chat/completions"
	if !strings.HasSuffix(p.config.BaseURL, "/v1") {
		url = "https://api.openai.com/v1/chat/completions"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	httpReq.Header.Set("OpenAI-Organization", "")

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

	// 解析响应
	var openaiResp openAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误
	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s - %s", openaiResp.Error.Type, openaiResp.Error.Message)
	}

	// 提取结果
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatResponse{
		Model:        openaiResp.Model,
		Content:      openaiResp.Choices[0].Message.Content,
		FinishReason: openaiResp.Choices[0].FinishReason,
		Tokens:       openaiResp.Usage.TotalTokens,
	}, nil
}

// ChatStream 发送流式聊天请求
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest, callback func(*StreamingChunk)) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "gpt-4"
		}
	}

	// 构建请求
	openaiReq := openAIRequest{
		Model:       model,
		Messages:    make([]openAIMessage, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	for i, msg := range req.Messages {
		openaiReq.Messages[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/chat/completions"
	if !strings.HasSuffix(p.config.BaseURL, "/v1") {
		url = "https://api.openai.com/v1/chat/completions"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	decoder := json.NewDecoder(resp.Body)
	totalTokens := 0

	for decoder.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var chunk openAIStreamResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream: %w", err)
		}

		content := ""
		done := false
		if len(chunk.Choices) > 0 {
			content = chunk.Choices[0].Delta.Content
			if chunk.Choices[0].FinishReason != "" {
				done = true
			}
		}

		if chunk.Usage != nil {
			totalTokens = chunk.Usage.TotalTokens
		}

		callback(&StreamingChunk{
			Content:    content,
			Done:       done,
			TotalTokens: totalTokens,
		})

		if done {
			break
		}
	}

	return nil
}

// GetModels 获取可用模型列表
func (p *OpenAIProvider) GetModels() []ModelInfo {
	return []ModelInfo{
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai"},
		{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Provider: "openai"},
		{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
		{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai"},
	}
}
