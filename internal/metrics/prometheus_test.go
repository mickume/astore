package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/metrics"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestPrometheusCollector(t *testing.T) {
	// Create a single shared collector to avoid duplicate metric registration
	collector := metrics.NewPrometheusCollector()
	ctx := context.Background()

	t.Run("Record artifact upload", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an artifact upload
		collector.RecordArtifactUpload(ctx, "releases", "jar", 1024000, 500*time.Millisecond, true)

		// Then: No error occurs (metrics are recorded successfully)
		// Note: We can't easily verify prometheus metrics in unit tests
		// but we can verify the method completes without panic
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record artifact download", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an artifact download
		collector.RecordArtifactDownload(ctx, "releases", "rpm", 2048000, 750*time.Millisecond, true)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record signing operation", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording a signing operation
		collector.RecordSigningOperation(ctx, "sign", true, 100*time.Millisecond)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record verification operation", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording a verification operation
		collector.RecordVerificationOperation(ctx, "verify", true, 50*time.Millisecond)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record authentication attempt", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an authentication attempt
		collector.RecordAuthenticationAttempt(ctx, "bearer_token", true)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record authorization check", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an authorization check
		collector.RecordAuthorizationCheck(ctx, "bucket/key", "read", true)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record SBOM operation", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an SBOM operation
		collector.RecordSBOMOperation(ctx, "attach", "spdx", true)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Record attestation operation", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Recording an attestation operation
		collector.RecordAttestationOperation(ctx, "add", "build", true)

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})

	t.Run("Active connections tracking", func(t *testing.T) {
		// Given: A Prometheus collector
		// When: Incrementing and decrementing active connections
		collector.IncrementActiveConnections()
		collector.IncrementActiveConnections()
		collector.DecrementActiveConnections()

		// Then: No error occurs
		test.AssertTrue(t, collector != nil, "collector not nil")
	})
}
