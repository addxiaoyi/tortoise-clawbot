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

// ============ Ollama Provider (本地) ============

// OllamaProvider Ollama 本地模型提供商
type OllamaProvider struct {
	config *ProviderConfig
	client *http.Client
}

// Ollama 请求结构
type ollamaRequest struct {
	Model       string           `json:"model"`
	Messages    []ollamaMessage `json:"messages,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Options     *ollamaOptions  `json:"options,omitempty"`
	Stream      *bool           `json:"stream,omitempty"`
	KeepAlive   string          `json:"keep_alive,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	NumPredict int `json:"num_predict,omitempty"`
}

// Ollama 响应结构
type ollamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
	Error              string `json:"error,omitempty"`
}

// Ollama 流式响应
type ollamaStreamResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	Error              string `json:"error,omitempty"`
}

// NewOllamaProvider 创建 Ollama Provider
func NewOllamaProvider(config *ProviderConfig) (*OllamaProvider, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OllamaProvider{
		config: config,
		client: &http.Client{
			Timeout: 300 * time.Second, // Ollama 可能需要更长时间
		},
	}, nil
}

// Name 返回提供商名称
func (p *OllamaProvider) Name() string {
	return "Ollama (本地)"
}

// ID 返回提供商 ID
func (p *OllamaProvider) ID() string {
	return "ollama"
}

// IsEnabled 是否启用
func (p *OllamaProvider) IsEnabled() bool {
	return p.config.Enabled
}

// Chat 发送聊天请求
func (p *OllamaProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "llama2"
		}
	}

	// 构建 Ollama 消息格式
	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		role := msg.Role
		if role == "system" {
			role = "system"
		} else if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		messages[i] = ollamaMessage{
			Role:    role,
			Content: msg.Content,
		}
	}

	// 设置 max_tokens
	var options *ollamaOptions
	if req.MaxTokens > 0 {
		options = &ollamaOptions{NumPredict: req.MaxTokens}
	}

	ollamaReq := ollamaRequest{
		Model:       model,
		Messages:    messages,
		Temperature: req.Temperature,
		Options:     options,
		KeepAlive:   "5m",
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/api/chat"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	return &ChatResponse{
		Model:        ollamaResp.Model,
		Content:      ollamaResp.Response,
		FinishReason: "stop",
		Tokens:       ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
	}, nil
}

// ChatStream 发送流式聊天请求
func (p *OllamaProvider) ChatStream(ctx context.Context, req *ChatRequest, callback func(*StreamingChunk)) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "llama2"
		}
	}

	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		role := msg.Role
		if role == "system" {
			role = "system"
		} else if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		messages[i] = ollamaMessage{
			Role:    role,
			Content: msg.Content,
		}
	}

	var options *ollamaOptions
	if req.MaxTokens > 0 {
		options = &ollamaOptions{NumPredict: req.MaxTokens}
	}

	stream := true
	ollamaReq := ollamaRequest{
		Model:       model,
		Messages:    messages,
		Temperature: req.Temperature,
		Options:     options,
		Stream:      &stream,
		KeepAlive:   "5m",
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/api/chat"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, string(body))
	}

	// 处理流式响应 (NDJSON 格式)
	decoder := json.NewDecoder(resp.Body)
	totalTokens := 0

	for decoder.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var chunk ollamaStreamResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode stream: %w", err)
		}

		if chunk.Error != "" {
			return fmt.Errorf("Ollama stream error: %s", chunk.Error)
		}

		totalTokens = chunk.PromptEvalCount + chunk.EvalCount

		callback(&StreamingChunk{
			Content:    chunk.Response,
			Done:       chunk.Done,
			TotalTokens: totalTokens,
		})

		if chunk.Done {
			break
		}
	}

	return nil
}

// GetModels 获取可用模型列表
// 注意：Ollama 需要调用 API 获取实际模型列表，这里返回常见模型
func (p *OllamaProvider) GetModels() []ModelInfo {
	return []ModelInfo{
		{ID: "llama2", Name: "Llama 2", Provider: "ollama"},
		{ID: "llama3", Name: "Llama 3", Provider: "ollama"},
		{ID: "llama3.1", Name: "Llama 3.1", Provider: "ollama"},
		{ID: "mistral", Name: "Mistral", Provider: "ollama"},
		{ID: "codellama", Name: "Code Llama", Provider: "ollama"},
		{ID: "qwen2", Name: "Qwen 2", Provider: "ollama"},
		{ID: "phi3", Name: "Phi-3", Provider: "ollama"},
		{ID: "gemma", Name: "Gemma", Provider: "ollama"},
	}
}

// ListLocalModels 调用 Ollama API 获取本地模型列表
func (p *OllamaProvider) ListLocalModels() ([]string, error) {
	url := p.config.BaseURL + "/api/tags"

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			Model      string `json:"model"`
			Size       int64  `json:"size"`
			ModifiedAt string `json:"modified_at"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}
