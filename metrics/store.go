package metrics

import (
	"sync"
	"time"
)

type RequestMetric struct {
	Path       string        `json:"path"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
}

var (
	requests []RequestMetric
	mu       sync.RWMutex
)

func Record(path string, statusCode int, duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	requests = append(requests, RequestMetric{
		Path:       path,
		StatusCode: statusCode,
		Duration:   duration,
		Timestamp:  time.Now(),
	})
}

func getLast24hRequests() []RequestMetric {
	mu.RLock()
	defer mu.RUnlock()

	cutoff := time.Now().Add(-24 * time.Hour)

	var result []RequestMetric

	for _, r := range requests {
		if r.Timestamp.After(cutoff) {
			result = append(result, r)
		}
	}

	return result
}

func GetTotalRequests24h() int {
	return len(getLast24hRequests())
}

func GetSuccessfulRequests24h() int {
	reqs := getLast24hRequests()

	count := 0

	for _, r := range reqs {
		if r.StatusCode >= 200 && r.StatusCode < 400 {
			count++
		}
	}

	return count
}

func GetFailedRequests24h() int {
	reqs := getLast24hRequests()

	count := 0

	for _, r := range reqs {
		if r.StatusCode >= 400 {
			count++
		}
	}

	return count
}

func GetAverageRequestTimePerEndpoint() map[string]float64 {
	reqs := getLast24hRequests()

	totalDurations := make(map[string]time.Duration)
	counts := make(map[string]int)

	for _, r := range reqs {
		totalDurations[r.Path] += r.Duration
		counts[r.Path]++
	}

	averages := make(map[string]float64)

	for path, total := range totalDurations {
		avg := total.Seconds() / float64(counts[path])
		averages[path] = avg
	}

	return averages
}

func GetRequestsPerMinutePerEndpoint() map[string]float64 {
	reqs := getLast24hRequests()

	counts := make(map[string]int)

	for _, r := range reqs {
		counts[r.Path]++
	}

	result := make(map[string]float64)

	for path, count := range counts {
		result[path] = float64(count) / (24 * 60)
	}

	return result
}
