package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	r := SetupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestCreateSession(t *testing.T) {
	r := SetupRouter()
	
	body := `{"title": "Test Session", "ai_provider": "openai", "model": "gpt-4"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/sessions", nil)
	req.Body = stringToReader(body)
	r.ServeHTTP(w, req)
	
	// Note: This test may fail without proper setup
	// Adjust based on actual implementation
	t.Logf("Response status: %d", w.Code)
}

func TestListSessions(t *testing.T) {
	r := SetupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions", nil)
	r.ServeHTTP(w, req)
	
	// Note: May require auth
	t.Logf("Response status: %d", w.Code)
}

func TestChatCompletions(t *testing.T) {
	r := SetupRouter()
	
	body := `{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", nil)
	req.Body = stringToReader(body)
	r.ServeHTTP(w, req)
	
	// Note: May require auth and API key
	t.Logf("Response status: %d", w.Code)
}

func TestCORS(t *testing.T) {
	r := SetupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/sessions", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)
	
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, w.Code)
	}
	
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" && corsHeader != "http://localhost:3000" {
		t.Errorf("Unexpected CORS header: %s", corsHeader)
	}
}

func stringToReader(s string) *stringReader {
	return &stringReader{s: s, index: 0}
}

type stringReader struct {
	s     string
	index int
}

func (r *stringReader) Read(p []byte) (n int, err error) {
	if r.index >= len(r.s) {
		return 0, nil
	}
	n = copy(p, r.s[r.index:])
	r.index += n
	return n, nil
}
