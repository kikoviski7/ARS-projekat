package metrics

import (
	"testing"
	"time"
)

func TestRecordRequest(t *testing.T) {

	requests = nil

	Record("/test", 200, 100*time.Millisecond)

	if len(requests) != 1 {
		t.Errorf("expected 1 request, got %d", len(requests))
	}
}

func TestSuccessfulRequests(t *testing.T) {

	requests = nil

	Record("/test", 200, 50*time.Millisecond)
	Record("/test", 201, 50*time.Millisecond)
	Record("/test", 302, 50*time.Millisecond)

	result := GetSuccessfulRequests24h()

	if result != 3 {
		t.Errorf("expected 3 successful requests, got %d", result)
	}
}

func TestFailedRequests(t *testing.T) {

	requests = nil

	Record("/test", 404, 50*time.Millisecond)
	Record("/test", 500, 50*time.Millisecond)

	result := GetFailedRequests24h()

	if result != 2 {
		t.Errorf("expected 2 failed requests, got %d", result)
	}
}
