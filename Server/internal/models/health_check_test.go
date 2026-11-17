package models

import (
	"testing"
	"time"
)

func TestHealthCheckCreation(t *testing.T) {

	httpCode := 200
	responseTime := int64(120)

	checkedAt := time.Now()

	healthCheck := HealthCheck{
		ID:             "check-123",
		ServiceID:      "service-123",
		Status:         "UP",
		HTTPStatusCode: &httpCode,
		ResponseTimeMs: &responseTime,
		CheckedAt:      checkedAt,
	}

	if healthCheck.Status != "UP" {
		t.Errorf("Expected status UP, got %s", healthCheck.Status)
	}

	if healthCheck.HTTPStatusCode == nil {
		t.Fatal("Expected HTTP status code, got nil")
	}

	if *healthCheck.HTTPStatusCode != 200 {
		t.Errorf("Expected HTTP status 200, got %d", *healthCheck.HTTPStatusCode)
	}

	if *healthCheck.ResponseTimeMs != 120 {
		t.Errorf("Expected response time 120, got %d", *healthCheck.ResponseTimeMs)
	}

	if healthCheck.ErrorMessage != nil {
		t.Errorf("Expected error message to be nil")
	}
}
