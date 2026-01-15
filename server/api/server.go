// REST API server

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/tortoise/server/gateway"
)

// Config holds API server configuration
type Config struct {
	Gateway *gateway.Gateway
	Port    int
}

// Server is the REST API server
type Server struct {
	cfg      *Config
	mux      *http.ServeMux
	server   *http.Server
}

// New creates a new API server
func New(cfg *Config) *Server {
	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}

	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// API v1 routes
	api := s.mux.PathPrefix("/api/v1").HandlerFunc(s.handleAPI)

	// Sessions
	api.HandleFunc("/sessions", s.handleSessions)
	api.HandleFunc("/sessions/", s.handleSessionByID)

	// Messages
	api.HandleFunc("/sessions/{sessionId}/messages", s.handleMessages)

	// Tools
	api.HandleFunc("/tools", s.handleTools)
	api.HandleFunc("/tools/", s.handleToolInvoke)

	// Memory
	api.HandleFunc("/memory", s.handleMemory)

	// Plugins
	api.HandleFunc("/plugins", s.handlePlugins)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "0.1.0",
	})
}

// handleAPI is middleware for API versioning
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Continue to specific handlers
	s.mux.ServeHTTP(w, r)
}

// handleSessions handles session management
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listSessions(w, r)
	case "POST":
		s.createSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listSessions lists all sessions
func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	stats := s.cfg.Gateway.Stats()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": []interface{}{},
		"stats":    stats,
	})
}

// createSession creates a new session
func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := map[string]interface{}{
		"sessionId": "sess_" + generateID(),
		"createdAt": time.Now().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// handleSessionByID handles single session operations
func (s *Server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	sessionID := extractPathParam(r.URL.Path, "/sessions/")

	switch r.Method {
	case "GET":
		s.getSession(w, r, sessionID)
	case "DELETE":
		s.deleteSession(w, r, sessionID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSession retrieves a session by ID
func (s *Server) getSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session := map[string]interface{}{
		"sessionId":     sessionID,
		"status":        "active",
		"messageCount":  0,
		"createdAt":     time.Now().Format(time.RFC3339),
		"lastActiveAt":  time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(session)
}

// deleteSession deletes a session
func (s *Server) deleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.WriteHeader(http.StatusNoContent)
}

// handleMessages handles message operations
func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := extractPathParam(r.URL.Path, "/messages")

	switch r.Method {
	case "GET":
		s.listMessages(w, r, sessionID)
	case "POST":
		s.sendMessage(w, r, sessionID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listMessages lists messages for a session
func (s *Server) listMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": []interface{}{},
	})
}

// sendMessage sends a message
func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"messageId": "msg_" + generateID(),
		"role":      "assistant",
		"content":    "Response from Tortoise",
		"metadata": map[string]interface{}{
			"model":   "gpt-4o",
			"tokens":  map[string]int{"prompt": 10, "completion": 20},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleTools handles tool operations
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listTools(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listTools lists available tools
func (s *Server) listTools(w http.ResponseWriter, r *http.Request) {
	tools := []map[string]interface{}{
		{
			"name":        "web_search",
			"description": "Search the web for information",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":      map[string]string{"type": "string"},
					"numResults": map[string]interface{}{"type": "integer", "default": 5},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "calculator",
			"description": "Perform mathematical calculations",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]string{"type": "string"},
				},
				"required": []string{"expression"},
			},
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"tools": tools})
}

// handleToolInvoke handles tool invocation
func (s *Server) handleToolInvoke(w http.ResponseWriter, r *http.Request) {
	toolName := extractPathParam(r.URL.Path, "/tools/")

	if r.Method == "POST" {
		s.invokeTool(w, r, toolName)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// invokeTool invokes a tool
func (s *Server) invokeTool(w http.ResponseWriter, r *http.Request, toolName string) {
	var req InvokeToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"toolName":       toolName,
		"success":        true,
		"result":         map[string]interface{}{"output": "Tool result"},
		"executionTimeMs": 50,
	}

	json.NewEncoder(w).Encode(result)
}

// handleMemory handles memory operations
func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.searchMemory(w, r)
	case "POST":
		s.storeMemory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// searchMemory searches memory
func (s *Server) searchMemory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limit := 10

	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{
				"id":         "mem_1",
				"content":    "Sample memory",
				"similarity": 0.95,
			},
		},
		"query": query,
		"limit": limit,
	})
}

// storeMemory stores a memory
func (s *Server) storeMemory(w http.ResponseWriter, r *http.Request) {
	var req StoreMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"id":      "mem_" + generateID(),
		"success": true,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// handlePlugins handles plugin operations
func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listPlugins(w, r)
	case "POST":
		s.installPlugin(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listPlugins lists installed plugins
func (s *Server) listPlugins(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plugins": []interface{}{},
	})
}

// installPlugin installs a plugin
func (s *Server) installPlugin(w http.ResponseWriter, r *http.Request) {
	var req InstallPluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"success": true,
		"plugin": map[string]interface{}{
			"id":      req.PluginID,
			"version": "1.0.0",
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// Request types
type CreateSessionRequest struct {
	UserID  string                 `json:"userId"`
	Config  map[string]interface{} `json:"config"`
}

type SendMessageRequest struct {
	Content     string                   `json:"content"`
	Type        string                   `json:"type"`
	Attachments []map[string]interface{} `json:"attachments"`
	Tools       []string                 `json:"tools"`
	Stream      bool                     `json:"stream"`
}

type InvokeToolRequest struct {
	SessionID string                 `json:"sessionId"`
	Arguments map[string]interface{} `json:"arguments"`
}

type StoreMemoryRequest struct {
	SessionID  string    `json:"sessionId"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	Tags       []string  `json:"tags"`
	Importance float64   `json:"importance"`
}

type InstallPluginRequest struct {
	Source   string `json:"source"`
	PluginID string `json:"pluginId"`
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("%08x", time.Now().UnixNano())
}

func extractPathParam(path, prefix string) string {
	if len(path) > len(prefix) {
		return path[len(prefix):]
	}
	return ""
}
