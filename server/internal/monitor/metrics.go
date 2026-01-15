package monitor

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 指标收集器
type Metrics struct {
	// 请求计数
	requestCount    uint64
	requestSuccess uint64
	requestError   uint64

	// 响应时间 (毫秒)
	responseTimes []int64
	responseMu     sync.Mutex

	// AI 请求
	aiRequests   uint64
	aiTokens     uint64
	aiLatency    int64

	// 内存
	memoryUsed    uint64
	memoryAlloc    uint64

	// WebSocket 连接
	wsConnections uint64

	// 启动时间
	startTime time.Time

	// 系统信息
	systemInfo SystemInfo
}

// SystemInfo 系统信息
type SystemInfo struct {
	NumCPU     int    `json:"num_cpu"`
	GoVersion  string `json:"go_version"`
	NumGoroutine int `json:"num_goroutine"`
	NumCgoCall   int64 `json:"num_cgo_call"`
}

// 全局指标实例
var globalMetrics *Metrics
var metricsOnce sync.Once

// GetMetrics 获取全局指标
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			startTime: time.Now(),
		}
	})
	return globalMetrics
}

// ============ 指标收集方法 ============

// RecordRequest 记录请求
func (m *Metrics) RecordRequest(success bool, latencyMs int64) {
	atomic.AddUint64(&m.requestCount, 1)
	if success {
		atomic.AddUint64(&m.requestSuccess, 1)
	} else {
		atomic.AddUint64(&m.requestError, 1)
	}
	m.recordResponseTime(latencyMs)
}

// recordResponseTime 记录响应时间
func (m *Metrics) recordResponseTime(ms int64) {
	m.responseMu.Lock()
	defer m.responseMu.Unlock()

	// 保留最近 1000 条记录
	if len(m.responseTimes) >= 1000 {
		m.responseTimes = m.responseTimes[1:]
	}
	m.responseTimes = append(m.responseTimes, ms)
}

// RecordAIRequest 记录 AI 请求
func (m *Metrics) RecordAIRequest(tokens int, latencyMs int64) {
	atomic.AddUint64(&m.aiRequests, 1)
	atomic.AddUint64(&m.aiTokens, uint64(tokens))
	
	// 更新平均延迟
	oldLatency := atomic.LoadInt64(&m.aiLatency)
	newLatency := (oldLatency + latencyMs) / 2
	atomic.StoreInt64(&m.aiLatency, newLatency)
}

// RecordWSConnection 记录 WebSocket 连接
func (m *Metrics) RecordWSConnection(connected bool) {
	if connected {
		atomic.AddUint64(&m.wsConnections, 1)
	} else {
		val := atomic.LoadUint64(&m.wsConnections)
		if val > 0 {
			atomic.StoreUint64(&m.wsConnections, val-1)
		}
	}
}

// UpdateMemory 更新内存信息
func (m *Metrics) UpdateMemory() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	atomic.StoreUint64(&m.memoryUsed, memStats.Sys)
	atomic.StoreUint64(&m.memoryAlloc, memStats.Alloc)

	// 更新系统信息
	m.systemInfo.NumCPU = runtime.NumCPU()
	m.systemInfo.NumGoroutine = runtime.NumGoroutine()
	m.systemInfo.NumCgoCall = runtime.NumCgoCall()
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() MetricsStats {
	// 更新内存
	m.UpdateMemory()

	return MetricsStats{
		Requests: RequestStats{
			Total:    atomic.LoadUint64(&m.requestCount),
			Success:  atomic.LoadUint64(&m.requestSuccess),
			Error:    atomic.LoadUint64(&m.requestError),
			QPS:      m.calculateQPS(),
		},
		Response: ResponseStats{
			P50:    m.percentile(50),
			P90:    m.percentile(90),
			P99:    m.percentile(99),
			Avg:    m.averageResponseTime(),
		},
		AI: AIStats{
			Requests: atomic.LoadUint64(&m.aiRequests),
			Tokens:   atomic.LoadUint64(&m.aiTokens),
			AvgLatency: atomic.LoadInt64(&m.aiLatency),
		},
		Memory: MemoryStats{
			Used:  atomic.LoadUint64(&m.memoryUsed),
			Alloc: atomic.LoadUint64(&m.memoryAlloc),
		},
		WebSocket: WSStats{
			Connections: atomic.LoadUint64(&m.wsConnections),
		},
		Uptime: time.Since(m.startTime).String(),
		System: m.systemInfo,
	}
}

// ============ 统计结构 ============

// MetricsStats 完整统计
type MetricsStats struct {
	Requests  RequestStats  `json:"requests"`
	Response ResponseStats `json:"response"`
	AI       AIStats       `json:"ai"`
	Memory   MemoryStats   `json:"memory"`
	WebSocket WSStats     `json:"websocket"`
	Uptime   string        `json:"uptime"`
	System   SystemInfo    `json:"system"`
}

// RequestStats 请求统计
type RequestStats struct {
	Total   uint64   `json:"total"`
	Success uint64   `json:"success"`
	Error   uint64   `json:"error"`
	QPS     float64  `json:"qps"`
}

// ResponseStats 响应时间统计
type ResponseStats struct {
	P50  int64 `json:"p50"`
	P90  int64 `json:"p90"`
	P99  int64 `json:"p99"`
	Avg  int64 `json:"avg"`
}

// AIStats AI 统计
type AIStats struct {
	Requests   uint64 `json:"requests"`
	Tokens     uint64 `json:"tokens"`
	AvgLatency int64  `json:"avg_latency_ms"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Used  uint64 `json:"used_bytes"`
	Alloc uint64 `json:"alloc_bytes"`
}

// WSStats WebSocket 统计
type WSStats struct {
	Connections uint64 `json:"connections"`
}

// ============ 辅助方法 ============

// calculateQPS 计算 QPS
func (m *Metrics) calculateQPS() float64 {
	m.responseMu.Lock()
	defer m.responseMu.Unlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	// 计算最近 1 分钟的 QPS
	now := time.Now()
	var recent int
	for _, t := range m.responseTimes {
		if now.Sub(time.UnixMilli(t)).Seconds() < 60 {
			recent++
		}
	}

	return float64(recent) / 60.0
}

// percentile 计算百分位数
func (m *Metrics) percentile(p int) int64 {
	m.responseMu.Lock()
	defer m.responseMu.Unlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	// 复制并排序
	times := make([]int64, len(m.responseTimes))
	copy(times, m.responseTimes)

	// 简单排序
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[j] < times[i] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	idx := (p * len(times)) / 100
	if idx >= len(times) {
		idx = len(times) - 1
	}
	return times[idx]
}

// averageResponseTime 计算平均响应时间
func (m *Metrics) averageResponseTime() int64 {
	m.responseMu.Lock()
	defer m.responseMu.Unlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	var sum int64
	for _, t := range m.responseTimes {
		sum += t
	}
	return sum / int64(len(m.responseTimes))
}

// ============ 中间件 ============

// MetricsMiddleware 指标中间件
func MetricsMiddleware() func(next func()) {
	return func(next func()) {
		start := time.Now()
		metrics := GetMetrics()
		
		metrics.RecordRequest(true, time.Since(start).Milliseconds())
		next()
	}
}
