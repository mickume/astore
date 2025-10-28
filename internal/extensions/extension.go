package extensions

import (
	"context"

	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	"zotregistry.io/zot/pkg/storage"
)

// Extension defines the interface that all Zot Artifact Store extensions must implement.
// Extensions provide additional functionality beyond the core Zot registry capabilities.
type Extension interface {
	// Name returns the unique name of the extension
	Name() string

	// IsEnabled checks if the extension is enabled in the configuration
	IsEnabled(cfg *config.Config) bool

	// Setup initializes the extension with configuration and dependencies
	Setup(cfg *config.Config, storeController storage.StoreController, log log.Logger) error

	// RegisterRoutes registers HTTP routes for the extension
	// This is called during server initialization to add extension-specific endpoints
	RegisterRoutes(router *mux.Router, storeController storage.StoreController) error

	// Shutdown performs cleanup when the extension is being shut down
	Shutdown(ctx context.Context) error
}

// ExtensionConfig defines common configuration for all extensions
type ExtensionConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

// Registry manages all registered extensions
type Registry struct {
	extensions map[string]Extension
	logger     log.Logger
}

// NewRegistry creates a new extension registry
func NewRegistry(logger log.Logger) *Registry {
	return &Registry{
		extensions: make(map[string]Extension),
		logger:     logger,
	}
}

// Register adds an extension to the registry
func (r *Registry) Register(ext Extension) error {
	name := ext.Name()
	if _, exists := r.extensions[name]; exists {
		r.logger.Warn().Str("extension", name).Msg("extension already registered, overwriting")
	}
	r.extensions[name] = ext
	r.logger.Info().Str("extension", name).Msg("registered extension")
	return nil
}

// Get retrieves an extension by name
func (r *Registry) Get(name string) (Extension, bool) {
	ext, exists := r.extensions[name]
	return ext, exists
}

// GetAll returns all registered extensions
func (r *Registry) GetAll() map[string]Extension {
	return r.extensions
}

// SetupAll initializes all enabled extensions
func (r *Registry) SetupAll(cfg *config.Config, storeController storage.StoreController) error {
	for name, ext := range r.extensions {
		if !ext.IsEnabled(cfg) {
			r.logger.Info().Str("extension", name).Msg("extension disabled, skipping setup")
			continue
		}

		r.logger.Info().Str("extension", name).Msg("setting up extension")
		if err := ext.Setup(cfg, storeController, r.logger); err != nil {
			r.logger.Error().Err(err).Str("extension", name).Msg("failed to setup extension")
			return err
		}
		r.logger.Info().Str("extension", name).Msg("extension setup complete")
	}
	return nil
}

// RegisterAllRoutes registers HTTP routes for all enabled extensions
func (r *Registry) RegisterAllRoutes(router *mux.Router, cfg *config.Config, storeController storage.StoreController) error {
	for name, ext := range r.extensions {
		if !ext.IsEnabled(cfg) {
			continue
		}

		r.logger.Info().Str("extension", name).Msg("registering routes for extension")
		if err := ext.RegisterRoutes(router, storeController); err != nil {
			r.logger.Error().Err(err).Str("extension", name).Msg("failed to register routes")
			return err
		}
	}
	return nil
}

// ShutdownAll shuts down all extensions gracefully
func (r *Registry) ShutdownAll(ctx context.Context) error {
	for name, ext := range r.extensions {
		r.logger.Info().Str("extension", name).Msg("shutting down extension")
		if err := ext.Shutdown(ctx); err != nil {
			r.logger.Error().Err(err).Str("extension", name).Msg("error during extension shutdown")
			// Continue shutting down other extensions even if one fails
		}
	}
	return nil
}
