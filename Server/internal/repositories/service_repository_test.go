package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestCreateService(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create a test user first.
	email := fmt.Sprintf("service-test-%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Name:         "Service Test User",
		Email:        email,
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Delete the test user after the test.
	// Related services are deleted through ON DELETE CASCADE.
	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	service := &models.Service{
		UserID:             user.ID,
		Name:               "My Portfolio",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	err = serviceRepo.CreateService(ctx, service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Verify database-generated values.
	if service.ID == "" {
		t.Error("Expected service ID, got empty string")
	}

	if service.CurrentStatus != "UNKNOWN" {
		t.Errorf("Expected status UNKNOWN, got %s", service.CurrentStatus)
	}

	if !service.IsEnabled {
		t.Error("Expected service to be enabled")
	}

	if service.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if service.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}
func TestGetServiceByID(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create a test user.
	email := fmt.Sprintf("get-service-%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Name:         "Service Test User",
		Email:        email,
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create a test service.
	service := &models.Service{
		UserID:             user.ID,
		Name:               "My Portfolio",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Retrieve the service using its ID and owner's ID.
	foundService, err := serviceRepo.GetServiceByID(ctx, service.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	// Verify the retrieved service.
	if foundService.ID != service.ID {
		t.Errorf("Expected ID %s, got %s", service.ID, foundService.ID)
	}

	if foundService.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, foundService.UserID)
	}

	if foundService.Name != service.Name {
		t.Errorf("Expected Name %s, got %s", service.Name, foundService.Name)
	}

	if foundService.URL != service.URL {
		t.Errorf("Expected URL %s, got %s", service.URL, foundService.URL)
	}
}
func TestGetServicesByUserID(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create a test user.
	email := fmt.Sprintf("get-services-%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Name:         "Multiple Services User",
		Email:        email,
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create the first service.
	service1 := &models.Service{
		UserID:             user.ID,
		Name:               "Portfolio",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service1); err != nil {
		t.Fatalf("Failed to create first service: %v", err)
	}

	// Create the second service.
	service2 := &models.Service{
		UserID:             user.ID,
		Name:               "GitHub",
		URL:                "https://github.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service2); err != nil {
		t.Fatalf("Failed to create second service: %v", err)
	}

	// Retrieve all services belonging to the user.
	services, err := serviceRepo.GetServicesByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get services: %v", err)
	}

	// Verify the number of services.
	if len(services) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(services))
	}

	// Verify that both services belong to the user.
	for _, service := range services {
		if service.UserID != user.ID {
			t.Errorf("Expected UserID %s, got %s", user.ID, service.UserID)
		}
	}
}
func TestUpdateService(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create a test user.
	email := fmt.Sprintf("update-service-%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Name:         "Update Service User",
		Email:        email,
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create a service.
	service := &models.Service{
		UserID:             user.ID,
		Name:               "Old Portfolio",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	originalUpdatedAt := service.UpdatedAt

	// Update service details.
	service.Name = "Updated Portfolio"
	service.URL = "https://updated-example.com"
	service.MonitoringInterval = 120
	service.TimeoutSeconds = 15
	service.FailureThreshold = 4
	service.RecoveryThreshold = 3
	service.IsEnabled = false

	time.Sleep(10 * time.Millisecond)

	if err := serviceRepo.UpdateService(ctx, service); err != nil {
		t.Fatalf("Failed to update service: %v", err)
	}

	// Retrieve the updated service.
	updatedService, err := serviceRepo.GetServiceByID(ctx, service.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated service: %v", err)
	}

	// Verify updated values.
	if updatedService.Name != "Updated Portfolio" {
		t.Errorf("Expected updated name, got %s", updatedService.Name)
	}

	if updatedService.URL != "https://updated-example.com" {
		t.Errorf("Expected updated URL, got %s", updatedService.URL)
	}

	if updatedService.MonitoringInterval != 120 {
		t.Errorf("Expected monitoring interval 120, got %d", updatedService.MonitoringInterval)
	}

	if updatedService.TimeoutSeconds != 15 {
		t.Errorf("Expected timeout 15, got %d", updatedService.TimeoutSeconds)
	}

	if updatedService.FailureThreshold != 4 {
		t.Errorf("Expected failure threshold 4, got %d", updatedService.FailureThreshold)
	}

	if updatedService.RecoveryThreshold != 3 {
		t.Errorf("Expected recovery threshold 3, got %d", updatedService.RecoveryThreshold)
	}

	if updatedService.IsEnabled {
		t.Error("Expected service to be disabled")
	}

	if !updatedService.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be later than original UpdatedAt")
	}
}
func TestDeleteService(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create a test user.
	email := fmt.Sprintf("delete-service-%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Name:         "Delete Service User",
		Email:        email,
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create a service.
	service := &models.Service{
		UserID:             user.ID,
		Name:               "Service To Delete",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Delete the service.
	err = serviceRepo.DeleteService(ctx, service.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete service: %v", err)
	}

	// Verify that the service no longer exists for this user.
	deletedService, err := serviceRepo.GetServiceByID(ctx, service.ID, user.ID)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}

	if deletedService != nil {
		t.Errorf("Expected nil service, got %v", deletedService)
	}
}
func TestGetServiceByIDWrongUser(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create the service owner.
	owner := &models.User{
		Name:         "Service Owner",
		Email:        fmt.Sprintf("owner-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, owner); err != nil {
		t.Fatalf("Failed to create owner: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", owner.ID)
		if err != nil {
			t.Errorf("Failed to clean up owner: %v", err)
		}
	})

	// Create another user.
	otherUser := &models.User{
		Name:         "Another User",
		Email:        fmt.Sprintf("other-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, otherUser); err != nil {
		t.Fatalf("Failed to create another user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", otherUser.ID)
		if err != nil {
			t.Errorf("Failed to clean up another user: %v", err)
		}
	})

	// Create a service belonging to the owner.
	service := &models.Service{
		UserID:             owner.ID,
		Name:               "Owner's Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Another user tries to access the owner's service.
	foundService, err := serviceRepo.GetServiceByID(
		ctx,
		service.ID,
		otherUser.ID,
	)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}

	if foundService != nil {
		t.Errorf("Expected nil service, got %v", foundService)
	}
}
func TestUpdateServiceWrongUser(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create the service owner.
	owner := &models.User{
		Name:         "Service Owner",
		Email:        fmt.Sprintf("update-owner-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, owner); err != nil {
		t.Fatalf("Failed to create owner: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", owner.ID)
		if err != nil {
			t.Errorf("Failed to clean up owner: %v", err)
		}
	})

	// Create another user.
	otherUser := &models.User{
		Name:         "Another User",
		Email:        fmt.Sprintf("update-other-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, otherUser); err != nil {
		t.Fatalf("Failed to create another user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", otherUser.ID)
		if err != nil {
			t.Errorf("Failed to clean up another user: %v", err)
		}
	})

	// Create a service belonging to the owner.
	service := &models.Service{
		UserID:             owner.ID,
		Name:               "Original Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Another user attempts to update the service.
	service.Name = "Hacked Website"
	service.UserID = otherUser.ID

	err = serviceRepo.UpdateService(ctx, service)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}

	// Verify that the original service was not modified.
	foundService, err := serviceRepo.GetServiceByID(
		ctx,
		service.ID,
		owner.ID,
	)

	if err != nil {
		t.Fatalf("Failed to retrieve original service: %v", err)
	}

	if foundService.Name != "Original Website" {
		t.Errorf("Expected original name to remain unchanged, got %s", foundService.Name)
	}
}
func TestDeleteServiceWrongUser(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	// Create the service owner.
	owner := &models.User{
		Name:         "Service Owner",
		Email:        fmt.Sprintf("delete-owner-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, owner); err != nil {
		t.Fatalf("Failed to create owner: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", owner.ID)
		if err != nil {
			t.Errorf("Failed to clean up owner: %v", err)
		}
	})

	// Create another user.
	otherUser := &models.User{
		Name:         "Another User",
		Email:        fmt.Sprintf("delete-other-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "test-hash",
	}

	if err := userRepo.CreateUser(ctx, otherUser); err != nil {
		t.Fatalf("Failed to create another user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", otherUser.ID)
		if err != nil {
			t.Errorf("Failed to clean up another user: %v", err)
		}
	})

	// Create a service belonging to the owner.
	service := &models.Service{
		UserID:             owner.ID,
		Name:               "Owner Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	if err := serviceRepo.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Another user attempts to delete the owner's service.
	err = serviceRepo.DeleteService(ctx, service.ID, otherUser.ID)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}

	// Verify that the original service still exists.
	foundService, err := serviceRepo.GetServiceByID(
		ctx,
		service.ID,
		owner.ID,
	)

	if err != nil {
		t.Fatalf("Expected owner's service to still exist, got error: %v", err)
	}

	if foundService.ID != service.ID {
		t.Errorf("Expected service ID %s, got %s", service.ID, foundService.ID)
	}
}
func TestGetServiceByIDNotFound(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	user := &models.User{
		Name:         "Test User",
		Email:        fmt.Sprintf("service-not-found-%d@example.com", time.Now().UnixNano()),
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

	// This is a valid UUID, but no service has this ID.
	missingServiceID := "00000000-0000-0000-0000-000000000011"

	service, err := serviceRepo.GetServiceByID(ctx, missingServiceID, user.ID)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}

	if service != nil {
		t.Errorf("Expected nil service, got %+v", service)
	}
}
func TestUpdateServiceNotFound(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	user := &models.User{
		Name:         "Test User",
		Email:        fmt.Sprintf("update-not-found-%d@example.com", time.Now().UnixNano()),
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

	// Create a service object with an ID that does not exist.
	service := &models.Service{
		ID:                 "00000000-0000-0000-0000-000000000012",
		UserID:             user.ID,
		Name:               "Updated Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
		IsEnabled:          true,
	}

	err = serviceRepo.UpdateService(ctx, service)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}
}
func TestDeleteServiceNotFound(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	user := &models.User{
		Name:         "Test User",
		Email:        fmt.Sprintf("delete-not-found-%d@example.com", time.Now().UnixNano()),
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

	missingServiceID := "00000000-0000-0000-0000-000000000013"

	err = serviceRepo.DeleteService(ctx, missingServiceID, user.ID)

	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Expected ErrServiceNotFound, got %v", err)
	}
}
func TestGetServicesByUserIDNoServices(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	serviceRepo := NewServiceRepository(db)

	user := &models.User{
		Name:         "Empty Services User",
		Email:        fmt.Sprintf("empty-services-%d@example.com", time.Now().UnixNano()),
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

	services, err := serviceRepo.GetServicesByUserID(ctx, user.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(services))
	}

	if services == nil {
		t.Error("Expected an empty slice, got nil")
	}
}
func TestCreateServiceInvalidUser(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()
	serviceRepo := NewServiceRepository(db)

	service := &models.Service{
		UserID:             "00000000-0000-0000-0000-000000000099",
		Name:               "Invalid Owner Website",
		URL:                "https://example.com",
		MonitoringInterval: 60,
		TimeoutSeconds:     10,
		FailureThreshold:   3,
		RecoveryThreshold:  2,
	}

	err = serviceRepo.CreateService(ctx, service)

	if err == nil {
		t.Fatal("Expected an error when creating a service with an invalid user ID")
	}
}
