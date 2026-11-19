// Package intelligence provides AI-assisted incident summaries, root cause explanations,
// and remediation suggestions, with full fallback when LLMs are offline.
package intelligence

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AegisOpsAI/aegisops/internal/models"
	"github.com/AegisOpsAI/aegisops/internal/normalizer"
	"github.com/AegisOpsAI/aegisops/internal/rca"
)

// IncidentSummary encapsulates human-readable incident briefings.
type IncidentSummary struct {
	Summary            string   `json:"summary"`
	KeyFindings        []string `json:"keyFindings"`
	AffectedSystems    []string `json:"affectedSystems"`
	RecommendedActions []string `json:"recommendedActions"`
	GeneratedByAI      bool     `json:"generatedByAi"`
}

// LLMProvider defines the provider-agnostic contract for AI integrations.
type LLMProvider interface {
	Name() string
	GenerateIncidentSummary(ctx context.Context, prompt string) (*IncidentSummary, error)
	ExplainRootCause(ctx context.Context, incident *models.Incident, rcaResult *rca.Result) (string, error)
	SuggestRemediation(ctx context.Context, incident *models.Incident, rcaResult *rca.Result) ([]string, error)
}

// Service manages AI intelligence queries with deterministic fallback.
type Service struct {
	provider LLMProvider
	enabled  bool
	logger   *slog.Logger
}

// NewService creates an intelligence service.
func NewService(provider LLMProvider, enabled bool, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		provider: provider,
		enabled:  enabled,
		logger:   logger.With(slog.String("component", "intelligence-service")),
	}
}

// SummarizeIncident produces an incident summary using the LLM or falling back to deterministic synthesis.
func (s *Service) SummarizeIncident(ctx context.Context, incident *models.Incident, events []*normalizer.NormalizedEvent, rcaResult *rca.Result) *IncidentSummary {
	if s.enabled && s.provider != nil {
		prompt := s.buildPrompt(incident, events, rcaResult)
		summary, err := s.provider.GenerateIncidentSummary(ctx, prompt)
		if err == nil && summary != nil {
			summary.GeneratedByAI = true
			return summary
		}
		s.logger.Warn("AI provider failed or unavailable; falling back to deterministic summary",
			slog.String("error", err.Error()),
		)
	}

	// Deterministic fallback summary:
	return s.fallbackSummary(incident, events, rcaResult)
}

func (s *Service) buildPrompt(incident *models.Incident, events []*normalizer.NormalizedEvent, rcaResult *rca.Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Incident: %s\nSeverity: %s\nSource: %s\nDescription: %s\n\n",
		incident.Title, incident.Severity, incident.Source, incident.Description))

	if rcaResult != nil && len(rcaResult.ProbableCauses) > 0 {
		sb.WriteString("RCA Findings:\n")
		for _, c := range rcaResult.ProbableCauses {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s (Likelihood: %.0f%%)\n",
				c.Component, c.FailureType, c.Description, c.Likelihood*100))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Correlated Events (%d total):\n", len(events)))
	for i, ev := range events {
		if i >= 10 { // limit to top 10
			sb.WriteString("... and more events\n")
			break
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", ev.Severity, ev.Title, ev.Description))
	}

	return sb.String()
}

func (s *Service) fallbackSummary(incident *models.Incident, events []*normalizer.NormalizedEvent, rcaResult *rca.Result) *IncidentSummary {
	findings := []string{
		fmt.Sprintf("Incident triggered with severity %s from %s", incident.Severity, incident.Source),
		fmt.Sprintf("Total correlated events observed: %d", len(events)),
	}

	var recommendations []string
	if rcaResult != nil && len(rcaResult.ProbableCauses) > 0 {
		primary := rcaResult.ProbableCauses[0]
		findings = append(findings, fmt.Sprintf("Primary suspect: %s (%s)", primary.Component, primary.Description))
		recommendations = append(recommendations, fmt.Sprintf("Inspect %s logs and metrics for anomaly verification", primary.Component))
	} else {
		recommendations = append(recommendations, "Review operational dashboard and service health checks")
	}

	affected := []string{incident.Source}

	return &IncidentSummary{
		Summary:            fmt.Sprintf("Automated analysis for incident '%s'. Cause suspected at %s.", incident.Title, incident.Source),
		KeyFindings:        findings,
		AffectedSystems:    affected,
		RecommendedActions: recommendations,
		GeneratedByAI:      false,
	}
}
