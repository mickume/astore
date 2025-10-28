package models

import "time"

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// ArtifactMetric represents metrics for artifact operations
type ArtifactMetric struct {
	Operation string        // upload, download, delete
	Bucket    string        // bucket name
	Type      string        // artifact type (jar, rpm, tar.gz, etc.)
	Size      int64         // artifact size in bytes
	Duration  time.Duration // operation duration
	Success   bool          // operation success status
	Timestamp time.Time     // when the metric was recorded
}

// SupplyChainMetric represents metrics for supply chain operations
type SupplyChainMetric struct {
	Operation string        // sign, verify, attach_sbom, add_attestation
	Success   bool          // operation success status
	Duration  time.Duration // operation duration
	Timestamp time.Time     // when the metric was recorded
}

// RBACMetric represents metrics for RBAC operations
type RBACMetric struct {
	Operation string    // authentication, authorization
	Method    string    // bearer_token, keycloak
	Resource  string    // resource being accessed
	Action    string    // read, write, delete, list
	Success   bool      // operation success status
	Timestamp time.Time // when the metric was recorded
}

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents the health status of a service component
type HealthCheck struct {
	Component string            `json:"component"`
	Status    HealthStatus      `json:"status"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// SystemHealth represents overall system health
type SystemHealth struct {
	Status     HealthStatus   `json:"status"`
	Checks     []HealthCheck  `json:"checks"`
	Version    string         `json:"version"`
	Uptime     time.Duration  `json:"uptime"`
	StartTime  time.Time      `json:"start_time"`
	CheckedAt  time.Time      `json:"checked_at"`
}
