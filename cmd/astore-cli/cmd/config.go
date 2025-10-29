package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage astore CLI configuration`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a default configuration file in the home directory.

Examples:
  # Initialize config with interactive prompts
  astore config init

  # Initialize with server URL
  astore config init --server https://artifacts.example.com`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration values`,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".astore.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists: %s\n", configPath)
		fmt.Print("Overwrite? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			fmt.Println("Initialization cancelled")
			return nil
		}
	}

	// Get server URL
	serverURL := viper.GetString("server")
	if serverURL == "" {
		fmt.Print("Server URL: ")
		fmt.Scanln(&serverURL)
	}

	// Get token (optional)
	token := viper.GetString("token")
	if token == "" {
		fmt.Print("Authentication token (optional): ")
		fmt.Scanln(&token)
	}

	// Create config content
	config := fmt.Sprintf(`# Zot Artifact Store CLI Configuration

# Server URL (required)
server: %s

# Authentication token (optional)
%s

# Request timeout in seconds (default: 60)
timeout: 60

# Skip TLS certificate verification (default: false)
insecure: false
`, serverURL, func() string {
		if token != "" {
			return "token: " + token
		}
		return "# token: your-token-here"
	}())

	// Write config file
	err = os.WriteFile(configPath, []byte(config), 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created: %s\n", configPath)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("Current configuration:")
	fmt.Printf("  Server:   %s\n", viper.GetString("server"))
	fmt.Printf("  Token:    %s\n", func() string {
		token := viper.GetString("token")
		if token == "" {
			return "(not set)"
		}
		if len(token) > 10 {
			return token[:10] + "..."
		}
		return token
	}())
	fmt.Printf("  Timeout:  %d seconds\n", viper.GetInt("timeout"))
	fmt.Printf("  Insecure: %v\n", viper.GetBool("insecure"))

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("\nConfig file: %s\n", viper.ConfigFileUsed())
	}

	return nil
}
