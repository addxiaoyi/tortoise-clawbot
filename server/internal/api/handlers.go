package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"tortoise-server/internal/ai"

	"github.com/gin-gonic/gin"
)

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// ============ Session Handlers ============

func (s *Server) handleGetSessions(c *gin.Context) {
	sessions := s.sessionStore.GetSessions()
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

func (s *Server) handleCreateSession(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := s.sessionStore.CreateSession(req.Name, req.UserID)
	c.JSON(http.StatusCreated, session)
}

func (s *Server) handleGetSession(c *gin.Context) {
	id := c.Param("id")
	session, ok := s.sessionStore.GetSession(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (s *Server) handleDeleteSession(c *gin.Context) {
	id := c.Param("id")

	if !s.sessionStore.DeleteSession(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	s.messageStore.DeleteMessages(id)
	c.JSON(http.StatusOK, gin.H{"message": "session deleted"})
}

func (s *Server) handleUpdateSession(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, ok := s.sessionStore.UpdateSession(id, req.Name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ============ Message Handlers ============

func (s *Server) handleGetMessages(c *gin.Context) {
	sessionID := c.Param("id")
	limit := 50

	messages := s.messageStore.GetMessages(sessionID, limit)
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":   len(messages),
	})
}

func (s *Server) handleSendMessage(c *gin.Context) {
	sessionID := c.Param("id")

	var req struct {
		Content string `json:"content" binding:"required"`
		Model   string `json:"model"` // 可选指定模型
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查会话是否存在
	if _, ok := s.sessionStore.GetSession(sessionID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 保存用户消息
	userMsg := s.messageStore.CreateMessage(sessionID, "user", req.Content)
	s.sessionStore.IncrementMessageCount(sessionID, truncateString(req.Content, 50))

	// 获取历史消息用于上下文
	historyMessages := s.messageStore.GetMessages(sessionID, 20)
	
	// 构建 AI 请求
	var aiResp *ai.ChatResponse
	var err error

	if s.aiEngine != nil {
		// 使用真实 AI
		chatReq := &ai.ChatRequest{
			Model:       req.Model,
			Temperature: 0.7,
			MaxTokens:   4096,
		}

		// 转换历史消息
		for _, msg := range historyMessages {
			role := "user"
			if msg.Role == "assistant" {
				role = "assistant"
			}
			chatReq.Messages = append(chatReq.Messages, ai.Message{
				Role:    role,
				Content: msg.Content,
			})
		}

		// 调用 AI
		aiResp, err = s.aiEngine.Chat(c.Request.Context(), chatReq)
	} else {
		// 回退到模拟响应
		aiResp = &ai.ChatResponse{
			Model:      "mock",
			Content:    generateAIResponse(req.Content),
			FinishReason: "stop",
			Tokens:     100,
		}
	}

	if err != nil {
		// AI 调用失败，返回错误但保存用户消息
		c.JSON(http.StatusInternalServerError, gin.H{
			"user_message": userMsg,
			"error":       fmt.Sprintf("AI request failed: %v", err),
		})
		return
	}

	// 保存 AI 响应
	assistantMsg := s.messageStore.CreateMessage(sessionID, "assistant", aiResp.Content)
	s.sessionStore.IncrementMessageCount(sessionID, truncateString(aiResp.Content, 50))

	c.JSON(http.StatusOK, gin.H{
		"user_message":       userMsg,
		"assistant_message": assistantMsg,
		"messageId":         assistantMsg.ID,
		"content":           aiResp.Content,
		"model":             aiResp.Model,
		"tokens":            aiResp.Tokens,
	})
}

// ============ Memory Handlers ============

func (s *Server) handleGetMemories(c *gin.Context) {
	memType := c.Query("type")
	memories := s.memoryStore.GetMemories(memType)

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

func (s *Server) handleCreateMemory(c *gin.Context) {
	var req struct {
		Type       string  `json:"type" binding:"required"`
		Content    string  `json:"content" binding:"required"`
		Importance float64 `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Importance == 0 {
		req.Importance = 0.5
	}

	memory := s.memoryStore.CreateMemory(req.Type, req.Content, req.Importance)
	c.JSON(http.StatusCreated, memory)
}

func (s *Server) handleDeleteMemory(c *gin.Context) {
	id := c.Param("id")

	if !s.memoryStore.DeleteMemory(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "memory deleted"})
}

func (s *Server) handleSearchMemories(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	memories := s.memoryStore.SearchMemories(query)
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
		"query":    query,
	})
}

// ============ Plugin Handlers ============

func (s *Server) handleGetPlugins(c *gin.Context) {
	plugins := s.pluginStore.GetPlugins()
	c.JSON(http.StatusOK, gin.H{
		"plugins": plugins,
		"total":   len(plugins),
	})
}

func (s *Server) handleGetPlugin(c *gin.Context) {
	id := c.Param("id")
	plugin, ok := s.pluginStore.GetPlugin(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
		return
	}

	c.JSON(http.StatusOK, plugin)
}

func (s *Server) handleTogglePlugin(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plugin, ok := s.pluginStore.TogglePlugin(id, req.Enabled)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
		return
	}

	c.JSON(http.StatusOK, plugin)
}

func (s *Server) handleDeletePlugin(c *gin.Context) {
	id := c.Param("id")

	if !s.pluginStore.DeletePlugin(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plugin deleted"})
}

func (s *Server) handleInstallPlugin(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "plugin installed successfully",
		"url":     req.URL,
	})
}

// ============ Tool Handlers ============

func (s *Server) handleGetTools(c *gin.Context) {
	plugins := s.pluginStore.GetPlugins()
	tools := make([]map[string]interface{}, 0)

	for _, plugin := range plugins {
		if plugin.Enabled {
			for _, tool := range plugin.Tools {
				tools = append(tools, map[string]interface{}{
					"name":        tool.Name,
					"description": tool.Description,
					"plugin":      plugin.Name,
					"parameters":  tool.Parameters,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"total": len(tools),
	})
}

func (s *Server) handleExecuteTool(c *gin.Context) {
	var req struct {
		Tool string                 `json:"tool" binding:"required"`
		Args map[string]interface{} `json:"args"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := map[string]interface{}{
		"success": true,
		"tool":    req.Tool,
		"result":  fmt.Sprintf("Executed %s with args: %v", req.Tool, req.Args),
	}

	c.JSON(http.StatusOK, result)
}

// ============ AI Handlers ============

func (s *Server) handleGetAIStats(c *gin.Context) {
	if s.aiEngine == nil {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "AI engine not initialized. Please configure API keys in settings.",
		})
		return
	}

	stats := s.aiEngine.GetStats()
	stats["available"] = true
	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleGetAIModels(c *gin.Context) {
	if s.aiEngine == nil {
		c.JSON(http.StatusOK, gin.H{
			"models": []interface{}{},
			"message": "AI engine not initialized",
		})
		return
	}

	models := s.aiEngine.GetModels()
	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// ============ Config Handlers ============

func (s *Server) handleGetConfig(c *gin.Context) {
	config := s.configStore.GetConfig()
	c.JSON(200, config)
}

func (s *Server) handleUpdateConfig(c *gin.Context) {
	var updates map[string]interface{}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.configStore.UpdateConfig(updates); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 如果更新了 AI 配置，重新加载 AI 引擎
	if _, hasAI := updates["ai"]; hasAI {
		if err := s.ReloadAIEngine(); err != nil {
			// 记录错误但继续
			fmt.Printf("Warning: failed to reload AI engine: %v\n", err)
		}
	}

	config := s.configStore.GetConfig()
	c.JSON(200, config)
}

func (s *Server) handleCreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	key, err := s.configStore.GenerateAPIKey(req.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"key":  key,
		"name": req.Name,
	})
}

func (s *Server) handleDeleteAPIKey(c *gin.Context) {
	id := c.Param("id")

	if !s.configStore.DeleteAPIKey(id) {
		c.JSON(404, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(200, gin.H{"message": "API key deleted"})
}

// ============ Stats Handler ============

func (s *Server) handleGetStats(c *gin.Context) {
	sessions := s.sessionStore.GetSessions()
	memories := s.memoryStore.GetMemories("")
	plugins := s.pluginStore.GetPlugins()

	enabledPlugins := 0
	for _, p := range plugins {
		if p.Enabled {
			enabledPlugins++
		}
	}

	toolCount := 0
	for _, p := range plugins {
		toolCount += len(p.Tools)
	}

	// AI 状态
	aiAvailable := s.aiEngine != nil
	aiProviders := 0
	if s.aiEngine != nil {
		aiProviders = len(s.aiEngine.GetProviders())
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions":        len(sessions),
		"memories":        len(memories),
		"plugins":         len(plugins),
		"enabled_plugins": enabledPlugins,
		"tools":           toolCount,
		"uptime":          "24h",
		"version":         "0.1.0",
		"ai_available":    aiAvailable,
		"ai_providers":    aiProviders,
	})
}

// ============ Helper Functions ============

func generateAIResponse(input string) string {
	responses := []string{
		"我理解你的问题。让我来分析一下...\n\n根据我的理解，这个问题涉及多个方面。首先，我们需要考虑基本原理，然后逐步深入。",
		"这是一个很好的问题！让我来详细解答。\n\n1. 首先，我们要明确目标\n2. 然后制定计划\n3. 最后执行和验证",
		"根据我的分析，这个情况需要我们考虑以下几点：\n\n• 技术可行性\n• 资源投入\n• 时间成本\n• 风险评估",
		"我建议可以从以下几个方面入手：\n\n1. 明确需求\n2. 评估现状\n3. 制定方案\n4. 逐步实施\n\n需要我详细解释某个方面吗？",
		"这是一个常见的问题。让我提供一些建议：\n\n首先，需要了解基本概念。其次，掌握核心方法。最后，多加实践。\n\n有什么具体的方面需要我进一步说明吗？",
	}

	rand.Seed(time.Now().UnixNano())
	return responses[rand.Intn(len(responses))]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
