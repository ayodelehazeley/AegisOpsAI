// Package normalizer provides event normalization for AegisOpsAI.
// It converts events from heterogeneous sources into a common schema
// for unified processing by the correlation and detection engines.
package normalizer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Severity represents the severity level of a normalized event.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
	SeverityUnknown  Severity = "unknown"
)

// NormalizedEvent is the canonical event schema used throughout AegisOpsAI.
// All incoming events are normalized into this structure regardless of source.
type NormalizedEvent struct {
	// ID is a unique identifier for this event.
	ID string `json:"id"`
	// Source is the origin system (e.g., "prometheus", "alertmanager", "webhook").
	Source string `json:"source"`
	// Type categorizes the event (e.g., "alert", "metric", "heartbeat", "change").
	Type string `json:"type"`
	// Severity indicates the urgency level.
	Severity Severity `json:"severity"`
	// Title is a short human-readable summary.
	Title string `json:"title"`
	// Description provides additional detail about the event.
	Description string `json:"description"`
	// Labels are key-value metadata for grouping, routing, and correlation.
	Labels map[string]string `json:"labels,omitempty"`
	// Metrics are key-value numeric data associated with the event.
	Metrics map[string]float64 `json:"metrics,omitempty"`
	// Timestamp is when the event originally occurred.
	Timestamp time.Time `json:"timestamp"`
	// RawPayload is the original event data, preserved for audit and debugging.
	RawPayload json.RawMessage `json:"raw_payload,omitempty"`
}

// Normalizer defines the interface for normalizing raw event payloads
// from a specific source into the canonical NormalizedEvent schema.
type Normalizer interface {
	// Normalize converts a raw event payload from the given source into a NormalizedEvent.
	Normalize(ctx context.Context, source string, payload json.RawMessage) (*NormalizedEvent, error)
}

// SourceNormalizer handles normalization for a single event source type.
type SourceNormalizer interface {
	// NormalizePayload converts a raw payload into one or more NormalizedEvents.
	NormalizePayload(ctx context.Context, payload json.RawMessage) ([]*NormalizedEvent, error)
	// SourceName returns the name of the source this normalizer handles.
	SourceName() string
}

// Registry maps source names to their specific normalizers.
type Registry struct {
	normalizers map[string]SourceNormalizer
}

// NewRegistry creates a new normalizer Registry with the given source normalizers.
func NewRegistry(normalizers ...SourceNormalizer) *Registry {
	r := &Registry{
		normalizers: make(map[string]SourceNormalizer, len(normalizers)),
	}
	for _, n := range normalizers {
		r.normalizers[n.SourceName()] = n
	}
	return r
}

// Register adds a source normalizer to the registry.
func (r *Registry) Register(n SourceNormalizer) {
	r.normalizers[n.SourceName()] = n
}

// Normalize detects the source and applies the appropriate normalizer.
// If no source-specific normalizer is registered, it falls back to the generic normalizer.
func (r *Registry) Normalize(ctx context.Context, source string, payload json.RawMessage) (*NormalizedEvent, error) {
	normalizedSource := strings.ToLower(strings.TrimSpace(source))

	if sn, ok := r.normalizers[normalizedSource]; ok {
		events, err := sn.NormalizePayload(ctx, payload)
		if err != nil {
			return nil, fmt.Errorf("normalize %s event: %w", normalizedSource, err)
		}
		if len(events) == 0 {
			return nil, fmt.Errorf("normalizer for %s returned no events", normalizedSource)
		}
		// Return the first event; multi-event payloads should be handled by callers
		// using NormalizeAll.
		return events[0], nil
	}

	// Fall back to generic normalization.
	return normalizeGeneric(ctx, normalizedSource, payload)
}

// NormalizeAll normalizes a payload into potentially multiple events.
// Useful for sources like Alertmanager that send batched alerts.
func (r *Registry) NormalizeAll(ctx context.Context, source string, payload json.RawMessage) ([]*NormalizedEvent, error) {
	normalizedSource := strings.ToLower(strings.TrimSpace(source))

	if sn, ok := r.normalizers[normalizedSource]; ok {
		events, err := sn.NormalizePayload(ctx, payload)
		if err != nil {
			return nil, fmt.Errorf("normalize %s events: %w", normalizedSource, err)
		}
		return events, nil
	}

	event, err := normalizeGeneric(ctx, normalizedSource, payload)
	if err != nil {
		return nil, err
	}
	return []*NormalizedEvent{event}, nil
}

// normalizeGeneric provides a best-effort normalization for unknown sources.
func normalizeGeneric(_ context.Context, source string, payload json.RawMessage) (*NormalizedEvent, error) {
	// Attempt to extract common fields from a JSON object.
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		// Not valid JSON — wrap as-is.
		return &NormalizedEvent{
			ID:         uuid.New().String(),
			Source:     source,
			Type:       "raw",
			Severity:   SeverityUnknown,
			Title:      fmt.Sprintf("Raw event from %s", source),
			Timestamp:  time.Now().UTC(),
			RawPayload: payload,
		}, nil
	}

	event := &NormalizedEvent{
		ID:         uuid.New().String(),
		Source:     source,
		Type:       extractString(raw, "type", "event"),
		Severity:   parseSeverity(extractString(raw, "severity", "unknown")),
		Title:      extractString(raw, "title", extractString(raw, "summary", extractString(raw, "name", "Unknown event"))),
		Description: extractString(raw, "description", extractString(raw, "message", "")),
		Labels:     extractStringMap(raw, "labels"),
		Timestamp:  time.Now().UTC(),
		RawPayload: payload,
	}

	// Try to parse a timestamp from the payload.
	if ts, ok := raw["timestamp"]; ok {
		if tsStr, ok := ts.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				event.Timestamp = parsed
			}
		}
	}

	return event, nil
}

// parseSeverity converts a string to a Severity constant.
func parseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical", "fatal", "emergency", "page":
		return SeverityCritical
	case "high", "error", "severe", "major":
		return SeverityHigh
	case "medium", "warning", "warn":
		return SeverityMedium
	case "low", "minor":
		return SeverityLow
	case "info", "informational", "notice", "debug":
		return SeverityInfo
	default:
		return SeverityUnknown
	}
}

// extractString extracts a string value from a map, returning the default if not found.
func extractString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return defaultVal
}

// extractStringMap extracts a map[string]string from a map's key, if present.
func extractStringMap(m map[string]interface{}, key string) map[string]string {
	v, ok := m[key]
	if !ok {
		return nil
	}

	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]string, len(obj))
	for k, val := range obj {
		if s, ok := val.(string); ok {
			result[k] = s
		} else {
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}
