package handler

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type observabilityMiddleware struct {
	requestCount atomic.Uint64
	errorCount   atomic.Uint64
}

func newObservabilityMiddleware() *observabilityMiddleware {
	return &observabilityMiddleware{}
}

func (m *observabilityMiddleware) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		correlationID := recorder.Header().Get("X-Correlation-Id")
		if correlationID == "" {
			correlationID = correlationIDFromRequest(r)
		}

		errorRate := float64(totalErrors) / float64(maxUint64(totalRequests, 1))
		log.Printf(
			"correlation_id=%s event=request_completed method=%s path=%s status=%d latency_ms=%d request_count=%d error_count=%d error_rate=%.4f",
			correlationID,
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			time.Since(startedAt).Milliseconds(),
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
