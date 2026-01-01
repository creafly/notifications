package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notifications_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	NotificationsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notifications_total",
			Help: "Total number of notifications",
		},
		[]string{"type", "status"},
	)

	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"type", "channel"},
	)

	WebSocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notifications_websocket_connections",
			Help: "Number of active WebSocket connections",
		},
	)

	WebSocketMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"direction", "type"},
	)

	InvitationsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notifications_invitations_total",
			Help: "Total number of invitations",
		},
		[]string{"status"},
	)

	InvitationOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_invitation_operations_total",
			Help: "Total number of invitation operations",
		},
		[]string{"operation", "status"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notifications_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notifications_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)

	OutboxEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_outbox_events_total",
			Help: "Total number of outbox events",
		},
		[]string{"event_type", "status"},
	)

	KafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_kafka_messages_consumed_total",
			Help: "Total number of Kafka messages consumed",
		},
		[]string{"topic", "status"},
	)
)

func RecordNotificationSent(notifType, channel string) {
	NotificationsSent.WithLabelValues(notifType, channel).Inc()
}

func RecordWebSocketConnection(delta float64) {
	WebSocketConnections.Add(delta)
}

func RecordWebSocketMessage(direction, msgType string) {
	WebSocketMessages.WithLabelValues(direction, msgType).Inc()
}

func RecordInvitationOperation(operation, status string) {
	InvitationOperations.WithLabelValues(operation, status).Inc()
}

func RecordError(errorType string) {
	ErrorsTotal.WithLabelValues(errorType).Inc()
}

func RecordOutboxEvent(eventType, status string) {
	OutboxEventsTotal.WithLabelValues(eventType, status).Inc()
}

func RecordKafkaMessage(topic, status string) {
	KafkaMessagesConsumed.WithLabelValues(topic, status).Inc()
}
