package rbac

import (
	"context"

	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	"zotregistry.io/zot/pkg/storage"
)

// RBACExtension provides role-based access control with Keycloak integration
type RBACExtension struct {
	config          *Config
	logger          log.Logger
	storeController storage.StoreController
}

// Config holds the RBAC extension configuration
type Config struct {
	Enabled      bool            `json:"enabled" mapstructure:"enabled"`
	Keycloak     KeycloakConfig  `json:"keycloak" mapstructure:"keycloak"`
	AuditLogging bool            `json:"auditLogging" mapstructure:"auditLogging"`
}

// KeycloakConfig holds Keycloak-specific configuration
type KeycloakConfig struct {
	URL          string `json:"url" mapstructure:"url"`
	Realm        string `json:"realm" mapstructure:"realm"`
	ClientID     string `json:"clientId" mapstructure:"clientId"`
	ClientSecret string `json:"clientSecret" mapstructure:"clientSecret"`
}

// NewRBACExtension creates a new RBAC extension
func NewRBACExtension() *RBACExtension {
	return &RBACExtension{}
}

// Name returns the extension name
func (e *RBACExtension) Name() string {
	return "rbac"
}

// IsEnabled checks if the extension is enabled
func (e *RBACExtension) IsEnabled(cfg *config.Config) bool {
	// TODO: Check actual config when extension config is implemented
	return true
}

// Setup initializes the extension
func (e *RBACExtension) Setup(cfg *config.Config, storeController storage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// TODO: Load extension-specific configuration
	// TODO: Initialize Keycloak client
	// TODO: Set up policy engine

	e.config = &Config{
		Enabled:      true,
		AuditLogging: true,
	}

	e.logger.Info().Msg("RBAC extension initialized")
	return nil
}

// RegisterRoutes registers RBAC-related routes
func (e *RBACExtension) RegisterRoutes(router *mux.Router, storeController storage.StoreController) error {
	// RBAC routes will be implemented in Phase 3
	// TODO: Implement RBAC endpoints:
	// - POST /rbac/policies - Create policy
	// - GET /rbac/policies - List policies
	// - GET /rbac/policies/{id} - Get policy
	// - PUT /rbac/policies/{id} - Update policy
	// - DELETE /rbac/policies/{id} - Delete policy
	// - GET /rbac/audit - Get audit logs
	// - POST /rbac/authorize - Check authorization

	e.logger.Info().Msg("RBAC routes registration (stub - to be implemented in Phase 3)")
	return nil
}

// Shutdown performs cleanup
func (e *RBACExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("RBAC extension shutdown")
	// TODO: Close Keycloak connections
	// TODO: Flush audit logs
	return nil
}
