package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	user *models.User,
) error {

	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {

	user := &models.User{}

	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	err := r.DB.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}
func (r *UserRepository) GetUserByID(
	ctx context.Context,
	userID string,
) (*models.User, error) {

	user := &models.User{}

	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	err := r.DB.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(
	ctx context.Context,
	user *models.User,
) error {

	query := `
        UPDATE users
        SET name = $1,
            email = $2,
            updated_at = NOW()
        WHERE id = $3
        RETURNING updated_at
    `

	err := r.DB.QueryRow(
		ctx,
		query,
		user.Name,
		user.Email,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
func (r *UserRepository) DeleteUser(
	ctx context.Context,
	userID string,
) error {

	query := `
        DELETE FROM users
        WHERE id = $1
    `

	result, err := r.DB.Exec(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
