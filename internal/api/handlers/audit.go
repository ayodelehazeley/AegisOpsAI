package handlers

import (
	"log/slog"
	"net/http"

	"github.com/AegisOpsAI/aegisops/internal/audit"
)

// AuditHandler serves audit trail querying.
type AuditHandler struct {
	svc    *audit.Service
	logger *slog.Logger
}

// NewAuditHandler creates an AuditHandler instance.
func NewAuditHandler(svc *audit.Service, logger *slog.Logger) *AuditHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditHandler{
		svc:    svc,
		logger: logger.With(slog.String("component", "audit-handler")),
	}
}

// ListLogs retrieves the audit log stream.
func (h *AuditHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	pagination := ParsePagination(r)
	logs, err := h.svc.QueryLogs(r.Context(), pagination.Limit(), pagination.Offset())
	if err != nil {
		h.logger.Error("failed to query audit logs", slog.String("error", err.Error()))
		JSONError(w, http.StatusInternalServerError, "failed to query audit trail")
		return
	}

	JSON(w, http.StatusOK, logs)
}
