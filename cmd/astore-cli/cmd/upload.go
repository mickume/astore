package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/spf13/cobra"
)

var (
	uploadContentType string
	uploadMetadata    []string
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload <local-file> <bucket/key>",
	Short: "Upload an artifact to the store",
	Long: `Upload a local file to the artifact store.

Examples:
  # Upload a file
  astore upload app.tar.gz releases/app-1.0.0.tar.gz

  # Upload with content type
  astore upload --content-type application/gzip app.tar.gz releases/app-1.0.0.tar.gz

  # Upload with metadata
  astore upload --metadata version=1.0.0 --metadata author=ci app.tar.gz releases/app-1.0.0.tar.gz`,
	Args: cobra.ExactArgs(2),
	RunE: runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVar(&uploadContentType, "content-type", "", "content type of the artifact")
	uploadCmd.Flags().StringArrayVarP(&uploadMetadata, "metadata", "m", []string{}, "metadata key=value pairs")
}

func runUpload(cmd *cobra.Command, args []string) error {
	localFile := args[0]
	remotePath := args[1]

	// Parse bucket and key
	bucket, key, err := getBucketAndKey(remotePath)
	if err != nil {
		return fmt.Errorf("invalid remote path: %w", err)
	}

	// Check if local file exists
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot upload directory: %s", localFile)
	}

	// Read file
	data, err := os.ReadFile(localFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse metadata
	metadata := make(map[string]string)
	for _, kv := range uploadMetadata {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key=value)", kv)
		}
		metadata[parts[0]] = parts[1]
	}

	// Determine content type
	contentType := uploadContentType
	if contentType == "" {
		contentType = guessContentType(localFile)
	}

	// Create client
	c, err := getClient()
	if err != nil {
		return err
	}

	// Prepare upload options
	opts := &client.UploadOptions{
		ContentType:      contentType,
		Metadata:         metadata,
		ProgressCallback: getProgressCallback(int64(len(data)), "Upload"),
	}

	// Upload
	ctx := context.Background()
	if verbose {
		fmt.Printf("Uploading %s to %s/%s (%s)\n", localFile, bucket, key, formatSize(int64(len(data))))
	}

	err = c.Upload(ctx, bucket, key, strings.NewReader(string(data)), int64(len(data)), opts)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if verbose {
		fmt.Println() // New line after progress
	}
	fmt.Printf("✓ Uploaded %s to %s/%s\n", localFile, bucket, key)

	return nil
}

func guessContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".tar":
		return "application/x-tar"
	case ".gz", ".gzip":
		return "application/gzip"
	case ".tgz":
		return "application/gzip"
	case ".zip":
		return "application/zip"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	default:
		return "application/octet-stream"
	}
}
