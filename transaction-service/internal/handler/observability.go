package handler

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_service_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_service_http_errors_total",
			Help: "Total number of HTTP error responses (4xx and 5xx).",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(requestCount, errorCount, requestDuration)
}

type observabilityMiddleware struct {
	requestCount atomic.Uint64
	errorCount   atomic.Uint64
}

func newObservabilityMiddleware() *observabilityMiddleware {
	return &observabilityMiddleware{}
}

func (m *observabilityMiddleware) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		// Allow local frontend during development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Email, X-User-ID, X-Correlation-Id, X-Session-ID, X-User-Session-ID")

		startedAt := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		totalRequests := m.requestCount.Add(1)
		totalErrors := m.errorCount.Load()
		if recorder.statusCode >= http.StatusBadRequest {
			totalErrors = m.errorCount.Add(1)
		}

		// Record Prometheus metrics
		statusStr := strconv.Itoa(recorder.statusCode)
		requestCount.WithLabelValues(r.Method, r.URL.Path, statusStr).Inc()
		if recorder.statusCode >= http.StatusBadRequest {
			errorCount.WithLabelValues(r.Method, r.URL.Path, statusStr).Inc()
		}
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(startedAt).Seconds())

		correlationID := recorder.Header().Get("X-Correlation-Id")
		if correlationID == "" {
			correlationID = correlationIDFromRequest(r)
		}

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		latencyMs := time.Since(startedAt).Milliseconds()
		errorRate := float64(totalErrors) / float64(maxUint64(totalRequests, 1))

		// Structured logging with timestamp, correlation_id, user_id, endpoint, status, latency, error
		log.Printf(
			"timestamp=%s correlation_id=%s user_id=%s endpoint=%s method=%s status=%d latency_ms=%d request_count=%d error_count=%d error_rate=%.4f",
			startedAt.Format(time.RFC3339Nano),
			correlationID,
			userID,
			r.URL.Path,
			r.Method,
			recorder.statusCode,
			latencyMs,
			totalRequests,
			totalErrors,
			errorRate,
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func maxUint64(value uint64, fallback uint64) uint64 {
	if value == 0 {
		return fallback
	}

	return value
}
