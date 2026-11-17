package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
)

func setupNotificationTest(t *testing.T) (
	context.Context,
	*NotificationRepository,
	*models.User,
	*models.Service,
	*models.Incident,
) {
	t.Helper()

	_, ctx, user, service, incidentRepo := setupIncidentTest(t)

	incident := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now(),
	}

	if err := incidentRepo.CreateIncident(ctx, incident); err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	notificationRepo := NewNotificationRepository(incidentRepo.DB)

	return ctx, notificationRepo, user, service, incident
}

func createTestNotification(
	t *testing.T,
	ctx context.Context,
	repo *NotificationRepository,
	user *models.User,
	incident *models.Incident,
) *models.Notification {
	t.Helper()

	notification := &models.Notification{
		IncidentID: incident.ID,
		UserID:     user.ID,
		Type:       "DOWN",
	}

	if err := repo.CreateNotification(ctx, notification); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	return notification
}

func TestCreateNotification(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	notification := createTestNotification(t, ctx, repo, user, incident)

	if notification.ID == "" {
		t.Error("Expected generated notification ID")
	}

	if notification.Status != "PENDING" {
		t.Errorf("Expected PENDING, got %s", notification.Status)
	}

	if notification.RetryCount != 0 {
		t.Errorf("Expected retry count 0, got %d", notification.RetryCount)
	}

	if notification.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", notification.MaxRetries)
	}
}

func TestGetNotificationByID(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	notification := createTestNotification(t, ctx, repo, user, incident)

	result, err := repo.GetNotificationByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to get notification: %v", err)
	}

	if result.ID != notification.ID {
		t.Errorf("Expected ID %s, got %s", notification.ID, result.ID)
	}

	if result.IncidentID != incident.ID {
		t.Errorf("Expected incident ID %s, got %s", incident.ID, result.IncidentID)
	}

	if result.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, result.UserID)
	}
}

func TestGetNotificationByIDNotFound(t *testing.T) {
	ctx, repo, _, _, _ := setupNotificationTest(t)

	missingID := "00000000-0000-0000-0000-000000000041"

	result, err := repo.GetNotificationByID(ctx, missingID)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("Expected ErrNotificationNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}
}

func TestGetPendingNotifications(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	dueNotification := createTestNotification(t, ctx, repo, user, incident)

	futureNotification := createTestNotification(t, ctx, repo, user, incident)

	futureTime := time.Now().Add(10 * time.Minute)

	err := repo.ScheduleNotificationRetry(
		ctx,
		futureNotification.ID,
		&futureTime,
		"Temporary email failure",
	)
	if err != nil {
		t.Fatalf("Failed to schedule future retry: %v", err)
	}

	results, err := repo.GetPendingNotifications(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get pending notifications: %v", err)
	}

	foundDueNotification := false
	foundFutureNotification := false

	for _, result := range results {
		if result.ID == dueNotification.ID {
			foundDueNotification = true
		}

		if result.ID == futureNotification.ID {
			foundFutureNotification = true
		}
	}

	if !foundDueNotification {
		t.Error("Expected the due notification to be returned")
	}

	if foundFutureNotification {
		t.Error("Future notification should not be returned")
	}
}

func TestGetPendingNotificationsNoResults(t *testing.T) {
	ctx, repo, _, _, _ := setupNotificationTest(t)

	results, err := repo.GetPendingNotifications(ctx, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 notifications, got %d", len(results))
	}

	if results == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestMarkNotificationSent(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	notification := createTestNotification(t, ctx, repo, user, incident)

	err := repo.MarkNotificationSent(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to mark notification as sent: %v", err)
	}

	result, err := repo.GetNotificationByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve notification: %v", err)
	}

	if result.Status != "SENT" {
		t.Errorf("Expected SENT, got %s", result.Status)
	}

	if result.SentAt == nil {
		t.Error("Expected sent_at to be set")
	}
}

func TestMarkNotificationSentNotFound(t *testing.T) {
	ctx, repo, _, _, _ := setupNotificationTest(t)

	err := repo.MarkNotificationSent(
		ctx,
		"00000000-0000-0000-0000-000000000042",
	)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("Expected ErrNotificationNotFound, got %v", err)
	}
}

func TestScheduleNotificationRetry(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	notification := createTestNotification(t, ctx, repo, user, incident)

	retryTime := time.Now().Add(5 * time.Minute)
	errorMessage := "SMTP server unavailable"

	err := repo.ScheduleNotificationRetry(
		ctx,
		notification.ID,
		&retryTime,
		errorMessage,
	)
	if err != nil {
		t.Fatalf("Failed to schedule retry: %v", err)
	}

	result, err := repo.GetNotificationByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve notification: %v", err)
	}

	if result.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", result.RetryCount)
	}

	if result.NextRetryAt == nil {
		t.Fatal("Expected next_retry_at to be set")
	}

	if result.ErrorMessage == nil || *result.ErrorMessage != errorMessage {
		t.Error("Expected error message to match")
	}

	if result.Status != "PENDING" {
		t.Errorf("Expected PENDING, got %s", result.Status)
	}
}

func TestScheduleNotificationRetryNotFound(t *testing.T) {
	ctx, repo, _, _, _ := setupNotificationTest(t)

	retryTime := time.Now().Add(5 * time.Minute)

	err := repo.ScheduleNotificationRetry(
		ctx,
		"00000000-0000-0000-0000-000000000043",
		&retryTime,
		"Test error",
	)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("Expected ErrNotificationNotFound, got %v", err)
	}
}

func TestMarkNotificationFailed(t *testing.T) {
	ctx, repo, user, _, incident := setupNotificationTest(t)

	notification := createTestNotification(t, ctx, repo, user, incident)

	errorMessage := "Maximum retry attempts reached"

	err := repo.MarkNotificationFailed(ctx, notification.ID, errorMessage)
	if err != nil {
		t.Fatalf("Failed to mark notification as failed: %v", err)
	}

	result, err := repo.GetNotificationByID(ctx, notification.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve notification: %v", err)
	}

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED, got %s", result.Status)
	}

	if result.ErrorMessage == nil || *result.ErrorMessage != errorMessage {
		t.Error("Expected error message to match")
	}
}

func TestMarkNotificationFailedNotFound(t *testing.T) {
	ctx, repo, _, _, _ := setupNotificationTest(t)

	err := repo.MarkNotificationFailed(
		ctx,
		"00000000-0000-0000-0000-000000000044",
		"Test failure",
	)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("Expected ErrNotificationNotFound, got %v", err)
	}
}

func TestCreateNotificationInvalidIncident(t *testing.T) {
	ctx, repo, user, _, _ := setupNotificationTest(t)

	notification := &models.Notification{
		IncidentID: "00000000-0000-0000-0000-000000000099",
		UserID:     user.ID,
		Type:       "DOWN",
	}

	err := repo.CreateNotification(ctx, notification)

	if err == nil {
		t.Fatal("Expected error for nonexistent incident ID")
	}
}
