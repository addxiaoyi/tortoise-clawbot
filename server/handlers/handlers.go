package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"tortoise/database"
	"tortoise/services/ai"
	"tortoise/services/channel"
	"tortoise/services/plugin"
	"tortoise/websocket"
)

// Dependencies 依赖注入
type Dependencies struct {
	DB             database.DB
	AIService      *ai.AIService
	ChannelService *channel.ChannelService
	PluginService  *plugin.PluginService
	WSHub          *websocket.Hub
}

// Handler 处理器
type Handler struct {
	deps Dependencies
}

// NewHandler 创建处理器
func NewHandler(deps Dependencies) *Handler {
	return &Handler{deps: deps}
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "0.1.0",
	})
}

// ========== 会话处理 ==========

// ListSessions 列出会话
func (h *Handler) ListSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	
	sessions, err := h.deps.DB.ListSessions(userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// CreateSession 创建会话
func (h *Handler) CreateSession(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Title      string `json:"title"`
		AIProvider string `json:"ai_provider"`
		Model      string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	session := &database.Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		Title:      req.Title,
		AIProvider: req.AIProvider,
		Model:      req.Model,
		CreatedAt:  time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if session.Title == "" {
		session.Title = "新会话"
	}
	
	if err := h.deps.DB.CreateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, session)
}

// GetSession 获取会话
func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")
	
	session, err := h.deps.DB.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	
	c.JSON(http.StatusOK, session)
}

// DeleteSession 删除会话
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.DB.DeleteSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListMessages 列出消息
func (h *Handler) ListMessages(c *gin.Context) {
	sessionID := c.Param("id")
	
	messages, err := h.deps.DB.ListMessages(sessionID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ========== AI 聊天 ==========

// ChatCompletions 聊天补全
func (h *Handler) ChatCompletions(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Model       string  `json:"model"`
		Messages    []ai.ChatMessage `json:"messages"`
		Temperature *float64 `json:"temperature"`
		MaxTokens   *int    `json:"max_tokens"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 获取会话 ID（可选）
	sessionID := c.GetHeader("X-Session-ID")
	
	// 保存用户消息
	if sessionID != "" {
		for _, msg := range req.Messages {
			if msg.Role == "user" {
				message := &database.Message{
					ID:        uuid.New().String(),
					SessionID: sessionID,
					Role:      msg.Role,
					Content:   msg.Content,
					CreatedAt: time.Now(),
				}
				h.deps.DB.CreateMessage(message)
			}
		}
		
		// 更新会话
		session, _ := h.deps.DB.GetSession(sessionID)
		if session != nil {
			session.UpdatedAt = time.Now()
			session.AIProvider = "openai"
			session.Model = req.Model
			h.deps.DB.UpdateSession(session)
		}
	}
	
	// 调用 AI
	chatReq := ai.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	
	resp, err := h.deps.AIService.Chat(c.Request.Context(), "openai", chatReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 保存 AI 响应
	if sessionID != "" && len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		message := &database.Message{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   choice.Message.Content,
			Model:     resp.Model,
			CreatedAt: time.Now(),
		}
		message.SetMetadata(map[string]interface{}{
			"tokens":       resp.Usage.TotalTokens,
			"finish_reason": choice.FinishReason,
		})
		h.deps.DB.CreateMessage(message)
	}
	
	c.JSON(http.StatusOK, resp)
}

// ChatCompletionsStream 流式聊天补全
func (h *Handler) ChatCompletionsStream(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Model       string  `json:"model"`
		Messages    []ai.ChatMessage `json:"messages"`
		Temperature *float64 `json:"temperature"`
		MaxTokens   *int    `json:"max_tokens"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 获取会话 ID
	sessionID := c.GetHeader("X-Session-ID")
	
	// 保存用户消息
	if sessionID != "" {
		for _, msg := range req.Messages {
			if msg.Role == "user" {
				message := &database.Message{
					ID:        uuid.New().String(),
					SessionID: sessionID,
					Role:      msg.Role,
					Content:   msg.Content,
					CreatedAt: time.Now(),
				}
				h.deps.DB.CreateMessage(message)
			}
		}
	}
	
	// 设置 SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	
	// 流式调用 AI
	chatReq := ai.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	
	respChan, errChan := h.deps.AIService.StreamChat(c.Request.Context(), "openai", chatReq)
	
	// 收集完整响应用于保存
	fullContent := ""
	
	c.Stream(func(w io.Writer) bool {
		select {
		case resp, ok := <-respChan:
			if !ok {
				// 保存完整响应
				if sessionID != "" && fullContent != "" {
					message := &database.Message{
						ID:        uuid.New().String(),
						SessionID: sessionID,
						Role:      "assistant",
						Content:   fullContent,
						Model:     req.Model,
						CreatedAt: time.Now(),
					}
					h.deps.DB.CreateMessage(message)
				}
				return false
			}
			
			// 发送数据
			if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
				content := resp.Choices[0].Delta.Content
				fullContent += content
				
				data, _ := json.Marshal(resp)
				c.SSEvent("message", string(data))
			}
			return true
			
		case err := <-errChan:
			c.SSEvent("error", err.Error())
			return false
		}
	})
}

// ========== 渠道处理 ==========

// ListChannels 列出渠道
func (h *Handler) ListChannels(c *gin.Context) {
	userID := c.GetString("user_id")
	
	channels, err := h.deps.DB.ListChannels(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// CreateChannel 创建渠道
func (h *Handler) CreateChannel(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	channel := &database.Channel{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      req.Type,
		Name:      req.Name,
		Config:    req.Config,
		Status:    "disconnected",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := h.deps.DB.CreateChannel(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, channel)
}

// GetChannel 获取渠道
func (h *Handler) GetChannel(c *gin.Context) {
	id := c.Param("id")
	
	channel, err := h.deps.DB.GetChannel(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	c.JSON(http.StatusOK, channel)
}

// UpdateChannel 更新渠道
func (h *Handler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	
	var req struct {
		Name   string `json:"name"`
		Config string `json:"config"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	channel, err := h.deps.DB.GetChannel(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Config != "" {
		channel.Config = req.Config
	}
	if req.Status != "" {
		channel.Status = req.Status
	}
	channel.UpdatedAt = time.Now()
	
	if err := h.deps.DB.UpdateChannel(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, channel)
}

// DeleteChannel 删除渠道
func (h *Handler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.DB.DeleteChannel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ConnectChannel 连接渠道
func (h *Handler) ConnectChannel(c *gin.Context) {
	id := c.Param("id")
	
	channel, err := h.deps.DB.GetChannel(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// 连接渠道
	if err := h.deps.ChannelService.Connect(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	channel.Status = "connected"
	channel.UpdatedAt = time.Now()
	h.deps.DB.UpdateChannel(channel)
	
	c.JSON(http.StatusOK, channel)
}

// DisconnectChannel 断开渠道
func (h *Handler) DisconnectChannel(c *gin.Context) {
	id := c.Param("id")
	
	channel, err := h.deps.DB.GetChannel(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// 断开渠道
	if err := h.deps.ChannelService.Disconnect(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	channel.Status = "disconnected"
	channel.UpdatedAt = time.Now()
	h.deps.DB.UpdateChannel(channel)
	
	c.JSON(http.StatusOK, channel)
}

// ========== 记忆处理 ==========

// ListMemories 列出记忆
func (h *Handler) ListMemories(c *gin.Context) {
	userID := c.GetString("user_id")
	
	memories, err := h.deps.DB.ListMemories(userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"memories": memories})
}

// CreateMemory 创建记忆
func (h *Handler) CreateMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Type    string   `json:"type"`
		Tags    []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	memory := &database.Memory{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	memory.SetTags(req.Tags)
	
	if err := h.deps.DB.CreateMemory(memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, memory)
}

// GetMemory 获取记忆
func (h *Handler) GetMemory(c *gin.Context) {
	id := c.Param("id")
	
	memory, err := h.deps.DB.GetMemory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		return
	}
	
	c.JSON(http.StatusOK, memory)
}

// UpdateMemory 更新记忆
func (h *Handler) UpdateMemory(c *gin.Context) {
	id := c.Param("id")
	
	var req struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Type    string   `json:"type"`
		Tags    []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	memory, err := h.deps.DB.GetMemory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		return
	}
	
	if req.Title != "" {
		memory.Title = req.Title
	}
	if req.Content != "" {
		memory.Content = req.Content
	}
	if req.Type != "" {
		memory.Type = req.Type
	}
	if req.Tags != nil {
		memory.SetTags(req.Tags)
	}
	memory.UpdatedAt = time.Now()
	
	if err := h.deps.DB.UpdateMemory(memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, memory)
}

// DeleteMemory 删除记忆
func (h *Handler) DeleteMemory(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.DB.DeleteMemory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SearchMemory 搜索记忆
func (h *Handler) SearchMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	query := c.Query("q")
	
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}
	
	memories, err := h.deps.DB.SearchMemories(userID, query, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"memories": memories})
}

// ========== 插件处理 ==========

// ListPlugins 列出插件
func (h *Handler) ListPlugins(c *gin.Context) {
	plugins := h.deps.PluginService.ListPlugins()
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// InstallPlugin 安装插件
func (h *Handler) InstallPlugin(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.deps.PluginService.Install(req.ID, req.URL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// EnablePlugin 启用插件
func (h *Handler) EnablePlugin(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.PluginService.Enable(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DisablePlugin 禁用插件
func (h *Handler) DisablePlugin(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.PluginService.Disable(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UninstallPlugin 卸载插件
func (h *Handler) UninstallPlugin(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.deps.PluginService.Uninstall(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ========== 配置处理 ==========

// GetConfig 获取配置
func (h *Handler) GetConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	
	config := make(map[string]string)
	
	// 读取所有配置
	keys := []string{"ai_provider", "ai_model", "theme", "language"}
	for _, key := range keys {
		value, _ := h.deps.DB.GetConfig(userID, key)
		if value != "" {
			config[key] = value
		}
	}
	
	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新配置
func (h *Handler) UpdateConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var config map[string]string
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	for key, value := range config {
		if err := h.deps.DB.SetConfig(userID, key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to set %s: %s", key, err.Error())})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ========== WebSocket ==========

// WebSocket WebSocket 处理
func (h *Handler) WebSocket(c *gin.Context) {
	// 升级到 WebSocket
	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 注册到 Hub
	client := websocket.NewClient(h.deps.WSHub, conn)
	h.deps.WSHub.Register <- client
	
	// 启动读取
	go client.ReadPump()
	go client.WritePump()
}
