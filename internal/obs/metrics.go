package obs

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds Prometheus metrics
type Metrics struct {
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	cacheHits          prometheus.Counter
	cacheMisses        prometheus.Counter
	databaseOperations *prometheus.HistogramVec
	activeConnections  prometheus.Gauge
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	m := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		httpRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"method", "path"},
		),
		cacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
		),
		cacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
		),
		databaseOperations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_operation_duration_seconds",
				Help:    "Database operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Current number of active connections",
			},
		),
	}

	// Register metrics
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.cacheHits,
		m.cacheMisses,
		m.databaseOperations,
		m.activeConnections,
	)

	return m
}

// MetricsMiddleware creates a Gin middleware for metrics collection
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Track in-flight requests
		metrics.httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Inc()
		defer metrics.httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler returns the Prometheus metrics endpoint handler
func MetricsHandler(metrics *Metrics) gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// RecordCacheHit increments the cache hit counter
func (m *Metrics) RecordCacheHit() {
	m.cacheHits.Inc()
}

// RecordCacheMiss increments the cache miss counter
func (m *Metrics) RecordCacheMiss() {
	m.cacheMisses.Inc()
}

// RecordDatabaseOperation records database operation duration
func (m *Metrics) RecordDatabaseOperation(operation, table string, duration time.Duration) {
	m.databaseOperations.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// SetActiveConnections sets the active connections gauge
func (m *Metrics) SetActiveConnections(count int) {
	m.activeConnections.Set(float64(count))
}
