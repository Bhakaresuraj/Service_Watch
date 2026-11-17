package models

import "time"

type Service struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Name               string     `json:"name"`
	URL                string     `json:"url"`
	MonitoringInterval int        `json:"monitoring_interval"`
	TimeoutSeconds     int        `json:"timeout_seconds"`
	FailureThreshold   int        `json:"failure_threshold"`
	RecoveryThreshold  int        `json:"recovery_threshold"`
	CurrentStatus      string     `json:"current_status"`
	IsEnabled          bool       `json:"is_enabled"`
	LastCheckedAt      *time.Time `json:"last_checked_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
