package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ============ 定时任务调度器 ============

// Task 任务接口
type Task interface {
	// Run 执行任务
	Run(ctx context.Context) error

	// GetName 获取任务名称
	GetName() string

	// GetSchedule 获取调度表达式
	GetSchedule() string
}

// TaskFunc 函数类型任务
type TaskFunc struct {
	Name     string
	Schedule string
	Fn       func(ctx context.Context) error
}

func (t *TaskFunc) Run(ctx context.Context) error {
	return t.Fn(ctx)
}

func (t *TaskFunc) GetName() string {
	return t.Name
}

func (t *TaskFunc) GetSchedule() string {
	return t.Schedule
}

// Manager 调度器管理器
type Manager struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager 创建调度器管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		tasks: make(map[string]*Task),
		ctx:   ctx,
		cancel: cancel,
	}
}

// Start 启动调度器
func (m *Manager) Start() error {
	log.Printf("✅ 任务调度器已启动")
	return nil
}

// Stop 停止调度器
func (m *Manager) Stop() {
	m.cancel()
	log.Printf("🛑 任务调度器已停止")
}

// Add 添加任务
func (m *Manager) Add(task Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := task.GetName()
	if _, exists := m.tasks[name]; exists {
		return fmt.Errorf("task already exists: %s", name)
	}

	m.tasks[name] = &task
	log.Printf("✅ 定时任务已添加: %s (调度: %s)", name, task.GetSchedule())
	return nil
}

// Remove 移除任务
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[name]; !exists {
		return fmt.Errorf("task not found: %s", name)
	}

	delete(m.tasks, name)
	log.Printf("🗑️ 定时任务已移除: %s", name)
	return nil
}

// List 列出所有任务
func (m *Manager) List() []TaskInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]TaskInfo, 0, len(m.tasks))
	for name, task := range m.tasks {
		infos = append(infos, TaskInfo{
			Name:     name,
			Schedule: task.GetSchedule(),
		})
	}
	return infos
}

// RunNow 立即执行任务
func (m *Manager) RunNow(name string) error {
	m.mu.RLock()
	task, exists := m.tasks[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task not found: %s", name)
	}

	go func() {
		taskCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		log.Printf("📋 立即执行定时任务: %s", name)
		start := time.Now()

		if err := task.Run(taskCtx); err != nil {
			log.Printf("❌ 定时任务执行失败: %s - %v", name, err)
		} else {
			log.Printf("✅ 定时任务完成: %s (耗时: %v)", name, time.Since(start))
		}
	}()

	return nil
}

// TaskInfo 任务信息
type TaskInfo struct {
	Name     string
	Schedule string
}

// CleanupMemoryTask 清理过期记忆任务
type CleanupMemoryTask struct {
	MaxAge time.Duration
}

func (t *CleanupMemoryTask) Run(ctx context.Context) error {
	log.Printf("🧹 开始清理过期记忆...")
	// 实现清理逻辑
	return nil
}

func (t *CleanupMemoryTask) GetName() string {
	return "cleanup-memory"
}

func (t *CleanupMemoryTask) GetSchedule() string {
	return "0 3 * * *" // 每天凌晨 3 点
}

// BackupTask 数据备份任务
type BackupTask struct {
	DataDir string
}

func (t *BackupTask) Run(ctx context.Context) error {
	log.Printf("💾 开始数据备份...")
	// 实现备份逻辑
	return nil
}

func (t *BackupTask) GetName() string {
	return "backup-data"
}

func (t *BackupTask) GetSchedule() string {
	return "0 2 * * *" // 每天凌晨 2 点
}

// HealthCheckTask 健康检查任务
type HealthCheckTask struct{}

func (t *HealthCheckTask) Run(ctx context.Context) error {
	log.Printf("🏥 执行健康检查...")
	// 实现健康检查逻辑
	return nil
}

func (t *HealthCheckTask) GetName() string {
	return "health-check"
}

func (t *HealthCheckTask) GetSchedule() string {
	return "*/5 * * * *" // 每 5 分钟
}

// RegisterBuiltins 注册内置任务
func (m *Manager) RegisterBuiltins() {
	m.Add(&HealthCheckTask{})
	m.Add(&BackupTask{DataDir: "./data"})
	m.Add(&CleanupMemoryTask{MaxAge: 30 * 24 * time.Hour})

	log.Printf("📦 已注册 %d 个内置定时任务", len(m.tasks))
}
