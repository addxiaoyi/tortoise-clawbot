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

	"tortoise/config"
)

// AIService AI服务
type AIService struct {
	providers map[string]*Provider
}

// Provider AI提供商
type Provider struct {
	Name    string
	BaseURL string
	APIKey  string
	Models  []string
	client  *http.Client
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role         string                 `json:"role"`
	Content      string                 `json:"content"`
	Name         string                 `json:"name,omitempty"`
	FunctionCall *FunctionCall          `json:"function_call,omitempty"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Function 函数定义
type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

// StreamResponse 流式响应
type StreamResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []StreamChoice      `json:"choices"`
}

// StreamChoice 流式选择
type StreamChoice struct {
	Index        int               `json:"index"`
	Delta        StreamDelta       `json:"delta"`
	FinishReason string           `json:"finish_reason"`
}

// StreamDelta 流式增量
type StreamDelta struct {
	Role         string        `json:"role,omitempty"`
	Content      string        `json:"content,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// NewAIService 创建AI服务
func NewAIService(cfg config.AIConfig) *AIService {
	providers := make(map[string]*Provider)
	
	for _, p := range cfg.Providers {
		providers[p.Name] = &Provider{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			APIKey:  p.APIKey,
			Models:  p.Models,
			client: &http.Client{
				Timeout: 120 * time.Second,
			},
		}
	}
	
	return &AIService{providers: providers}
}

// Chat 聊天
func (s *AIService) Chat(ctx context.Context, providerName string, req ChatRequest) (*ChatResponse, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}
	
	return s.chatOpenAI(ctx, provider, req)
}

// chatOpenAI OpenAI格式聊天
func (s *AIService) chatOpenAI(ctx context.Context, p *Provider, req ChatRequest) (*ChatResponse, error) {
	// 转换消息格式
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		m := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.Name != "" {
			m["name"] = msg.Name
		}
		if msg.FunctionCall != nil {
			m["function_call"] = msg.FunctionCall
		}
		messages[i] = m
	}
	
	// 构建请求体
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}
	
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		body["max_tokens"] = *req.MaxTokens
	}
	if req.TopP != nil {
		body["top_p"] = *req.TopP
	}
	if req.Stream {
		body["stream"] = true
	}
	if len(req.Stop) > 0 {
		body["stop"] = req.Stop
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	
	// 发送请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}
	
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	
	return &chatResp, nil
}

// StreamChat 流式聊天
func (s *AIService) StreamChat(ctx context.Context, providerName string, req ChatRequest) (<-chan *StreamResponse, <-chan error) {
	respChan := make(chan *StreamResponse)
	errChan := make(chan error, 1)
	
	go func() {
		defer close(respChan)
		defer close(errChan)
		
		provider, ok := s.providers[providerName]
		if !ok {
			errChan <- fmt.Errorf("provider not found: %s", providerName)
			return
		}
		
		// 转换消息格式
		messages := make([]map[string]interface{}, len(req.Messages))
		for i, msg := range req.Messages {
			m := map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}
			if msg.Name != "" {
				m["name"] = msg.Name
			}
			messages[i] = m
		}
		
		// 构建请求体
		body := map[string]interface{}{
			"model":    req.Model,
			"messages": messages,
			"stream":   true,
		}
		
		if req.Temperature != nil {
			body["temperature"] = *req.Temperature
		}
		if req.MaxTokens != nil {
			body["max_tokens"] = *req.MaxTokens
		}
		
		jsonBody, err := json.Marshal(body)
		if err != nil {
			errChan <- err
			return
		}
		
		// 发送请求
		httpReq, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			errChan <- err
			return
		}
		
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
		
		resp, err := provider.client.Do(httpReq)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
			return
		}
		
		// 读取流
		reader := resp.Body
		buf := make([]byte, 0, 4096)
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				chunk := make([]byte, 1024)
				n, err := reader.Read(chunk)
				if n > 0 {
					buf = append(buf, chunk[:n]...)
					
					// 处理已接收的数据
					for {
						lines := strings.Split(string(buf), "\n")
						if len(lines) < 2 {
							break
						}
						
						// 找到完整的行
						var completeLines string
						for i := 0; i < len(lines)-1; i++ {
							line := strings.TrimSpace(lines[i])
							if line == "" || line == "data: [DONE]" {
								continue
							}
							
							if strings.HasPrefix(line, "data: ") {
								data := line[6:]
								var streamResp StreamResponse
								if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
									select {
									case respChan <- &streamResp:
									case <-ctx.Done():
										return
									}
								}
							}
						}
						
						// 保留未处理完的数据
						buf = []byte(lines[len(lines)-1])
						
						if len(lines) <= 1 {
							break
						}
					}
				}
				
				if err == io.EOF {
					return
				}
				if err != nil {
					errChan <- err
					return
				}
			}
		}
	}()
	
	return respChan, errChan
}

// ListProviders 列出提供商
func (s *AIService) ListProviders() []string {
	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}
	return providers
}

// ListModels 列出模型
func (s *AIService) ListModels(providerName string) []string {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil
	}
	return provider.Models
}
