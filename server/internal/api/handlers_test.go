package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tortoise-server/internal/store"
)

func setupTestServer() *Server {
	memStore := store.NewMemoryStore()
	sessionStore := store.NewSessionStore()
	messageStore := store.NewMessageStore()
	pluginStore := store.NewPluginStore()
	configStore := store.NewConfigStore()

	return NewServer(memStore, sessionStore, messageStore, pluginStore, configStore)
}

func TestHandleHealth(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", resp["status"])
	}
}

func TestCreateSession(t *testing.T) {
	server := setupTestServer()

	body := `{"name": "Test Session", "user_id": "user123"}`
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var session map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if session["name"] != "Test Session" {
		t.Errorf("expected name 'Test Session', got %v", session["name"])
	}
}

func TestGetSessions(t *testing.T) {
	server := setupTestServer()

	// 先创建会话
	body := `{"name": "Test Session", "user_id": "user123"}`
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	// 获取会话列表
	req = httptest.NewRequest("GET", "/api/v1/sessions", nil)
	w = httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	sessions, ok := resp["sessions"].([]interface{})
	if !ok {
		t.Fatal("expected sessions array")
	}

	if len(sessions) == 0 {
		t.Error("expected at least one session")
	}
}

func TestSendMessage(t *testing.T) {
	server := setupTestServer()

	// 创建会话
	body := `{"name": "Test Session", "user_id": "user123"}`
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	var session map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &session)
	sessionID := session["id"].(string)

	// 发送消息
	body = `{"content": "Hello, AI!"}`
	req = httptest.NewRequest("POST", "/api/v1/sessions/"+sessionID+"/messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["content"] == nil || resp["content"] == "" {
		t.Error("expected content in response")
	}
}

func TestDeleteSession(t *testing.T) {
	server := setupTestServer()

	// 创建会话
	body := `{"name": "Test Session", "user_id": "user123"}`
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	var session map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &session)
	sessionID := session["id"].(string)

	// 删除会话
	req = httptest.NewRequest("DELETE", "/api/v1/sessions/"+sessionID, nil)
	w = httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// 验证会话已删除
	req = httptest.NewRequest("GET", "/api/v1/sessions/"+sessionID, nil)
	w = httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestGetConfig(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/config", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &config); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if config["ai"] == nil {
		t.Error("expected ai config")
	}
	if config["channels"] == nil {
		t.Error("expected channels config")
	}
}

func TestUpdateConfig(t *testing.T) {
	server := setupTestServer()

	// 更新 AI 配置
	body := `{
		"ai": {
			"providers": [
				{
					"id": "openai",
					"enabled": true,
					"api_key": "sk-test123"
				}
			]
		}
	}`
	req := httptest.NewRequest("PATCH", "/api/v1/config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCreateMemory(t *testing.T) {
	server := setupTestServer()

	body := `{"type": "working", "content": "Test memory", "importance": 0.8}`
	req := httptest.NewRequest("POST", "/api/v1/memories", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var memory map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &memory); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if memory["content"] != "Test memory" {
		t.Errorf("expected content 'Test memory', got %v", memory["content"])
	}
}

func TestGetStats(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if stats["sessions"] == nil {
		t.Error("expected sessions count")
	}
	if stats["version"] != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %v", stats["version"])
	}
}

func TestCORS(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("OPTIONS", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header Access-Control-Allow-Origin: *")
	}
}

// Benchmark tests
func BenchmarkSendMessage(b *testing.B) {
	server := setupTestServer()

	// 创建会话
	body := `{"name": "Benchmark Session", "user_id": "user123"}`
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	var session map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &session)
	sessionID := session["id"].(string)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := `{"content": "Hello!"}`
		req := httptest.NewRequest("POST", "/api/v1/sessions/"+sessionID+"/messages", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.engine.ServeHTTP(w, req)
	}
}

// Mock AI Engine for testing
type MockAIEngine struct{}

func (m *MockAIEngine) Chat(ctx context.Context, req *struct{}) (*struct{}, error) {
	return nil, nil
}
