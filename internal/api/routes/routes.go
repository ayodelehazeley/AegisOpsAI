// Package routes configures and registers all HTTP endpoints for AegisOpsAI.
package routes

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/AegisOpsAI/aegisops/internal/api/handlers"
	"github.com/AegisOpsAI/aegisops/internal/api/middleware"
	"github.com/AegisOpsAI/aegisops/internal/auth"
	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Handlers holds all initialized handler dependencies for route registration.
type Handlers struct {
	Health       *handlers.HealthHandler
	Auth         *handlers.AuthHandler
	Incidents    *handlers.IncidentHandler
	Events       *handlers.EventHandler
	Remediations *handlers.RemediationHandler
	Audit        *handlers.AuditHandler
	Alertmanager http.Handler
	Webhook      http.Handler
}

// Register mounts all platform routes onto the provided chi.Mux.
func Register(r chi.Router, h *Handlers, authSvc *auth.Service, logger *slog.Logger) {
	// Public Health Endpoint
	r.Get("/healthz", h.Health.HealthCheck)
	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	// Public Integrations / Webhooks
	r.Route("/integrations", func(r chi.Router) {
		if h.Alertmanager != nil {
			r.Method(http.MethodPost, "/alertmanager", h.Alertmanager)
		}
		if h.Webhook != nil {
			r.Method(http.MethodPost, "/webhook", h.Webhook)
		}
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth Endpoints
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", h.Auth.Login)
			r.Post("/register", h.Auth.Register)
			r.Post("/refresh", h.Auth.Refresh)
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			if authSvc != nil {
				r.Use(middleware.Auth(authSvc))
			}

			// Incidents
			r.Route("/incidents", func(r chi.Router) {
				r.Get("/", h.Incidents.ListIncidents)
				r.Post("/", h.Incidents.CreateIncident)
				r.Get("/{id}", h.Incidents.GetIncident)
				r.Patch("/{id}/status", h.Incidents.UpdateIncidentStatus)
			})

			// Events
			r.Route("/events", func(r chi.Router) {
				r.Get("/", h.Events.ListEvents)
				r.Post("/ingest", h.Events.IngestEvent)
				r.Get("/{id}", h.Events.GetEvent)
			})

			// Remediations (Requires Operator or Admin role)
			r.Route("/remediations", func(r chi.Router) {
				r.Post("/", h.Remediations.Propose)
				r.Post("/{id}/approve", h.Remediations.Approve)
				r.Post("/{id}/execute", h.Remediations.Execute)
			})

			// Audit Logs (Requires Admin role)
			r.Route("/audit", func(r chi.Router) {
				r.With(middleware.RoleRequired(models.RoleAdmin)).Get("/logs", h.Audit.ListLogs)
			})
		})
	})
}
