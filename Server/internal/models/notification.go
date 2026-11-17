package models

import "time"

type Notification struct {
	ID           string     `json:"id"`
	IncidentID   string     `json:"incident_id"`
	UserID       string     `json:"user_id"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	RetryCount   int        `json:"retry_count"`
	MaxRetries   int        `json:"max_retries"`
	NextRetryAt  *time.Time `json:"next_retry_at"`
	SentAt       *time.Time `json:"sent_at"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
