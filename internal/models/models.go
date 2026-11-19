// Package models defines the core domain models for AegisOpsAI.
//
// All models use UUID primary keys and include standard audit timestamps.
// JSON tags follow camelCase convention; db tags map to PostgreSQL column names.
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Severity represents the severity level of an incident, event, or alert.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// ValidSeverities contains all valid severity values.
var ValidSeverities = []Severity{
	SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo,
}

// IsValid checks whether the severity value is a recognized severity level.
func (s Severity) IsValid() bool {
	for _, v := range ValidSeverities {
		if s == v {
			return true
		}
	}
	return false
}

// IncidentStatus represents the lifecycle status of an incident.
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "open"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusIdentified    IncidentStatus = "identified"
	IncidentStatusMonitoring    IncidentStatus = "monitoring"
	IncidentStatusResolved      IncidentStatus = "resolved"
)

// ValidIncidentStatuses contains all valid incident status values.
var ValidIncidentStatuses = []IncidentStatus{
	IncidentStatusOpen, IncidentStatusInvestigating, IncidentStatusIdentified,
	IncidentStatusMonitoring, IncidentStatusResolved,
}

// IsValid checks whether the incident status value is recognized.
func (s IncidentStatus) IsValid() bool {
	for _, v := range ValidIncidentStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// RemediationStatus represents the lifecycle status of a remediation action.
type RemediationStatus string

const (
	RemediationStatusPending   RemediationStatus = "pending"
	RemediationStatusApproved  RemediationStatus = "approved"
	RemediationStatusExecuting RemediationStatus = "executing"
	RemediationStatusCompleted RemediationStatus = "completed"
	RemediationStatusFailed    RemediationStatus = "failed"
	RemediationStatusRejected  RemediationStatus = "rejected"
)

// ValidRemediationStatuses contains all valid remediation status values.
var ValidRemediationStatuses = []RemediationStatus{
	RemediationStatusPending, RemediationStatusApproved, RemediationStatusExecuting,
	RemediationStatusCompleted, RemediationStatusFailed, RemediationStatusRejected,
}

// IsValid checks whether the remediation status value is recognized.
func (s RemediationStatus) IsValid() bool {
	for _, v := range ValidRemediationStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// AlertStatus represents the status of an alert.
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Role represents a user's authorization level within the system.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// ValidRoles contains all valid user role values.
var ValidRoles = []Role{RoleAdmin, RoleOperator, RoleViewer}

// IsValid checks whether the role value is recognized.
func (r Role) IsValid() bool {
	for _, v := range ValidRoles {
		if r == v {
			return true
		}
	}
	return false
}

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID `json:"id"           db:"id"`
	Email        string    `json:"email"         db:"email"`
	PasswordHash string    `json:"-"             db:"password_hash"`
	DisplayName  string    `json:"displayName"   db:"display_name"`
	Role         Role      `json:"role"          db:"role"`
	CreatedAt    time.Time `json:"createdAt"     db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt"     db:"updated_at"`
}

// Incident represents a tracked operational incident.
type Incident struct {
	ID          uuid.UUID      `json:"id"          db:"id"`
	Title       string         `json:"title"       db:"title"`
	Description string         `json:"description" db:"description"`
	Severity    Severity       `json:"severity"    db:"severity"`
	Status      IncidentStatus `json:"status"      db:"status"`
	Source      string         `json:"source"      db:"source"`
	CreatedAt   time.Time      `json:"createdAt"   db:"created_at"`
	UpdatedAt   time.Time      `json:"updatedAt"   db:"updated_at"`
	ResolvedAt  *time.Time     `json:"resolvedAt"  db:"resolved_at"`
	AssigneeID  *uuid.UUID     `json:"assigneeId"  db:"assignee_id"`
}

// Event represents a raw or normalized operational event ingested into the system.
type Event struct {
	ID         uuid.UUID       `json:"id"         db:"id"`
	Source     string          `json:"source"     db:"source"`
	Type       string          `json:"type"       db:"type"`
	Severity   Severity        `json:"severity"   db:"severity"`
	Payload    json.RawMessage `json:"payload"    db:"payload"`
	Timestamp  time.Time       `json:"timestamp"  db:"timestamp"`
	IncidentID *uuid.UUID      `json:"incidentId" db:"incident_id"`
	Normalized bool            `json:"normalized" db:"normalized"`
	CreatedAt  time.Time       `json:"createdAt"  db:"created_at"`
}

// Alert represents an alert generated from a monitoring source.
type Alert struct {
	ID         uuid.UUID         `json:"id"         db:"id"`
	Source     string            `json:"source"     db:"source"`
	Name       string            `json:"name"       db:"name"`
	Labels     map[string]string `json:"labels"     db:"labels"`
	Severity   Severity          `json:"severity"   db:"severity"`
	Status     AlertStatus       `json:"status"     db:"status"`
	FiredAt    time.Time         `json:"firedAt"    db:"fired_at"`
	ResolvedAt *time.Time        `json:"resolvedAt" db:"resolved_at"`
	EventID    *uuid.UUID        `json:"eventId"    db:"event_id"`
}

// AuditLog represents an auditable action performed by a user.
type AuditLog struct {
	ID           uuid.UUID       `json:"id"           db:"id"`
	UserID       uuid.UUID       `json:"userId"       db:"user_id"`
	Action       string          `json:"action"       db:"action"`
	ResourceType string          `json:"resourceType" db:"resource_type"`
	ResourceID   string          `json:"resourceId"   db:"resource_id"`
	Details      json.RawMessage `json:"details"      db:"details"`
	Timestamp    time.Time       `json:"timestamp"    db:"timestamp"`
}

// Remediation represents a corrective action taken in response to an incident.
type Remediation struct {
	ID          uuid.UUID         `json:"id"          db:"id"`
	IncidentID  uuid.UUID         `json:"incidentId"  db:"incident_id"`
	Description string            `json:"description" db:"description"`
	Type        string            `json:"type"        db:"type"`
	Status      RemediationStatus `json:"status"      db:"status"`
	ApprovedBy  *uuid.UUID        `json:"approvedBy"  db:"approved_by"`
	ExecutedAt  *time.Time        `json:"executedAt"  db:"executed_at"`
	Result      *string           `json:"result"      db:"result"`
	CreatedAt   time.Time         `json:"createdAt"   db:"created_at"`
}

// --- Pagination ---

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// DefaultPagination returns pagination parameters with sensible defaults.
func DefaultPagination() PaginationParams {
	return PaginationParams{Page: 1, PageSize: 20}
}

// Offset calculates the SQL OFFSET from page and page size.
func (p PaginationParams) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

// Limit returns the page size, clamped to a maximum of 100.
func (p PaginationParams) Limit() int {
	if p.PageSize <= 0 {
		return 20
	}
	if p.PageSize > 100 {
		return 100
	}
	return p.PageSize
}

// PaginatedResponse wraps a list of items with pagination metadata.
type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
}
