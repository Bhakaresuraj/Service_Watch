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

func TestCreateHealthCheck(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	healthRepo := NewHealthCheckRepository(db)

	user := &models.User{
		Name:         "Health Test User",
		Email:        fmt.Sprintf("health-create-%d@example.com", time.Now().UnixNano()),
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
		Name:               "Test Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}
	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	httpCode := 200
	responseTime := int64(120)

	healthCheck := &models.HealthCheck{
		ServiceID:      service.ID,
		Status:         "UP",
		HTTPStatusCode: &httpCode,
		ResponseTimeMs: &responseTime,
		CheckedAt:      time.Now(),
	}

	if err := healthRepo.CreateHealthCheck(ctx, healthCheck); err != nil {
		t.Fatalf("Failed to create health check: %v", err)
	}

	if healthCheck.ID == "" {
		t.Error("Expected generated health check ID")
	}
	if healthCheck.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestGetHealthChecksByServiceID(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	healthRepo := NewHealthCheckRepository(db)

	user := &models.User{
		Name:         "History Test User",
		Email:        fmt.Sprintf("health-history-%d@example.com", time.Now().UnixNano()),
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
		Name:               "History Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}
	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	firstTime := time.Now().Add(-2 * time.Minute)
	secondTime := time.Now().Add(-1 * time.Minute)

	firstCheck := &models.HealthCheck{
		ServiceID: service.ID,
		Status:    "UP",
		CheckedAt: firstTime,
	}
	secondCheck := &models.HealthCheck{
		ServiceID: service.ID,
		Status:    "DOWN",
		CheckedAt: secondTime,
	}

	if err := healthRepo.CreateHealthCheck(ctx, firstCheck); err != nil {
		t.Fatalf("Failed to create first health check: %v", err)
	}
	if err := healthRepo.CreateHealthCheck(ctx, secondCheck); err != nil {
		t.Fatalf("Failed to create second health check: %v", err)
	}

	results, err := healthRepo.GetHealthChecksByServiceID(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to get health checks: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 health checks, got %d", len(results))
	}

	if results[0].ID != secondCheck.ID {
		t.Error("Expected newest health check first")
	}
	if results[1].ID != firstCheck.ID {
		t.Error("Expected oldest health check second")
	}
}

func TestGetHealthChecksByServiceIDNoResults(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	healthRepo := NewHealthCheckRepository(db)

	user := &models.User{
		Name:         "Empty History User",
		Email:        fmt.Sprintf("health-empty-%d@example.com", time.Now().UnixNano()),
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
		Name:               "Empty History Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}
	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	results, err := healthRepo.GetHealthChecksByServiceID(ctx, service.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
	if results == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestGetHealthCheckByID(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	healthRepo := NewHealthCheckRepository(db)

	user := &models.User{
		Name:         "Single Check User",
		Email:        fmt.Sprintf("health-single-%d@example.com", time.Now().UnixNano()),
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
		Name:               "Single Check Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}
	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	httpCode := 200
	responseTime := int64(150)
	errorMessage := "test message"

	healthCheck := &models.HealthCheck{
		ServiceID:      service.ID,
		Status:         "UP",
		HTTPStatusCode: &httpCode,
		ResponseTimeMs: &responseTime,
		ErrorMessage:   &errorMessage,
		CheckedAt:      time.Now(),
	}

	if err := healthRepo.CreateHealthCheck(ctx, healthCheck); err != nil {
		t.Fatalf("Failed to create health check: %v", err)
	}

	result, err := healthRepo.GetHealthCheckByID(ctx, healthCheck.ID)
	if err != nil {
		t.Fatalf("Failed to get health check: %v", err)
	}

	if result.ID != healthCheck.ID {
		t.Errorf("Expected ID %s, got %s", healthCheck.ID, result.ID)
	}
	if result.ServiceID != service.ID {
		t.Errorf("Expected service ID %s, got %s", service.ID, result.ServiceID)
	}
	if result.Status != "UP" {
		t.Errorf("Expected status UP, got %s", result.Status)
	}
	if result.HTTPStatusCode == nil || *result.HTTPStatusCode != 200 {
		t.Error("Expected HTTP status code 200")
	}
	if result.ResponseTimeMs == nil || *result.ResponseTimeMs != 150 {
		t.Error("Expected response time 150")
	}
	if result.ErrorMessage == nil || *result.ErrorMessage != errorMessage {
		t.Error("Expected error message to match")
	}
}

func TestGetHealthCheckByIDNotFound(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	healthRepo := NewHealthCheckRepository(db)

	missingID := "00000000-0000-0000-0000-000000000021"

	result, err := healthRepo.GetHealthCheckByID(ctx, missingID)

	if !errors.Is(err, ErrHealthCheckNotFound) {
		t.Errorf("Expected ErrHealthCheckNotFound, got %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}
}

func TestCreateHealthCheckInvalidService(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	healthRepo := NewHealthCheckRepository(db)

	healthCheck := &models.HealthCheck{
		ServiceID: "00000000-0000-0000-0000-000000000099",
		Status:    "UP",
		CheckedAt: time.Now(),
	}

	err = healthRepo.CreateHealthCheck(ctx, healthCheck)

	if err == nil {
		t.Fatal("Expected error for nonexistent service ID")
	}
}

func TestCreateHealthCheckNullableFields(t *testing.T) {
	db, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)
	healthRepo := NewHealthCheckRepository(db)

	user := &models.User{
		Name:         "Nullable Test User",
		Email:        fmt.Sprintf("health-nullable-%d@example.com", time.Now().UnixNano()),
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
		Name:               "Nullable Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}
	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	healthCheck := &models.HealthCheck{
		ServiceID: service.ID,
		Status:    "DOWN",
		CheckedAt: time.Now(),
	}

	if err := healthRepo.CreateHealthCheck(ctx, healthCheck); err != nil {
		t.Fatalf("Failed to create health check: %v", err)
	}

	result, err := healthRepo.GetHealthCheckByID(ctx, healthCheck.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve health check: %v", err)
	}

	if result.HTTPStatusCode != nil {
		t.Errorf("Expected nil HTTP status code, got %v", *result.HTTPStatusCode)
	}
	if result.ResponseTimeMs != nil {
		t.Errorf("Expected nil response time, got %v", *result.ResponseTimeMs)
	}
	if result.ErrorMessage != nil {
		t.Errorf("Expected nil error message, got %v", *result.ErrorMessage)
	}
}
