// Package scheduler provides background job orchestration for periodic telemetry polling,
// stale incident sweeps, and automated system health monitoring.
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// JobFunc defines the function signature for scheduled background tasks.
type JobFunc func(ctx context.Context) error

// Job represents a configured recurring background job.
type Job struct {
	Name     string
	Interval time.Duration
	Fn       JobFunc
}

// Scheduler coordinates recurring background operations.
type Scheduler struct {
	jobs   []Job
	logger *slog.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// NewScheduler creates an empty scheduler instance.
func NewScheduler(logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		logger: logger.With(slog.String("component", "scheduler")),
	}
}

// RegisterJob adds a new recurring task to the scheduler before Start is called.
func (s *Scheduler) RegisterJob(name string, interval time.Duration, fn JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, Job{
		Name:     name,
		Interval: interval,
		Fn:       fn,
	})
}

// Start launches all registered recurring jobs in individual goroutines.
func (s *Scheduler) Start(parentCtx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(ctx, job)
	}

	s.logger.Info("scheduler started", slog.Int("job_count", len(s.jobs)))
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	defer s.wg.Done()
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	s.logger.Info("job scheduled", slog.String("name", job.Name), slog.Duration("interval", job.Interval))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("job stopped by context cancellation", slog.String("name", job.Name))
			return
		case <-ticker.C:
			s.logger.Debug("running scheduled job", slog.String("name", job.Name))
			if err := job.Fn(ctx); err != nil {
				s.logger.Error("scheduled job failed",
					slog.String("name", job.Name),
					slog.String("error", err.Error()),
				)
			}
		}
	}
}

// Stop gracefully stops all background tasks.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("all background jobs terminated")
}
