// Package policy provides role-based access control, execution policies,
// and safety boundaries for automated actions within AegisOpsAI.
package policy

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/AegisOpsAI/aegisops/internal/models"
)

// Standard policy errors.
var (
	ErrActionDenied         = errors.New("action denied by security policy")
	ErrApprovalRequired     = errors.New("action requires explicit human approval before execution")
	ErrUnauthorizedRole     = errors.New("user role is not authorized to approve or execute this action")
)

// Policy defines an execution rule governing operations.
type Policy struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Resource         string        `json:"resource"`
	Action           string        `json:"action"`
	RequiresApproval bool          `json:"requiresApproval"`
	AllowedRoles     []models.Role `json:"allowedRoles"`
}

// Engine evaluates operations against registered policies.
type Engine struct {
	policies []Policy
	logger   *slog.Logger
	mu       sync.RWMutex
}

// NewEngine creates a policy engine initialized with strict default security guardrails.
func NewEngine(logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		policies: []Policy{
			{
				ID:               "pol-remediation-guard",
				Name:             "Remediation Guardrail",
				Resource:         "infrastructure",
				Action:           "remediation:*",
				RequiresApproval: true,
				AllowedRoles:     []models.Role{models.RoleAdmin, models.RoleOperator},
			},
			{
				ID:               "pol-audit-view",
				Name:             "Audit Trail Access",
				Resource:         "audit_logs",
				Action:           "read",
				RequiresApproval: false,
				AllowedRoles:     []models.Role{models.RoleAdmin},
			},
		},
		logger: logger.With(slog.String("component", "policy-engine")),
	}
}

// CanExecute evaluates whether a given user role can perform an action on a resource.
func (e *Engine) CanExecute(_ context.Context, role models.Role, resource, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, p := range e.policies {
		if p.Resource == resource || p.Resource == "*" {
			roleAllowed := false
			for _, allowed := range p.AllowedRoles {
				if allowed == role {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				return false, ErrUnauthorizedRole
			}
			return true, nil
		}
	}
	return true, nil
}

// RequiresHumanApproval determines if an action is guarded by mandatory human review.
func (e *Engine) RequiresHumanApproval(resource, action string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, p := range e.policies {
		if (p.Resource == resource || p.Resource == "*") && p.RequiresApproval {
			return true
		}
	}
	return true // Default secure: require approval
}
