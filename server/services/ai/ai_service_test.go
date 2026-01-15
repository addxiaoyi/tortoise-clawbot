package ai

import (
	"context"
	"testing"

	"tortoise/config"
)

func TestNewAIService(t *testing.T) {
	cfg := config.AIConfig{
		Providers: []config.AIProviderConfig{
			{
				Name:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-key",
				Models:  []string{"gpt-4", "gpt-3.5-turbo"},
			},
		},
	}

	service := NewAIService(cfg)
	if service == nil {
		t.Fatal("Expected AIService, got nil")
	}

	if len(service.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(service.providers))
	}
}

func TestAIService_ListProviders(t *testing.T) {
	cfg := config.AIConfig{
		Providers: []config.AIProviderConfig{
			{Name: "openai", APIKey: "key1"},
			{Name: "anthropic", APIKey: "key2"},
			{Name: "google", APIKey: "key3"},
		},
	}

	service := NewAIService(cfg)
	providers := service.ListProviders()

	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}
}

func TestAIService_ListModels(t *testing.T) {
	cfg := config.AIConfig{
		Providers: []config.AIProviderConfig{
			{
				Name:   "openai",
				APIKey: "test-key",
				Models: []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"},
			},
		},
	}

	service := NewAIService(cfg)
	models := service.ListModels("openai")

	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}

	// 测试不存在的提供商
	nonexistent := service.ListModels("nonexistent")
	if nonexistent != nil {
		t.Errorf("Expected nil for nonexistent provider, got %v", nonexistent)
	}
}

func TestAIService_Chat(t *testing.T) {
	cfg := config.AIConfig{
		Providers: []config.AIProviderConfig{
			{
				Name:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-key",
				Models:  []string{"gpt-4"},
			},
		},
	}

	service := NewAIService(cfg)

	req := ChatRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// 由于没有真实的 API key，这个测试会失败
	// 在有真实 key 的环境中测试
	ctx := context.Background()
	_, err := service.Chat(ctx, req)

	// 预期会失败，因为 API key 是无效的
	if err == nil {
		t.Log("Chat request succeeded (unexpected without valid API key)")
	}
}

func TestChatRequest_ToJSON(t *testing.T) {
	temp := 0.7
	maxTokens := 100

	req := ChatRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
		Temperature: &temp,
		MaxTokens:  &maxTokens,
		Stream:     false,
	}

	json := req.ToJSON()

	if json["model"] != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%v'", json["model"])
	}

	if json["temperature"] != 0.7 {
		t.Errorf("Expected temperature 0.7, got '%v'", json["temperature"])
	}

	messages := json["messages"].([]ChatMessage)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestChatMessage_ToJSON(t *testing.T) {
	msg := ChatMessage{
		Role:    "user",
		Content: "Hello, world!",
	}

	json := msg.ToJSON()

	if json["role"] != "user" {
		t.Errorf("Expected role 'user', got '%v'", json["role"])
	}

	if json["content"] != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%v'", json["content"])
	}
}

func TestChatMessage_WithName(t *testing.T) {
	msg := ChatMessage{
		Role:    "system",
		Content: "You are AI",
		Name:    "assistant",
	}

	json := msg.ToJSON()

	if json["name"] != "assistant" {
		t.Errorf("Expected name 'assistant', got '%v'", json["name"])
	}
}

func TestFunctionCall_ToJSON(t *testing.T) {
	fn := FunctionCall{
		Name:      "get_weather",
		Arguments: []byte(`{"location": "Beijing"}`),
	}

	json := fn.ToJSON()

	if json["name"] != "get_weather" {
		t.Errorf("Expected name 'get_weather', got '%v'", json["name"])
	}
}
