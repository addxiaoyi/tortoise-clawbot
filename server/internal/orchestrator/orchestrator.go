package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============ Multi-Agent Orchestrator ============

// Orchestrator 多代理编排器
type Orchestrator struct {
	config     *OrchestratorConfig
	agents     map[string]*Agent
	tasks     map[string]*Task
	workflows  map[string]*Workflow
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// OrchestratorConfig 编排器配置
type OrchestratorConfig struct {
	MaxConcurrentTasks int                // 最大并发任务数
	TaskTimeout       time.Duration     // 任务超时
	RetryAttempts     int               // 重试次数
	RetryDelay       time.Duration     // 重试延迟
	AgentPoolSize    int               // 代理池大小
	Strategy         OrchestrationStrategy // 编排策略
}

// OrchestrationStrategy 编排策略
type OrchestrationStrategy string

const (
	StrategySequential  OrchestrationStrategy = "sequential"  // 顺序执行
	StrategyParallel   OrchestrationStrategy = "parallel"   // 并行执行
	StrategyHierarchical OrchestrationStrategy = "hierarchical" // 层级编排
	StrategyDynamic    OrchestrationStrategy = "dynamic"    // 动态选择
)

// Agent 代理
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Role        AgentRole     `json:"role"`
	Model       string         `json:"model"`
	Capabilities []string     `json:"capabilities"`
	Skills      []string      `json:"skills"`
	Status      AgentStatus   `json:"status"`
	TaskCount   int           `json:"task_count"`
	CompletedTasks int        `json:"completed_tasks"`
	FailedTasks  int         `json:"failed_tasks"`
	AvgLatency  time.Duration `json:"avg_latency"`
	CreatedAt   time.Time     `json:"created_at"`
	LastActiveAt time.Time    `json:"last_active_at"`
	Config      AgentConfig   `json:"config"`
}

// AgentRole 代理角色
type AgentRole string

const (
	RoleOrchestrator AgentRole = "orchestrator" // 编排者
	RoleCoordinator  AgentRole = "coordinator"  // 协调者
	RoleSpecialist  AgentRole = "specialist"  // 专家
	RoleWorker      AgentRole = "worker"      // 工作者
	RoleReviewer    AgentRole = "reviewer"    // 审查者
)

// AgentStatus 代理状态
type AgentStatus string

const (
	StatusIdle       AgentStatus = "idle"
	StatusBusy      AgentStatus = "busy"
	StatusOffline   AgentStatus = "offline"
	StatusError     AgentStatus = "error"
)

// AgentConfig 代理配置
type AgentConfig struct {
	Temperature float64        `json:"temperature"`
	MaxTokens  int            `json:"max_tokens"`
	SystemPrompt string       `json:"system_prompt"`
	Tools      []string       `json:"tools"`
}

// Task 任务
type Task struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Type         TaskType      `json:"type"`
	Status       TaskStatus    `json:"status"`
	Priority     int           `json:"priority"`
	Input        interface{}   `json:"input"`
	Output       interface{}   `json:"output"`
	Error        string        `json:"error,omitempty"`
	AssignedAgent string       `json:"assigned_agent,omitempty"`
	ParentTaskID string       `json:"parent_task_id,omitempty"`
	SubTasks     []string     `json:"sub_tasks"`
	Dependencies []string     `json:"dependencies"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   time.Time     `json:"started_at,omitempty"`
	CompletedAt time.Time     `json:"completed_at,omitempty"`
	Timeout     time.Duration `json:"timeout"`
	RetryCount  int           `json:"retry_count"`
	Result      *TaskResult   `json:"result,omitempty"`
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSimple     TaskType = "simple"
	TaskTypeComplex    TaskType = "complex"
	TaskTypeParallel   TaskType = "parallel"
	TaskTypeSequential TaskType = "sequential"
	TaskTypeHierarchical TaskType = "hierarchical"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusComplete TaskStatus = "complete"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskResult 任务结果
type TaskResult struct {
	Success     bool        `json:"success"`
	Output     interface{} `json:"output"`
	Latency    time.Duration `json:"latency"`
	TokensUsed int         `json:"tokens_used"`
	Steps      []Step      `json:"steps"`
}

// Step 步骤
type Step struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	AgentID     string        `json:"agent_id"`
	Input       interface{}   `json:"input"`
	Output      interface{}   `json:"output"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Latency     time.Duration `json:"latency"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
}

// Workflow 工作流
type Workflow struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Status      WorkflowStatus `json:"status"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	TaskTemplate string  `json:"task_template"`
	AgentRole   AgentRole `json:"agent_role"`
	DependsOn   []string `json:"depends_on"`
	Retryable   bool     `json:"retryable"`
	Timeout     time.Duration `json:"timeout"`
}

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusPaused   WorkflowStatus = "paused"
	WorkflowStatusComplete WorkflowStatus = "complete"
	WorkflowStatusFailed  WorkflowStatus = "failed"
)

// NewOrchestrator 创建编排器
func NewOrchestrator(config *OrchestratorConfig) *Orchestrator {
	if config.MaxConcurrentTasks == 0 {
		config.MaxConcurrentTasks = 10
	}
	if config.TaskTimeout == 0 {
		config.TaskTimeout = 5 * time.Minute
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Orchestrator{
		config:    config,
		agents:   make(map[string]*Agent),
		tasks:    make(map[string]*Task),
		workflows: make(map[string]*Workflow),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterAgent 注册代理
func (o *Orchestrator) RegisterAgent(agent *Agent) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if _, exists := o.agents[agent.ID]; exists {
		return fmt.Errorf("代理已存在: %s", agent.ID)
	}
	
	agent.CreatedAt = time.Now()
	agent.Status = StatusIdle
	
	o.agents[agent.ID] = agent
	log.Printf("✅ 代理已注册: %s (%s)", agent.Name, agent.Role)
	
	return nil
}

// UnregisterAgent 注销代理
func (o *Orchestrator) UnregisterAgent(agentID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if _, exists := o.agents[agentID]; !exists {
		return fmt.Errorf("代理不存在: %s", agentID)
	}
	
	delete(o.agents, agentID)
	log.Printf("🛑 代理已注销: %s", agentID)
	
	return nil
}

// GetAgent 获取代理
func (o *Orchestrator) GetAgent(agentID string) *Agent {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.agents[agentID]
}

// ListAgents 列出代理
func (o *Orchestrator) ListAgents() []*Agent {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	agents := make([]*Agent, 0, len(o.agents))
	for _, a := range o.agents {
		agents = append(agents, a)
	}
	return agents
}

// CreateTask 创建任务
func (o *Orchestrator) CreateTask(task *Task) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	
	task.Status = TaskStatusPending
	task.CreatedAt = time.Now()
	
	o.tasks[task.ID] = task
	
	return nil
}

// SubmitTask 提交任务
func (o *Orchestrator) SubmitTask(task *Task) (*TaskResult, error) {
	if err := o.CreateTask(task); err != nil {
		return nil, err
	}
	
	return o.ExecuteTask(task.ID)
}

// ExecuteTask 执行任务
func (o *Orchestrator) ExecuteTask(taskID string) (*TaskResult, error) {
	o.mu.Lock()
	task, exists := o.tasks[taskID]
	if !exists {
		o.mu.Unlock()
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.Status = TaskStatusRunning
	task.StartedAt = time.Now()
	o.mu.Unlock()
	
	// 选择最佳代理
	agent, err := o.selectAgent(task)
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		return nil, err
	}
	
	// 执行任务
	result, err := o.runTaskWithAgent(task, agent)
	
	if err != nil {
		// 重试逻辑
		if task.RetryCount < o.config.RetryAttempts {
			task.RetryCount++
			time.Sleep(o.config.RetryDelay)
			return o.ExecuteTask(taskID)
		}
		
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		
		agent.FailedTasks++
		agent.Status = StatusError
	} else {
		task.Status = TaskStatusComplete
		task.CompletedAt = time.Now()
		task.Result = result
		
		agent.CompletedTasks++
		agent.Status = StatusIdle
	}
	
	agent.LastActiveAt = time.Now()
	
	return result, err
}

// selectAgent 选择最佳代理
func (o *Orchestrator) selectAgent(task *Task) (*Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	var candidates []*Agent
	
	// 根据任务类型和需求筛选
	for _, agent := range o.agents {
		if agent.Status == StatusOffline {
			continue
		}
		
		// 检查代理是否空闲
		if agent.Status == StatusBusy && o.config.AgentPoolSize > 0 {
			if agent.TaskCount >= o.config.AgentPoolSize {
				continue
			}
		}
		
		// 检查代理能力
		if o.matchCapabilities(agent, task) {
			candidates = append(candidates, agent)
		}
	}
	
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的代理")
	}
	
	// 根据策略选择
	switch o.config.Strategy {
	case StrategySequential:
		return candidates[0], nil
	case StrategyParallel:
		// 选择负载最低的
		minTasks := candidates[0]
		for _, a := range candidates[1:] {
			if a.TaskCount < minTasks.TaskCount {
				minTasks = a
			}
		}
		return minTasks, nil
	case StrategyDynamic:
		// 根据多个因素动态选择
		return o.dynamicSelect(candidates, task)
	default:
		return candidates[0], nil
	}
}

// matchCapabilities 匹配能力
func (o *Orchestrator) matchCapabilities(agent *Agent, task *Task) bool {
	// 简化实现 - 实际应更复杂
	return true
}

// dynamicSelect 动态选择
func (o *Orchestrator) dynamicSelect(candidates []*Agent, task *Task) (*Agent, error) {
	best := candidates[0]
	bestScore := o.scoreAgent(best, task)
	
	for _, agent := range candidates[1:] {
		score := o.scoreAgent(agent, task)
		if score > bestScore {
			best = agent
			bestScore = score
		}
	}
	
	return best, nil
}

// scoreAgent 给代理评分
func (o *Orchestrator) scoreAgent(agent *Agent, task *Task) float64 {
	var score float64 = 100
	
	// 任务完成率
	completed := float64(agent.CompletedTasks)
	failed := float64(agent.FailedTasks)
	total := completed + failed
	if total > 0 {
		score *= completed / total
	}
	
	// 负载
	score -= float64(agent.TaskCount) * 5
	
	// 平均延迟
	if agent.AvgLatency > 0 {
		score -= float64(agent.AvgLatency) / float64(time.Minute) * 10
	}
	
	return math.Max(0, score)
}

// runTaskWithAgent 使用代理运行任务
func (o *Orchestrator) runTaskWithAgent(task *Task, agent *Agent) (*TaskResult, error) {
	agent.TaskCount++
	agent.Status = StatusBusy
	
	startTime := time.Now()
	
	step := &Step{
		ID:        uuid.New().String(),
		Name:      task.Name,
		AgentID:   agent.ID,
		StartTime: startTime,
	}
	
	// 执行任务
	output, err := o.executeTaskLogic(task, agent)
	
	step.EndTime = time.Now()
	step.Latency = step.EndTime.Sub(startTime)
	step.Output = output
	
	if err != nil {
		step.Success = false
		step.Error = err.Error()
	} else {
		step.Success = true
	}
	
	agent.AvgLatency = (agent.AvgLatency + step.Latency) / 2
	agent.Status = StatusIdle
	
	return &TaskResult{
		Success:  err == nil,
		Output:   output,
		Latency:  step.Latency,
		Steps:    []*Step{step},
	}, err
}

// executeTaskLogic 执行任务逻辑
func (o *Orchestrator) executeTaskLogic(task *Task, agent *Agent) (interface{}, error) {
	// 模拟任务执行
	ctx, cancel := context.WithTimeout(o.ctx, task.Timeout)
	defer cancel()
	
	// 根据任务类型执行
	switch task.Type {
	case TaskTypeSimple:
		return o.executeSimpleTask(task, agent)
	case TaskTypeParallel:
		return o.executeParallelTasks(task, agent)
	case TaskTypeSequential:
		return o.executeSequentialTasks(task, agent)
	case TaskTypeHierarchical:
		return o.executeHierarchicalTask(task, agent)
	default:
		return o.executeSimpleTask(task, agent)
	}
}

// executeSimpleTask 执行简单任务
func (o *Orchestrator) executeSimpleTask(task *Task, agent *Agent) (interface{}, error) {
	// 模拟处理
	time.Sleep(100 * time.Millisecond)
	
	return map[string]interface{}{
		"agent": agent.Name,
		"task":  task.Name,
		"result": "completed",
	}, nil
}

// executeParallelTasks 执行并行任务
func (o *Orchestrator) executeParallelTasks(task *Task, agent *Agent) (interface{}, error) {
	o.mu.Lock()
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]interface{}, 0)
	errors := make([]error, 0)
	
	for _, subTaskID := range task.SubTasks {
		subTask, ok := o.tasks[subTaskID]
		if !ok {
			continue
		}
		
		wg.Add(1)
		go func(t *Task) {
			defer wg.Done()
			
			subAgent, err := o.selectAgent(t)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
			
			result, err := o.runTaskWithAgent(t, subAgent)
			mu.Lock()
			if err == nil {
				results = append(results, result.Output)
			} else {
				errors = append(errors, err)
			}
			mu.Unlock()
		}(subTask)
	}
	
	o.mu.Unlock()
	wg.Wait()
	
	if len(errors) > 0 {
		return results, errors[0]
	}
	
	return results, nil
}

// executeSequentialTasks 执行顺序任务
func (o *Orchestrator) executeSequentialTasks(task *Task, agent *Agent) (interface{}, error) {
	results := make([]interface{}, 0)
	
	for _, subTaskID := range task.SubTasks {
		subTask, ok := o.tasks[subTaskID]
		if !ok {
			continue
		}
		
		// 检查依赖
		if !o.checkDependencies(subTask) {
			return nil, fmt.Errorf("依赖未满足: %s", subTaskID)
		}
		
		subAgent, err := o.selectAgent(subTask)
		if err != nil {
			return nil, err
		}
		
		result, err := o.runTaskWithAgent(subTask, subAgent)
		if err != nil {
			return nil, err
		}
		
		results = append(results, result.Output)
	}
	
	return results, nil
}

// executeHierarchicalTask 执行层级任务
func (o *Orchestrator) executeHierarchicalTask(task *Task, agent *Agent) (interface{}, error) {
	// 层级任务 - 编排者分解任务，然后分配给下级
	coordAgent, err := o.selectAgent(&Task{Type: TaskTypeSimple})
	if err != nil {
		return nil, err
	}
	
	// 协调者分解任务
	subTasks := o.decomposeTask(task, coordAgent)
	
	// 并行执行子任务
	return o.executeParallelTasks(&Task{SubTasks: subTasks}, agent)
}

// decomposeTask 分解任务
func (o *Orchestrator) decomposeTask(task *Task, agent *Agent) []string {
	subTaskIDs := make([]string, 0)
	
	// 模拟任务分解
	for i := 0; i < 3; i++ {
		subTask := &Task{
			ID:          uuid.New().String(),
			Name:        fmt.Sprintf("%s-part-%d", task.Name, i+1),
			Description: task.Description,
			Type:        TaskTypeSimple,
			Priority:    task.Priority,
			ParentTaskID: task.ID,
		}
		
		o.CreateTask(subTask)
		subTaskIDs = append(subTaskIDs, subTask.ID)
	}
	
	return subTaskIDs
}

// checkDependencies 检查依赖
func (o *Orchestrator) checkDependencies(task *Task) bool {
	for _, depID := range task.Dependencies {
		depTask, ok := o.tasks[depID]
		if !ok {
			continue
		}
		if depTask.Status != TaskStatusComplete {
			return false
		}
	}
	return true
}

// CreateWorkflow 创建工作流
func (o *Orchestrator) CreateWorkflow(workflow *Workflow) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}
	
	workflow.CreatedAt = time.Now()
	workflow.Status = WorkflowStatusDraft
	
	o.workflows[workflow.ID] = workflow
	
	return nil
}

// ExecuteWorkflow 执行工作流
func (o *Orchestrator) ExecuteWorkflow(workflowID string) (*WorkflowResult, error) {
	o.mu.Lock()
	workflow, exists := o.workflows[workflowID]
	if !exists {
		o.mu.Unlock()
		return nil, fmt.Errorf("工作流不存在: %s", workflowID)
	}
	
	workflow.Status = WorkflowStatusActive
	o.mu.Unlock()
	
	result := &WorkflowResult{
		WorkflowID: workflowID,
		StartTime:  time.Now(),
		Steps:      make([]*WorkflowStepResult, 0),
	}
	
	// 执行步骤
	for _, step := range workflow.Steps {
		if !o.checkStepDependencies(step, result.Steps) {
			result.Success = false
			result.Error = fmt.Sprintf("步骤 %s 依赖未满足", step.Name)
			break
		}
		
		stepResult, err := o.executeWorkflowStep(workflow, step)
		result.Steps = append(result.Steps, stepResult)
		
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			if !step.Retryable {
				break
			}
		}
	}
	
	result.EndTime = time.Now()
	result.Latency = result.EndTime.Sub(result.StartTime)
	
	if result.Error == "" {
		workflow.Status = WorkflowStatusComplete
		result.Success = true
	} else {
		workflow.Status = WorkflowStatusFailed
	}
	
	workflow.UpdatedAt = time.Now()
	
	return result, nil
}

// WorkflowResult 工作流结果
type WorkflowResult struct {
	WorkflowID string              `json:"workflow_id"`
	Success    bool               `json:"success"`
	Error      string             `json:"error,omitempty"`
	Steps      []*WorkflowStepResult `json:"steps"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Latency   time.Duration      `json:"latency"`
}

// WorkflowStepResult 工作流步骤结果
type WorkflowStepResult struct {
	StepID    string        `json:"step_id"`
	TaskID    string        `json:"task_id"`
	Success   bool          `json:"success"`
	Output    interface{}   `json:"output,omitempty"`
	Latency   time.Duration `json:"latency"`
}

// checkStepDependencies 检查步骤依赖
func (o *Orchestrator) checkStepDependencies(step *WorkflowStep, completedSteps []*WorkflowStepResult) bool {
	for _, depID := range step.DependsOn {
		found := false
		for _, s := range completedSteps {
			if s.StepID == depID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// executeWorkflowStep 执行工作流步骤
func (o *Orchestrator) executeWorkflowStep(workflow *Workflow, step *WorkflowStep) (*WorkflowStepResult, error) {
	startTime := time.Now()
	
	// 创建任务
	task := &Task{
		ID:          uuid.New().String(),
		Name:        step.Name,
		Description: step.TaskTemplate,
		Type:        TaskTypeSimple,
		Timeout:     step.Timeout,
	}
	
	o.CreateTask(task)
	
	// 选择代理
	agent, err := o.selectAgentByRole(step.AgentRole)
	if err != nil {
		return &WorkflowStepResult{
			StepID:  step.ID,
			TaskID:  task.ID,
			Success: false,
		}, err
	}
	
	// 执行
	result, err := o.runTaskWithAgent(task, agent)
	
	return &WorkflowStepResult{
		StepID:  step.ID,
		TaskID:  task.ID,
		Success: err == nil,
		Output:  result.Output,
		Latency: time.Since(startTime),
	}, err
}

// selectAgentByRole 根据角色选择代理
func (o *Orchestrator) selectAgentByRole(role AgentRole) (*Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	for _, agent := range o.agents {
		if agent.Role == role && agent.Status != StatusOffline {
			return agent, nil
		}
	}
	
	// 如果没有特定角色，返回任何可用代理
	for _, agent := range o.agents {
		if agent.Status != StatusOffline {
			return agent, nil
		}
	}
	
	return nil, fmt.Errorf("没有可用的 %s 代理", role)
}

// GetTask 获取任务
func (o *Orchestrator) GetTask(taskID string) *Task {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.tasks[taskID]
}

// ListTasks 列出任务
func (o *Orchestrator) ListTasks() []*Task {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(o.tasks))
	for _, t := range o.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetWorkflow 获取工作流
func (o *Orchestrator) GetWorkflow(workflowID string) *Workflow {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.workflows[workflowID]
}

// ListWorkflows 列出工作流
func (o *Orchestrator) ListWorkflows() []*Workflow {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	workflows := make([]*Workflow, 0, len(o.workflows))
	for _, w := range o.workflows {
		workflows = append(workflows, w)
	}
	return workflows
}

// Stop 停止编排器
func (o *Orchestrator) Stop() {
	o.cancel()
	log.Printf("🛑 编排器已停止")
}

// GetStats 获取统计信息
type OrchestratorStats struct {
	TotalAgents    int           `json:"total_agents"`
	BusyAgents     int           `json:"busy_agents"`
	IdleAgents     int           `json:"idle_agents"`
	TotalTasks     int           `json:"total_tasks"`
	PendingTasks   int           `json:"pending_tasks"`
	RunningTasks   int           `json:"running_tasks"`
	CompletedTasks int           `json:"completed_tasks"`
	FailedTasks    int           `json:"failed_tasks"`
	TotalWorkflows int           `json:"total_workflows"`
}

func (o *Orchestrator) GetStats() *OrchestratorStats {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	stats := &OrchestratorStats{
		TotalAgents:  len(o.agents),
		TotalTasks:   len(o.tasks),
		TotalWorkflows: len(o.workflows),
	}
	
	for _, agent := range o.agents {
		switch agent.Status {
		case StatusBusy:
			stats.BusyAgents++
		case StatusIdle:
			stats.IdleAgents++
		}
	}
	
	for _, task := range o.tasks {
		switch task.Status {
		case TaskStatusPending:
			stats.PendingTasks++
		case TaskStatusRunning:
			stats.RunningTasks++
		case TaskStatusComplete:
			stats.CompletedTasks++
		case TaskStatusFailed:
			stats.FailedTasks++
		}
	}
	
	return stats
}

// MarshalJSON 自定义 JSON 序列化
func (o *Orchestrator) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.GetStats())
}
