package handlers

import (
	"encoding/json"
	"net/http"

	"projekat/metrics"
)

type MetricsResponse struct {
	TotalRequests24h      int                `json:"total_requests_24h"`
	SuccessfulRequests24h int                `json:"successful_requests_24h"`
	FailedRequests24h     int                `json:"failed_requests_24h"`
	AverageRequestTimes   map[string]float64 `json:"average_request_times_seconds"`
	RequestsPerMinute     map[string]float64 `json:"requests_per_minute"`
}


func GetMetrics(w http.ResponseWriter, r *http.Request) {

	response := MetricsResponse{
		TotalRequests24h:      metrics.GetTotalRequests24h(),
		SuccessfulRequests24h: metrics.GetSuccessfulRequests24h(),
		FailedRequests24h:     metrics.GetFailedRequests24h(),
		AverageRequestTimes:   metrics.GetAverageRequestTimePerEndpoint(),
		RequestsPerMinute:     metrics.GetRequestsPerMinutePerEndpoint(),
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
