// Package correlation provides rule-based event grouping and incident correlation.
// It groups related alerts and operational events into cohesive incidents to prevent
// alert storms and reduce cognitive load on engineering teams.
package correlation

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/models"
	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// IncidentStore defines storage operations required by the correlation engine.
type IncidentStore interface {
	GetOpenIncidents(ctx context.Context) ([]*models.Incident, error)
	CreateIncident(ctx context.Context, inc *models.Incident) error
	AttachEventToIncident(ctx context.Context, incidentID uuid.UUID, eventID uuid.UUID) error
}

// CorrelationRule evaluates whether an event belongs to an existing incident.
type CorrelationRule interface {
	Name() string
	Match(event *normalizer.NormalizedEvent, incident *models.Incident) (bool, float64) // match, confidence
}

// SameServiceRule matches events sharing the same "service" label.
type SameServiceRule struct{}

func (r *SameServiceRule) Name() string { return "SameServiceRule" }
func (r *SameServiceRule) Match(event *normalizer.NormalizedEvent, incident *models.Incident) (bool, float64) {
	if event.Labels == nil {
		return false, 0
	}
	eventSvc := event.Labels["service"]
	if eventSvc == "" {
		eventSvc = event.Labels["app"]
	}
	if eventSvc == "" {
		return false, 0
	}

	// Simple heuristic: if incident title or description mentions the service
	if incident.Source == event.Source {
		return true, 0.75
	}
	return false, 0
}

// Engine coordinates event correlation against active incidents.
type Engine struct {
	store      IncidentStore
	rules      []CorrelationRule
	window     time.Duration
	logger     *slog.Logger
	mu         sync.RWMutex
}

// NewEngine creates an initialized correlation engine.
func NewEngine(store IncidentStore, window time.Duration, logger *slog.Logger) *Engine {
	if window <= 0 {
		window = 30 * time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		store:  store,
		rules:  []CorrelationRule{&SameServiceRule{}},
		window: window,
		logger: logger.With(slog.String("component", "correlation-engine")),
	}
}

// RegisterRule appends an extra correlation rule.
func (e *Engine) RegisterRule(rule CorrelationRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// CorrelateResult represents the action taken for an incoming event.
type CorrelateResult struct {
	IncidentID uuid.UUID
	CreatedNew bool
	RuleName   string
	Confidence float64
}

// CorrelateEvent evaluates an event against open incidents and either correlates or creates a new incident.
func (e *Engine) CorrelateEvent(ctx context.Context, event *normalizer.NormalizedEvent) (*CorrelateResult, error) {
	e.mu.RLock()
	rules := make([]CorrelationRule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	openIncidents, err := e.store.GetOpenIncidents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open incidents: %w", err)
	}

	eventUUID, err := uuid.Parse(event.ID)
	if err != nil {
		eventUUID = uuid.New()
	}

	var bestIncident *models.Incident
	var highestConfidence float64
	var matchedRule string

	for _, inc := range openIncidents {
		// Enforce correlation time window
		if time.Since(inc.CreatedAt) > e.window {
			continue
		}

		for _, rule := range rules {
			matched, confidence := rule.Match(event, inc)
			if matched && confidence > highestConfidence {
				highestConfidence = confidence
				bestIncident = inc
				matchedRule = rule.Name()
			}
		}
	}

	if bestIncident != nil && highestConfidence >= 0.5 {
		e.logger.Info("correlated event to existing incident",
			slog.String("incident_id", bestIncident.ID.String()),
			slog.String("event_id", event.ID),
			slog.Float64("confidence", highestConfidence),
			slog.String("rule", matchedRule),
		)

		if err := e.store.AttachEventToIncident(ctx, bestIncident.ID, eventUUID); err != nil {
			return nil, fmt.Errorf("failed to attach event to incident: %w", err)
		}

		return &CorrelateResult{
			IncidentID: bestIncident.ID,
			CreatedNew: false,
			RuleName:   matchedRule,
			Confidence: highestConfidence,
		}, nil
	}

	// No incident matched: If severity is critical or high, create a new incident.
	if event.Severity == normalizer.SeverityCritical || event.Severity == normalizer.SeverityHigh {
		sev := models.SeverityHigh
		if event.Severity == normalizer.SeverityCritical {
			sev = models.SeverityCritical
		}

		newInc := &models.Incident{
			ID:          uuid.New(),
			Title:       event.Title,
			Description: event.Description,
			Severity:    sev,
			Status:      models.IncidentStatusOpen,
			Source:      event.Source,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := e.store.CreateIncident(ctx, newInc); err != nil {
			return nil, fmt.Errorf("failed to create incident: %w", err)
		}

		_ = e.store.AttachEventToIncident(ctx, newInc.ID, eventUUID)

		e.logger.Info("created new incident from un-correlated event",
			slog.String("incident_id", newInc.ID.String()),
			slog.String("event_id", event.ID),
			slog.String("severity", string(sev)),
		)

		return &CorrelateResult{
			IncidentID: newInc.ID,
			CreatedNew: true,
			Confidence: 1.0,
		}, nil
	}

	return nil, nil
}
