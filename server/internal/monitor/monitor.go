package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// ============ Monitor 监控系统 ============

// Monitor 监控系统
type Monitor struct {
	mu           sync.RWMutex
	enabled      bool
	tracerProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
	httpServer   *http.Server
	
	// 指标
	metrics      *Metrics
	
	// OpenTelemetry 配置
	otlpEndpoint string
}

// Metrics 指标收集器
type Metrics struct {
	// 请求计数
	RequestCount   uint64
	RequestSuccess uint64
	RequestError   uint64
	
	// AI 调用
	AIRequestCount   uint64
	AIRequestTokens  uint64
	AIRequestLatency uint64 // ms
	
	// 渠道消息
	ChannelMessages uint64
	
	// 活跃会话
	ActiveSessions int64
	
	// 内存使用
	MemoryUsage    uint64
	GoRoutines     int
}

// NewMonitor 创建监控系统
func NewMonitor() *Monitor {
	return &Monitor{
		enabled: true,
		metrics: &Metrics{},
	}
}

// Start 启动监控
func (m *Monitor) Start(port int, otlpEndpoint string) error {
	if !m.enabled {
		return nil
	}

	m.otlpEndpoint = otlpEndpoint

	// 初始化 OpenTelemetry
	if otlpEndpoint != "" {
		if err := m.initTelemetry(); err != nil {
			log.Printf("⚠️ OpenTelemetry 初始化失败: %v", err)
		}
	}

	// 启动指标 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.handleMetrics)
	mux.HandleFunc("/health", m.handleHealth)

	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("📊 监控服务器已启动: http://localhost:%d/metrics", port)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("监控服务器错误: %v", err)
		}
	}()

	// 启动指标收集循环
	go m.collectLoop()

	return nil
}

// Stop 停止监控
func (m *Monitor) Stop() error {
	if m.tracerProvider != nil {
		if err := m.tracerProvider.Shutdown(context.Background()); err != nil {
			return err
		}
	}
	if m.meterProvider != nil {
		if err := m.meterProvider.Shutdown(context.Background()); err != nil {
			return err
		}
	}
	if m.httpServer != nil {
		return m.httpServer.Close()
	}
	return nil
}

// initTelemetry 初始化 OpenTelemetry
func (m *Monitor) initTelemetry() error {
	ctx := context.Background()

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("tortoise"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("创建资源失败: %w", err)
	}

	// 初始化追踪导出器
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(m.otlpEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("创建追踪导出器失败: %w", err)
	}

	// 初始化追踪提供者
	m.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(m.tracerProvider)

	// 初始化指标导出器
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(m.otlpEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("创建指标导出器失败: %w", err)
	}

	// 初始化指标提供者
	m.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(m.meterProvider)

	log.Printf("✅ OpenTelemetry 已初始化 (endpoint: %s)", m.otlpEndpoint)
	return nil
}

// handleMetrics 处理指标请求
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	metrics := m.metrics
	m.mu.RUnlock()

	// 更新运行时指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.MemoryUsage = memStats.Alloc
	metrics.GoRoutines = runtime.NumGoroutine()

	// Prometheus 格式输出
	output := fmt.Sprintf(`# HELP tortoise_requests_total Total number of requests
# TYPE tortoise_requests_total counter
tortoise_requests_total %d

# HELP tortoise_requests_success_total Total number of successful requests
# TYPE tortoise_requests_success_total counter
tortoise_requests_success_total %d

# HELP tortoise_requests_error_total Total number of failed requests
# TYPE tortoise_requests_error_total counter
tortoise_requests_error_total %d

# HELP tortoise_ai_requests_total Total number of AI requests
# TYPE tortoise_ai_requests_total counter
tortoise_ai_requests_total %d

# HELP tortoise_ai_tokens_total Total number of AI tokens used
# TYPE tortoise_ai_tokens_total counter
tortoise_ai_tokens_total %d

# HELP tortoise_active_sessions Current number of active sessions
# TYPE tortoise_active_sessions gauge
tortoise_active_sessions %d

# HELP tortoise_memory_usage_bytes Current memory usage
# TYPE tortoise_memory_usage_bytes gauge
tortoise_memory_usage_bytes %d

# HELP tortoise_goroutines Current number of goroutines
# TYPE tortoise_goroutines gauge
tortoise_goroutines %d
`,
		metrics.RequestCount,
		metrics.RequestSuccess,
		metrics.RequestError,
		metrics.AIRequestCount,
		metrics.AIRequestTokens,
		metrics.ActiveSessions,
		metrics.MemoryUsage,
		metrics.GoRoutines,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(output))
}

// handleHealth 健康检查
func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().Format(time.RFC3339),
		"uptime":     time.Since(startTime).String(),
	}
	json.NewEncoder(w).Encode(status)
}

// collectLoop 定期收集指标
func (m *Monitor) collectLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		m.mu.Lock()
		m.metrics.MemoryUsage = memStats.Alloc
		m.metrics.GoRoutines = runtime.NumGoroutine()
		m.mu.Unlock()
	}
}

// ============ 指标记录方法 ============

// RecordRequest 记录请求
func (m *Monitor) RecordRequest(success bool) {
	atomic.AddUint64(&m.metrics.RequestCount, 1)
	if success {
		atomic.AddUint64(&m.metrics.RequestSuccess, 1)
	} else {
		atomic.AddUint64(&m.metrics.RequestError, 1)
	}
}

// RecordAIRequest 记录 AI 请求
func (m *Monitor) RecordAIRequest(tokens int, latencyMs uint64) {
	atomic.AddUint64(&m.metrics.AIRequestCount, 1)
	atomic.AddUint64(&m.metrics.AIRequestTokens, uint64(tokens))
	atomic.AddUint64(&m.metrics.AIRequestLatency, latencyMs)
}

// RecordChannelMessage 记录渠道消息
func (m *Monitor) RecordChannelMessage() {
	atomic.AddUint64(&m.metrics.ChannelMessages, 1)
}

// SetActiveSessions 设置活跃会话数
func (m *Monitor) SetActiveSessions(count int64) {
	atomic.StoreInt64(&m.metrics.ActiveSessions, count)
}

// IncActiveSessions 增加活跃会话
func (m *Monitor) IncActiveSessions() {
	atomic.AddInt64(&m.metrics.ActiveSessions, 1)
}

// DecActiveSessions 减少活跃会话
func (m *Monitor) DecActiveSessions() {
	atomic.AddInt64(&m.metrics.ActiveSessions, -1)
}

// GetMetrics 获取当前指标
func (m *Monitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := *m.metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.MemoryUsage = memStats.Alloc
	metrics.GoRoutines = runtime.NumGoroutine()

	return &metrics
}

var startTime = time.Now()
