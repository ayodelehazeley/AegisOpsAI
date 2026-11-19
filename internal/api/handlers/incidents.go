package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// IncidentHandler serves incident management REST endpoints.
type IncidentHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewIncidentHandler creates an IncidentHandler instance.
func NewIncidentHandler(db *pgxpool.Pool, logger *slog.Logger) *IncidentHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &IncidentHandler{
		db:     db,
		logger: logger.With(slog.String("component", "incident-handler")),
	}
}

// ListIncidents retrieves incidents with optional status/severity filters and pagination.
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	pagination := ParsePagination(r)
	statusFilter := r.URL.Query().Get("status")
	severityFilter := r.URL.Query().Get("severity")

	query := `
		SELECT id, title, description, severity, status, source, assignee_id, created_at, updated_at, resolved_at
		FROM incidents
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR severity = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := h.db.Query(r.Context(), query, statusFilter, severityFilter, pagination.Limit(), pagination.Offset())
	if err != nil {
		h.logger.Error("failed to query incidents", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "failed to query incidents")
		return
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(
			&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status,
			&inc.Source, &inc.AssigneeID, &inc.CreatedAt, &inc.UpdatedAt, &inc.ResolvedAt,
		); err != nil {
			h.logger.Error("failed to scan incident row", slog.String("error", err.Error()))
			JSONError(w, http.StatusInternalServerError, "failed to scan incidents")
			return
		}
		incidents = append(incidents, inc)
	}

	JSON(w, http.StatusOK, incidents)
}

// GetIncident returns a single incident by UUID.
func (h *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid incident uuid")
		return
	}

	var inc models.Incident
	query := `
		SELECT id, title, description, severity, status, source, assignee_id, created_at, updated_at, resolved_at
		FROM incidents
		WHERE id = $1
	`
	err = h.db.QueryRow(r.Context(), query, id).Scan(
		&inc.ID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status,
		&inc.Source, &inc.AssigneeID, &inc.CreatedAt, &inc.UpdatedAt, &inc.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			JSONError(w, http.StatusNotFound, "incident not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	JSON(w, http.StatusOK, inc)
}

// CreateIncident creates a new incident manually.
func (h *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var inc models.Incident
	if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if inc.Title == "" {
		JSONError(w, http.StatusBadRequest, "title is required")
		return
	}

	inc.ID = uuid.New()
	now := time.Now().UTC()
	inc.CreatedAt = now
	inc.UpdatedAt = now
	if inc.Status == "" {
		inc.Status = models.IncidentStatusOpen
	}
	if inc.Severity == "" {
		inc.Severity = models.SeverityMedium
	}

	query := `
		INSERT INTO incidents (id, title, description, severity, status, source, assignee_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := h.db.Exec(r.Context(), query,
		inc.ID, inc.Title, inc.Description, inc.Severity, inc.Status,
		inc.Source, inc.AssigneeID, inc.CreatedAt, inc.UpdatedAt,
	)
	if err != nil {
		h.logger.Error("failed to create incident", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "failed to persist incident")
		return
	}

	JSON(w, http.StatusCreated, inc)
}

// UpdateIncidentStatus updates the lifecycle state of an incident.
func (h *IncidentHandler) UpdateIncidentStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid incident uuid")
		return
	}

	var payload struct {
		Status models.IncidentStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	now := time.Now().UTC()
	var resolvedAt *time.Time
	if payload.Status == models.IncidentStatusResolved {
		resolvedAt = &now
	}

	query := `
		UPDATE incidents
		SET status = $1, updated_at = $2, resolved_at = COALESCE($3, resolved_at)
		WHERE id = $4
	`
	res, err := h.db.Exec(r.Context(), query, payload.Status, now, resolvedAt, id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update status")
		return
	}
	if res.RowsAffected() == 0 {
		JSONError(w, http.StatusNotFound, "incident not found")
		return
	}

	JSON(w, http.StatusOK, map[string]any{"id": id, "status": payload.Status, "updatedAt": now})
}
