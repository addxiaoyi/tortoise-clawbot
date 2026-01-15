package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============ Memory API Handlers ============

// MemoryHandler 记忆处理器
type MemoryHandler struct{}

// NewMemoryHandler 创建记忆处理器
func NewMemoryHandler() *MemoryHandler {
	return &MemoryHandler{}
}

// RegisterRoutes 注册路由
func (h *MemoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 记忆 CRUD
	r.GET("/memories", h.ListMemories)
	r.POST("/memories", h.CreateMemory)
	r.GET("/memories/:id", h.GetMemory)
	r.PUT("/memories/:id", h.UpdateMemory)
	r.DELETE("/memories/:id", h.DeleteMemory)
	
	// 搜索
	r.GET("/memories/search", h.SearchMemories)
	r.GET("/memories/similar", h.FindSimilar)
	
	// 记忆类型
	r.GET("/memories/episodic", h.ListEpisodicMemories)
	r.GET("/memories/semantic", h.ListSemanticMemories)
	r.GET("/memories/procedural", h.ListProceduralMemories)
	
	// 固定记忆
	r.GET("/memories/pinned", h.ListPinnedMemories)
	r.POST("/memories/pinned/:id", h.PinMemory)
	r.DELETE("/memories/pinned/:id", h.UnpinMemory)
	
	// 统计
	r.GET("/stats", h.GetMemoryStats)
	r.GET("/forget", h.TriggerForget)
}

// MemoryResponse 记忆响应
type MemoryResponse struct {
	ID          string                 `json:"id"`
	Type       string                 `json:"type"`
	Content    string                 `json:"content"`
	Importance float64                `json:"importance"`
	AccessCount int                  `json:"access_count"`
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
	LastAccess string                `json:"last_access"`
	Tags       []string              `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Embedding  []float64            `json:"embedding,omitempty"`
	Source     string                `json:"source"`    // user, ai, system
	Archived   bool                  `json:"archived"`
}

// ListMemories 列出记忆
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "episodic",
			Content:     "用户询问如何配置 Telegram Bot",
			Importance:  0.8,
			AccessCount: 5,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Tags:        []string{"telegram", "config"},
		},
		{
			ID:          uuid.New().String(),
			Type:        "semantic",
			Content:     "Tortoise 是一个 AI 代理框架",
			Importance:  0.9,
			AccessCount: 100,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Tags:        []string{"tortoise", "ai"},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

// CreateMemory 创建记忆
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	var req struct {
		Type       string                 `json:"type" binding:"required"`
		Content    string                 `json:"content" binding:"required"`
		Importance float64                `json:"importance"`
		Tags       []string              `json:"tags"`
		Metadata   map[string]interface{} `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	memory := MemoryResponse{
		ID:          uuid.New().String(),
		Type:        req.Type,
		Content:     req.Content,
		Importance:  req.Importance,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		AccessCount: 0,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusCreated, memory)
}

// GetMemory 获取记忆
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	id := c.Param("id")
	
	memory := MemoryResponse{
		ID:          id,
		Type:        "episodic",
		Content:     "Sample memory content",
		Importance:  0.8,
		AccessCount: 5,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, memory)
}

// UpdateMemory 更新记忆
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	id := c.Param("id")
	
	var req struct {
		Content    string   `json:"content"`
		Importance float64  `json:"importance"`
		Tags       []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	memory := MemoryResponse{
		ID:          id,
		Type:        "episodic",
		Content:     req.Content,
		Importance:  req.Importance,
		Tags:        req.Tags,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, memory)
}

// DeleteMemory 删除记忆
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted", "id": id})
}

// SearchMemories 搜索记忆
func (h *MemoryHandler) SearchMemories(c *gin.Context) {
	query := c.Query("q")
	memoryType := c.Query("type")
	limit := 10
	
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        memoryType,
			Content:     "匹配: " + query,
			Importance:  0.9,
			AccessCount: 5,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
		"query":    query,
	})
}

// FindSimilar 查找相似记忆
func (h *MemoryHandler) FindSimilar(c *gin.Context) {
	content := c.Query("content")
	limit := 10
	
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "semantic",
			Content:     "相似记忆: " + content,
			Importance:  0.95,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
		"query":    content,
	})
}

// ListEpisodicMemories 列出情景记忆
func (h *MemoryHandler) ListEpisodicMemories(c *gin.Context) {
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "episodic",
			Content:     "用户在下午 3 点询问天气",
			Importance:  0.7,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

// ListSemanticMemories 列出语义记忆
func (h *MemoryHandler) ListSemanticMemories(c *gin.Context) {
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "semantic",
			Content:     "用户名叫张三",
			Importance:  0.9,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

// ListProceduralMemories 列出程序性记忆
func (h *MemoryHandler) ListProceduralMemories(c *gin.Context) {
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "procedural",
			Content:     "用户偏好使用暗色主题",
			Importance:  0.8,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

// ListPinnedMemories 列出固定记忆
func (h *MemoryHandler) ListPinnedMemories(c *gin.Context) {
	memories := []MemoryResponse{
		{
			ID:          uuid.New().String(),
			Type:        "pinned",
			Content:     "重要规则：永远不要删除用户数据",
			Importance:  1.0,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"total":    len(memories),
	})
}

// PinMemory 固定记忆
func (h *MemoryHandler) PinMemory(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Memory pinned", "id": id})
}

// UnpinMemory 取消固定
func (h *MemoryHandler) UnpinMemory(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Memory unpinned", "id": id})
}

// MemoryStats 记忆统计
type MemoryStats struct {
	TotalMemories   int                    `json:"total_memories"`
	ByType         map[string]int         `json:"by_type"`
	TotalAccesses  int64                  `json:"total_accesses"`
	AvgImportance  float64                `json:"avg_importance"`
	ForgetCandidates int                  `json:"forget_candidates"`
	RetentionRate  float64                `json:"retention_rate"`
	TotalSize      int64                  `json:"total_size_bytes"`
}

// GetMemoryStats 获取记忆统计
func (h *MemoryHandler) GetMemoryStats(c *gin.Context) {
	stats := MemoryStats{
		TotalMemories:   1000,
		ByType:         map[string]int{"episodic": 500, "semantic": 300, "procedural": 200},
		TotalAccesses:  10000,
		AvgImportance:  0.75,
		ForgetCandidates: 50,
		RetentionRate:  0.95,
		TotalSize:      1024000,
	}
	
	c.JSON(http.StatusOK, stats)
}

// TriggerForget 触发遗忘
func (h *MemoryHandler) TriggerForget(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":           "Forget process triggered",
		"memories_forgotten": 10,
		"memories_retained": 990,
	})
}
