package metrics_test

import (
	"context"
	"os"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/metrics"
	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/candlekeep/zot-artifact-store/test"
	"zotregistry.io/zot/pkg/log"
)

func TestHealthChecker(t *testing.T) {
	t.Run("Check health with healthy components", func(t *testing.T) {
		// Given: A health checker with initialized components
		tmpFile, _ := os.CreateTemp("", "health-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		logger := log.NewLogger("debug", "")
		checker := metrics.NewHealthChecker(store, logger, "1.0.0-test")

		ctx := context.Background()

		// When: Checking health
		health := checker.CheckHealth(ctx)

		// Then: Overall status is healthy
		test.AssertEqual(t, models.HealthStatusHealthy, health.Status, "overall health status")
		test.AssertTrue(t, len(health.Checks) > 0, "health checks present")
		test.AssertEqual(t, "1.0.0-test", health.Version, "version")
	})

	t.Run("Check readiness with initialized store", func(t *testing.T) {
		// Given: A health checker with initialized metadata store
		tmpFile, _ := os.CreateTemp("", "health-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		logger := log.NewLogger("debug", "")
		checker := metrics.NewHealthChecker(store, logger, "1.0.0-test")

		ctx := context.Background()

		// When: Checking readiness
		ready := checker.CheckReadiness(ctx)

		// Then: Service is ready
		test.AssertTrue(t, ready, "service should be ready")
	})

	t.Run("Check readiness with nil store", func(t *testing.T) {
		// Given: A health checker without metadata store
		logger := log.NewLogger("debug", "")
		checker := metrics.NewHealthChecker(nil, logger, "1.0.0-test")

		ctx := context.Background()

		// When: Checking readiness
		ready := checker.CheckReadiness(ctx)

		// Then: Service is not ready
		test.AssertFalse(t, ready, "service should not be ready")
	})

	t.Run("Check liveness", func(t *testing.T) {
		// Given: A health checker
		logger := log.NewLogger("debug", "")
		checker := metrics.NewHealthChecker(nil, logger, "1.0.0-test")

		ctx := context.Background()

		// When: Checking liveness
		alive := checker.CheckLiveness(ctx)

		// Then: Service is alive (basic liveness always true if process running)
		test.AssertTrue(t, alive, "service should be alive")
	})

	t.Run("Metadata store check fails with unhealthy status", func(t *testing.T) {
		// Given: A health checker without initialized store
		logger := log.NewLogger("debug", "")
		checker := metrics.NewHealthChecker(nil, logger, "1.0.0-test")

		ctx := context.Background()

		// When: Checking health
		health := checker.CheckHealth(ctx)

		// Then: Metadata store check shows unhealthy
		test.AssertEqual(t, models.HealthStatusUnhealthy, health.Status, "overall health should be unhealthy")

		// Find metadata store check
		var metadataCheck *models.HealthCheck
		for i := range health.Checks {
			if health.Checks[i].Component == "metadata_store" {
				metadataCheck = &health.Checks[i]
				break
			}
		}

		test.AssertTrue(t, metadataCheck != nil, "metadata store check present")
		test.AssertEqual(t, models.HealthStatusUnhealthy, metadataCheck.Status, "metadata store unhealthy")
	})
}
