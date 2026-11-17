
package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrHealthCheckNotFound = errors.New("health check not found")

type HealthCheckRepository struct {
	DB *pgxpool.Pool
}

func NewHealthCheckRepository(db *pgxpool.Pool) *HealthCheckRepository {
	return &HealthCheckRepository{DB: db}
}

// CreateHealthCheck saves a monitoring result.
func (r *HealthCheckRepository) CreateHealthCheck(
	ctx context.Context,
	healthCheck *models.HealthCheck,
) error {
	query := `
		INSERT INTO health_checks (
			service_id,
			status,
			http_status_code,
			response_time_ms,
			error_message,
			checked_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		healthCheck.ServiceID,
		healthCheck.Status,
		healthCheck.HTTPStatusCode,
		healthCheck.ResponseTimeMs,
		healthCheck.ErrorMessage,
		healthCheck.CheckedAt,
	).Scan(
		&healthCheck.ID,
		&healthCheck.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create health check: %w", err)
	}

	return nil
}

// GetHealthChecksByServiceID retrieves monitoring history.
func (r *HealthCheckRepository) GetHealthChecksByServiceID(
	ctx context.Context,
	serviceID string,
) ([]models.HealthCheck, error) {
	query := `
		SELECT
			id,
			service_id,
			status,
			http_status_code,
			response_time_ms,
			error_message,
			checked_at,
			created_at
		FROM health_checks
		WHERE service_id = $1
		ORDER BY checked_at DESC
	`

	rows, err := r.DB.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health checks: %w", err)
	}
	defer rows.Close()

	healthChecks := make([]models.HealthCheck, 0)

	for rows.Next() {
		var healthCheck models.HealthCheck

		err := rows.Scan(
			&healthCheck.ID,
			&healthCheck.ServiceID,
			&healthCheck.Status,
			&healthCheck.HTTPStatusCode,
			&healthCheck.ResponseTimeMs,
			&healthCheck.ErrorMessage,
			&healthCheck.CheckedAt,
			&healthCheck.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan health check: %w", err)
		}

		healthChecks = append(healthChecks, healthCheck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating health checks: %w", err)
	}

	return healthChecks, nil
}

// GetHealthCheckByID retrieves one health check.
func (r *HealthCheckRepository) GetHealthCheckByID(
	ctx context.Context,
	healthCheckID string,
) (*models.HealthCheck, error) {
	healthCheck := &models.HealthCheck{}

	query := `
		SELECT
			id,
			service_id,
			status,
			http_status_code,
			response_time_ms,
			error_message,
			checked_at,
			created_at
		FROM health_checks
		WHERE id = $1
	`

	err := r.DB.QueryRow(ctx, query, healthCheckID).Scan(
		&healthCheck.ID,
		&healthCheck.ServiceID,
		&healthCheck.Status,
		&healthCheck.HTTPStatusCode,
		&healthCheck.ResponseTimeMs,
		&healthCheck.ErrorMessage,
		&healthCheck.CheckedAt,
		&healthCheck.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHealthCheckNotFound
		}

		return nil, fmt.Errorf("failed to get health check by ID: %w", err)
	}

	return healthCheck, nil
}
