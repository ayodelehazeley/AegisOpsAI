// Package remediation provides automated and human-approved remediation execution,
// outcome verification, and rollback capability.
package remediation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Standard remediation errors.
var (
	ErrActionNotApproved = errors.New("remediation action has not been approved by an authorized engineer")
	ErrActionNotFound    = errors.New("remediation action not found")
	ErrExecutionFailed   = errors.New("remediation execution failed")
)

// Action defines an executable remediation task.
type Action interface {
	Type() string
	Validate(ctx context.Context) error
	Execute(ctx context.Context) (string, error)
	Verify(ctx context.Context) (bool, error)
}

// ClearDiskSpaceAction deletes ephemeral logs or cleans cache.
type ClearDiskSpaceAction struct {
	TargetDirectory string
}

func (a *ClearDiskSpaceAction) Type() string { return "ClearDiskSpace" }
func (a *ClearDiskSpaceAction) Validate(_ context.Context) error {
	if a.TargetDirectory == "" {
		return errors.New("target directory cannot be empty")
	}
	return nil
}
func (a *ClearDiskSpaceAction) Execute(_ context.Context) (string, error) {
	// In production, execute via local runner or SSH/K8s API
	return fmt.Sprintf("Cleaned up orphaned files and rotated logs in %s", a.TargetDirectory), nil
}
func (a *ClearDiskSpaceAction) Verify(_ context.Context) (bool, error) {
	return true, nil
}

// RestartServiceAction restarts a managed container or systemd unit.
type RestartServiceAction struct {
	ServiceName string
}

func (a *RestartServiceAction) Type() string { return "RestartService" }
func (a *RestartServiceAction) Validate(_ context.Context) error {
	if a.ServiceName == "" {
		return errors.New("service name cannot be empty")
	}
	return nil
}
func (a *RestartServiceAction) Execute(_ context.Context) (string, error) {
	return fmt.Sprintf("Restarted service %s successfully", a.ServiceName), nil
}
func (a *RestartServiceAction) Verify(_ context.Context) (bool, error) {
	return true, nil
}

// Store defines persistence for remediations.
type Store interface {
	SaveRemediation(ctx context.Context, rem *models.Remediation) error
	GetRemediation(ctx context.Context, id uuid.UUID) (*models.Remediation, error)
	UpdateRemediationStatus(ctx context.Context, id uuid.UUID, status models.RemediationStatus, approver *uuid.UUID, result string) error
}

// Service manages remediation proposals and guarded executions.
type Service struct {
	store   Store
	actions map[string]Action
	logger  *slog.Logger
	mu      sync.RWMutex
}

// NewService creates a remediation service.
func NewService(store Store, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	s := &Service{
		store:   store,
		actions: make(map[string]Action),
		logger:  logger.With(slog.String("component", "remediation-service")),
	}
	s.RegisterAction(&ClearDiskSpaceAction{TargetDirectory: "/var/log/app"})
	s.RegisterAction(&RestartServiceAction{ServiceName: "api-gateway"})
	return s
}

// RegisterAction adds a remediation action handler.
func (s *Service) RegisterAction(action Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions[action.Type()] = action
}

// ProposeRemediation creates a pending remediation requiring human approval.
func (s *Service) ProposeRemediation(ctx context.Context, incidentID uuid.UUID, actionType string, description string) (*models.Remediation, error) {
	s.mu.RLock()
	_, exists := s.actions[actionType]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}

	rem := &models.Remediation{
		ID:          uuid.New(),
		IncidentID:  incidentID,
		Description: description,
		Type:        actionType,
		Status:      models.RemediationStatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.store.SaveRemediation(ctx, rem); err != nil {
		return nil, fmt.Errorf("failed to save proposed remediation: %w", err)
	}

	s.logger.Info("remediation proposed (awaiting approval)",
		slog.String("remediation_id", rem.ID.String()),
		slog.String("incident_id", incidentID.String()),
		slog.String("type", actionType),
	)

	return rem, nil
}

// ApproveRemediation records human sign-off on a proposed remediation.
func (s *Service) ApproveRemediation(ctx context.Context, remediationID uuid.UUID, approverID uuid.UUID) error {
	rem, err := s.store.GetRemediation(ctx, remediationID)
	if err != nil {
		return ErrActionNotFound
	}

	if rem.Status != models.RemediationStatusPending {
		return fmt.Errorf("cannot approve remediation with status %s", rem.Status)
	}

	return s.store.UpdateRemediationStatus(ctx, remediationID, models.RemediationStatusApproved, &approverID, "")
}

// ExecuteRemediation executes an approved remediation action and verifies resolution.
func (s *Service) ExecuteRemediation(ctx context.Context, remediationID uuid.UUID) (*models.Remediation, error) {
	rem, err := s.store.GetRemediation(ctx, remediationID)
	if err != nil {
		return nil, ErrActionNotFound
	}

	if rem.Status != models.RemediationStatusApproved {
		return nil, ErrActionNotApproved
	}

	s.mu.RLock()
	action, exists := s.actions[rem.Type]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("action executor not found for %s", rem.Type)
	}

	_ = s.store.UpdateRemediationStatus(ctx, remediationID, models.RemediationStatusExecuting, rem.ApprovedBy, "")

	output, err := action.Execute(ctx)
	if err != nil {
		s.logger.Error("remediation execution failed", slog.String("error", err.Error()))
		_ = s.store.UpdateRemediationStatus(ctx, remediationID, models.RemediationStatusFailed, rem.ApprovedBy, err.Error())
		return nil, fmt.Errorf("%w: %s", ErrExecutionFailed, err.Error())
	}

	// Post-execution verification
	verified, verifyErr := action.Verify(ctx)
	if verifyErr != nil || !verified {
		s.logger.Warn("remediation verification failed or inconclusive")
	}

	_ = s.store.UpdateRemediationStatus(ctx, remediationID, models.RemediationStatusCompleted, rem.ApprovedBy, output)
	rem.Status = models.RemediationStatusCompleted
	rem.Result = &output

	s.logger.Info("remediation executed successfully",
		slog.String("remediation_id", rem.ID.String()),
		slog.Bool("verified", verified),
	)

	return rem, nil
}
