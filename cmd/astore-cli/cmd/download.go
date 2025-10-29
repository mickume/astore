package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/spf13/cobra"
)

var (
	downloadOutput string
	downloadRange  string
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <bucket/key> [local-file]",
	Short: "Download an artifact from the store",
	Long: `Download an artifact from the store to a local file.

If local-file is not specified, the artifact key name will be used.

Examples:
  # Download to current directory
  astore download releases/app-1.0.0.tar.gz

  # Download to specific file
  astore download releases/app-1.0.0.tar.gz ./app.tar.gz

  # Download with output flag
  astore download --output ./app.tar.gz releases/app-1.0.0.tar.gz

  # Download specific byte range
  astore download --range bytes=0-1023 releases/app-1.0.0.tar.gz`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "output file path")
	downloadCmd.Flags().StringVar(&downloadRange, "range", "", "byte range to download (e.g., bytes=0-1023)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	// Parse bucket and key
	bucket, key, err := getBucketAndKey(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Determine output file
	var outputFile string
	if downloadOutput != "" {
		outputFile = downloadOutput
	} else if len(args) > 1 {
		outputFile = args[1]
	} else {
		outputFile = filepath.Base(key)
	}

	// Create client
	c, err := getClient()
	if err != nil {
		return err
	}

	// Get object metadata first for size
	ctx := context.Background()
	obj, err := c.GetObjectMetadata(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to get object metadata: %w", err)
	}

	if verbose {
		fmt.Printf("Downloading %s/%s to %s (%s)\n", bucket, key, outputFile, formatSize(obj.Size))
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Prepare download options
	opts := &client.DownloadOptions{
		Range:            downloadRange,
		ProgressCallback: getProgressCallback(obj.Size, "Download"),
	}

	// Download
	err = c.Download(ctx, bucket, key, file, opts)
	if err != nil {
		os.Remove(outputFile) // Clean up partial download
		return fmt.Errorf("download failed: %w", err)
	}

	if verbose {
		fmt.Println() // New line after progress
	}
	fmt.Printf("✓ Downloaded %s/%s to %s\n", bucket, key, outputFile)

	return nil
}
