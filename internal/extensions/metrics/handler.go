package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/candlekeep/zot-artifact-store/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"zotregistry.io/zot/pkg/log"
)

// Handler handles metrics and health check API requests
type Handler struct {
	prometheusMetrics *metrics.PrometheusCollector
	healthChecker     *metrics.HealthChecker
	tracingProvider   *metrics.TracingProvider
	logger            log.Logger
}

// NewHandler creates a new metrics handler
func NewHandler(
	prometheusMetrics *metrics.PrometheusCollector,
	healthChecker *metrics.HealthChecker,
	tracingProvider *metrics.TracingProvider,
	logger log.Logger,
) *Handler {
	return &Handler{
		prometheusMetrics: prometheusMetrics,
		healthChecker:     healthChecker,
		tracingProvider:   tracingProvider,
		logger:            logger,
	}
}

// RegisterRoutes registers metrics and health check routes
func (h *Handler) RegisterRoutes(router *mux.Router, config *Config) {
	// Prometheus metrics endpoint
	if config.Prometheus.Enabled {
		router.Handle(config.Prometheus.Path, promhttp.Handler()).Methods("GET")
	}

	// Health check endpoints
	if config.Health.Enabled {
		router.HandleFunc(config.Health.HealthPath, h.GetHealth).Methods("GET")
		router.HandleFunc(config.Health.ReadinessPath, h.GetReadiness).Methods("GET")
		router.HandleFunc(config.Health.LivenessPath, h.GetLiveness).Methods("GET")
	}
}

// GetHealth returns comprehensive health information
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	health := h.healthChecker.CheckHealth(ctx)

	h.writeJSON(w, http.StatusOK, health)
}

// GetReadiness returns readiness status (for Kubernetes readiness probes)
func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ready := h.healthChecker.CheckReadiness(ctx)

	if ready {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ready",
		})
	} else {
		h.writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not_ready",
		})
	}
}

// GetLiveness returns liveness status (for Kubernetes liveness probes)
func (h *Handler) GetLiveness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alive := h.healthChecker.CheckLiveness(ctx)

	if alive {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "alive",
		})
	} else {
		h.writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "dead",
		})
	}
}

// === Helper Functions ===

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
