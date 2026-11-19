// Package alertmanager handles webhooks sent from Prometheus Alertmanager,
// transforming incoming alert notifications into normalized events.
package alertmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// WebhookPayload represents the JSON payload dispatched by Alertmanager.
type WebhookPayload struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Status            string            `json:"status"` // "firing" or "resolved"
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

// Alert represents an individual alert within an Alertmanager webhook payload.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// Handler handles Alertmanager webhook HTTP requests.
type Handler struct {
	logger        *slog.Logger
	onEventParsed func(event *normalizer.NormalizedEvent) error
}

// NewHandler creates a new Alertmanager webhook HTTP handler.
func NewHandler(onEventParsed func(event *normalizer.NormalizedEvent) error, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:        logger.With(slog.String("component", "alertmanager-webhook")),
		onEventParsed: onEventParsed,
	}
}

// ServeHTTP parses the Alertmanager webhook and forwards each alert as a normalized event.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		h.logger.Error("failed to read webhook body", slog.String("error", err.Error()))
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("failed to unmarshal alertmanager payload", slog.String("error", err.Error()))
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	for _, alert := range payload.Alerts {
		event := h.parseAlert(alert, body)
		if h.onEventParsed != nil {
			if err := h.onEventParsed(event); err != nil {
				h.logger.Error("failed to forward normalized alert", slog.String("error", err.Error()))
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"accepted","alerts":%d}`, len(payload.Alerts))
}

func (h *Handler) parseAlert(alert Alert, rawPayload []byte) *normalizer.NormalizedEvent {
	severity := normalizer.SeverityMedium
	if sev, ok := alert.Labels["severity"]; ok {
		switch sev {
		case "critical", "page":
			severity = normalizer.SeverityCritical
		case "warning", "high":
			severity = normalizer.SeverityHigh
		case "low":
			severity = normalizer.SeverityLow
		case "info":
			severity = normalizer.SeverityInfo
		}
	}

	title := alert.Annotations["summary"]
	if title == "" {
		title = alert.Labels["alertname"]
	}
	if title == "" {
		title = "Alertmanager Notification"
	}

	desc := alert.Annotations["description"]
	if desc == "" {
		desc = alert.Annotations["message"]
	}

	ts := alert.StartsAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	return &normalizer.NormalizedEvent{
		ID:          uuid.New().String(),
		Source:      "alertmanager",
		Type:        alert.Status, // "firing" or "resolved"
		Severity:    severity,
		Title:       title,
		Description: desc,
		Labels:      alert.Labels,
		Timestamp:   ts,
		RawPayload:  rawPayload,
	}
}
