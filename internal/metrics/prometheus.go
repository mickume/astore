package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusCollector collects and exposes Prometheus metrics
type PrometheusCollector struct {
	// Artifact metrics
	artifactUploads   *prometheus.CounterVec
	artifactDownloads *prometheus.CounterVec
	artifactDeletes   *prometheus.CounterVec
	artifactSizes     *prometheus.HistogramVec
	artifactDuration  *prometheus.HistogramVec

	// Supply chain metrics
	signingOperations      *prometheus.CounterVec
	verificationOperations *prometheus.CounterVec
	sbomOperations         *prometheus.CounterVec
	attestationOperations  *prometheus.CounterVec
	supplyChainDuration    *prometheus.HistogramVec

	// RBAC metrics
	authenticationAttempts *prometheus.CounterVec
	authorizationChecks    *prometheus.CounterVec

	// System metrics
	activeConnections prometheus.Gauge
	totalRequests     *prometheus.CounterVec
	errorRate         *prometheus.CounterVec
}

// NewPrometheusCollector creates a new Prometheus metrics collector
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		// Artifact metrics
		artifactUploads: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "artifact_uploads_total",
				Help: "Total number of artifact uploads",
			},
			[]string{"bucket", "type", "status"},
		),
		artifactDownloads: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "artifact_downloads_total",
				Help: "Total number of artifact downloads",
			},
			[]string{"bucket", "type", "status"},
		),
		artifactDeletes: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "artifact_deletes_total",
				Help: "Total number of artifact deletes",
			},
			[]string{"bucket", "status"},
		),
		artifactSizes: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "artifact_size_bytes",
				Help:    "Size of artifacts in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 10, 8), // 1KB to ~10GB
			},
			[]string{"bucket", "type", "operation"},
		),
		artifactDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "artifact_operation_duration_seconds",
				Help:    "Duration of artifact operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "bucket"},
		),

		// Supply chain metrics
		signingOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "supplychain_signing_operations_total",
				Help: "Total number of artifact signing operations",
			},
			[]string{"status"},
		),
		verificationOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "supplychain_verification_operations_total",
				Help: "Total number of signature verification operations",
			},
			[]string{"status"},
		),
		sbomOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "supplychain_sbom_operations_total",
				Help: "Total number of SBOM operations",
			},
			[]string{"operation", "format", "status"},
		),
		attestationOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "supplychain_attestation_operations_total",
				Help: "Total number of attestation operations",
			},
			[]string{"operation", "type", "status"},
		),
		supplyChainDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "supplychain_operation_duration_seconds",
				Help:    "Duration of supply chain operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),

		// RBAC metrics
		authenticationAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rbac_authentication_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"method", "status"},
		),
		authorizationChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rbac_authorization_checks_total",
				Help: "Total number of authorization checks",
			},
			[]string{"resource", "action", "result"},
		),

		// System metrics
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_active_connections",
				Help: "Number of active connections",
			},
		),
		totalRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "system_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		errorRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "system_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "component"},
		),
	}
}

// RecordArtifactUpload records metrics for artifact upload operations
func (c *PrometheusCollector) RecordArtifactUpload(ctx context.Context, bucket, artifactType string, size int64, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.artifactUploads.WithLabelValues(bucket, artifactType, status).Inc()
	c.artifactSizes.WithLabelValues(bucket, artifactType, "upload").Observe(float64(size))
	c.artifactDuration.WithLabelValues("upload", bucket).Observe(duration.Seconds())
}

// RecordArtifactDownload records metrics for artifact download operations
func (c *PrometheusCollector) RecordArtifactDownload(ctx context.Context, bucket, artifactType string, size int64, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.artifactDownloads.WithLabelValues(bucket, artifactType, status).Inc()
	c.artifactSizes.WithLabelValues(bucket, artifactType, "download").Observe(float64(size))
	c.artifactDuration.WithLabelValues("download", bucket).Observe(duration.Seconds())
}

// RecordArtifactDelete records metrics for artifact delete operations
func (c *PrometheusCollector) RecordArtifactDelete(ctx context.Context, bucket string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.artifactDeletes.WithLabelValues(bucket, status).Inc()
}

// RecordSigningOperation records metrics for artifact signing operations
func (c *PrometheusCollector) RecordSigningOperation(ctx context.Context, operation string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.signingOperations.WithLabelValues(status).Inc()
	c.supplyChainDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordVerificationOperation records metrics for signature verification operations
func (c *PrometheusCollector) RecordVerificationOperation(ctx context.Context, operation string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.verificationOperations.WithLabelValues(status).Inc()
	c.supplyChainDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordSBOMOperation records metrics for SBOM operations
func (c *PrometheusCollector) RecordSBOMOperation(ctx context.Context, operation, format string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.sbomOperations.WithLabelValues(operation, format, status).Inc()
}

// RecordAttestationOperation records metrics for attestation operations
func (c *PrometheusCollector) RecordAttestationOperation(ctx context.Context, operation, attestationType string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.attestationOperations.WithLabelValues(operation, attestationType, status).Inc()
}

// RecordAuthenticationAttempt records metrics for authentication attempts
func (c *PrometheusCollector) RecordAuthenticationAttempt(ctx context.Context, method string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	c.authenticationAttempts.WithLabelValues(method, status).Inc()
}

// RecordAuthorizationCheck records metrics for authorization checks
func (c *PrometheusCollector) RecordAuthorizationCheck(ctx context.Context, resource, action string, allowed bool) {
	result := "allowed"
	if !allowed {
		result = "denied"
	}

	c.authorizationChecks.WithLabelValues(resource, action, result).Inc()
}

// RecordHTTPRequest records metrics for HTTP requests
func (c *PrometheusCollector) RecordHTTPRequest(method, endpoint string, statusCode int) {
	c.totalRequests.WithLabelValues(method, endpoint, string(rune(statusCode))).Inc()
}

// RecordError records error metrics
func (c *PrometheusCollector) RecordError(errorType, component string) {
	c.errorRate.WithLabelValues(errorType, component).Inc()
}

// IncrementActiveConnections increments the active connections gauge
func (c *PrometheusCollector) IncrementActiveConnections() {
	c.activeConnections.Inc()
}

// DecrementActiveConnections decrements the active connections gauge
func (c *PrometheusCollector) DecrementActiveConnections() {
	c.activeConnections.Dec()
}
