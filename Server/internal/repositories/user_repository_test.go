package repositories

import (
	"context"
	"errors"
	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
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

	repo := NewUserRepository(db)

	email := "test-create-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user := &models.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: "test-password-hash",
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1",
			email,
		)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.ID == "" {
		t.Error("Expected generated user ID, got empty ID")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Expected updated_at to be set")
	}
}

func TestGetUserByEmail(t *testing.T) {
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

	repo := NewUserRepository(db)
	// Generate a unique email for every test run.
	email := "test-email-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user := &models.User{
		Name:         "Email Test User",
		Email:        email,
		PasswordHash: "test-password-hash",
	}

	// Register cleanup before inserting the user.
	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1",
			email,
		)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create the test user.
	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Retrieve the user using their email.
	foundUser, err := repo.GetUserByEmail(
		context.Background(),
		user.Email,
	)

	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	// Verify the retrieved user.
	if foundUser.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, foundUser.ID)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	if foundUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
	}
}

func TestGetUserByID(t *testing.T) {
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

	repo := NewUserRepository(db)

	// Generate a unique email.
	email := "test-id-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user := &models.User{
		Name:         "ID Test User",
		Email:        email,
		PasswordHash: "test-password-hash",
	}

	// Register cleanup before creating the user.
	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1",
			email,
		)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create the test user.
	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Retrieve the user using their ID.
	foundUser, err := repo.GetUserByID(
		context.Background(),
		user.ID,
	)

	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	// Verify the retrieved user.
	if foundUser.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, foundUser.ID)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	if foundUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
	}
}

func TestUpdateUser(t *testing.T) {
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

	repo := NewUserRepository(db)

	email := "test-update-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user := &models.User{
		Name:         "Original Name",
		Email:        email,
		PasswordHash: "test-password-hash",
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1 OR id = $2",
			email,
			user.ID,
		)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create a user first.
	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	originalUpdatedAt := user.UpdatedAt

	// Change the user's name and email.
	updatedEmail := "updated-" + email

	user.Name = "Updated Name"
	user.Email = updatedEmail

	// Update the user in PostgreSQL.
	err = repo.UpdateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify the updated details from PostgreSQL.
	foundUser, err := repo.GetUserByID(
		context.Background(),
		user.ID,
	)
	if err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}

	if foundUser.Name != "Updated Name" {
		t.Errorf("Expected name Updated Name, got %s", foundUser.Name)
	}

	if foundUser.Email != updatedEmail {
		t.Errorf("Expected email %s, got %s", updatedEmail, foundUser.Email)
	}

	if !user.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected updated_at to be later than the original timestamp")
	}
}

func TestDeleteUser(t *testing.T) {
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

	repo := NewUserRepository(db)

	email := "test-delete-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user := &models.User{
		Name:         "Delete Test User",
		Email:        email,
		PasswordHash: "test-password-hash",
	}

	// Clean up if the test fails before deletion.
	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM users WHERE email = $1 OR id = $2",
			email,
			user.ID,
		)
		if err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})

	// Create the user first.
	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Delete the user.
	err = repo.DeleteUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify that the user no longer exists.
	foundUser, err := repo.GetUserByID(
		context.Background(),
		user.ID,
	)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if foundUser != nil {
		t.Error("Expected deleted user to be nil")
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
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

	repo := NewUserRepository(db)

	// Search for an email that should not exist.
	_, err = repo.GetUserByEmail(
		context.Background(),
		"missing-user@example.com",
	)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
func TestGetUserByIDNotFound(t *testing.T) {
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

	repo := NewUserRepository(db)

	missingID := "00000000-0000-0000-0000-000000000001"

	user, err := repo.GetUserByID(context.Background(), missingID)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}
func TestUpdateUserNotFound(t *testing.T) {
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

	repo := NewUserRepository(db)

	user := &models.User{
		ID:    "00000000-0000-0000-0000-000000000002",
		Name:  "Missing User",
		Email: "missing-user@example.com",
	}

	err = repo.UpdateUser(context.Background(), user)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
func TestDeleteUserNotFound(t *testing.T) {
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

	repo := NewUserRepository(db)

	missingID := "00000000-0000-0000-0000-000000000003"

	err = repo.DeleteUser(context.Background(), missingID)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
func TestCreateUserDuplicateEmail(t *testing.T) {
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

	repo := NewUserRepository(db)

	email := "duplicate-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	firstUser := &models.User{
		Name:         "First User",
		Email:        email,
		PasswordHash: "first-hash",
	}

	secondUser := &models.User{
		Name:         "Second User",
		Email:        email,
		PasswordHash: "second-hash",
	}

	t.Cleanup(func() {
		_, err := db.Exec(context.Background(),
			"DELETE FROM users WHERE email = $1", email)
		if err != nil {
			t.Errorf("Failed to clean up test users: %v", err)
		}
	})

	// First user should be created successfully.
	if err := repo.CreateUser(context.Background(), firstUser); err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Second user should fail because the email already exists.
	err = repo.CreateUser(context.Background(), secondUser)

	if err == nil {
		t.Fatal("Expected duplicate email error, got nil")
	}
}
