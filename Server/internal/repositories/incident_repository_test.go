package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupIncidentTest(t *testing.T) (
	*pgxpool.Pool,
	context.Context,
	*models.User,
	*models.Service,
	*IncidentRepository,
) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	incidentRepo := NewIncidentRepository(db)

	user := &models.User{
		Name:         "Incident Test User",
		Email:        fmt.Sprintf("incident-test-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up user: %v", err)
		}
	})

	service := &models.Service{
		UserID:             user.ID,
		Name:               "Incident Test Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	return db, ctx, user, service, incidentRepo
}

func TestCreateIncident(t *testing.T) {
	_, _, _, service, repo := setupIncidentTest(t)

	incident := &models.Incident{
		ServiceID:   service.ID,
		StartedAt:   time.Now(),
		Description: stringPointer("Website is unavailable"),
	}

	if err := repo.CreateIncident(context.Background(), incident); err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	if incident.ID == "" {
		t.Error("Expected generated incident ID")
	}

	if incident.Status != "OPEN" {
		t.Errorf("Expected status OPEN, got %s", incident.Status)
	}

	if incident.CreatedAt.IsZero() || incident.UpdatedAt.IsZero() {
		t.Error("Expected created_at and updated_at to be set")
	}
}

func TestGetOpenIncidentByServiceID(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	incident := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now(),
	}

	if err := repo.CreateIncident(ctx, incident); err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	result, err := repo.GetOpenIncidentByServiceID(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to get open incident: %v", err)
	}

	if result.ID != incident.ID {
		t.Errorf("Expected incident ID %s, got %s", incident.ID, result.ID)
	}

	if result.Status != "OPEN" {
		t.Errorf("Expected OPEN, got %s", result.Status)
	}
}

func TestGetOpenIncidentByServiceIDNotFound(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	result, err := repo.GetOpenIncidentByServiceID(ctx, service.ID)

	if !errors.Is(err, ErrIncidentNotFound) {
		t.Errorf("Expected ErrIncidentNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}
}

func TestGetIncidentsByServiceID(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	first := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now().Add(-2 * time.Minute),
	}

	second := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now().Add(-1 * time.Minute),
	}

	if err := repo.CreateIncident(ctx, first); err != nil {
		t.Fatalf("Failed to create first incident: %v", err)
	}

	if err := repo.ResolveIncident(ctx, first.ID); err != nil {
		t.Fatalf("Failed to resolve first incident: %v", err)
	}

	if err := repo.CreateIncident(ctx, second); err != nil {
		t.Fatalf("Failed to create second incident: %v", err)
	}

	results, err := repo.GetIncidentsByServiceID(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to get incidents: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 incidents, got %d", len(results))
	}

	if results[0].ID != second.ID {
		t.Error("Expected newest incident first")
	}

	if results[1].ID != first.ID {
		t.Error("Expected oldest incident second")
	}
}

func TestGetIncidentsByServiceIDNoResults(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	results, err := repo.GetIncidentsByServiceID(ctx, service.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 incidents, got %d", len(results))
	}

	if results == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestGetIncidentByID(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	description := "Website is down"

	incident := &models.Incident{
		ServiceID:   service.ID,
		StartedAt:   time.Now(),
		Description: &description,
	}

	if err := repo.CreateIncident(ctx, incident); err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	result, err := repo.GetIncidentByID(ctx, incident.ID)
	if err != nil {
		t.Fatalf("Failed to get incident: %v", err)
	}

	if result.ID != incident.ID {
		t.Errorf("Expected ID %s, got %s", incident.ID, result.ID)
	}

	if result.ServiceID != service.ID {
		t.Errorf("Expected service ID %s, got %s", service.ID, result.ServiceID)
	}

	if result.Description == nil || *result.Description != description {
		t.Error("Expected incident description to match")
	}
}

func TestGetIncidentByIDNotFound(t *testing.T) {
	_, ctx, _, _, repo := setupIncidentTest(t)

	missingID := "00000000-0000-0000-0000-000000000031"

	result, err := repo.GetIncidentByID(ctx, missingID)

	if !errors.Is(err, ErrIncidentNotFound) {
		t.Errorf("Expected ErrIncidentNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}
}

func TestResolveIncident(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	incident := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now(),
	}

	if err := repo.CreateIncident(ctx, incident); err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	if err := repo.ResolveIncident(ctx, incident.ID); err != nil {
		t.Fatalf("Failed to resolve incident: %v", err)
	}

	result, err := repo.GetIncidentByID(ctx, incident.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve resolved incident: %v", err)
	}

	if result.Status != "RESOLVED" {
		t.Errorf("Expected RESOLVED, got %s", result.Status)
	}

	if result.ResolvedAt == nil {
		t.Error("Expected resolved_at to be set")
	}
}

func TestResolveIncidentNotFound(t *testing.T) {
	_, ctx, _, _, repo := setupIncidentTest(t)

	missingID := "00000000-0000-0000-0000-000000000032"

	err := repo.ResolveIncident(ctx, missingID)

	if !errors.Is(err, ErrIncidentNotFound) {
		t.Errorf("Expected ErrIncidentNotFound, got %v", err)
	}
}

func TestCreateDuplicateOpenIncident(t *testing.T) {
	_, ctx, _, service, repo := setupIncidentTest(t)

	first := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now(),
	}

	if err := repo.CreateIncident(ctx, first); err != nil {
		t.Fatalf("Failed to create first incident: %v", err)
	}

	second := &models.Incident{
		ServiceID: service.ID,
		StartedAt: time.Now(),
	}

	err := repo.CreateIncident(ctx, second)

	if err == nil {
		t.Fatal("Expected error when creating duplicate open incident")
	}
}

func TestCreateIncidentInvalidService(t *testing.T) {
	_, ctx, _, _, repo := setupIncidentTest(t)

	incident := &models.Incident{
		ServiceID: "00000000-0000-0000-0000-000000000099",
		StartedAt: time.Now(),
	}

	err := repo.CreateIncident(ctx, incident)

	if err == nil {
		t.Fatal("Expected error for nonexistent service ID")
	}
}

func stringPointer(value string) *string {
	return &value
}
