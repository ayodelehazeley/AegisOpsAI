package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// StreamConfig defines the configuration for a single JetStream stream.
type StreamConfig struct {
	// Name is the stream name.
	Name string
	// Subjects is the list of subjects this stream captures.
	Subjects []string
	// Retention is the message retention policy.
	Retention nats.RetentionPolicy
	// MaxAge is the maximum age of messages in the stream.
	MaxAge time.Duration
	// MaxMsgs is the maximum number of messages in the stream. -1 for unlimited.
	MaxMsgs int64
	// MaxBytes is the maximum total bytes for the stream. -1 for unlimited.
	MaxBytes int64
	// Storage is the storage type (file or memory).
	Storage nats.StorageType
	// Replicas is the number of stream replicas for HA.
	Replicas int
	// Description is a human-readable description of the stream.
	Description string
}

// DefaultStreams returns the default stream configurations for AegisOpsAI.
func DefaultStreams() []StreamConfig {
	return []StreamConfig{
		{
			Name:        "EVENTS",
			Subjects:    []string{"events.>"},
			Retention:   nats.WorkQueuePolicy,
			MaxAge:      72 * time.Hour,
			MaxMsgs:     -1,
			MaxBytes:    -1,
			Storage:     nats.FileStorage,
			Replicas:    1,
			Description: "Raw incoming events from all sources (prometheus, alertmanager, webhooks, etc.)",
		},
		{
			Name:        "INCIDENTS",
			Subjects:    []string{"incidents.>"},
			Retention:   nats.LimitsPolicy,
			MaxAge:      30 * 24 * time.Hour,
			MaxMsgs:     -1,
			MaxBytes:    -1,
			Storage:     nats.FileStorage,
			Replicas:    1,
			Description: "Incident lifecycle events (created, updated, escalated, resolved, closed)",
		},
		{
			Name:        "ACTIONS",
			Subjects:    []string{"actions.>"},
			Retention:   nats.WorkQueuePolicy,
			MaxAge:      7 * 24 * time.Hour,
			MaxMsgs:     -1,
			MaxBytes:    -1,
			Storage:     nats.FileStorage,
			Replicas:    1,
			Description: "Remediation action events (proposed, approved, executing, completed, failed)",
		},
	}
}

// JetStreamManager defines the interface for managing JetStream streams.
// This abstraction enables testing without a real NATS connection.
type JetStreamManager interface {
	// StreamInfo returns information about a stream, or nil if it doesn't exist.
	StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	// AddStream creates a new stream with the given configuration.
	AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	// UpdateStream updates an existing stream with the given configuration.
	UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
}

// SetupStreams creates or updates the required JetStream streams for AegisOpsAI.
// It is idempotent — existing streams are updated to match the desired configuration.
func SetupStreams(ctx context.Context, js JetStreamManager, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("component", "stream_setup")

	streams := DefaultStreams()

	for _, s := range streams {
		cfg := &nats.StreamConfig{
			Name:        s.Name,
			Subjects:    s.Subjects,
			Retention:   s.Retention,
			MaxAge:      s.MaxAge,
			MaxMsgs:     s.MaxMsgs,
			MaxBytes:    s.MaxBytes,
			Storage:     s.Storage,
			Replicas:    s.Replicas,
			Description: s.Description,
			Discard:     nats.DiscardOld,
			Duplicates:  2 * time.Minute,
		}

		info, err := js.StreamInfo(s.Name)
		if err != nil && err != nats.ErrStreamNotFound {
			logger.ErrorContext(ctx, "failed to query stream info",
				"stream", s.Name,
				"error", err,
			)
			return fmt.Errorf("query stream info for %s: %w", s.Name, err)
		}

		if info == nil {
			// Stream does not exist — create it.
			created, err := js.AddStream(cfg)
			if err != nil {
				logger.ErrorContext(ctx, "failed to create stream",
					"stream", s.Name,
					"error", err,
				)
				return fmt.Errorf("create stream %s: %w", s.Name, err)
			}
			logger.InfoContext(ctx, "stream created",
				"stream", s.Name,
				"subjects", s.Subjects,
				"retention", s.Retention.String(),
				"state_messages", created.State.Msgs,
			)
		} else {
			// Stream exists — update it to ensure configuration matches.
			updated, err := js.UpdateStream(cfg)
			if err != nil {
				logger.ErrorContext(ctx, "failed to update stream",
					"stream", s.Name,
					"error", err,
				)
				return fmt.Errorf("update stream %s: %w", s.Name, err)
			}
			logger.InfoContext(ctx, "stream updated",
				"stream", s.Name,
				"subjects", s.Subjects,
				"state_messages", updated.State.Msgs,
			)
		}
	}

	logger.InfoContext(ctx, "all streams configured", "count", len(streams))
	return nil
}
