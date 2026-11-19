// Package database provides Postgres-backed repositories for domain models.
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AegisOpsAI/aegisops/internal/models"
	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// Repository encapsulates data access across PostgreSQL tables.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new database repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// StoreEvent implements ingestion.EventStore.
func (r *Repository) StoreEvent(ctx context.Context, event *normalizer.NormalizedEvent) error {
	id, err := uuid.Parse(event.ID)
	if err != nil {
		id = uuid.New()
	}

	rawPayload := event.RawPayload
	if len(rawPayload) == 0 {
		rawPayload, _ = json.Marshal(event)
	}

	query := `
		INSERT INTO events (id, source, type, severity, payload, timestamp, normalized, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`
	_, err = r.pool.Exec(ctx, query,
		id, event.Source, event.Type, event.Severity, rawPayload,
		event.Timestamp, true, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	return nil
}

// GetOpenIncidents implements correlation.IncidentStore.
func (r *Repository) GetOpenIncidents(ctx context.Context) ([]*models.Incident, error) {
	query := `
		SELECT id, title, description, severity, status, source, assignee_id, created_at, updated_at, resolved_at
		FROM incidents
		WHERE status IN ('open', 'investigating', 'identified', 'monitoring')
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(
			&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status,
			&inc.Source, &inc.AssigneeID, &inc.CreatedAt, &inc.UpdatedAt, &inc.ResolvedAt,
		); err != nil {
			return nil, err
		}
		incidents = append(incidents, &inc)
	}
	return incidents, nil
}

// CreateIncident implements correlation.IncidentStore.
func (r *Repository) CreateIncident(ctx context.Context, inc *models.Incident) error {
	query := `
		INSERT INTO incidents (id, title, description, severity, status, source, assignee_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		inc.ID, inc.Title, inc.Description, inc.Severity, inc.Status,
		inc.Source, inc.AssigneeID, inc.CreatedAt, inc.UpdatedAt,
	)
	return err
}

// AttachEventToIncident implements correlation.IncidentStore.
func (r *Repository) AttachEventToIncident(ctx context.Context, incidentID uuid.UUID, eventID uuid.UUID) error {
	query := `
		UPDATE events
		SET incident_id = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, incidentID, eventID)
	return err
}

// SaveRemediation implements remediation.Store.
func (r *Repository) SaveRemediation(ctx context.Context, rem *models.Remediation) error {
	query := `
		INSERT INTO remediations (id, incident_id, description, type, status, approved_by, executed_at, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		rem.ID, rem.IncidentID, rem.Description, rem.Type, rem.Status,
		rem.ApprovedBy, rem.ExecutedAt, rem.Result, rem.CreatedAt,
	)
	return err
}

// GetRemediation implements remediation.Store.
func (r *Repository) GetRemediation(ctx context.Context, id uuid.UUID) (*models.Remediation, error) {
	query := `
		SELECT id, incident_id, description, type, status, approved_by, executed_at, result, created_at
		FROM remediations
		WHERE id = $1
	`
	var rem models.Remediation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rem.ID, &rem.IncidentID, &rem.Description, &rem.Type, &rem.Status,
		&rem.ApprovedBy, &rem.ExecutedAt, &rem.Result, &rem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rem, nil
}

// UpdateRemediationStatus implements remediation.Store.
func (r *Repository) UpdateRemediationStatus(ctx context.Context, id uuid.UUID, status models.RemediationStatus, approver *uuid.UUID, result string) error {
	now := time.Now().UTC()
	var execTime *time.Time
	if status == models.RemediationStatusCompleted || status == models.RemediationStatusFailed {
		execTime = &now
	}

	query := `
		UPDATE remediations
		SET status = $1, approved_by = COALESCE($2, approved_by), result = COALESCE(NULLIF($3, ''), result), executed_at = COALESCE($4, executed_at)
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, status, approver, result, execTime, id)
	return err
}
