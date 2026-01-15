// Tortoise Monitoring and Observability

package monitoring

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tortoise_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tortoise_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// WebSocket metrics
	wsConnectionsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tortoise_ws_connections_total",
			Help: "Total number of WebSocket connections",
		},
	)

	wsConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tortoise_ws_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	wsMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_ws_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"direction", "type"},
	)

	// Session metrics
	sessionsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tortoise_sessions_total",
			Help: "Total number of sessions created",
		},
	)

	sessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tortoise_sessions_active",
			Help: "Number of active sessions",
		},
	)

	sessionsDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tortoise_session_duration_seconds",
			Help:    "Session duration in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400},
		},
	)

	// Message metrics
	messagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_messages_total",
			Help: "Total number of messages",
		},
		[]string{"direction", "type"},
	)

	messageSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tortoise_message_size_bytes",
			Help:    "Message size in bytes",
			Buckets: []float64{64, 256, 1024, 4096, 16384, 65536},
		},
		[]string{"direction"},
	)

	// AI model metrics
	modelRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_model_requests_total",
			Help: "Total number of model API requests",
		},
		[]string{"model", "status"},
	)

	modelLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tortoise_model_latency_seconds",
			Help:    "Model API latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model"},
	)

	modelTokens = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_model_tokens_total",
			Help: "Total number of tokens processed",
		},
		[]string{"model", "type"},
	)

	// Tool metrics
	toolExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_tool_executions_total",
			Help: "Total number of tool executions",
		},
		[]string{"tool", "status"},
	)

	toolExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tortoise_tool_execution_duration_seconds",
			Help:    "Tool execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool"},
	)

	// Memory metrics
	memoryStored = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tortoise_memory_stored_total",
			Help: "Total number of memories stored",
		},
	)

	memorySearches = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tortoise_memory_searches_total",
			Help: "Total number of memory searches",
		},
	)

	memorySearchLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tortoise_memory_search_duration_seconds",
			Help:    "Memory search duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tortoise_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)
)

// Recorder provides methods to record metrics
type Recorder struct{}

// NewRecorder creates a new metrics recorder
func NewRecorder() *Recorder {
	return &Recorder{}
}

// RecordHTTPRequest records an HTTP request
func (r *Recorder) RecordHTTPRequest(method, endpoint string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode(status)).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordWSConnection records a WebSocket connection
func (r *Recorder) RecordWSConnection(active bool) {
	wsConnectionsTotal.Inc()
	if active {
		wsConnectionsActive.Inc()
	} else {
		wsConnectionsActive.Dec()
	}
}

// RecordWSMessage records a WebSocket message
func (r *Recorder) RecordWSMessage(direction, msgType string, size int64) {
	wsMessagesTotal.WithLabelValues(direction, msgType).Inc()
	messageSize.WithLabelValues(direction).Observe(float64(size))
}

// RecordSession records session metrics
func (r *Recorder) RecordSession(duration time.Duration) {
	sessionsTotal.Inc()
	sessionsActive.Inc()
	sessionsDuration.Observe(duration.Seconds())
}

// RecordMessage records message metrics
func (r *Recorder) RecordMessage(direction, msgType string, size int64) {
	messagesTotal.WithLabelValues(direction, msgType).Inc()
	messageSize.WithLabelValues(direction).Observe(float64(size))
}

// RecordModelRequest records AI model request metrics
func (r *Recorder) RecordModelRequest(model string, status string, duration time.Duration, tokens int) {
	modelRequestsTotal.WithLabelValues(model, status).Inc()
	modelLatency.WithLabelValues(model).Observe(duration.Seconds())
	modelTokens.WithLabelValues(model, "prompt").Inc()
	modelTokens.WithLabelValues(model, "completion").Inc()
}

// RecordToolExecution records tool execution metrics
func (r *Recorder) RecordToolExecution(tool string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	toolExecutionsTotal.WithLabelValues(tool, status).Inc()
	toolExecutionDuration.WithLabelValues(tool).Observe(duration.Seconds())
}

// RecordMemory records memory metrics
func (r *Recorder) RecordMemory(duration time.Duration) {
	memoryStored.Inc()
	memorySearches.Inc()
	memorySearchLatency.Observe(duration.Seconds())
}

// RecordError records error metrics
func (r *Recorder) RecordError(errorType, component string) {
	errorsTotal.WithLabelValues(errorType, component).Inc()
}

func statusCode(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

// Context key for recorder
type contextKey string

const recorderKey contextKey = "recorder"

// WithRecorder adds a recorder to context
func WithRecorder(ctx context.Context, r *Recorder) context.Context {
	return context.WithValue(ctx, recorderKey, r)
}

// FromContext retrieves a recorder from context
func FromContext(ctx context.Context) *Recorder {
	if r, ok := ctx.Value(recorderKey).(*Recorder); ok {
		return r
	}
	return NewRecorder()
}
