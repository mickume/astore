package metrics

import (
	"context"

	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	"zotregistry.io/zot/pkg/storage"
)

// MetricsExtension provides enhanced Prometheus metrics and OpenTelemetry tracing
type MetricsExtension struct {
	config          *Config
	logger          log.Logger
	storeController storage.StoreController
}

// Config holds the metrics extension configuration
type Config struct {
	Enabled    bool                  `json:"enabled" mapstructure:"enabled"`
	Prometheus PrometheusConfig      `json:"prometheus" mapstructure:"prometheus"`
	Tracing    TracingConfig         `json:"tracing" mapstructure:"tracing"`
}

// PrometheusConfig holds Prometheus-specific configuration
type PrometheusConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	Path       string `json:"path" mapstructure:"path"`
	Namespace  string `json:"namespace" mapstructure:"namespace"`
}

// TracingConfig holds OpenTelemetry tracing configuration
type TracingConfig struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	Endpoint     string `json:"endpoint" mapstructure:"endpoint"`
	ServiceName  string `json:"serviceName" mapstructure:"serviceName"`
}

// NewMetricsExtension creates a new metrics extension
func NewMetricsExtension() *MetricsExtension {
	return &MetricsExtension{}
}

// Name returns the extension name
func (e *MetricsExtension) Name() string {
	return "metrics"
}

// IsEnabled checks if the extension is enabled
func (e *MetricsExtension) IsEnabled(cfg *config.Config) bool {
	// TODO: Check actual config when extension config is implemented
	return true
}

// Setup initializes the extension
func (e *MetricsExtension) Setup(cfg *config.Config, storeController storage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// TODO: Load extension-specific configuration
	// TODO: Initialize Prometheus metrics
	// TODO: Set up OpenTelemetry tracing

	e.config = &Config{
		Enabled: true,
		Prometheus: PrometheusConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "zot_artifact_store",
		},
		Tracing: TracingConfig{
			Enabled:     true,
			Endpoint:    "localhost:4317",
			ServiceName: "zot-artifact-store",
		},
	}

	e.logger.Info().Msg("Enhanced metrics extension initialized")
	return nil
}

// RegisterRoutes registers metrics and health check routes
func (e *MetricsExtension) RegisterRoutes(router *mux.Router, storeController storage.StoreController) error {
	// Metrics routes will be implemented in Phase 6
	// TODO: Implement metrics endpoints:
	// - GET /metrics - Prometheus metrics
	// - GET /health - Health check
	// - GET /ready - Readiness check
	// - GET /live - Liveness check

	e.logger.Info().Msg("Enhanced metrics routes registration (stub - to be implemented in Phase 6)")
	return nil
}

// Shutdown performs cleanup
func (e *MetricsExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("Enhanced metrics extension shutdown")
	// TODO: Flush metrics
	// TODO: Close tracing connections
	return nil
}
