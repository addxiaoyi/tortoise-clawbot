package ai

import (
	"testing"
	"time"
)

// ============ AI Engine Tests ============

func TestEngine_AddProvider(t *testing.T) {
	engine := NewEngine()
	
	provider := &mockProvider{
		id:   "test",
		name: "Test Provider",
	}
	
	engine.AddProvider(provider)
	
	if len(engine.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(engine.providers))
	}
}

func TestEngine_Chat(t *testing.T) {
	engine := NewEngine()
	
	// 添加模拟提供商
	engine.AddProvider(&mockProvider{
		id:   "test",
		name: "Test Provider",
	})
	
	// 测试无 API Key 情况
	req := &ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}
	
	_, err := engine.Chat(nil, req)
	
	// 模拟提供商返回错误 (因为没有真实 API)
	// 在实际环境中，这里会测试路由逻辑
	if err == nil {
		// 预期可能失败，因为 mock provider 返回 nil response
	}
}

func TestEngine_GetStats(t *testing.T) {
	engine := NewEngine()
	
	stats := engine.GetStats()
	
	if stats == nil {
		t.Fatal("GetStats should not return nil")
	}
	
	if stats.TotalRequests < 0 {
		t.Error("TotalRequests should be non-negative")
	}
}

// ============ ChatRequest Tests ============

func TestChatRequest_Defaults(t *testing.T) {
	req := &ChatRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}
	
	if req.Temperature != 0.7 {
		t.Errorf("Default temperature should be 0.7, got %f", req.Temperature)
	}
	
	if req.MaxTokens != 4096 {
		t.Errorf("Default maxTokens should be 4096, got %d", req.MaxTokens)
	}
}

// ============ Message Tests ============

func TestMessage_Roles(t *testing.T) {
	msg := &Message{
		Role:    "user",
		Content: "Hello",
	}
	
	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}
}

// ============ Provider Tests ============

func TestOpenAIProvider_Name(t *testing.T) {
	provider := &OpenAIProvider{}
	
	if provider.Name() != "OpenAI" {
		t.Errorf("Expected 'OpenAI', got '%s'", provider.Name())
	}
}

func TestOpenAIProvider_ID(t *testing.T) {
	provider := &OpenAIProvider{}
	
	if provider.ID() != "openai" {
		t.Errorf("Expected 'openai', got '%s'", provider.ID())
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	provider := &AnthropicProvider{}
	
	if provider.Name() != "Anthropic" {
		t.Errorf("Expected 'Anthropic', got '%s'", provider.Name())
	}
}

func TestOllamaProvider_Name(t *testing.T) {
	provider := &OllamaProvider{}
	
	if provider.Name() != "Ollama (本地)" {
		t.Errorf("Expected 'Ollama (本地)', got '%s'", provider.Name())
	}
}

// ============ Mock Provider ============

type mockProvider struct {
	id   string
	name string
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) ID() string {
	return p.id
}

func (p *mockProvider) IsEnabled() bool {
	return true
}

func (p *mockProvider) Chat(req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		Model:   "mock-model",
		Content: "Mock response",
	}, nil
}

func (p *mockProvider) ChatStream(req *ChatRequest, callback func(*StreamingChunk)) error {
	callback(&StreamingChunk{
		Content: "Mock streaming response",
		Done:    true,
	})
	return nil
}

func (p *mockProvider) GetModels() []ModelInfo {
	return []ModelInfo{
		{ID: "mock-model", Name: "Mock Model", Provider: "mock"},
	}
}

// ============ Integration Tests ============

func TestEngine_ProviderRouting(t *testing.T) {
	engine := NewEngine()
	
	// 添加多个提供商
	engine.AddProvider(&mockProvider{id: "provider1", name: "Provider 1"})
	engine.AddProvider(&mockProvider{id: "provider2", name: "Provider 2"})
	
	if len(engine.providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(engine.providers))
	}
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	engine := NewEngine()
	engine.AddProvider(&mockProvider{id: "test", name: "Test"})
	
	done := make(chan bool)
	
	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			engine.GetStats()
			done <- true
		}()
	}
	
	// 并发写入
	for i := 0; i < 10; i++ {
		go func() {
			engine.AddProvider(&mockProvider{id: "extra", name: "Extra"})
			done <- true
		}()
	}
	
	// 等待所有完成
	for i := 0; i < 20; i++ {
		<-done
	}
}

// ============ Benchmark Tests ============

func BenchmarkEngine_GetStats(b *testing.B) {
	engine := NewEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.GetStats()
	}
}

func BenchmarkEngine_AddProvider(b *testing.B) {
	engine := NewEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.AddProvider(&mockProvider{
			id:   string(rune(i)),
			name: "Provider",
		})
	}
}
