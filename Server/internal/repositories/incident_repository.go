package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bhakaresuraj/Service_Watch/Server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrIncidentNotFound = errors.New("incident not found")

type IncidentRepository struct {
	DB *pgxpool.Pool
}

func NewIncidentRepository(db *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{DB: db}
}

// CreateIncident creates a new open incident.
func (r *IncidentRepository) CreateIncident(
	ctx context.Context,
	incident *models.Incident,
) error {
	query := `
		INSERT INTO incidents (
			service_id,
			status,
			started_at,
			description
		)
		VALUES ($1, 'OPEN', $2, $3)
		RETURNING id, status, created_at, updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		incident.ServiceID,
		incident.StartedAt,
		incident.Description,
	).Scan(
		&incident.ID,
		&incident.Status,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetOpenIncidentByServiceID retrieves the open incident for a service.
func (r *IncidentRepository) GetOpenIncidentByServiceID(
	ctx context.Context,
	serviceID string,
) (*models.Incident, error) {
	incident := &models.Incident{}

	query := `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			description,
			created_at,
			updated_at
		FROM incidents
		WHERE service_id = $1 AND status = 'OPEN'
	`

	err := r.DB.QueryRow(ctx, query, serviceID).Scan(
		&incident.ID,
		&incident.ServiceID,
		&incident.Status,
		&incident.StartedAt,
		&incident.ResolvedAt,
		&incident.Description,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrIncidentNotFound
		}

		return nil, fmt.Errorf("failed to get open incident: %w", err)
	}

	return incident, nil
}

// GetIncidentsByServiceID retrieves all incidents for a service.
func (r *IncidentRepository) GetIncidentsByServiceID(
	ctx context.Context,
	serviceID string,
) ([]models.Incident, error) {
	query := `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			description,
			created_at,
			updated_at
		FROM incidents
		WHERE service_id = $1
		ORDER BY started_at DESC
	`

	rows, err := r.DB.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}
	defer rows.Close()

	incidents := make([]models.Incident, 0)

	for rows.Next() {
		var incident models.Incident

		err := rows.Scan(
			&incident.ID,
			&incident.ServiceID,
			&incident.Status,
			&incident.StartedAt,
			&incident.ResolvedAt,
			&incident.Description,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, incident)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating incidents: %w", err)
	}

	return incidents, nil
}

// GetIncidentByID retrieves one incident.
func (r *IncidentRepository) GetIncidentByID(
	ctx context.Context,
	incidentID string,
) (*models.Incident, error) {
	incident := &models.Incident{}

	query := `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			description,
			created_at,
			updated_at
		FROM incidents
		WHERE id = $1
	`

	err := r.DB.QueryRow(ctx, query, incidentID).Scan(
		&incident.ID,
		&incident.ServiceID,
		&incident.Status,
		&incident.StartedAt,
		&incident.ResolvedAt,
		&incident.Description,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrIncidentNotFound
		}

		return nil, fmt.Errorf("failed to get incident by ID: %w", err)
	}

	return incident, nil
}

// ResolveIncident marks an open incident as resolved.
func (r *IncidentRepository) ResolveIncident(
	ctx context.Context,
	incidentID string,
) error {
	query := `
		UPDATE incidents
		SET
			status = 'RESOLVED',
			resolved_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status = 'OPEN'
		RETURNING id
	`

	var id string

	err := r.DB.QueryRow(ctx, query, incidentID).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrIncidentNotFound
		}

		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	return nil
}
