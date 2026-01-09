package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 在线用户数量
	onlineUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chat_online_users_total",
		Help: "当前在线用户数量",
	})

	// WebSocket 连接数量
	webSocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chat_websocket_connections_total",
		Help: "当前 WebSocket 连接数量",
	})

	// 消息发送总数
	messagesSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_messages_sent_total",
			Help: "发送的消息总数",
		},
		[]string{"type"}, // type: private, group, broadcast
	)

	// 消息接收总数
	messagesReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_messages_received_total",
			Help: "接收的消息总数",
		},
		[]string{"type"}, // type: private, group, broadcast
	)

	// WebSocket 连接建立总数
	webSocketConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "chat_websocket_connections_created_total",
		Help: "WebSocket 连接建立总数",
	})

	// WebSocket 连接关闭总数
	webSocketDisconnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "chat_websocket_connections_closed_total",
		Help: "WebSocket 连接关闭总数",
	})

	// WebSocket 连接持续时间
	webSocketConnectionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chat_websocket_connection_duration_seconds",
		Help:    "WebSocket 连接持续时间",
		Buckets: prometheus.DefBuckets,
	})

	// HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chat_http_request_duration_seconds",
			Help:    "HTTP 请求持续时间",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// SetOnlineUsers 设置在线用户数量
func SetOnlineUsers(count float64) {
	onlineUsers.Set(count)
}

// SetWebSocketConnections 设置 WebSocket 连接数量
func SetWebSocketConnections(count float64) {
	webSocketConnections.Set(count)
}

// IncrementWebSocketConnection 增加 WebSocket 连接计数
func IncrementWebSocketConnection() {
	webSocketConnectionsTotal.Inc()
}

// IncrementWebSocketDisconnection 增加 WebSocket 断开连接计数
func IncrementWebSocketDisconnection() {
	webSocketDisconnectionsTotal.Inc()
}

// ObserveWebSocketConnectionDuration 记录 WebSocket 连接持续时间
func ObserveWebSocketConnectionDuration(duration float64) {
	webSocketConnectionDuration.Observe(duration)
}

// IncrementMessagesSent 增加发送消息计数
func IncrementMessagesSent(msgType string) {
	messagesSentTotal.WithLabelValues(msgType).Inc()
}

// IncrementMessagesReceived 增加接收消息计数
func IncrementMessagesReceived(msgType string) {
	messagesReceivedTotal.WithLabelValues(msgType).Inc()
}

// IncrementHTTPRequest 增加 HTTP 请求计数
func IncrementHTTPRequest(method, path, status string) {
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// ObserveHTTPRequestDuration 记录 HTTP 请求持续时间
func ObserveHTTPRequestDuration(method, path string, duration float64) {
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// GetRegistry 获取 Prometheus 注册表
func GetRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

// MustRegister 注册自定义指标
func MustRegister(registry *prometheus.Registry, collectors ...prometheus.Collector) {
	registry.MustRegister(collectors...)
}

// MetricsCollector 自定义指标收集器
type MetricsCollector struct {
	mu sync.RWMutex
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// Describe 实现 prometheus.Collector 接口
func (mc *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	onlineUsers.Describe(ch)
	webSocketConnections.Describe(ch)
}

// Collect 实现 prometheus.Collector 接口
func (mc *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	onlineUsers.Collect(ch)
	webSocketConnections.Collect(ch)
}
