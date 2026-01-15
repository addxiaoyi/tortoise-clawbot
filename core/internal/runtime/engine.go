package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Config 运行时配置
type Config struct {
	MaxWorkers     int
	QueueSize     int
	MemoryPoolSize int
}

// Task 任务定义
type Task struct {
	ID       string
	Type     string
	Priority int
	Handler  func() error
	Deadline time.Time
}

// Engine 高性能运行时引擎
type Engine struct {
	config     Config
	taskQueue chan *Task
	workerPool chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	stats      Stats
	ctx        context.Context
	cancel     context.CancelFunc
	running    atomic.Bool
}

// Stats 运行时统计
type Stats struct {
	TasksProcessed  atomic.Int64
	TasksPending   atomic.Int64
	WorkersActive  atomic.Int64
	QueueFull     atomic.Int64
	AvgLatencyUs  atomic.Int64
}

// NewEngine 创建运行时引擎
func NewEngine(cfg Config) *Engine {
	if cfg.MaxWorkers == 0 {
		cfg.MaxWorkers = 10000
	}
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 100000
	}

	ctx, cancel := context.WithCancel(context.Background())

	e := &Engine{
		config:     cfg,
		taskQueue:  make(chan *Task, cfg.QueueSize),
		workerPool: make(chan struct{}, cfg.MaxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
	e.running.Store(true)

	// 启动 worker pool
	e.startWorkers()

	return e
}

// startWorkers 启动 worker pool
func (e *Engine) startWorkers() {
	for i := 0; i < e.config.MaxWorkers; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
}

// worker 工作协程
func (e *Engine) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case task := <-e.taskQueue:
			if task == nil {
				continue
			}

			e.stats.WorkersActive.Add(1)
			start := time.Now()

			// 执行任务
			if task.Handler != nil {
				task.Handler()
			}

			// 更新统计
			e.stats.TasksProcessed.Add(1)
			e.stats.WorkersActive.Add(-1)
			latency := time.Since(start).Microseconds()
			e.stats.AvgLatencyUs.Store(latency)
		}
	}
}

// Submit 提交任务
func (e *Engine) Submit(task *Task) error {
	if !e.running.Load() {
		return ErrEngineStopped
	}

	select {
	case e.taskQueue <- task:
		e.stats.TasksPending.Add(1)
		return nil
	default:
		e.stats.QueueFull.Add(1)
		return ErrQueueFull
	}
}

// SubmitWithTimeout 带超时提交
func (e *Engine) SubmitWithTimeout(task *Task, timeout time.Duration) error {
	if !e.running.Load() {
		return ErrEngineStopped
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case e.taskQueue <- task:
		e.stats.TasksPending.Add(1)
		return nil
	case <-timer.C:
		e.stats.QueueFull.Add(1)
		return ErrQueueFull
	}
}

// Schedule 调度任务
func (e *Engine) Schedule(task *Task) {
	go func() {
		if task.Deadline.IsZero() {
			e.Submit(task)
		} else {
			delay := time.Until(task.Deadline)
			if delay > 0 {
				time.Sleep(delay)
			}
			e.Submit(task)
		}
	}()
}

// Stop 停止引擎
func (e *Engine) Stop() {
	e.running.Store(false)
	e.cancel()
	e.wg.Wait()
	close(e.taskQueue)
}

// Stats 获取统计信息
func (e *Engine) Stats() (s Stats) {
	s = Stats{
		TasksProcessed: Stats{TasksProcessed: atomic.Int64{}},
	}
	s.TasksProcessed.Store(e.stats.TasksProcessed.Load())
	s.TasksPending.Store(e.stats.TasksPending.Load())
	s.WorkersActive.Store(e.stats.WorkersActive.Load())
	s.QueueFull.Store(e.stats.QueueFull.Load())
	s.AvgLatencyUs.Store(e.stats.AvgLatencyUs.Load())
	return
}

// Errors
var (
	ErrEngineStopped = &RuntimeError{Code: "ENGINE_STOPPED", Message: "引擎已停止"}
	ErrQueueFull     = &RuntimeError{Code: "QUEUE_FULL", Message: "任务队列已满"}
)

// RuntimeError 运行时错误
type RuntimeError struct {
	Code    string
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Code + ": " + e.Message
}
