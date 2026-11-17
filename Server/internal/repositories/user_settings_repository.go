package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserSettingsNotFound = errors.New("user settings not found")

type UserSettingsRepository struct {
	DB *pgxpool.Pool
}

func NewUserSettingsRepository(db *pgxpool.Pool) *UserSettingsRepository {
	return &UserSettingsRepository{DB: db}
}

// CreateUserSettings creates default notification settings for a user.
func (r *UserSettingsRepository) CreateUserSettings(
	ctx context.Context,
	settings *models.UserSettings,
) error {
	query := `
		INSERT INTO user_settings (user_id)
		VALUES ($1)
		RETURNING email_notifications_enabled,
		          down_alerts_enabled,
		          recovery_alerts_enabled,
		          created_at,
		          updated_at
	`

	err := r.DB.QueryRow(ctx, query, settings.UserID).Scan(
		&settings.EmailNotificationsEnabled,
		&settings.DownAlertsEnabled,
		&settings.RecoveryAlertsEnabled,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user settings: %w", err)
	}

	return nil
}

// GetUserSettingsByUserID retrieves settings belonging to a user.
func (r *UserSettingsRepository) GetUserSettingsByUserID(
	ctx context.Context,
	userID string,
) (*models.UserSettings, error) {
	settings := &models.UserSettings{}

	query := `
		SELECT user_id,
		       email_notifications_enabled,
		       down_alerts_enabled,
		       recovery_alerts_enabled,
		       created_at,
		       updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	err := r.DB.QueryRow(ctx, query, userID).Scan(
		&settings.UserID,
		&settings.EmailNotificationsEnabled,
		&settings.DownAlertsEnabled,
		&settings.RecoveryAlertsEnabled,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserSettingsNotFound
		}

		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return settings, nil
}

// UpdateUserSettings updates a user's notification preferences.
func (r *UserSettingsRepository) UpdateUserSettings(
	ctx context.Context,
	settings *models.UserSettings,
) error {
	query := `
		UPDATE user_settings
		SET email_notifications_enabled = $2,
		    down_alerts_enabled = $3,
		    recovery_alerts_enabled = $4,
		    updated_at = NOW()
		WHERE user_id = $1
		RETURNING updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		settings.UserID,
		settings.EmailNotificationsEnabled,
		settings.DownAlertsEnabled,
		settings.RecoveryAlertsEnabled,
	).Scan(&settings.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserSettingsNotFound
		}

		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}
