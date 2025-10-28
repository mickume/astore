package s3api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/candlekeep/zot-artifact-store/internal/api/s3"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	zotStorage "zotregistry.io/zot/pkg/storage"
)

// S3APIExtension provides S3-compatible API for binary artifact storage
type S3APIExtension struct {
	config          *Config
	logger          log.Logger
	storeController zotStorage.StoreController
	metadataStore   *storage.MetadataStore
	handler         *s3.Handler
	dataDir         string
}

// Config holds the S3 API extension configuration
type Config struct {
	Enabled            bool   `json:"enabled" mapstructure:"enabled"`
	BasePath           string `json:"basePath" mapstructure:"basePath"`
	MaxUploadSize      int64  `json:"maxUploadSize" mapstructure:"maxUploadSize"`
	EnableMultipart    bool   `json:"enableMultipart" mapstructure:"enableMultipart"`
	EnablePresignedURL bool   `json:"enablePresignedURL" mapstructure:"enablePresignedURL"`
	DataDir            string `json:"dataDir" mapstructure:"dataDir"`
	MetadataDBPath     string `json:"metadataDBPath" mapstructure:"metadataDBPath"`
}

// NewS3APIExtension creates a new S3 API extension
func NewS3APIExtension() *S3APIExtension {
	return &S3APIExtension{}
}

// Name returns the extension name
func (e *S3APIExtension) Name() string {
	return "s3api"
}

// IsEnabled checks if the extension is enabled
func (e *S3APIExtension) IsEnabled(cfg *config.Config) bool {
	// TODO: Check actual config when extension config is implemented
	return true
}

// Setup initializes the extension
func (e *S3APIExtension) Setup(cfg *config.Config, storeController zotStorage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// Load extension-specific configuration
	e.config = &Config{
		Enabled:            true,
		BasePath:           "/s3",
		MaxUploadSize:      5 * 1024 * 1024 * 1024, // 5GB default
		EnableMultipart:    true,
		EnablePresignedURL: true,
		DataDir:            cfg.Storage.RootDirectory,
		MetadataDBPath:     filepath.Join(cfg.Storage.RootDirectory, "metadata.db"),
	}

	// If no root directory is set, use a default
	if e.config.DataDir == "" {
		e.config.DataDir = "/tmp/zot-artifacts"
	}
	if e.config.MetadataDBPath == "" {
		e.config.MetadataDBPath = "/tmp/zot-artifacts/metadata.db"
	}

	e.dataDir = e.config.DataDir

	// Initialize metadata store
	metadataStore, err := storage.NewMetadataStore(e.config.MetadataDBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	e.metadataStore = metadataStore

	// Create S3 API handler
	e.handler = s3.NewHandler(metadataStore, e.dataDir, logger)
	// TODO: Integrate with Zot storage controller when needed

	e.logger.Info().
		Str("dataDir", e.config.DataDir).
		Str("metadataDB", e.config.MetadataDBPath).
		Msg("S3 API extension initialized")

	return nil
}

// RegisterRoutes registers S3-compatible API routes
func (e *S3APIExtension) RegisterRoutes(router *mux.Router, storeController zotStorage.StoreController) error {
	if e.handler == nil {
		return fmt.Errorf("S3 API handler not initialized")
	}

	e.handler.RegisterRoutes(router)
	e.logger.Info().Msg("S3 API routes registered")

	return nil
}

// Shutdown performs cleanup
func (e *S3APIExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("S3 API extension shutdown")

	if e.metadataStore != nil {
		if err := e.metadataStore.Close(); err != nil {
			e.logger.Error().Err(err).Msg("failed to close metadata store")
			return err
		}
	}

	return nil
}
