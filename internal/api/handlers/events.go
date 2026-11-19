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

// IngestionService defines ingestion forwarding capability.
type IngestionService interface {
	PublishRawEvent(source string, payload []byte) error
}

// EventHandler serves event-related endpoints.
type EventHandler struct {
	db       *pgxpool.Pool
	ingester IngestionService
	logger   *slog.Logger
}

// NewEventHandler creates an EventHandler instance.
func NewEventHandler(db *pgxpool.Pool, ingester IngestionService, logger *slog.Logger) *EventHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &EventHandler{
		db:       db,
		ingester: ingester,
		logger:   logger.With(slog.String("component", "event-handler")),
	}
}

// ListEvents lists ingested events with pagination.
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	pagination := ParsePagination(r)
	query := `
		SELECT id, source, type, severity, payload, timestamp, incident_id, normalized, created_at
		FROM events
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := h.db.Query(r.Context(), query, pagination.Limit(), pagination.Offset())
	if err != nil {
		h.logger.Error("failed to query events", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "failed to query events")
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var ev models.Event
		if err := rows.Scan(
			&ev.ID, &ev.Source, &ev.Type, &ev.Severity, &ev.Payload,
			&ev.Timestamp, &ev.IncidentID, &ev.Normalized, &ev.CreatedAt,
		); err != nil {
			JSONError(w, http.StatusInternalServerError, "failed to scan events")
			return
		}
		events = append(events, ev)
	}

	JSON(w, http.StatusOK, events)
}

// GetEvent retrieves a single event by UUID.
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid event uuid")
		return
	}

	var ev models.Event
	query := `
		SELECT id, source, type, severity, payload, timestamp, incident_id, normalized, created_at
		FROM events
		WHERE id = $1
	`
	err = h.db.QueryRow(r.Context(), query, id).Scan(
		&ev.ID, &ev.Source, &ev.Type, &ev.Severity, &ev.Payload,
		&ev.Timestamp, &ev.IncidentID, &ev.Normalized, &ev.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			JSONError(w, http.StatusNotFound, "event not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	JSON(w, http.StatusOK, ev)
}

// IngestEvent accepts a direct event payload and queues it.
func (h *EventHandler) IngestEvent(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	source := r.URL.Query().Get("source")
	if source == "" {
		source = "api-direct"
	}

	if h.ingester != nil {
		if err := h.ingester.PublishRawEvent(source, body); err != nil {
			h.logger.Error("failed to publish raw event", slog.String("error", err.Error()))
			JSONError(w, http.StatusInternalServerError, "failed to enqueue event")
			return
		}
	}

	JSON(w, http.StatusAccepted, map[string]string{
		"status": "queued",
		"source": source,
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
