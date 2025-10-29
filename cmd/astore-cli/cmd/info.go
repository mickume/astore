package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info <bucket/key>",
	Short: "Get information about an artifact",
	Long: `Get detailed information about an artifact including metadata.

Examples:
  # Get artifact information
  astore info releases/app-1.0.0.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	// Parse bucket and key
	bucket, key, err := getBucketAndKey(remotePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Create client
	c, err := getClient()
	if err != nil {
		return err
	}

	// Get object metadata
	ctx := context.Background()
	obj, err := c.GetObjectMetadata(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Display information
	fmt.Printf("Artifact: %s/%s\n", bucket, key)
	fmt.Printf("Size:         %s (%d bytes)\n", formatSize(obj.Size), obj.Size)
	fmt.Printf("Content-Type: %s\n", obj.ContentType)
	if obj.ETag != "" {
		fmt.Printf("ETag:         %s\n", obj.ETag)
	}
	if !obj.LastModified.IsZero() {
		fmt.Printf("Last Modified: %s\n", obj.LastModified.Format("2006-01-02 15:04:05 MST"))
	}

	// Display metadata
	if len(obj.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range obj.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}
