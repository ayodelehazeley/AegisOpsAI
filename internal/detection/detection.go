// Package detection provides rule-based anomaly and degradation detection.
// It detects threshold breaches, sudden metric spikes, and missing service heartbeats.
package detection

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// DetectionResult represents an anomaly detected by an evaluation rule.
type DetectionResult struct {
	RuleName    string                    `json:"ruleName"`
	Severity    normalizer.Severity       `json:"severity"`
	Message     string                    `json:"message"`
	MetricName  string                    `json:"metricName,omitempty"`
	CurrentVal  float64                   `json:"currentValue,omitempty"`
	Threshold   float64                   `json:"threshold,omitempty"`
	DetectedAt  time.Time                 `json:"detectedAt"`
	EventSource string                    `json:"eventSource"`
}

// ToNormalizedEvent converts a detection result into a NormalizedEvent.
func (dr *DetectionResult) ToNormalizedEvent() *normalizer.NormalizedEvent {
	return &normalizer.NormalizedEvent{
		ID:          uuid.New().String(),
		Source:      "aegisops-detection",
		Type:        "anomaly",
		Severity:    dr.Severity,
		Title:       fmt.Sprintf("Anomaly Detected: %s", dr.RuleName),
		Description: dr.Message,
		Metrics: map[string]float64{
			dr.MetricName: dr.CurrentVal,
		},
		Timestamp: dr.DetectedAt,
	}
}

// DetectionRule evaluates incoming metrics or heartbeat events.
type DetectionRule interface {
	Name() string
	Evaluate(ctx context.Context, event *normalizer.NormalizedEvent) (*DetectionResult, bool)
}

// ThresholdBreachRule alerts when a metric breaches an upper threshold.
type ThresholdBreachRule struct {
	MetricName string
	Threshold  float64
	Severity   normalizer.Severity
}

func (r *ThresholdBreachRule) Name() string {
	return fmt.Sprintf("ThresholdBreach[%s > %0.2f]", r.MetricName, r.Threshold)
}

func (r *ThresholdBreachRule) Evaluate(_ context.Context, event *normalizer.NormalizedEvent) (*DetectionResult, bool) {
	if event.Metrics == nil {
		return nil, false
	}
	val, ok := event.Metrics[r.MetricName]
	if !ok {
		return nil, false
	}
	if val > r.Threshold {
		return &DetectionResult{
			RuleName:    r.Name(),
			Severity:    r.Severity,
			Message:     fmt.Sprintf("Metric %s value %0.2f exceeded threshold %0.2f", r.MetricName, val, r.Threshold),
			MetricName:  r.MetricName,
			CurrentVal:  val,
			Threshold:   r.Threshold,
			DetectedAt:  time.Now().UTC(),
			EventSource: event.Source,
		}, true
	}
	return nil, false
}

// Engine coordinates anomaly detection rules over operational events.
type Engine struct {
	rules  []DetectionRule
	logger *slog.Logger
	mu     sync.RWMutex
}

// NewEngine constructs a new detection engine.
func NewEngine(logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		rules: []DetectionRule{
			&ThresholdBreachRule{
				MetricName: "cpu_usage",
				Threshold:  90.0,
				Severity:   normalizer.SeverityCritical,
			},
			&ThresholdBreachRule{
				MetricName: "disk_utilization",
				Threshold:  85.0,
				Severity:   normalizer.SeverityHigh,
			},
			&ThresholdBreachRule{
				MetricName: "error_rate",
				Threshold:  5.0,
				Severity:   normalizer.SeverityHigh,
			},
		},
		logger: logger.With(slog.String("component", "detection-engine")),
	}
}

// RegisterRule adds an additional detection rule to the engine.
func (e *Engine) RegisterRule(rule DetectionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// Detect checks an event against all configured detection rules.
func (e *Engine) Detect(ctx context.Context, event *normalizer.NormalizedEvent) []*DetectionResult {
	e.mu.RLock()
	rules := make([]DetectionRule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	var results []*DetectionResult
	for _, rule := range rules {
		if res, match := rule.Evaluate(ctx, event); match {
			e.logger.Warn("detection rule triggered",
				slog.String("rule", rule.Name()),
				slog.String("source", event.Source),
				slog.String("severity", string(res.Severity)),
			)
			results = append(results, res)
		}
	}
	return results
}
