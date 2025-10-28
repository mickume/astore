package models

import (
	"time"
)

// User represents an authenticated user
type User struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Roles    []string          `json:"roles"`
	Groups   []string          `json:"groups"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Resource    string            `json:"resource"` // bucket name or "bucket/key" pattern
	Actions     []string          `json:"actions"`  // read, write, delete, list
	Effect      PolicyEffect      `json:"effect"`   // allow or deny
	Principals  []string          `json:"principals,omitempty"`
	Conditions  map[string]string `json:"conditions,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// PolicyEffect defines whether a policy allows or denies access
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
)

// Action represents an API action
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
	ActionList   Action = "list"
	ActionAdmin  Action = "admin"
)

// AuditLog represents an access audit log entry
type AuditLog struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	UserID    string            `json:"userId"`
	Username  string            `json:"username"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Method    string            `json:"method"`
	Status    int               `json:"status"`
	IPAddress string            `json:"ipAddress"`
	UserAgent string            `json:"userAgent"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Permission represents a user's permission for a resource
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []Action `json:"actions"`
}

// AuthContext holds authentication and authorization context
type AuthContext struct {
	User        *User
	Permissions []Permission
	IsAnonymous bool
}
