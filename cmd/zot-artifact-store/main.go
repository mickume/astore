package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/candlekeep/zot-artifact-store/internal/extensions"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/metrics"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/rbac"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/s3api"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/supplychain"
	"zotregistry.io/zot/pkg/api"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
)

const version = "0.1.0-dev"

func main() {
	fmt.Printf("Zot Artifact Store v%s\n", version)
	fmt.Println("Starting...")

	// Create logger
	logger := log.NewLogger("info", "")

	// Create default configuration
	cfg := config.New()
	cfg.HTTP.Address = "0.0.0.0"
	cfg.HTTP.Port = "8080"

	// Create and configure Zot controller
	ctlr := api.NewController(cfg)
	ctlr.Log = logger

	// Initialize extension registry
	logger.Info().Msg("Initializing extension registry")
	extRegistry := extensions.NewRegistry(logger)

	// Register all extensions
	logger.Info().Msg("Registering extensions")
	if err := extRegistry.Register(s3api.NewS3APIExtension()); err != nil {
		logger.Error().Err(err).Msg("Failed to register S3 API extension")
		os.Exit(1)
	}
	if err := extRegistry.Register(rbac.NewRBACExtension()); err != nil {
		logger.Error().Err(err).Msg("Failed to register RBAC extension")
		os.Exit(1)
	}
	if err := extRegistry.Register(supplychain.NewSupplyChainExtension()); err != nil {
		logger.Error().Err(err).Msg("Failed to register supply chain extension")
		os.Exit(1)
	}
	if err := extRegistry.Register(metrics.NewMetricsExtension()); err != nil {
		logger.Error().Err(err).Msg("Failed to register metrics extension")
		os.Exit(1)
	}

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info().Msg("Received shutdown signal")
		// Shutdown extensions first
		if err := extRegistry.ShutdownAll(context.Background()); err != nil {
			logger.Error().Err(err).Msg("Error shutting down extensions")
		}
		cancel()
	}()

	// TODO: Load configuration from file
	// TODO: Initialize storage backends
	// TODO: Setup extensions with storage controller (when available)
	// TODO: Register extension routes (when router is available)

	logger.Info().Str("address", cfg.HTTP.Address).Str("port", cfg.HTTP.Port).Msg("Starting Zot Artifact Store")
	logger.Info().Int("extensions", len(extRegistry.GetAll())).Msg("Extensions registered")

	// Start the server
	if err := ctlr.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("Server error")
		os.Exit(1)
	}

	logger.Info().Msg("Server stopped gracefully")
}
