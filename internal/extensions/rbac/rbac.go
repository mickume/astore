package rbac

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/candlekeep/zot-artifact-store/internal/auth"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	zotStorage "zotregistry.io/zot/pkg/storage"
)

// RBACExtension provides role-based access control with Keycloak integration
type RBACExtension struct {
	config          *Config
	logger          log.Logger
	storeController zotStorage.StoreController
	metadataStore   *storage.MetadataStore
	jwtValidator    *auth.JWTValidator
	policyEngine    *auth.PolicyEngine
	auditLogger     *auth.AuditLogger
	middleware      *auth.Middleware
	handler         *Handler
}

// Config holds the RBAC extension configuration
type Config struct {
	Enabled            bool           `json:"enabled" mapstructure:"enabled"`
	Keycloak           KeycloakConfig `json:"keycloak" mapstructure:"keycloak"`
	AuditLogging       bool           `json:"auditLogging" mapstructure:"auditLogging"`
	AllowAnonymousGet  bool           `json:"allowAnonymousGet" mapstructure:"allowAnonymousGet"`
	MetadataDBPath     string         `json:"metadataDBPath" mapstructure:"metadataDBPath"`
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
	// RBAC is optional, defaults to disabled
	return e.config != nil && e.config.Enabled
}

// Setup initializes the extension
func (e *RBACExtension) Setup(cfg *config.Config, storeController zotStorage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// Load extension-specific configuration with defaults
	e.config = &Config{
		Enabled:           false, // Disabled by default
		AuditLogging:      true,
		AllowAnonymousGet: false,
		Keycloak: KeycloakConfig{
			URL:   "http://localhost:8081",  // Default Keycloak URL
			Realm: "zot-artifact-store",     // Default realm
		},
	}

	// Set metadata DB path
	if cfg.Storage.RootDirectory != "" {
		e.config.MetadataDBPath = filepath.Join(cfg.Storage.RootDirectory, "metadata.db")
	} else {
		e.config.MetadataDBPath = "/tmp/zot-artifacts/metadata.db"
	}

	// Initialize metadata store (shared with S3 API extension)
	metadataStore, err := storage.NewMetadataStore(e.config.MetadataDBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	e.metadataStore = metadataStore

	// Initialize JWT validator for Keycloak
	e.jwtValidator = auth.NewJWTValidator(e.config.Keycloak.URL, e.config.Keycloak.Realm)

	// Initialize policy engine
	e.policyEngine = auth.NewPolicyEngine(e.config.AllowAnonymousGet)

	// Load existing policies from database
	if err := e.loadPolicies(); err != nil {
		logger.Warn().Err(err).Msg("failed to load policies from database")
	}

	// Initialize audit logger
	e.auditLogger = auth.NewAuditLogger(metadataStore, logger, e.config.AuditLogging)

	// Initialize auth middleware
	e.middleware = auth.NewMiddleware(e.jwtValidator, e.policyEngine, logger, e.config.Enabled)

	// Initialize RBAC API handler
	e.handler = NewHandler(e.policyEngine, e.auditLogger, e.metadataStore, logger)

	e.logger.Info().
		Str("keycloakURL", e.config.Keycloak.URL).
		Str("realm", e.config.Keycloak.Realm).
		Bool("enabled", e.config.Enabled).
		Bool("auditLogging", e.config.AuditLogging).
		Bool("allowAnonymousGet", e.config.AllowAnonymousGet).
		Msg("RBAC extension initialized")

	return nil
}

// RegisterRoutes registers RBAC-related routes
func (e *RBACExtension) RegisterRoutes(router *mux.Router, storeController zotStorage.StoreController) error {
	if e.handler == nil {
		return fmt.Errorf("RBAC handler not initialized")
	}

	e.handler.RegisterRoutes(router)
	e.logger.Info().Msg("RBAC routes registered")

	return nil
}

// Shutdown performs cleanup
func (e *RBACExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("RBAC extension shutdown")

	if e.metadataStore != nil {
		if err := e.metadataStore.Close(); err != nil {
			e.logger.Error().Err(err).Msg("failed to close metadata store")
			return err
		}
	}

	return nil
}

// GetMiddleware returns the authentication middleware
func (e *RBACExtension) GetMiddleware() *auth.Middleware {
	return e.middleware
}

// GetAuditLogger returns the audit logger
func (e *RBACExtension) GetAuditLogger() *auth.AuditLogger {
	return e.auditLogger
}

// loadPolicies loads existing policies from the database into the policy engine
func (e *RBACExtension) loadPolicies() error {
	policies, err := e.metadataStore.ListPolicies()
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	for _, policy := range policies {
		e.policyEngine.AddPolicy(policy)
	}

	e.logger.Info().Int("count", len(policies)).Msg("loaded policies from database")
	return nil
}
