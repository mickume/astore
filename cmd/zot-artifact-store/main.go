package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/candlekeep/zot-artifact-store/internal/extensions"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/metrics"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/rbac"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/s3api"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/supplychain"
	"gopkg.in/yaml.v2"
	"zotregistry.io/zot/pkg/api"
	"zotregistry.io/zot/pkg/api/config"
	"zotregistry.io/zot/pkg/log"
)

const version = "0.1.0-dev"

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	fmt.Printf("Zot Artifact Store v%s\n", version)
	fmt.Println("Starting...")

	// Load configuration
	var cfg *config.Config
	if *configFile != "" {
		fmt.Printf("Loading configuration from: %s\n", *configFile)
		var err error
		cfg, err = loadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("No configuration file specified, using defaults")
		// Create default configuration
		cfg = config.New()
		cfg.HTTP.Address = "0.0.0.0"
		cfg.HTTP.Port = "8080"
		cfg.Storage.RootDirectory = "/tmp/zot-artifacts"
	}

	// Create logger with configured level
	logLevel := "info"
	if cfg.Log != nil && cfg.Log.Level != "" {
		logLevel = cfg.Log.Level
	}
	logger := log.NewLogger(logLevel, cfg.Log.Output)

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

	// Initialize storage
	logger.Info().Msg("Initializing storage")
	if err := ctlr.InitImageStore(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize storage")
		os.Exit(1)
	}

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

// loadConfig loads configuration from a YAML file
func loadConfig(configPath string) (*config.Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	cfg := config.New()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if cfg.Storage.RootDirectory == "" {
		cfg.Storage.RootDirectory = "/tmp/zot-artifacts"
	}

	if cfg.HTTP.Port == "" {
		cfg.HTTP.Port = "8080"
	}

	if cfg.HTTP.Address == "" {
		cfg.HTTP.Address = "0.0.0.0"
	}

	return cfg, nil
}
