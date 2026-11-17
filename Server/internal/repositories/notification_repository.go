package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationRepository struct {
	DB *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

// CreateNotification creates a new pending notification.
func (r *NotificationRepository) CreateNotification(
	ctx context.Context,
	notification *models.Notification,
) error {
	query := `
		INSERT INTO notifications (incident_id, user_id, type)
		VALUES ($1, $2, $3)
		RETURNING id, status, retry_count, max_retries, created_at, updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		notification.IncidentID,
		notification.UserID,
		notification.Type,
	).Scan(
		&notification.ID,
		&notification.Status,
		&notification.RetryCount,
		&notification.MaxRetries,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetNotificationByID retrieves a notification by its ID.
func (r *NotificationRepository) GetNotificationByID(
	ctx context.Context,
	notificationID string,
) (*models.Notification, error) {
	notification := &models.Notification{}

	query := `
		SELECT id, incident_id, user_id, type, status,
		       retry_count, max_retries, next_retry_at,
		       sent_at, error_message, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	err := r.DB.QueryRow(ctx, query, notificationID).Scan(
		&notification.ID,
		&notification.IncidentID,
		&notification.UserID,
		&notification.Type,
		&notification.Status,
		&notification.RetryCount,
		&notification.MaxRetries,
		&notification.NextRetryAt,
		&notification.SentAt,
		&notification.ErrorMessage,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}

		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return notification, nil
}

// GetPendingNotifications retrieves pending notifications ready for processing.
func (r *NotificationRepository) GetPendingNotifications(
	ctx context.Context,
	limit int,
) ([]models.Notification, error) {
	query := `
		SELECT id, incident_id, user_id, type, status,
		       retry_count, max_retries, next_retry_at,
		       sent_at, error_message, created_at, updated_at
		FROM notifications
		WHERE status = 'PENDING'
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]models.Notification, 0)

	for rows.Next() {
		var notification models.Notification

		if err := rows.Scan(
			&notification.ID,
			&notification.IncidentID,
			&notification.UserID,
			&notification.Type,
			&notification.Status,
			&notification.RetryCount,
			&notification.MaxRetries,
			&notification.NextRetryAt,
			&notification.SentAt,
			&notification.ErrorMessage,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating notifications: %w", err)
	}

	return notifications, nil
}

// MarkNotificationSent marks a pending notification as successfully sent.
func (r *NotificationRepository) MarkNotificationSent(
	ctx context.Context,
	notificationID string,
) error {
	query := `
		UPDATE notifications
		SET status = 'SENT',
		    sent_at = NOW(),
		    next_retry_at = NULL,
		    error_message = NULL,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id
	`

	var id string

	err := r.DB.QueryRow(ctx, query, notificationID).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotificationNotFound
		}

		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return nil
}

// ScheduleNotificationRetry schedules another attempt after a failure.
func (r *NotificationRepository) ScheduleNotificationRetry(
	ctx context.Context,
	notificationID string,
	nextRetryAt *time.Time,
	errorMessage string,
) error {
	query := `
		UPDATE notifications
		SET retry_count = retry_count + 1,
		    next_retry_at = $2,
		    error_message = $3,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id
	`

	var id string

	err := r.DB.QueryRow(
		ctx,
		query,
		notificationID,
		nextRetryAt,
		errorMessage,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotificationNotFound
		}

		return fmt.Errorf("failed to schedule notification retry: %w", err)
	}

	return nil
}

// MarkNotificationFailed marks a notification as permanently failed.
func (r *NotificationRepository) MarkNotificationFailed(
	ctx context.Context,
	notificationID string,
	errorMessage string,
) error {
	query := `
		UPDATE notifications
		SET status = 'FAILED',
		    next_retry_at = NULL,
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id
	`

	var id string

	err := r.DB.QueryRow(
		ctx,
		query,
		notificationID,
		errorMessage,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotificationNotFound
		}

		return fmt.Errorf("failed to mark notification as failed: %w", err)
	}

	return nil
}
