package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// HealthHandler provides health check endpoints for the API server.
type HealthHandler struct {
	db     *pgxpool.Pool
	nc     *nats.Conn
	logger *slog.Logger
}

// NewHealthHandler creates a new HealthHandler with database and NATS dependencies.
func NewHealthHandler(db *pgxpool.Pool, nc *nats.Conn, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		nc:     nc,
		logger: logger,
	}
}

// ComponentStatus represents the health status of a single system component.
type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthResponse is the response body for the health check endpoint.
type HealthResponse struct {
	Status     string            `json:"status"`
	Components []ComponentStatus `json:"components"`
}

// HealthCheck handles GET /healthz requests by checking database and NATS connectivity.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5_000_000_000) // 5 seconds
	defer cancel()

	overall := "healthy"
	components := make([]ComponentStatus, 0, 2)

	// Check PostgreSQL connectivity.
	dbStatus := ComponentStatus{Name: "postgresql", Status: "healthy"}
	if err := h.db.Ping(ctx); err != nil {
		dbStatus.Status = "unhealthy"
		dbStatus.Message = err.Error()
		overall = "unhealthy"
		h.logger.Error("health check: database ping failed", slog.String("error", err.Error()))
	}
	components = append(components, dbStatus)

	// Check NATS connectivity.
	natsStatus := ComponentStatus{Name: "nats", Status: "healthy"}
	if h.nc == nil || !h.nc.IsConnected() {
		natsStatus.Status = "degraded"
		natsStatus.Message = "NATS broker not connected (using direct pipeline)"
		h.logger.Debug("health check: NATS running in degraded/direct mode")
	}
	components = append(components, natsStatus)

	resp := HealthResponse{
		Status:     overall,
		Components: components,
	}

	statusCode := http.StatusOK
	if overall != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	JSON(w, statusCode, resp)
}
