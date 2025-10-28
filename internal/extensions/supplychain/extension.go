package supplychain

import (
	"context"

	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
	"zotregistry.io/zot/pkg/storage"
)

// SupplyChainExtension provides artifact signing, SBOM, and attestation capabilities
type SupplyChainExtension struct {
	config          *Config
	logger          log.Logger
	storeController storage.StoreController
}

// Config holds the supply chain security extension configuration
type Config struct {
	Enabled      bool           `json:"enabled" mapstructure:"enabled"`
	Signing      SigningConfig  `json:"signing" mapstructure:"signing"`
	SBOM         SBOMConfig     `json:"sbom" mapstructure:"sbom"`
	Attestation  AttestConfig   `json:"attestation" mapstructure:"attestation"`
}

// SigningConfig holds artifact signing configuration
type SigningConfig struct {
	Providers []string `json:"providers" mapstructure:"providers"` // cosign, notary, gpg
	Verify    bool     `json:"verify" mapstructure:"verify"`
}

// SBOMConfig holds SBOM configuration
type SBOMConfig struct {
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
	// TODO: Check actual config when extension config is implemented
	return true
}

// Setup initializes the extension
func (e *SupplyChainExtension) Setup(cfg *config.Config, storeController storage.StoreController, logger log.Logger) error {
	e.logger = logger
	e.storeController = storeController

	// TODO: Load extension-specific configuration
	// TODO: Initialize signing providers (Cosign, Notary)
	// TODO: Set up SBOM parsers (SPDX, CycloneDX)

	e.config = &Config{
		Enabled: true,
		Signing: SigningConfig{
			Providers: []string{"cosign", "notary"},
			Verify:    true,
		},
		SBOM: SBOMConfig{
			Formats: []string{"spdx", "cyclonedx"},
			Require: false,
		},
		Attestation: AttestConfig{
			Enabled: true,
			Types:   []string{"provenance", "vulnerability"},
		},
	}

	e.logger.Info().Msg("Supply chain security extension initialized")
	return nil
}

// RegisterRoutes registers supply chain security routes
func (e *SupplyChainExtension) RegisterRoutes(router *mux.Router, storeController storage.StoreController) error {
	// Supply chain routes will be implemented in Phase 4
	// TODO: Implement supply chain endpoints:
	// Signing:
	// - POST /supplychain/sign/{bucket}/{artifact} - Sign artifact
	// - GET /supplychain/signatures/{bucket}/{artifact} - Get signatures
	// - POST /supplychain/verify/{bucket}/{artifact} - Verify signatures
	// SBOM:
	// - POST /supplychain/sbom/{bucket}/{artifact} - Attach SBOM
	// - GET /supplychain/sbom/{bucket}/{artifact} - Get SBOM
	// - GET /supplychain/sbom/{bucket}/{artifact}/dependencies - Get dependencies
	// Attestations:
	// - POST /supplychain/attestations/{bucket}/{artifact} - Add attestation
	// - GET /supplychain/attestations/{bucket}/{artifact} - Get attestations
	// - GET /supplychain/attestations/{bucket}/{artifact}/{type} - Get specific attestation type

	e.logger.Info().Msg("Supply chain security routes registration (stub - to be implemented in Phase 4)")
	return nil
}

// Shutdown performs cleanup
func (e *SupplyChainExtension) Shutdown(ctx context.Context) error {
	e.logger.Info().Msg("Supply chain security extension shutdown")
	// TODO: Close signing provider connections
	return nil
}
