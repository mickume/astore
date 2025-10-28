package rbac

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/auth"
	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/log"
)

// Handler handles RBAC API requests
type Handler struct {
	policyEngine  *auth.PolicyEngine
	auditLogger   *auth.AuditLogger
	metadataStore *storage.MetadataStore
	logger        log.Logger
}

// NewHandler creates a new RBAC handler
func NewHandler(policyEngine *auth.PolicyEngine, auditLogger *auth.AuditLogger, metadataStore *storage.MetadataStore, logger log.Logger) *Handler {
	return &Handler{
		policyEngine:  policyEngine,
		auditLogger:   auditLogger,
		metadataStore: metadataStore,
		logger:        logger,
	}
}

// RegisterRoutes registers RBAC API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Policy management
	router.HandleFunc("/rbac/policies", h.CreatePolicy).Methods("POST")
	router.HandleFunc("/rbac/policies", h.ListPolicies).Methods("GET")
	router.HandleFunc("/rbac/policies/{id}", h.GetPolicy).Methods("GET")
	router.HandleFunc("/rbac/policies/{id}", h.UpdatePolicy).Methods("PUT")
	router.HandleFunc("/rbac/policies/{id}", h.DeletePolicy).Methods("DELETE")

	// Authorization check
	router.HandleFunc("/rbac/authorize", h.CheckAuthorization).Methods("POST")

	// Audit logs
	router.HandleFunc("/rbac/audit", h.ListAuditLogs).Methods("GET")
}

// === Policy Operations ===

// CreatePolicy creates a new access control policy
func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var policy models.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	// Validate policy
	if policy.Resource == "" {
		http.Error(w, "Resource is required", http.StatusBadRequest)
		return
	}
	if len(policy.Actions) == 0 {
		http.Error(w, "At least one action is required", http.StatusBadRequest)
		return
	}
	if policy.Effect == "" {
		policy.Effect = models.PolicyEffectAllow
	}

	// Store policy in database
	if err := h.metadataStore.StorePolicy(&policy); err != nil {
		h.logger.Error().Err(err).Msg("failed to store policy")
		http.Error(w, "Failed to store policy", http.StatusInternalServerError)
		return
	}

	// Add policy to engine
	h.policyEngine.AddPolicy(&policy)

	h.logger.Info().Str("policyId", policy.ID).Str("resource", policy.Resource).Msg("policy created")

	h.writeJSON(w, http.StatusCreated, policy)
}

// ListPolicies lists all policies
func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.metadataStore.ListPolicies()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list policies")
		http.Error(w, "Failed to list policies", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"policies": policies,
		"count":    len(policies),
	})
}

// GetPolicy retrieves a single policy by ID
func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	policy, err := h.metadataStore.GetPolicy(policyID)
	if err != nil {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, http.StatusOK, policy)
}

// UpdatePolicy updates an existing policy
func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	var policy models.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID matches
	policy.ID = policyID
	policy.UpdatedAt = time.Now()

	// Store updated policy
	if err := h.metadataStore.StorePolicy(&policy); err != nil {
		h.logger.Error().Err(err).Msg("failed to update policy")
		http.Error(w, "Failed to update policy", http.StatusInternalServerError)
		return
	}

	// Update policy in engine (remove and re-add)
	h.policyEngine.RemovePolicy(policyID)
	h.policyEngine.AddPolicy(&policy)

	h.logger.Info().Str("policyId", policyID).Msg("policy updated")

	h.writeJSON(w, http.StatusOK, policy)
}

// DeletePolicy deletes a policy
func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	// Delete from database
	if err := h.metadataStore.DeletePolicy(policyID); err != nil {
		h.logger.Error().Err(err).Str("policyId", policyID).Msg("failed to delete policy")
		http.Error(w, "Failed to delete policy", http.StatusInternalServerError)
		return
	}

	// Remove from engine
	h.policyEngine.RemovePolicy(policyID)

	h.logger.Info().Str("policyId", policyID).Msg("policy deleted")

	w.WriteHeader(http.StatusNoContent)
}

// === Authorization ===

// AuthorizationRequest represents an authorization check request
type AuthorizationRequest struct {
	UserID   string       `json:"userId"`
	Resource string       `json:"resource"`
	Action   models.Action `json:"action"`
}

// AuthorizationResponse represents an authorization check response
type AuthorizationResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// CheckAuthorization checks if a user is authorized to perform an action
func (h *Handler) CheckAuthorization(w http.ResponseWriter, r *http.Request) {
	var req AuthorizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, we simulate a user lookup
	// In production, this would look up the actual user from the database
	user := &models.User{
		ID:       req.UserID,
		Username: req.UserID,
		Roles:    []string{}, // Would be loaded from database
		Groups:   []string{},
	}

	// Check authorization
	allowed, err := h.policyEngine.Authorize(user, req.Resource, req.Action)

	response := AuthorizationResponse{
		Allowed: allowed,
	}

	if err != nil {
		response.Reason = err.Error()
	}

	h.writeJSON(w, http.StatusOK, response)
}

// === Audit Logs ===

// ListAuditLogs retrieves audit logs with optional filtering
func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	userID := r.URL.Query().Get("userId")
	resource := r.URL.Query().Get("resource")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Parse time range
	var startTime, endTime time.Time
	if startStr := r.URL.Query().Get("startTime"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := r.URL.Query().Get("endTime"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	// Retrieve logs
	logs, err := h.auditLogger.GetAuditLogs(userID, resource, startTime, endTime, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to retrieve audit logs")
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// === Helper Functions ===

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
