// Package webhook provides a generic HTTP webhook receiver supporting HMAC
// signature verification and flexible field extraction into normalized events.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// Config configures generic webhook ingestion.
type Config struct {
	Secret        string `json:"secret,omitempty"`        // Secret for HMAC-SHA256 signature
	AuthToken     string `json:"authToken,omitempty"`     // Static bearer token if used
	DefaultSource string `json:"defaultSource,omitempty"` // Fallback source name
}

// Handler handles incoming webhook HTTP payloads.
type Handler struct {
	cfg           Config
	logger        *slog.Logger
	onEventParsed func(event *normalizer.NormalizedEvent) error
}

// NewHandler creates a new generic webhook handler.
func NewHandler(cfg Config, onEventParsed func(event *normalizer.NormalizedEvent) error, logger *slog.Logger) *Handler {
	if cfg.DefaultSource == "" {
		cfg.DefaultSource = "webhook"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		cfg:           cfg,
		logger:        logger.With(slog.String("component", "generic-webhook")),
		onEventParsed: onEventParsed,
	}
}

// ServeHTTP verifies the webhook and forwards the normalized event.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if err := h.authenticate(r, body); err != nil {
		h.logger.Warn("webhook authentication failed", slog.String("error", err.Error()))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var rawMap map[string]any
	if err := json.Unmarshal(body, &rawMap); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	event := h.normalizeMap(rawMap, body)
	if h.onEventParsed != nil {
		if err := h.onEventParsed(event); err != nil {
			h.logger.Error("failed to process webhook event", slog.String("error", err.Error()))
			http.Error(w, "internal processing error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintln(w, `{"status":"received"}`)
}

func (h *Handler) authenticate(r *http.Request, body []byte) error {
	if h.cfg.AuthToken != "" {
		header := r.Header.Get("Authorization")
		if header != "Bearer "+h.cfg.AuthToken && header != h.cfg.AuthToken {
			return errors.New("invalid authorization token")
		}
	}

	if h.cfg.Secret != "" {
		sig := r.Header.Get("X-Hub-Signature-256")
		if sig == "" {
			sig = r.Header.Get("X-Signature")
		}
		if sig == "" {
			return errors.New("missing signature header")
		}

		mac := hmac.New(sha256.New, []byte(h.cfg.Secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(sig), []byte(expected)) && !hmac.Equal([]byte(sig), []byte("sha256="+expected)) {
			return errors.New("signature mismatch")
		}
	}

	return nil
}

func (h *Handler) normalizeMap(m map[string]any, raw []byte) *normalizer.NormalizedEvent {
	title := "Generic Webhook Event"
	if t, ok := m["title"].(string); ok && t != "" {
		title = t
	} else if summary, ok := m["summary"].(string); ok && summary != "" {
		title = summary
	}

	desc := ""
	if d, ok := m["description"].(string); ok {
		desc = d
	} else if msg, ok := m["message"].(string); ok {
		desc = msg
	}

	sevStr := "medium"
	if s, ok := m["severity"].(string); ok {
		sevStr = s
	}

	source := h.cfg.DefaultSource
	if src, ok := m["source"].(string); ok && src != "" {
		source = src
	}

	return &normalizer.NormalizedEvent{
		ID:          uuid.New().String(),
		Source:      source,
		Type:        "webhook_notification",
		Severity:    normalizer.Severity(sevStr),
		Title:       title,
		Description: desc,
		Timestamp:   time.Now().UTC(),
		RawPayload:  raw,
	}
}
