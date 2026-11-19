// Package ingestion provides the event ingestion pipeline for AegisOpsAI.
// It receives raw events from various sources, publishes them to NATS JetStream,
// and consumes them for normalization and persistent storage.
package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// EventStore defines the interface for persisting normalized events.
type EventStore interface {
	// StoreEvent persists a normalized event to the database.
	StoreEvent(ctx context.Context, event *normalizer.NormalizedEvent) error
}

// MessagePublisher defines the interface for publishing messages to a stream.
type MessagePublisher interface {
	// Publish publishes data to the given subject.
	Publish(subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
}

// MessageSubscriber defines the interface for subscribing to stream messages.
type MessageSubscriber interface {
	// Subscribe creates a durable consumer subscription.
	Subscribe(subject string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
}

// RawEvent represents an incoming event before normalization.
type RawEvent struct {
	// ID is a unique identifier assigned upon ingestion.
	ID string `json:"id"`
	// Source indicates the origin system (e.g., "prometheus", "alertmanager", "webhook").
	Source string `json:"source"`
	// ReceivedAt is the timestamp when the event was received.
	ReceivedAt time.Time `json:"received_at"`
	// ContentType is the MIME type of the raw payload.
	ContentType string `json:"content_type"`
	// Payload is the raw event data.
	Payload json.RawMessage `json:"payload"`
}

// EventIngester orchestrates the event ingestion pipeline, receiving raw events,
// publishing them to NATS JetStream, and consuming them for normalization and storage.
type EventIngester struct {
	publisher  MessagePublisher
	subscriber MessageSubscriber
	store      EventStore
	normalizer normalizer.Normalizer
	logger     *slog.Logger
}

// NewEventIngester creates a new EventIngester with the given dependencies.
func NewEventIngester(
	publisher MessagePublisher,
	subscriber MessageSubscriber,
	store EventStore,
	norm normalizer.Normalizer,
	logger *slog.Logger,
) *EventIngester {
	if logger == nil {
		logger = slog.Default()
	}
	return &EventIngester{
		publisher:  publisher,
		subscriber: subscriber,
		store:      store,
		normalizer: norm,
		logger:     logger.With("component", "event_ingester"),
	}
}

// IngestEvent receives a raw event, assigns an ID and timestamp, and publishes
// it to the EVENTS JetStream stream for downstream processing.
func (ei *EventIngester) IngestEvent(ctx context.Context, source string, contentType string, payload json.RawMessage) (string, error) {
	if len(payload) == 0 {
		return "", errors.New("empty event payload")
	}

	eventID := uuid.New().String()

	raw := RawEvent{
		ID:          eventID,
		Source:      source,
		ReceivedAt:  time.Now().UTC(),
		ContentType: contentType,
		Payload:     payload,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		ei.logger.ErrorContext(ctx, "failed to marshal raw event",
			"event_id", eventID,
			"source", source,
			"error", err,
		)
		return "", fmt.Errorf("marshal raw event: %w", err)
	}

	if ei.publisher == nil {
		// Fallback when NATS is not connected: normalize and store directly
		normalized, nErr := ei.normalizer.Normalize(ctx, source, payload)
		if nErr == nil && normalized != nil && ei.store != nil {
			if normalized.ID == "" {
				normalized.ID = eventID
			}
			_ = ei.store.StoreEvent(ctx, normalized)
		}
		return eventID, nil
	}

	subject := fmt.Sprintf("events.%s.ingest", source)
	ack, err := ei.publisher.Publish(subject, data)
	if err != nil {
		ei.logger.ErrorContext(ctx, "failed to publish event to NATS",
			"event_id", eventID,
			"source", source,
			"subject", subject,
			"error", err,
		)
		return "", fmt.Errorf("publish event to NATS: %w", err)
	}

	ei.logger.InfoContext(ctx, "event published to stream",
		"event_id", eventID,
		"source", source,
		"subject", subject,
		"stream", ack.Stream,
		"sequence", ack.Sequence,
	)

	return eventID, nil
}

// StartConsumer subscribes to the EVENTS stream and processes incoming events
// by normalizing and storing them. It blocks until the context is cancelled.
// The returned subscription can be used to drain or unsubscribe.
func (ei *EventIngester) StartConsumer(ctx context.Context) (*nats.Subscription, error) {
	if ei.subscriber == nil {
		ei.logger.InfoContext(ctx, "skipping event consumer (running without NATS subscriber)")
		return nil, nil
	}

	handler := func(msg *nats.Msg) {
		ei.processMessage(ctx, msg)
	}

	sub, err := ei.subscriber.Subscribe(
		"events.>",
		handler,
		nats.Durable("aegisops-event-processor"),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
		nats.MaxDeliver(5),
		nats.DeliverAll(),
	)
	if err != nil {
		ei.logger.ErrorContext(ctx, "failed to subscribe to events stream",
			"error", err,
		)
		return nil, fmt.Errorf("subscribe to events stream: %w", err)
	}

	ei.logger.InfoContext(ctx, "event consumer started",
		"subject", "events.>",
		"durable", "aegisops-event-processor",
	)

	return sub, nil
}

// processMessage handles a single NATS message by unmarshalling the raw event,
// normalizing it, storing it, and acknowledging the message.
func (ei *EventIngester) processMessage(ctx context.Context, msg *nats.Msg) {
	var raw RawEvent
	if err := json.Unmarshal(msg.Data, &raw); err != nil {
		ei.logger.ErrorContext(ctx, "failed to unmarshal raw event from NATS",
			"subject", msg.Subject,
			"error", err,
		)
		// Terminal failure — do not retry malformed messages.
		if termErr := msg.Term(); termErr != nil {
			ei.logger.ErrorContext(ctx, "failed to terminate message", "error", termErr)
		}
		return
	}

	logger := ei.logger.With(
		"event_id", raw.ID,
		"source", raw.Source,
		"subject", msg.Subject,
	)

	normalized, err := ei.normalizer.Normalize(ctx, raw.Source, raw.Payload)
	if err != nil {
		logger.ErrorContext(ctx, "failed to normalize event", "error", err)
		// Normalization errors may be retryable (e.g., transient parsing issues).
		if nakErr := msg.Nak(); nakErr != nil {
			logger.ErrorContext(ctx, "failed to nak message", "error", nakErr)
		}
		return
	}

	// Preserve the original ingestion ID if the normalizer didn't set one.
	if normalized.ID == "" {
		normalized.ID = raw.ID
	}

	if err := ei.store.StoreEvent(ctx, normalized); err != nil {
		logger.ErrorContext(ctx, "failed to store normalized event", "error", err)
		if nakErr := msg.Nak(); nakErr != nil {
			logger.ErrorContext(ctx, "failed to nak message", "error", nakErr)
		}
		return
	}

	if err := msg.Ack(); err != nil {
		logger.ErrorContext(ctx, "failed to ack message", "error", err)
		return
	}

	logger.InfoContext(ctx, "event processed and stored",
		"normalized_type", normalized.Type,
		"severity", normalized.Severity,
	)
}

// PublishRawEvent satisfies handlers.IngestionService for ingesting raw payloads.
func (ei *EventIngester) PublishRawEvent(source string, payload []byte) error {
	_, err := ei.IngestEvent(context.Background(), source, "application/json", payload)
	return err
}
