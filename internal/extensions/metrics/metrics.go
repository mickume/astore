package metrics

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/candlekeep/zot-artifact-store/internal/metrics"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	zotStorage "zotregistry.io/zot/pkg/storage"
)

// MetricsExtension provides enhanced metrics and observability
type MetricsExtension struct {
	config            *Config
	logger            log.Logger
	storeController   zotStorage.StoreController
	metadataStore     *storage.MetadataStore
	prometheusMetrics *metrics.PrometheusCollector
	healthChecker     *metrics.HealthChecker
	tracingProvider   *metrics.TracingProvider
	handler           *Handler
}

// Config holds the metrics extension configuration
type Config struct {
	Enabled         bool          `json:"enabled" mapstructure:"enabled"`
	Prometheus      PrometheusCfg `json:"prometheus" mapstructure:"prometheus"`
	Tracing         TracingCfg    `json:"tracing" mapstructure:"tracing"`
	Health          HealthCfg     `json:"health" mapstructure:"health"`
	MetadataDBPath  string        `json:"metadataDBPath" mapstructure:"metadataDBPath"`
}

// PrometheusCfg holds Prometheus configuration
type PrometheusCfg struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Path    string `json:"path" mapstructure:"path"` // Default: /metrics
}

// TracingCfg holds OpenTelemetry tracing configuration
type TracingCfg struct {
	Enabled     bool   `json:"enabled" mapstructure:"enabled"`
	Endpoint    string `json:"endpoint" mapstructure:"endpoint"`     // OTLP endpoint
	ServiceName string `json:"serviceName" mapstructure:"serviceName"` // Service name for traces
}

// HealthCfg holds health check configuration
type HealthCfg struct {
	Enabled       bool   `json:"enabled" mapstructure:"enabled"`
	ReadinessPath string `json:"readinessPath" mapstructure:"readinessPath"` // Default: /health/ready
	LivenessPath  string `json:"livenessPath" mapstructure:"livenessPath"`   // Default: /health/live
	HealthPath    string `json:"healthPath" mapstructure:"healthPath"`       // Default: /health
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
	return e.config != nil && e.config.Enabled
}

// Setup initializes the extension
func (e *MetricsExtension) Setup(cfg *config.Config, storeController zotStorage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// Load extension-specific configuration with defaults
	e.config = &Config{
		Enabled: true, // Enabled by default
		Prometheus: PrometheusCfg{
			Enabled: true,
			Path:    "/metrics",
		},
		Tracing: TracingCfg{
			Enabled:     false, // Disabled by default (requires OTLP endpoint)
			Endpoint:    "",
			ServiceName: "zot-artifact-store",
		},
		Health: HealthCfg{
			Enabled:       true,
			ReadinessPath: "/health/ready",
			LivenessPath:  "/health/live",
			HealthPath:    "/health",
		},
	}

	// Set metadata DB path
	if cfg.Storage.RootDirectory != "" {
		e.config.MetadataDBPath = filepath.Join(cfg.Storage.RootDirectory, "metadata.db")
	} else {
		e.config.MetadataDBPath = "/tmp/zot-artifacts/metadata.db"
	}

	// Initialize metadata store (shared with other extensions)
	metadataStore, err := storage.NewMetadataStore(e.config.MetadataDBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	e.metadataStore = metadataStore

	// Initialize Prometheus metrics collector
	if e.config.Prometheus.Enabled {
		e.prometheusMetrics = metrics.NewPrometheusCollector()
		e.logger.Info().Msg("Prometheus metrics collector initialized")
	}

	// Initialize health checker
	if e.config.Health.Enabled {
		e.healthChecker = metrics.NewHealthChecker(e.metadataStore, logger, "1.0.0")
		e.logger.Info().Msg("Health checker initialized")
	}

	// Initialize tracing provider
	if e.config.Tracing.Enabled && e.config.Tracing.Endpoint != "" {
		ctx := context.Background()
		tracingProvider, err := metrics.NewTracingProvider(
			ctx,
			e.config.Tracing.Endpoint,
			e.config.Tracing.ServiceName,
			logger,
		)
		if err != nil {
			e.logger.Warn().Err(err).Msg("Failed to initialize tracing, continuing without it")
		} else {
			e.tracingProvider = tracingProvider
		}
	}

	// Initialize API handler
	e.handler = NewHandler(e.prometheusMetrics, e.healthChecker, e.tracingProvider, logger)

	e.logger.Info().
		Bool("enabled", e.config.Enabled).
		Bool("prometheusEnabled", e.config.Prometheus.Enabled).
		Bool("tracingEnabled", e.config.Tracing.Enabled).
		Bool("healthEnabled", e.config.Health.Enabled).
		Msg("Metrics extension initialized")

	return nil
}

// RegisterRoutes registers metrics and health check routes
func (e *MetricsExtension) RegisterRoutes(router *mux.Router, storeController zotStorage.StoreController) error {
	if e.handler == nil {
		return fmt.Errorf("metrics handler not initialized")
	}

	e.handler.RegisterRoutes(router, e.config)
	e.logger.Info().Msg("Metrics extension routes registered")

	return nil
}

// Shutdown performs cleanup
func (e *MetricsExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("Metrics extension shutdown")

	// Shutdown tracing provider
	if e.tracingProvider != nil {
		if err := e.tracingProvider.Shutdown(ctx); err != nil {
			e.logger.Error().Err(err).Msg("failed to shutdown tracing provider")
		}
	}

	// Close metadata store
	if e.metadataStore != nil {
		if err := e.metadataStore.Close(); err != nil {
			e.logger.Error().Err(err).Msg("failed to close metadata store")
			return err
		}
	}

	return nil
}

// GetPrometheusCollector returns the Prometheus metrics collector
func (e *MetricsExtension) GetPrometheusCollector() *metrics.PrometheusCollector {
	return e.prometheusMetrics
}

// GetHealthChecker returns the health checker
func (e *MetricsExtension) GetHealthChecker() *metrics.HealthChecker {
	return e.healthChecker
}

// GetTracingProvider returns the tracing provider
func (e *MetricsExtension) GetTracingProvider() *metrics.TracingProvider {
	return e.tracingProvider
}
