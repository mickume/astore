package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool

	// Global flags
	serverURL string
	token     string
	timeout   int
	insecure  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "astore",
	Short: "Zot Artifact Store CLI",
	Long: `astore is a command-line interface for the Zot Artifact Store.

It provides commands to upload, download, list, and manage artifacts
in your Zot Artifact Store instance with support for authentication,
progress tracking, and supply chain security features.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.astore.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "artifact store server URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "authentication token")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 60, "request timeout in seconds")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".astore" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".astore")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("ASTORE")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getClient creates and returns a configured client
func getClient() (*client.Client, error) {
	serverURL := viper.GetString("server")
	if serverURL == "" {
		return nil, fmt.Errorf("server URL is required (use --server flag or set in config)")
	}

	token := viper.GetString("token")
	timeoutSec := viper.GetInt("timeout")
	insecure := viper.GetBool("insecure")

	config := &client.Config{
		BaseURL:            serverURL,
		Token:              token,
		Timeout:            time.Duration(timeoutSec) * time.Second,
		InsecureSkipVerify: insecure,
	}

	return client.NewClient(config)
}

// formatSize formats bytes as human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getProgressCallback returns a progress callback function
func getProgressCallback(totalSize int64, operation string) func(int64) {
	if !verbose {
		return nil
	}

	return func(bytesTransferred int64) {
		percentage := float64(bytesTransferred) / float64(totalSize) * 100
		fmt.Printf("\r%s progress: %.1f%% (%s / %s)",
			operation,
			percentage,
			formatSize(bytesTransferred),
			formatSize(totalSize))
	}
}

// getBucketAndKey extracts bucket and key from path
// Supports formats: bucket/key, /bucket/key, or separate args
func getBucketAndKey(path string) (bucket, key string, err error) {
	if path == "" {
		return "", "", fmt.Errorf("invalid path: path cannot be empty")
	}

	path = filepath.ToSlash(path)
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return "", "", fmt.Errorf("invalid path: must be in format bucket/key")
	}

	// Split on first /
	idx := findFirstSlash(path)
	if idx == -1 {
		return "", "", fmt.Errorf("invalid path: must be in format bucket/key")
	}

	bucket = path[:idx]
	key = path[idx+1:]

	if bucket == "" || key == "" {
		return "", "", fmt.Errorf("invalid path: bucket and key cannot be empty")
	}

	return bucket, key, nil
}

func findFirstSlash(s string) int {
	for i, c := range s {
		if c == '/' {
			return i
		}
	}
	return -1
}
