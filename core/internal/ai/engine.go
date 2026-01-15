package ai

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ProviderType AI提供商类型
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOllama   ProviderType = "ollama"
	ProviderLocal    ProviderType = "local"
)

// Config AI引擎配置
type Config struct {
	Providers []string
	Routing   string // "latency" | "load" | "cost"
	Timeout  time.Duration
	RetryAttempts int
}

// Engine AI引擎
type Engine struct {
	config Config

	// 提供商
	providers map[ProviderType]Provider

	// 路由
	router *Router

	// 会话
	sessions map[string]*AISession

	// 统计
	stats Stats

	mu sync.RWMutex
}

// Stats AI统计
type Stats struct {
	RequestsTotal     atomic.Int64
	RequestsSuccess   atomic.Int64
	RequestsFailed   atomic.Int64
	AvgLatencyMs     atomic.Int64
	TokensUsed       atomic.Int64
	CostUsd          atomic.Int64
}

// Provider AI提供商接口
type Provider interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error)
	Latency() time.Duration
	Cost() float64
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	Temperature float32
	MaxTokens   int
	Stream      bool
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string
	Content string
	Name    string
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID        string
	Model    string
	Content  string
	FinishReason string
	Usage    Usage
}

// Usage 使用量
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens     int
}

// EmbeddingsRequest 向量请求
type EmbeddingsRequest struct {
	Model string
	Input string
}

// EmbeddingsResponse 向量响应
type EmbeddingsResponse struct {
	Embedding []float32
}

// AISession AI会话
type AISession struct {
	ID       string
	UserID   string
	Model    string
	Context  []ChatMessage
	CreatedAt time.Time
}

// Router 路由器
type Router struct {
	providers map[ProviderType]*ProviderInfo
	strategy  string
	mu        sync.RWMutex
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Type      ProviderType
	Latency  atomic.Int64 // 毫秒
	Load     atomic.Int64 // 负载
	Cost     float64      // 每1K token成本
	Healthy  bool
	LastUsed time.Time
}

// NewEngine 创建AI引擎
func NewEngine(cfg Config) *Engine {
	e := &Engine{
		config:    cfg,
		providers: make(map[ProviderType]Provider),
		sessions:  make(map[string]*AISession),
	}

	// 初始化路由器
	e.router = &Router{
		providers: make(map[ProviderType]*ProviderInfo),
		strategy:  cfg.Routing,
	}

	// 初始化提供商
	for _, p := range cfg.Providers {
		switch p {
		case "openai":
			e.providers[ProviderOpenAI] = &OpenAIProvider{}
			e.router.providers[ProviderOpenAI] = &ProviderInfo{Type: ProviderOpenAI, Healthy: true}
		case "anthropic":
			e.providers[ProviderAnthropic] = &AnthropicProvider{}
			e.router.providers[ProviderAnthropic] = &ProviderInfo{Type: ProviderAnthropic, Healthy: true}
		case "ollama":
			e.providers[ProviderOllama] = &OllamaProvider{}
			e.router.providers[ProviderOllama] = &ProviderInfo{Type: ProviderOllama, Healthy: true}
		}
	}

	return e
}

// Chat 聊天
func (e *Engine) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	e.stats.RequestsTotal.Add(1)

	// 选择提供商
	provider := e.selectProvider()

	// 调用提供商
	start := time.Now()
	resp, err := provider.Chat(ctx, req)
	latency := time.Since(start)

	// 更新统计
	if err != nil {
		e.stats.RequestsFailed.Add(1)
		return nil, err
	}

	e.stats.RequestsSuccess.Add(1)
	e.stats.AvgLatencyMs.Store(latency.Milliseconds())
	e.stats.TokensUsed.Add(int64(resp.Usage.TotalTokens))
	e.stats.CostUsd.Store(int64(resp.Usage.TotalTokens) * int64(provider.Cost() * 1000))

	// 更新提供商延迟
	e.router.mu.Lock()
	if info, ok := e.router.providers[provider.(Provider).Type()]; ok {
		info.Latency.Store(latency.Milliseconds())
		info.LastUsed = time.Now()
	}
	e.router.mu.Unlock()

	return resp, nil
}

// selectProvider 选择提供商
func (e *Engine) selectProvider() Provider {
	switch e.config.Routing {
	case "latency":
		return e.selectByLatency()
	case "load":
		return e.selectByLoad()
	case "cost":
		return e.selectByCost()
	default:
		return e.selectByLatency()
	}
}

func (e *Engine) selectByLatency() Provider {
	e.router.mu.RLock()
	defer e.router.mu.RUnlock()

	var best Provider
	var minLatency int64 = 1 << 62

	for _, info := range e.router.providers {
		if info.Healthy && info.Latency.Load() < minLatency {
			minLatency = info.Latency.Load()
			best = e.providers[info.Type]
		}
	}

	return best
}

func (e *Engine) selectByLoad() Provider {
	e.router.mu.RLock()
	defer e.router.mu.RUnlock()

	var best Provider
	var minLoad int64 = 1 << 62

	for _, info := range e.router.providers {
		if info.Healthy && info.Load.Load() < minLoad {
			minLoad = info.Load.Load()
			best = e.providers[info.Type]
		}
	}

	return best
}

func (e *Engine) selectByCost() Provider {
	e.router.mu.RLock()
	defer e.router.mu.RUnlock()

	var best Provider
	var minCost float64 = 1 << 62

	for _, info := range e.router.providers {
		if info.Healthy && info.Cost < minCost {
			minCost = info.Cost
			best = e.providers[info.Type]
		}
	}

	return best
}

// Embeddings 生成向量
func (e *Engine) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	provider := e.providers[ProviderOpenAI]
	return provider.Embeddings(ctx, req)
}

// CreateSession 创建AI会话
func (e *Engine) CreateSession(userID, model string) string {
	session := &AISession{
		ID:        uuid.New().String(),
		UserID:   userID,
		Model:    model,
		Context:  make([]ChatMessage, 0),
		CreatedAt: time.Now(),
	}

	e.mu.Lock()
	e.sessions[session.ID] = session
	e.mu.Unlock()

	return session.ID
}

// GetSession 获取AI会话
func (e *Engine) GetSession(sessionID string) (*AISession, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, ok := e.sessions[sessionID]
	return session, ok
}

// AppendContext 追加上下文
func (e *Engine) AppendContext(sessionID string, msg ChatMessage) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if session, ok := e.sessions[sessionID]; ok {
		session.Context = append(session.Context, msg)
		return true
	}
	return false
}

// Stats 获取统计
func (e *Engine) Stats() Stats {
	return Stats{
		RequestsTotal:   e.stats.RequestsTotal,
		RequestsSuccess: e.stats.RequestsSuccess,
		RequestsFailed: e.stats.RequestsFailed,
		AvgLatencyMs:   e.stats.AvgLatencyMs,
		TokensUsed:     e.stats.TokensUsed,
		CostUsd:       e.stats.CostUsd,
	}
}

// ============ 提供商实现 ============

type OpenAIProvider struct{}

func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 模拟响应
	return &ChatResponse{
		ID:            uuid.New().String(),
		Model:         "gpt-4",
		Content:       "This is a simulated response from OpenAI.",
		FinishReason: "stop",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:     30,
		},
	}, nil
}

func (p *OpenAIProvider) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// 模拟向量
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(i % 100 / 100)
	}
	return &EmbeddingsResponse{Embedding: embedding}, nil
}

func (p *OpenAIProvider) Latency() time.Duration { return 100 * time.Millisecond }
func (p *OpenAIProvider) Cost() float64         { return 0.00003 } // $0.03/1K tokens
func (p *OpenAIProvider) Type() ProviderType    { return ProviderOpenAI }

type AnthropicProvider struct{}

func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		ID:            uuid.New().String(),
		Model:         "claude-3",
		Content:       "This is a simulated response from Anthropic.",
		FinishReason: "stop",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 25,
			TotalTokens:     35,
		},
	}, nil
}

func (p *AnthropicProvider) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(i % 100 / 100)
	}
	return &EmbeddingsResponse{Embedding: embedding}, nil
}

func (p *AnthropicProvider) Latency() time.Duration { return 150 * time.Millisecond }
func (p *AnthropicProvider) Cost() float64         { return 0.00004 }
func (p *AnthropicProvider) Type() ProviderType    { return ProviderAnthropic }

type OllamaProvider struct{}

func (p *OllamaProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		ID:            uuid.New().String(),
		Model:         "llama2",
		Content:       "This is a simulated response from Ollama (local).",
		FinishReason: "stop",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 30,
			TotalTokens:     40,
		},
	}, nil
}

func (p *OllamaProvider) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	embedding := make([]float32, 4096)
	for i := range embedding {
		embedding[i] = float32(i % 100 / 100)
	}
	return &EmbeddingsResponse{Embedding: embedding}, nil
}

func (p *OllamaProvider) Latency() time.Duration { return 50 * time.Millisecond }
func (p *OllamaProvider) Cost() float64         { return 0.0 } // 免费
func (p *OllamaProvider) Type() ProviderType    { return ProviderOllama }
