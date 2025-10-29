package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	deleteForce  bool
	deleteBucket bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <bucket/key>",
	Short: "Delete an artifact or bucket",
	Long: `Delete an artifact or bucket from the store.

Examples:
  # Delete an artifact
  astore delete releases/app-1.0.0.tar.gz

  # Delete without confirmation
  astore delete --force releases/app-1.0.0.tar.gz

  # Delete a bucket
  astore delete --bucket mybucket`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteBucket, "bucket", false, "delete bucket instead of object")
}

func runDelete(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Create client
	c, err := getClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	if deleteBucket {
		// Delete bucket
		bucket := path

		// Confirm deletion
		if !deleteForce {
			fmt.Printf("Delete bucket '%s'? (y/N): ", bucket)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		if verbose {
			fmt.Printf("Deleting bucket: %s\n", bucket)
		}

		err := c.DeleteBucket(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to delete bucket: %w", err)
		}

		fmt.Printf("✓ Deleted bucket: %s\n", bucket)
		return nil
	}

	// Delete object
	bucket, key, err := getBucketAndKey(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Confirm deletion
	if !deleteForce {
		fmt.Printf("Delete %s/%s? (y/N): ", bucket, key)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	if verbose {
		fmt.Printf("Deleting %s/%s\n", bucket, key)
	}

	err = c.DeleteObject(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	fmt.Printf("✓ Deleted %s/%s\n", bucket, key)
	return nil
}
