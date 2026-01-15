package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"

	"tortoise-server/internal/ai"
	"tortoise-server/internal/channel"
	"tortoise-server/internal/store"
)

// Server HTTP/WebSocket 服务器
type Server struct {
	mux          *http.ServeMux
	memStore     *store.MemoryStore
	sessionStore *store.SessionStore
	messageStore *store.MessageStore
	pluginStore  *store.PluginStore
	configStore  *store.ConfigStore
	aiEngine     *ai.Engine
	channelMgr   *channel.Manager
	upgrader     websocket.Upgrader
	wsClients    map[string]*wsClient
	wsMu         sync.RWMutex
}

// wsClient WebSocket 客户端
type wsClient struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Server *Server
}

// NewServer 创建 API 服务器
func NewServer(
	memStore *store.MemoryStore,
	sessionStore *store.SessionStore,
	messageStore *store.MessageStore,
	pluginStore *store.PluginStore,
	configStore *store.ConfigStore,
) *Server {
	s := &Server{
		mux:          http.NewServeMux(),
		memStore:     memStore,
		sessionStore: sessionStore,
		messageStore: messageStore,
		pluginStore:  pluginStore,
		configStore:  configStore,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		wsClients: make(map[string]*wsClient),
	}

	// 初始化 AI 引擎
	s.aiEngine = ai.NewEngine()

	// 初始化渠道管理器
	s.channelMgr = channel.NewManager()

	// 注册路由
	s.registerRoutes()

	return s
}

// registerRoutes 注册所有路由
func (s *Server) registerRoutes() {
	// 健康检查
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// 会话管理
	s.mux.HandleFunc("GET /api/v1/sessions", s.handleGetSessions)
	s.mux.HandleFunc("POST /api/v1/sessions", s.handleCreateSession)
	s.mux.HandleFunc("DELETE /api/v1/sessions/{id}", s.handleDeleteSession)

	// 消息管理
	s.mux.HandleFunc("GET /api/v1/sessions/{id}/messages", s.handleGetMessages)
	s.mux.HandleFunc("POST /api/v1/sessions/{id}/messages", s.handleSendMessage)

	// 记忆管理
	s.mux.HandleFunc("GET /api/v1/memories", s.handleGetMemories)
	s.mux.HandleFunc("POST /api/v1/memories", s.handleCreateMemory)
	s.mux.HandleFunc("DELETE /api/v1/memories/{id}", s.handleDeleteMemory)
	s.mux.HandleFunc("GET /api/v1/memories/search", s.handleSearchMemories)

	// 插件管理
	s.mux.HandleFunc("GET /api/v1/plugins", s.handleGetPlugins)
	s.mux.HandleFunc("POST /api/v1/plugins/{id}/toggle", s.handleTogglePlugin)

	// 配置管理
	s.mux.HandleFunc("GET /api/v1/config", s.handleGetConfig)
	s.mux.HandleFunc("PATCH /api/v1/config", s.handleUpdateConfig)

	// 统计
	s.mux.HandleFunc("GET /api/v1/stats", s.handleGetStats)
	s.mux.HandleFunc("GET /api/v1/ai/stats", s.handleGetAIStats)

	// WebSocket
	s.mux.HandleFunc("GET /ws", s.handleWebSocket)
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	// 启动渠道管理器
	s.channelMgr.Start()

	// 应用配置到 AI 引擎
	s.applyAIConfig()

	// 应用渠道配置
	s.applyChannelConfig()

	log.Printf("✅ API 服务器已启动: %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// applyAIConfig 应用 AI 配置
func (s *Server) applyAIConfig() {
	cfg := s.configStore.GetConfig()

	for _, provider := range cfg.AI.Providers {
		if !provider.Enabled {
			continue
		}

		apiKey := s.configStore.GetSecret(provider.ID)
		if apiKey == "" && provider.ID != "ollama" {
			log.Printf("⚠️ %s API Key 未配置", provider.Name)
			continue
		}

		config := &ai.ProviderConfig{
			APIKey:  apiKey,
			Model:   provider.Model,
			BaseURL: provider.BaseURL,
			Enabled: provider.Enabled,
		}

		switch provider.ID {
		case "openai":
			p, err := ai.NewOpenAIProvider(config)
			if err != nil {
				log.Printf("❌ OpenAI Provider 初始化失败: %v", err)
				continue
			}
			s.aiEngine.AddProvider(p)
			log.Printf("✅ OpenAI Provider 已添加 (模型: %s)", provider.Model)

		case "anthropic":
			p, err := ai.NewAnthropicProvider(config)
			if err != nil {
				log.Printf("❌ Anthropic Provider 初始化失败: %v", err)
				continue
			}
			s.aiEngine.AddProvider(p)
			log.Printf("✅ Anthropic Provider 已添加 (模型: %s)", provider.Model)

		case "ollama":
			p, err := ai.NewOllamaProvider(config)
			if err != nil {
				log.Printf("❌ Ollama Provider 初始化失败: %v", err)
				continue
			}
			s.aiEngine.AddProvider(p)
			log.Printf("✅ Ollama Provider 已添加 (模型: %s)", provider.Model)
		}
	}
}

// applyChannelConfig 应用渠道配置
func (s *Server) applyChannelConfig() {
	cfg := s.configStore.GetConfig()

	// Telegram
	if cfg.Channels.Telegram.Enabled {
		token := s.configStore.GetChannelSecret("telegram")
		if token != "" {
			ch := channel.NewTelegramChannel(token, nil)
			ch.SetAIEngine(s.aiEngine)
			s.channelMgr.Connect(1, nil, map[string]string{
				"enabled": "true",
			})
			log.Printf("✅ Telegram Bot 已配置")
		}
	}

	// Discord
	if cfg.Channels.Discord.Enabled {
		token := s.configStore.GetChannelSecret("discord")
		if token != "" {
			ch := channel.NewDiscordChannel(token)
			ch.SetAIEngine(s.aiEngine)
			s.channelMgr.Connect(2, nil, map[string]string{
				"enabled": "true",
			})
			log.Printf("✅ Discord Bot 已配置")
		}
	}
}

// ============ HTTP Handlers ============

// handleHealth 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleGetSessions 获取会话列表
func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.sessionStore.List()
	json.NewEncoder(w).Encode(sessions)
}

// handleCreateSession 创建会话
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session := s.sessionStore.Create(req.Name, req.UserID)
	json.NewEncoder(w).Encode(session)
}

// handleDeleteSession 删除会话
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.sessionStore.Delete(id)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGetMessages 获取消息
func (s *Server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	messages := s.messageStore.GetBySession(sessionID)
	json.NewEncoder(w).Encode(messages)
}

// handleSendMessage 发送消息
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 保存用户消息
	userMsg := &store.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	s.messageStore.Create(userMsg)

	// 调用 AI
	cfg := s.configStore.GetConfig()
	model := cfg.AI.DefaultModel

	aiReq := &ai.ChatRequest{
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   4096,
		Messages: []ai.Message{
			{Role: "user", Content: req.Content},
		},
	}

	resp, err := s.aiEngine.Chat(r.Context(), aiReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 保存 AI 回复
	aiMsg := &store.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   resp.Content,
		CreatedAt: time.Now(),
	}
	s.messageStore.Create(aiMsg)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messageId": aiMsg.ID,
		"content":   aiMsg.Content,
	})
}

// handleGetMemories 获取记忆
func (s *Server) handleGetMemories(w http.ResponseWriter, r *http.Request) {
	memories := s.memStore.List()
	json.NewEncoder(w).Encode(memories)
}

// handleCreateMemory 创建记忆
func (s *Server) handleCreateMemory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type       string  `json:"type"`
		Content    string  `json:"content"`
		Importance int     `json:"importance"`
		Tags       []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mem := s.memStore.Create(req.Type, req.Content, req.Importance, req.Tags)
	json.NewEncoder(w).Encode(mem)
}

// handleDeleteMemory 删除记忆
func (s *Server) handleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.memStore.Delete(id)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleSearchMemories 搜索记忆
func (s *Server) handleSearchMemories(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	memories := s.memStore.Search(query)
	json.NewEncoder(w).Encode(memories)
}

// handleGetPlugins 获取插件列表
func (s *Server) handleGetPlugins(w http.ResponseWriter, r *http.Request) {
	plugins := s.pluginStore.List()
	json.NewEncoder(w).Encode(plugins)
}

// handleTogglePlugin 切换插件状态
func (s *Server) handleTogglePlugin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pluginStore := s.pluginStore.Get(id)
	if pluginStore == nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	s.pluginStore.Toggle(id)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGetConfig 获取配置
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.configStore.GetConfig()
	json.NewEncoder(w).Encode(cfg)
}

// handleUpdateConfig 更新配置
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.configStore.UpdateConfig(updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 重新应用配置
	s.applyAIConfig()
	s.applyChannelConfig()

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGetStats 获取统计
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"sessions":  s.sessionStore.Count(),
		"messages":  s.messageStore.Count(),
		"memories":  s.memStore.Count(),
		"plugins":   s.pluginStore.Count(),
		"channels":  len(s.channelMgr.List()),
	}
	json.NewEncoder(w).Encode(stats)
}

// handleGetAIStats 获取 AI 统计
func (s *Server) handleGetAIStats(w http.ResponseWriter, r *http.Request) {
	stats := s.aiEngine.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// ============ WebSocket Handler ============

// handleWebSocket 处理 WebSocket 连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &wsClient{
		ID:     uuid.New().String(),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Server: s,
	}

	s.wsMu.Lock()
	s.wsClients[client.ID] = client
	s.wsMu.Unlock()

	go client.writePump()
	go client.readPump()
}

// writePump 写泵
func (c *wsClient) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// readPump 读泵
func (c *wsClient) readPump() {
	defer func() {
		c.Server.wsMu.Lock()
		delete(c.Server.wsClients, c.ID)
		c.Server.wsMu.Unlock()
		c.Conn.Close()
		close(c.Send)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 处理消息
		go c.handleMessage(message)
	}
}

// handleMessage 处理 WebSocket 消息
func (c *wsClient) handleMessage(message []byte) {
	var req struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(message, &req); err != nil {
		return
	}

	switch req.Type {
	case "chat":
		c.handleChat(req.Payload)
	case "ping":
		c.Send <- []byte(`{"type":"pong"}`)
	}
}

// handleChat 处理聊天消息
func (c *wsClient) handleChat(payload json.RawMessage) {
	var req struct {
		SessionID string `json:"sessionId"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return
	}

	// 发送确认
	ack, _ := json.Marshal(map[string]interface{}{
		"type": "chat_ack",
		"payload": map[string]string{
			"sessionId": req.SessionID,
			"status":    "processing",
		},
	})
	c.Send <- ack

	// 调用 AI
	ctx := context.Background()
	aiReq := &ai.ChatRequest{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		Messages: []ai.Message{
			{Role: "user", Content: req.Content},
		},
	}

	resp, err := c.Server.aiEngine.Chat(ctx, aiReq)
	content := "AI 服务未配置"
	if err == nil {
		content = resp.Content
	}

	// 发送回复
	reply, _ := json.Marshal(map[string]interface{}{
		"type": "chat_reply",
		"payload": map[string]interface{}{
			"sessionId": req.SessionID,
			"content":   content,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
	c.Send <- reply
}
