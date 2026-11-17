package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
)

func setupUserSettingsRepositoryTest(t *testing.T) (
	context.Context,
	*UserSettingsRepository,
	*models.User,
) {
	t.Helper()

	db, ctx, user, _, _ := setupIncidentTest(t)

	repo := NewUserSettingsRepository(db)

	return ctx, repo, user
}

func TestCreateUserSettings(t *testing.T) {
	ctx, repo, user := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID: user.ID,
	}

	err := repo.CreateUserSettings(ctx, settings)
	if err != nil {
		t.Fatalf("Failed to create user settings: %v", err)
	}

	if settings.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, settings.UserID)
	}

	if !settings.EmailNotificationsEnabled {
		t.Error("Expected email notifications to be enabled by default")
	}

	if !settings.DownAlertsEnabled {
		t.Error("Expected DOWN alerts to be enabled by default")
	}

	if !settings.RecoveryAlertsEnabled {
		t.Error("Expected RECOVERY alerts to be enabled by default")
	}

	if settings.CreatedAt.IsZero() || settings.UpdatedAt.IsZero() {
		t.Error("Expected created_at and updated_at to be set")
	}
}

func TestGetUserSettingsByUserID(t *testing.T) {
	ctx, repo, user := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID: user.ID,
	}

	if err := repo.CreateUserSettings(ctx, settings); err != nil {
		t.Fatalf("Failed to create user settings: %v", err)
	}

	result, err := repo.GetUserSettingsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user settings: %v", err)
	}

	if result.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, result.UserID)
	}

	if !result.EmailNotificationsEnabled ||
		!result.DownAlertsEnabled ||
		!result.RecoveryAlertsEnabled {
		t.Error("Expected all notification settings to be enabled")
	}
}

func TestGetUserSettingsByUserIDNotFound(t *testing.T) {
	ctx, repo, _ := setupUserSettingsRepositoryTest(t)

	missingID := "00000000-0000-0000-0000-000000000051"

	result, err := repo.GetUserSettingsByUserID(ctx, missingID)

	if !errors.Is(err, ErrUserSettingsNotFound) {
		t.Errorf("Expected ErrUserSettingsNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}
}

func TestUpdateUserSettings(t *testing.T) {
	ctx, repo, user := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID: user.ID,
	}

	if err := repo.CreateUserSettings(ctx, settings); err != nil {
		t.Fatalf("Failed to create user settings: %v", err)
	}

	settings.EmailNotificationsEnabled = false
	settings.DownAlertsEnabled = false
	settings.RecoveryAlertsEnabled = true

	if err := repo.UpdateUserSettings(ctx, settings); err != nil {
		t.Fatalf("Failed to update user settings: %v", err)
	}

	result, err := repo.GetUserSettingsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated settings: %v", err)
	}

	if result.EmailNotificationsEnabled {
		t.Error("Expected email notifications to be disabled")
	}

	if result.DownAlertsEnabled {
		t.Error("Expected DOWN alerts to be disabled")
	}

	if !result.RecoveryAlertsEnabled {
		t.Error("Expected RECOVERY alerts to remain enabled")
	}

	if settings.UpdatedAt.IsZero() {
		t.Error("Expected updated_at to be set")
	}
}

func TestUpdateUserSettingsNotFound(t *testing.T) {
	ctx, repo, _ := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID:                    "00000000-0000-0000-0000-000000000052",
		EmailNotificationsEnabled: false,
		DownAlertsEnabled:         false,
		RecoveryAlertsEnabled:     false,
	}

	err := repo.UpdateUserSettings(ctx, settings)

	if !errors.Is(err, ErrUserSettingsNotFound) {
		t.Errorf("Expected ErrUserSettingsNotFound, got %v", err)
	}
}

func TestCreateUserSettingsInvalidUser(t *testing.T) {
	ctx, repo, _ := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID: "00000000-0000-0000-0000-000000000053",
	}

	err := repo.CreateUserSettings(ctx, settings)

	if err == nil {
		t.Fatal("Expected error for nonexistent user ID")
	}
}

func TestCreateDuplicateUserSettings(t *testing.T) {
	ctx, repo, user := setupUserSettingsRepositoryTest(t)

	settings := &models.UserSettings{
		UserID: user.ID,
	}

	if err := repo.CreateUserSettings(ctx, settings); err != nil {
		t.Fatalf("Failed to create first settings record: %v", err)
	}

	duplicate := &models.UserSettings{
		UserID: user.ID,
	}

	err := repo.CreateUserSettings(ctx, duplicate)

	if err == nil {
		t.Fatal("Expected error when creating duplicate user settings")
	}
}
