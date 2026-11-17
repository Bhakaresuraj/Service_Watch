package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceRepository struct {
	DB *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{
		DB: db,
	}
}

func (r *ServiceRepository) CreateService(ctx context.Context, service *models.Service) error {
	query := `
        INSERT INTO services (
            user_id,
            name,
            url,
            monitoring_interval,
            timeout_seconds,
            failure_threshold,
            recovery_threshold
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, current_status, is_enabled, created_at, updated_at
    `

	err := r.DB.QueryRow(
		ctx,
		query,
		service.UserID,
		service.Name,
		service.URL,
		service.MonitoringInterval,
		service.TimeoutSeconds,
		service.FailureThreshold,
		service.RecoveryThreshold,
	).Scan(
		&service.ID,
		&service.CurrentStatus,
		&service.IsEnabled,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}
func (r *ServiceRepository) GetServiceByID(
	ctx context.Context,
	serviceID string,
	userID string,
) (*models.Service, error) {

	service := &models.Service{}

	query := `
        SELECT
            id,
            user_id,
            name,
            url,
            monitoring_interval,
            timeout_seconds,
            failure_threshold,
            recovery_threshold,
            current_status,
            is_enabled,
            last_checked_at,
            created_at,
            updated_at
        FROM services
        WHERE id = $1 AND user_id = $2
    `

	err := r.DB.QueryRow(ctx, query, serviceID, userID).Scan(
		&service.ID,
		&service.UserID,
		&service.Name,
		&service.URL,
		&service.MonitoringInterval,
		&service.TimeoutSeconds,
		&service.FailureThreshold,
		&service.RecoveryThreshold,
		&service.CurrentStatus,
		&service.IsEnabled,
		&service.LastCheckedAt,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}

		return nil, fmt.Errorf("failed to get service by ID: %w", err)
	}

	return service, nil
}
func (r *ServiceRepository) GetServicesByUserID(
	ctx context.Context,
	userID string,
) ([]models.Service, error) {

	query := `
        SELECT
            id,
            user_id,
            name,
            url,
            monitoring_interval,
            timeout_seconds,
            failure_threshold,
            recovery_threshold,
            current_status,
            is_enabled,
            last_checked_at,
            created_at,
            updated_at
        FROM services
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get services by user ID: %w", err)
	}

	defer rows.Close()

	services := make([]models.Service, 0)

	for rows.Next() {
		var service models.Service

		err := rows.Scan(
			&service.ID,
			&service.UserID,
			&service.Name,
			&service.URL,
			&service.MonitoringInterval,
			&service.TimeoutSeconds,
			&service.FailureThreshold,
			&service.RecoveryThreshold,
			&service.CurrentStatus,
			&service.IsEnabled,
			&service.LastCheckedAt,
			&service.CreatedAt,
			&service.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating services: %w", err)
	}

	return services, nil
}
func (r *ServiceRepository) UpdateService(
	ctx context.Context,
	service *models.Service,
) error {

	query := `
        UPDATE services
        SET
            name = $1,
            url = $2,
            monitoring_interval = $3,
            timeout_seconds = $4,
            failure_threshold = $5,
            recovery_threshold = $6,
            is_enabled = $7,
            updated_at = NOW()
        WHERE id = $8 AND user_id = $9
        RETURNING updated_at
    `

	err := r.DB.QueryRow(
		ctx,
		query,
		service.Name,
		service.URL,
		service.MonitoringInterval,
		service.TimeoutSeconds,
		service.FailureThreshold,
		service.RecoveryThreshold,
		service.IsEnabled,
		service.ID,
		service.UserID,
	).Scan(&service.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrServiceNotFound
		}

		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}
func (r *ServiceRepository) DeleteService(
	ctx context.Context,
	serviceID string,
	userID string,
) error {

	query := `
        DELETE FROM services
        WHERE id = $1 AND user_id = $2
    `

	result, err := r.DB.Exec(ctx, query, serviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrServiceNotFound
	}

	return nil
}
