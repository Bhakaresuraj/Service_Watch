package models

import "time"

type HealthCheck struct {
	ID             string    `json:"id"`
	ServiceID      string    `json:"service_id"`
	Status         string    `json:"status"`
	HTTPStatusCode *int      `json:"http_status_code"`
	ResponseTimeMs *int64    `json:"response_time_ms"`
	ErrorMessage   *string   `json:"error_message"`
	CheckedAt      time.Time `json:"checked_at"`
	CreatedAt      time.Time `json:"created_at"`
}
