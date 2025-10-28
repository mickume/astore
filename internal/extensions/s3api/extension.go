package s3api

import (
	"context"

	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	"zotregistry.io/zot/pkg/storage"
)

// S3APIExtension provides S3-compatible API for binary artifact storage
type S3APIExtension struct {
	config          *Config
	logger          log.Logger
	storeController storage.StoreController
}

// Config holds the S3 API extension configuration
type Config struct {
	Enabled            bool   `json:"enabled" mapstructure:"enabled"`
	BasePath           string `json:"basePath" mapstructure:"basePath"`
	MaxUploadSize      int64  `json:"maxUploadSize" mapstructure:"maxUploadSize"`
	EnableMultipart    bool   `json:"enableMultipart" mapstructure:"enableMultipart"`
	EnablePresignedURL bool   `json:"enablePresignedURL" mapstructure:"enablePresignedURL"`
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
func (e *S3APIExtension) Setup(cfg *config.Config, storeController storage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// TODO: Load extension-specific configuration
	e.config = &Config{
		Enabled:           true,
		BasePath:          "/s3",
		MaxUploadSize:     5 * 1024 * 1024 * 1024, // 5GB default
		EnableMultipart:   true,
		EnablePresignedURL: true,
	}

	e.logger.Info().Msg("S3 API extension initialized")
	return nil
}

// RegisterRoutes registers S3-compatible API routes
func (e *S3APIExtension) RegisterRoutes(router *mux.Router, storeController storage.StoreController) error {
	// S3 API routes will be implemented in Phase 2
	// TODO: Implement S3-compatible endpoints:
	// - PUT /s3/{bucket}/{key} - Upload object
	// - GET /s3/{bucket}/{key} - Download object
	// - DELETE /s3/{bucket}/{key} - Delete object
	// - HEAD /s3/{bucket}/{key} - Get object metadata
	// - GET /s3/{bucket}?list-type=2 - List objects
	// - POST /s3/{bucket}/{key}?uploads - Initiate multipart upload
	// - PUT /s3/{bucket}/{key}?uploadId={id}&partNumber={n} - Upload part
	// - POST /s3/{bucket}/{key}?uploadId={id} - Complete multipart upload
	// - DELETE /s3/{bucket}/{key}?uploadId={id} - Abort multipart upload

	e.logger.Info().Msg("S3 API routes registration (stub - to be implemented in Phase 2)")
	return nil
}

// Shutdown performs cleanup
func (e *S3APIExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("S3 API extension shutdown")
	return nil
}
