// Package knowledge provides semantic search over historical incidents and runbooks
// using pgvector embeddings in PostgreSQL.
package knowledge

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Runbook represents a standard operating procedure or troubleshooting playbook.
type Runbook struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Steps       []string  `json:"steps"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SimilarIncident represents a historical incident retrieved via vector similarity.
type SimilarIncident struct {
	IncidentID uuid.UUID `json:"incidentId"`
	Title      string    `json:"title"`
	Similarity float64   `json:"similarity"`
	Resolution string    `json:"resolution"`
}

// Service provides knowledge storage and vector retrieval capabilities.
type Service struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewService creates a knowledge service.
func NewService(db *pgxpool.Pool, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		db:     db,
		logger: logger.With(slog.String("component", "knowledge-service")),
	}
}

// StoreIncidentKnowledge indexes a resolved incident and its embedding vector.
func (s *Service) StoreIncidentKnowledge(ctx context.Context, incidentID uuid.UUID, content string, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	query := `
		INSERT INTO knowledge_embeddings (id, incident_id, content, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(ctx, query, uuid.New(), incidentID, content, vec, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert incident knowledge embedding: %w", err)
	}
	s.logger.Info("indexed incident knowledge embedding", slog.String("incident_id", incidentID.String()))
	return nil
}

// SearchSimilarIncidents finds the top-K most similar historical incidents via cosine distance.
func (s *Service) SearchSimilarIncidents(ctx context.Context, embedding []float32, limit int) ([]SimilarIncident, error) {
	if limit <= 0 {
		limit = 5
	}
	vec := pgvector.NewVector(embedding)
	query := `
		SELECT k.incident_id, i.title, 1 - (k.embedding <=> $1) AS similarity, COALESCE(i.description, '')
		FROM knowledge_embeddings k
		JOIN incidents i ON k.incident_id = i.id
		ORDER BY k.embedding <=> $1 ASC
		LIMIT $2
	`
	rows, err := s.db.Query(ctx, query, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("vector similarity search failed: %w", err)
	}
	defer rows.Close()

	var results []SimilarIncident
	for rows.Next() {
		var item SimilarIncident
		if err := rows.Scan(&item.IncidentID, &item.Title, &item.Similarity, &item.Resolution); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// StoreRunbook persists a runbook in PostgreSQL.
func (s *Service) StoreRunbook(ctx context.Context, runbook *Runbook) error {
	if runbook.ID == uuid.Nil {
		runbook.ID = uuid.New()
	}
	now := time.Now().UTC()
	runbook.CreatedAt = now
	runbook.UpdatedAt = now

	query := `
		INSERT INTO runbooks (id, title, description, steps, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(ctx, query, runbook.ID, runbook.Title, runbook.Description, runbook.Steps, runbook.Tags, runbook.CreatedAt, runbook.UpdatedAt)
	if err != nil {
		return fmt.Errorf("store runbook: %w", err)
	}
	return nil
}

// GetRunbook retrieves a runbook by ID.
func (s *Service) GetRunbook(ctx context.Context, id uuid.UUID) (*Runbook, error) {
	query := `
		SELECT id, title, description, steps, tags, created_at, updated_at
		FROM runbooks
		WHERE id = $1
	`
	var rb Runbook
	err := s.db.QueryRow(ctx, query, id).Scan(
		&rb.ID, &rb.Title, &rb.Description, &rb.Steps, &rb.Tags, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get runbook %s: %w", id, err)
	}
	return &rb, nil
}
