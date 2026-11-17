package models

import "time"

type UserSettings struct {
	UserID                    string    `json:"user_id"`
	EmailNotificationsEnabled bool      `json:"email_notifications_enabled"`
	DownAlertsEnabled         bool      `json:"down_alerts_enabled"`
	RecoveryAlertsEnabled     bool      `json:"recovery_alerts_enabled"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}
