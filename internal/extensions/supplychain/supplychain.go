package supplychain

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/candlekeep/zot-artifact-store/internal/storage"
	scPkg "github.com/candlekeep/zot-artifact-store/internal/supplychain"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	zotStorage "zotregistry.io/zot/pkg/storage"
)

// SupplyChainExtension provides artifact signing, SBOM, and attestation capabilities
type SupplyChainExtension struct {
	config          *Config
	logger          log.Logger
	storeController zotStorage.StoreController
	metadataStore   *storage.MetadataStore
	signer          *scPkg.Signer
	handler         *Handler
}

// Config holds the supply chain security extension configuration
type Config struct {
	Enabled        bool           `json:"enabled" mapstructure:"enabled"`
	Signing        SigningConfig  `json:"signing" mapstructure:"signing"`
	SBOM           SBOMConfig     `json:"sbom" mapstructure:"sbom"`
	Attestation    AttestConfig   `json:"attestation" mapstructure:"attestation"`
	MetadataDBPath string         `json:"metadataDBPath" mapstructure:"metadataDBPath"`
	PrivateKeyPath string         `json:"privateKeyPath" mapstructure:"privateKeyPath"`
}

// SigningConfig holds artifact signing configuration
type SigningConfig struct {
	Enabled   bool     `json:"enabled" mapstructure:"enabled"`
	Providers []string `json:"providers" mapstructure:"providers"` // rsa, cosign, notary
	Verify    bool     `json:"verify" mapstructure:"verify"`
}

// SBOMConfig holds SBOM configuration
type SBOMConfig struct {
	Enabled bool     `json:"enabled" mapstructure:"enabled"`
	Formats []string `json:"formats" mapstructure:"formats"` // spdx, cyclonedx
	Require bool     `json:"require" mapstructure:"require"`
}

// AttestConfig holds attestation configuration
type AttestConfig struct {
	Enabled bool     `json:"enabled" mapstructure:"enabled"`
	Types   []string `json:"types" mapstructure:"types"` // provenance, vulnerability, quality
}

// NewSupplyChainExtension creates a new supply chain security extension
func NewSupplyChainExtension() *SupplyChainExtension {
	return &SupplyChainExtension{}
}

// Name returns the extension name
func (e *SupplyChainExtension) Name() string {
	return "supplychain"
}

// IsEnabled checks if the extension is enabled
func (e *SupplyChainExtension) IsEnabled(cfg *config.Config) bool {
	// Supply chain is optional, defaults to disabled
	return e.config != nil && e.config.Enabled
}

// Setup initializes the extension
func (e *SupplyChainExtension) Setup(cfg *config.Config, storeController zotStorage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// Load extension-specific configuration with defaults
	e.config = &Config{
		Enabled: false, // Disabled by default
		Signing: SigningConfig{
			Enabled:   true,
			Providers: []string{"rsa"},
			Verify:    true,
		},
		SBOM: SBOMConfig{
			Enabled: true,
			Formats: []string{"spdx", "cyclonedx"},
			Require: false,
		},
		Attestation: AttestConfig{
			Enabled: true,
			Types:   []string{"build", "test", "scan", "provenance"},
		},
	}

	// Set metadata DB path
	if cfg.Storage.RootDirectory != "" {
		e.config.MetadataDBPath = filepath.Join(cfg.Storage.RootDirectory, "metadata.db")
	} else {
		e.config.MetadataDBPath = "/tmp/zot-artifacts/metadata.db"
	}

	// Initialize metadata store (shared with S3 API and RBAC extensions)
	metadataStore, err := storage.NewMetadataStore(e.config.MetadataDBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	e.metadataStore = metadataStore

	// Initialize signer (generate default key pair if not configured)
	// In production, keys would be loaded from secure storage
	signer, _, _, err := scPkg.GenerateKeyPair(2048)
	if err != nil {
		return fmt.Errorf("failed to generate signing key: %w", err)
	}
	e.signer = signer

	// Initialize supply chain handler
	e.handler = NewHandler(e.metadataStore, e.signer, logger)

	e.logger.Info().
		Bool("enabled", e.config.Enabled).
		Bool("signingEnabled", e.config.Signing.Enabled).
		Bool("sbomEnabled", e.config.SBOM.Enabled).
		Bool("attestationEnabled", e.config.Attestation.Enabled).
		Msg("Supply chain security extension initialized")

	return nil
}

// RegisterRoutes registers supply chain security routes
func (e *SupplyChainExtension) RegisterRoutes(router *mux.Router, storeController zotStorage.StoreController) error {
	if e.handler == nil {
		return fmt.Errorf("supply chain handler not initialized")
	}

	e.handler.RegisterRoutes(router)
	e.logger.Info().Msg("Supply chain security routes registered")

	return nil
}

// Shutdown performs cleanup
func (e *SupplyChainExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("Supply chain security extension shutdown")

	if e.metadataStore != nil {
		if err := e.metadataStore.Close(); err != nil {
			e.logger.Error().Err(err).Msg("failed to close metadata store")
			return err
		}
	}

	return nil
}

// GetSigner returns the artifact signer
func (e *SupplyChainExtension) GetSigner() *scPkg.Signer {
	return e.signer
}
