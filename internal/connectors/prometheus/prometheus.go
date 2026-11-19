// Package prometheus provides an integration connector for Prometheus.
// It supports querying metrics, fetching active alerts, and converting them
// into the canonical AegisOps event format.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/normalizer"
)

// Config configures the Prometheus connector.
type Config struct {
	BaseURL    string        `json:"baseUrl"`
	Timeout    time.Duration `json:"timeout"`
	AuthHeader string        `json:"authHeader,omitempty"`
}

// Connector manages connections and queries to a Prometheus server.
type Connector struct {
	cfg    Config
	client *http.Client
	logger *slog.Logger
}

// NewConnector creates a new Prometheus connector.
func NewConnector(cfg Config, logger *slog.Logger) *Connector {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Connector{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger.With(slog.String("component", "prometheus-connector")),
	}
}

// AlertResponse represents the Prometheus /api/v1/alerts response payload.
type AlertResponse struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []AlertItem `json:"alerts"`
	} `json:"data"`
}

// AlertItem describes a single active alert returned by Prometheus.
type AlertItem struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    time.Time         `json:"activeAt"`
	Value       string            `json:"value"`
}

// QueryResult represents a Prometheus instant or range query response.
type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"` // [unix_time, "float_val_string"]
		} `json:"result"`
	} `json:"data"`
}

// FetchActiveAlerts retrieves active alerts from Prometheus and transforms them into NormalizedEvents.
func (c *Connector) FetchActiveAlerts(ctx context.Context) ([]*normalizer.NormalizedEvent, error) {
	endpoint := fmt.Sprintf("%s/api/v1/alerts", c.cfg.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create prometheus alerts request: %w", err)
	}

	if c.cfg.AuthHeader != "" {
		req.Header.Set("Authorization", c.cfg.AuthHeader)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute prometheus alerts request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus alerts returned status %d: %s", resp.StatusCode, string(body))
	}

	var alertResp AlertResponse
	if err := json.NewDecoder(resp.Body).Decode(&alertResp); err != nil {
		return nil, fmt.Errorf("decode prometheus alerts response: %w", err)
	}

	events := make([]*normalizer.NormalizedEvent, 0, len(alertResp.Data.Alerts))
	for _, alert := range alertResp.Data.Alerts {
		severity := normalizer.SeverityInfo
		if sev, ok := alert.Labels["severity"]; ok {
			switch sev {
			case "critical", "page":
				severity = normalizer.SeverityCritical
			case "warning", "high":
				severity = normalizer.SeverityHigh
			case "medium":
				severity = normalizer.SeverityMedium
			default:
				severity = normalizer.SeverityLow
			}
		}

		title := alert.Annotations["summary"]
		if title == "" {
			title = alert.Labels["alertname"]
		}

		desc := alert.Annotations["description"]
		metrics := make(map[string]float64)
		if alert.Value != "" {
			if v, err := strconv.ParseFloat(alert.Value, 64); err == nil {
				metrics["value"] = v
			}
		}

		rawBytes, _ := json.Marshal(alert)
		event := &normalizer.NormalizedEvent{
			ID:          uuid.New().String(),
			Source:      "prometheus",
			Type:        "alert",
			Severity:    severity,
			Title:       title,
			Description: desc,
			Labels:      alert.Labels,
			Metrics:     metrics,
			Timestamp:   alert.ActiveAt,
			RawPayload:  rawBytes,
		}
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now().UTC()
		}
		events = append(events, event)
	}

	c.logger.Info("fetched active alerts from prometheus", slog.Int("count", len(events)))
	return events, nil
}

// InstantQuery runs a PromQL query at the current time.
func (c *Connector) InstantQuery(ctx context.Context, promQL string) (*QueryResult, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query", c.cfg.BaseURL))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", promQL)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.cfg.AuthHeader != "" {
		req.Header.Set("Authorization", c.cfg.AuthHeader)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
