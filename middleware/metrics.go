package middleware

import (
	"net/http"
	"time"

	"projekat/handlers"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler, metricsCollector *handlers.MetricsResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		metricsCollector.RecordRequest(
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
	})
}
