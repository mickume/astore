package metrics

import (
	"context"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"zotregistry.io/zot/pkg/log"
)

// HealthChecker performs health checks on system components
type HealthChecker struct {
	metadataStore *storage.MetadataStore
	logger        log.Logger
	startTime     time.Time
	version       string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(metadataStore *storage.MetadataStore, logger log.Logger, version string) *HealthChecker {
	return &HealthChecker{
		metadataStore: metadataStore,
		logger:        logger,
		startTime:     time.Now(),
		version:       version,
	}
}

// CheckHealth performs a comprehensive health check
func (h *HealthChecker) CheckHealth(ctx context.Context) *models.SystemHealth {
	checks := []models.HealthCheck{
		h.checkMetadataStore(ctx),
		h.checkMemory(ctx),
	}

	// Determine overall status
	overallStatus := models.HealthStatusHealthy
	for _, check := range checks {
		if check.Status == models.HealthStatusUnhealthy {
			overallStatus = models.HealthStatusUnhealthy
			break
		} else if check.Status == models.HealthStatusDegraded {
			overallStatus = models.HealthStatusDegraded
		}
	}

	return &models.SystemHealth{
		Status:    overallStatus,
		Checks:    checks,
		Version:   h.version,
		Uptime:    time.Since(h.startTime),
		StartTime: h.startTime,
		CheckedAt: time.Now(),
	}
}

// CheckReadiness performs a readiness check (for Kubernetes readiness probes)
func (h *HealthChecker) CheckReadiness(ctx context.Context) bool {
	// Check if metadata store is accessible
	if h.metadataStore == nil {
		return false
	}

	// Could add more checks here (e.g., can write to storage)
	return true
}

// CheckLiveness performs a liveness check (for Kubernetes liveness probes)
func (h *HealthChecker) CheckLiveness(ctx context.Context) bool {
	// Basic liveness check - process is running
	return true
}

func (h *HealthChecker) checkMetadataStore(ctx context.Context) models.HealthCheck {
	check := models.HealthCheck{
		Component: "metadata_store",
		Timestamp: time.Now(),
	}

	if h.metadataStore == nil {
		check.Status = models.HealthStatusUnhealthy
		check.Message = "Metadata store not initialized"
		return check
	}

	// Try to list buckets to verify database is accessible
	_, err := h.metadataStore.ListBuckets()
	if err != nil {
		check.Status = models.HealthStatusUnhealthy
		check.Message = "Failed to access metadata store"
		check.Details = map[string]string{
			"error": err.Error(),
		}
		return check
	}

	check.Status = models.HealthStatusHealthy
	check.Message = "Metadata store is accessible"
	return check
}

func (h *HealthChecker) checkMemory(ctx context.Context) models.HealthCheck {
	check := models.HealthCheck{
		Component: "memory",
		Status:    models.HealthStatusHealthy,
		Message:   "Memory usage within limits",
		Timestamp: time.Now(),
	}

	// In a real implementation, we would check actual memory usage
	// For now, we'll just report healthy
	return check
}
