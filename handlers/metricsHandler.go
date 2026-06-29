package handlers

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsResponse struct {
	TotalRequests24h       *prometheus.CounterVec
	SuccessfulRequests24h  *prometheus.CounterVec
	FailedRequests24h      *prometheus.CounterVec
	AverageRequestDuration *prometheus.HistogramVec
	RequestsPerMinute      *prometheus.CounterVec
}

func NewMetrics() *MetricsResponse {
	return &MetricsResponse{
		TotalRequests24h: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		SuccessfulRequests24h: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_successful_total",
				Help: "Total number of successful HTTP requests (2xx, 3xx)",
			},
			[]string{"method", "endpoint"},
		),
		FailedRequests24h: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_failed_total",
				Help: "Total number of failed HTTP requests (4xx, 5xx)",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		AverageRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		RequestsPerMinute: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_per_minute",
				Help: "Total number of HTTP requests per minute",
			},
			[]string{"method", "endpoint"},
		),
	}
}

func (m *MetricsResponse) RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusGroup := string(rune(statusCode))
	m.TotalRequests24h.WithLabelValues(method, endpoint, statusGroup).Inc()
	m.RequestsPerMinute.WithLabelValues(method, endpoint).Inc()
	m.AverageRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if statusCode >= 200 && statusCode < 400 {
		m.SuccessfulRequests24h.WithLabelValues(method, endpoint).Inc()
	} else if statusCode >= 400 {
		m.FailedRequests24h.WithLabelValues(method, endpoint, statusGroup).Inc()
	}
}
