// Package audit maintains an immutable, traceable log of all operational decisions,
// incident transitions, and remediation executions.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Service provides audit logging and retrieval.
type Service struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewService creates a new audit service.
func NewService(db *pgxpool.Pool, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		db:     db,
		logger: logger.With(slog.String("component", "audit-service")),
	}
}

// Log records an operational event into the persistent audit trail.
func (s *Service) Log(ctx context.Context, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details any) error {
	var detailsBytes []byte
	var err error
	if details != nil {
		detailsBytes, err = json.Marshal(details)
		if err != nil {
			s.logger.Warn("failed to marshal audit details, storing raw string", slog.String("error", err.Error()))
			detailsBytes = []byte(fmt.Sprintf(`{"raw":"%v"}`, details))
		}
	} else {
		detailsBytes = []byte("{}")
	}

	entryID := uuid.New()
	now := time.Now().UTC()

	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = s.db.Exec(ctx, query, entryID, userID, action, resourceType, resourceID, detailsBytes, now)
	if err != nil {
		s.logger.Error("failed to record audit entry", slog.String("error", err.Error()))
		return fmt.Errorf("record audit log: %w", err)
	}

	s.logger.Info("audit event recorded",
		slog.String("action", action),
		slog.String("resource_type", resourceType),
	)
	return nil
}

// QueryLogs retrieves audit log entries with optional pagination.
func (s *Service) QueryLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, timestamp
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.ResourceType, &l.ResourceID, &l.Details, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
