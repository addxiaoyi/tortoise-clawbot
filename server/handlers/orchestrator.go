package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============ Orchestrator API Handlers ============

// OrchestratorHandler 多代理编排器处理器
type OrchestratorHandler struct {
	// orchestrator *orchestrator.Orchestrator
}

// NewOrchestratorHandler 创建编排器处理器
func NewOrchestratorHandler() *OrchestratorHandler {
	return &OrchestratorHandler{}
}

// RegisterRoutes 注册路由
func (h *OrchestratorHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 代理管理
	r.GET("/agents", h.ListAgents)
	r.POST("/agents", h.RegisterAgent)
	r.GET("/agents/:id", h.GetAgent)
	r.DELETE("/agents/:id", h.UnregisterAgent)
	
	// 任务管理
	r.GET("/tasks", h.ListTasks)
	r.POST("/tasks", h.SubmitTask)
	r.GET("/tasks/:id", h.GetTask)
	r.POST("/tasks/:id/cancel", h.CancelTask)
	
	// 工作流
	r.GET("/workflows", h.ListWorkflows)
	r.POST("/workflows", h.CreateWorkflow)
	r.GET("/workflows/:id", h.GetWorkflow)
	r.POST("/workflows/:id/execute", h.ExecuteWorkflow)
	r.DELETE("/workflows/:id", h.DeleteWorkflow)
	
	// 统计
	r.GET("/stats", h.GetStats)
}

// ============ Agent Handlers ============

// AgentRequest 注册代理请求
type AgentRequest struct {
	Name        string   `json:"name" binding:"required"`
	Role        string   `json:"role" binding:"required"`
	Model       string   `json:"model"`
	Capabilities []string `json:"capabilities"`
	Skills      []string `json:"skills"`
}

// AgentResponse 代理响应
type AgentResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Role           string   `json:"role"`
	Model          string   `json:"model"`
	Capabilities   []string `json:"capabilities"`
	Skills        []string `json:"skills"`
	Status        string   `json:"status"`
	TaskCount     int      `json:"task_count"`
	CreatedAt     string   `json:"created_at"`
	LastActiveAt  string   `json:"last_active_at"`
}

// ListAgents 列出代理
func (h *OrchestratorHandler) ListAgents(c *gin.Context) {
	// TODO: 从 orchestrator 获取
	agents := []AgentResponse{}
	
	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"total":  len(agents),
	})
}

// RegisterAgent 注册代理
func (h *OrchestratorHandler) RegisterAgent(c *gin.Context) {
	var req AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	agent := AgentResponse{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Role:         req.Role,
		Model:        req.Model,
		Capabilities: req.Capabilities,
		Skills:       req.Skills,
		Status:       "idle",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusCreated, agent)
}

// GetAgent 获取代理
func (h *OrchestratorHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	
	agent := AgentResponse{
		ID:           id,
		Name:         "agent-1",
		Role:         "specialist",
		Status:       "idle",
		Capabilities: []string{"search", "code"},
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, agent)
}

// UnregisterAgent 注销代理
func (h *OrchestratorHandler) UnregisterAgent(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered", "id": id})
}

// ============ Task Handlers ============

// TaskRequest 任务请求
type TaskRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Priority    int                    `json:"priority"`
	Input       map[string]interface{} `json:"input"`
	AgentID     string                 `json:"agent_id"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Priority    int                    `json:"priority"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	AgentID     string                 `json:"agent_id,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	StartedAt   string                 `json:"started_at,omitempty"`
	CompletedAt string                 `json:"completed_at,omitempty"`
}

// ListTasks 列出任务
func (h *OrchestratorHandler) ListTasks(c *gin.Context) {
	status := c.Query("status")
	
	tasks := []TaskResponse{}
	
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
		"filter": gin.H{"status": status},
	})
}

// SubmitTask 提交任务
func (h *OrchestratorHandler) SubmitTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	task := TaskResponse{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      "pending",
		Priority:    req.Priority,
		Input:       req.Input,
		AgentID:     req.AgentID,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusCreated, task)
}

// GetTask 获取任务
func (h *OrchestratorHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	
	task := TaskResponse{
		ID:        id,
		Name:      "Sample Task",
		Type:      "simple",
		Status:    "completed",
		CreatedAt: time.Now().Format(time.RFC3339),
		CompletedAt: time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, task)
}

// CancelTask 取消任务
func (h *OrchestratorHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled", "id": id})
}

// ============ Workflow Handlers ============

// WorkflowRequest 工作流请求
type WorkflowRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Steps       []WorkflowStepRequest  `json:"steps" binding:"required"`
}

// WorkflowStepRequest 工作流步骤请求
type WorkflowStepRequest struct {
	ID          string   `json:"id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	AgentRole   string   `json:"agent_role" binding:"required"`
	DependsOn   []string `json:"depends_on"`
	Input       string   `json:"input"`
	OutputKey   string   `json:"output_key"`
	RetryPolicy string   `json:"retry_policy"`
}

// WorkflowResponse 工作流响应
type WorkflowResponse struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Steps       []WorkflowStepResponse `json:"steps"`
	Status      string                  `json:"status"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

// WorkflowStepResponse 工作流步骤响应
type WorkflowStepResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AgentRole   string   `json:"agent_role"`
	Status      string   `json:"status"`
	DependsOn   []string `json:"depends_on"`
	StartedAt   string   `json:"started_at,omitempty"`
	CompletedAt string   `json:"completed_at,omitempty"`
	Output      string   `json:"output,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// ListWorkflows 列出工作流
func (h *OrchestratorHandler) ListWorkflows(c *gin.Context) {
	workflows := []WorkflowResponse{}
	
	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"total": len(workflows),
	})
}

// CreateWorkflow 创建工作流
func (h *OrchestratorHandler) CreateWorkflow(c *gin.Context) {
	var req WorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	steps := make([]WorkflowStepResponse, 0, len(req.Steps))
	for _, s := range req.Steps {
		steps = append(steps, WorkflowStepResponse{
			ID:        s.ID,
			Name:      s.Name,
			AgentRole: s.AgentRole,
			Status:    "pending",
			DependsOn: s.DependsOn,
		})
	}
	
	workflow := WorkflowResponse{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Steps:       steps,
		Status:      "draft",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusCreated, workflow)
}

// GetWorkflow 获取工作流
func (h *OrchestratorHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")
	
	workflow := WorkflowResponse{
		ID:     id,
		Name:   "Sample Workflow",
		Status: "draft",
		Steps: []WorkflowStepResponse{
			{ID: "1", Name: "Step 1", Status: "pending"},
			{ID: "2", Name: "Step 2", Status: "pending", DependsOn: []string{"1"}},
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, workflow)
}

// ExecuteWorkflow 执行工作流
func (h *OrchestratorHandler) ExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Workflow execution started",
		"workflow_id": id,
		"execution_id": uuid.New().String(),
	})
}

// DeleteWorkflow 删除工作流
func (h *OrchestratorHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted", "id": id})
}

// ============ Stats Handler ============

// StatsResponse 统计响应
type StatsResponse struct {
	TotalAgents    int `json:"total_agents"`
	BusyAgents     int `json:"busy_agents"`
	IdleAgents     int `json:"idle_agents"`
	TotalTasks     int `json:"total_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	RunningTasks   int `json:"running_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
	TotalWorkflows int `json:"total_workflows"`
}

// GetStats 获取统计
func (h *OrchestratorHandler) GetStats(c *gin.Context) {
	stats := StatsResponse{
		TotalAgents:    5,
		BusyAgents:     2,
		IdleAgents:     3,
		TotalTasks:     100,
		PendingTasks:   10,
		RunningTasks:   5,
		CompletedTasks: 80,
		FailedTasks:    5,
		TotalWorkflows: 10,
	}
	
	c.JSON(http.StatusOK, stats)
}
