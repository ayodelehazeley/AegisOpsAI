// Package rca provides deterministic, rule-based Root Cause Analysis (RCA)
// for infrastructure and operational incidents.
package rca

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/AegisOpsAI/aegisops/internal/models"
	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// Evidence represents an artifact or metric supporting an RCA hypothesis.
type Evidence struct {
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// Cause represents an identified probable root cause.
type Cause struct {
	Component   string   `json:"component"`
	FailureType string   `json:"failureType"`
	Description string   `json:"description"`
	Likelihood  float64  `json:"likelihood"`
	EvidenceIDs []string `json:"evidenceIds,omitempty"`
}

// Result holds the aggregated outcome of root cause analysis.
type Result struct {
	IncidentID     string     `json:"incidentId"`
	ProbableCauses []Cause    `json:"probableCauses"`
	EvidenceList   []Evidence `json:"evidence"`
	OverallScore   float64    `json:"overallScore"`
	AnalyzedAt     time.Time  `json:"analyzedAt"`
}

// Strategy defines an analysis heuristic for diagnosing incidents.
type Strategy interface {
	Name() string
	Analyze(ctx context.Context, incident *models.Incident, events []*normalizer.NormalizedEvent) ([]Cause, []Evidence)
}

// ResourceExhaustionStrategy detects disk, memory, or CPU exhaustion patterns.
type ResourceExhaustionStrategy struct{}

func (s *ResourceExhaustionStrategy) Name() string { return "ResourceExhaustionStrategy" }

func (s *ResourceExhaustionStrategy) Analyze(_ context.Context, _ *models.Incident, events []*normalizer.NormalizedEvent) ([]Cause, []Evidence) {
	var causes []Cause
	var evidence []Evidence

	for _, ev := range events {
		lowerTitle := strings.ToLower(ev.Title + " " + ev.Description)
		if strings.Contains(lowerTitle, "disk") || strings.Contains(lowerTitle, "no space left") {
			evidence = append(evidence, Evidence{
				Source:      ev.Source,
				Description: fmt.Sprintf("Disk warning event received: %s", ev.Title),
				Confidence:  0.90,
				Timestamp:   ev.Timestamp,
			})
			causes = append(causes, Cause{
				Component:   "Storage Subsystem",
				FailureType: "DiskSpaceExhaustion",
				Description: "Storage capacity fully consumed, leading to service I/O failures",
				Likelihood:  0.92,
			})
		} else if strings.Contains(lowerTitle, "oom") || strings.Contains(lowerTitle, "out of memory") {
			evidence = append(evidence, Evidence{
				Source:      ev.Source,
				Description: fmt.Sprintf("OOM condition observed: %s", ev.Title),
				Confidence:  0.95,
				Timestamp:   ev.Timestamp,
			})
			causes = append(causes, Cause{
				Component:   "Compute / Memory",
				FailureType: "OOMKill",
				Description: "Kernel terminated process due to memory limits",
				Likelihood:  0.95,
			})
		}
	}
	return causes, evidence
}

// TemporalCascadeStrategy identifies earlier events that likely triggered subsequent failures.
type TemporalCascadeStrategy struct{}

func (s *TemporalCascadeStrategy) Name() string { return "TemporalCascadeStrategy" }

func (s *TemporalCascadeStrategy) Analyze(_ context.Context, _ *models.Incident, events []*normalizer.NormalizedEvent) ([]Cause, []Evidence) {
	if len(events) < 2 {
		return nil, nil
	}

	// Sort events by timestamp ascending
	sorted := make([]*normalizer.NormalizedEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	firstEvent := sorted[0]
	evidence := []Evidence{
		{
			Source:      firstEvent.Source,
			Description: fmt.Sprintf("Earliest failure event in chain: %s (%s)", firstEvent.Title, firstEvent.Source),
			Confidence:  0.80,
			Timestamp:   firstEvent.Timestamp,
		},
	}

	cause := Cause{
		Component:   firstEvent.Source,
		FailureType: "OriginatingFailure",
		Description: fmt.Sprintf("Earliest failure in sequence triggered downstream cascading alerts: %s", firstEvent.Title),
		Likelihood:  0.80,
	}

	return []Cause{cause}, evidence
}

// Engine coordinates RCA strategies against incidents and evidence.
type Engine struct {
	strategies []Strategy
	logger     *slog.Logger
}

// NewEngine creates an RCA engine with standard diagnostic strategies.
func NewEngine(logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		strategies: []Strategy{
			&ResourceExhaustionStrategy{},
			&TemporalCascadeStrategy{},
		},
		logger: logger.With(slog.String("component", "rca-engine")),
	}
}

// Analyze performs root cause analysis on an incident and its collected events.
func (e *Engine) Analyze(ctx context.Context, incident *models.Incident, events []*normalizer.NormalizedEvent) *Result {
	res := &Result{
		IncidentID: incident.ID.String(),
		AnalyzedAt: time.Now().UTC(),
	}

	causeMap := make(map[string]Cause)
	var allEvidence []Evidence

	for _, strat := range e.strategies {
		causes, evidence := strat.Analyze(ctx, incident, events)
		allEvidence = append(allEvidence, evidence...)
		for _, c := range causes {
			key := c.Component + ":" + c.FailureType
			if existing, exists := causeMap[key]; exists {
				if c.Likelihood > existing.Likelihood {
					causeMap[key] = c
				}
			} else {
				causeMap[key] = c
			}
		}
	}

	for _, c := range causeMap {
		res.ProbableCauses = append(res.ProbableCauses, c)
	}

	// Sort causes by likelihood descending
	sort.Slice(res.ProbableCauses, func(i, j int) bool {
		return res.ProbableCauses[i].Likelihood > res.ProbableCauses[j].Likelihood
	})

	res.EvidenceList = allEvidence
	if len(res.ProbableCauses) > 0 {
		res.OverallScore = res.ProbableCauses[0].Likelihood
	} else {
		res.OverallScore = 0.50
		res.ProbableCauses = append(res.ProbableCauses, Cause{
			Component:   incident.Source,
			FailureType: "UnclassifiedAnomaly",
			Description: "Deterministic analysis found no known signature; fallback to general telemetry review.",
			Likelihood:  0.50,
		})
	}

	e.logger.Info("completed RCA analysis",
		slog.String("incident_id", incident.ID.String()),
		slog.Int("probable_causes", len(res.ProbableCauses)),
		slog.Int("evidence_count", len(res.EvidenceList)),
	)

	return res
}
