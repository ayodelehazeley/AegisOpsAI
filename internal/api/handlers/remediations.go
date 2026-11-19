package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/api/middleware"
	"github.com/AegisOpsAI/aegisops/internal/models"
	"github.com/AegisOpsAI/aegisops/internal/remediation"
)

// RemediationHandler serves remediation execution endpoints.
type RemediationHandler struct {
	svc    *remediation.Service
	logger *slog.Logger
}

// NewRemediationHandler creates a RemediationHandler instance.
func NewRemediationHandler(svc *remediation.Service, logger *slog.Logger) *RemediationHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RemediationHandler{
		svc:    svc,
		logger: logger.With(slog.String("component", "remediation-handler")),
	}
}

// ProposeRequest is the JSON payload to suggest a remediation action.
type ProposeRequest struct {
	IncidentID  uuid.UUID `json:"incidentId"`
	ActionType  string    `json:"actionType"`
	Description string    `json:"description"`
}

// Propose handles POST /remediations.
func (h *RemediationHandler) Propose(w http.ResponseWriter, r *http.Request) {
	var req ProposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rem, err := h.svc.ProposeRemediation(r.Context(), req.IncidentID, req.ActionType, req.Description)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, rem)
}

// Approve handles POST /remediations/{id}/approve.
func (h *RemediationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid remediation uuid")
		return
	}

	// Extract user ID from auth context
	claims := middleware.GetUserClaims(r.Context())
	var approverID uuid.UUID
	if claims != nil {
		approverID = claims.UserID
	} else {
		approverID = uuid.New()
	}

	if err := h.svc.ApproveRemediation(r.Context(), id, approverID); err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]any{"id": id, "status": models.RemediationStatusApproved, "approvedBy": approverID})
}

// Execute handles POST /remediations/{id}/execute.
func (h *RemediationHandler) Execute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid remediation uuid")
		return
	}

	rem, err := h.svc.ExecuteRemediation(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, rem)
}
