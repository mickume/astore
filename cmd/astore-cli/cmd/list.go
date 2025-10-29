package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/spf13/cobra"
)

var (
	listPrefix   string
	listMaxKeys  int
	listBuckets  bool
	listLongForm bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [bucket]",
	Short: "List buckets or objects",
	Long: `List buckets or objects in the artifact store.

Examples:
  # List all buckets
  astore list --buckets

  # List objects in a bucket
  astore list releases

  # List with prefix filter
  astore list releases --prefix app/

  # List with details
  astore list releases --long`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listPrefix, "prefix", "", "filter objects by prefix")
	listCmd.Flags().IntVar(&listMaxKeys, "max-keys", 1000, "maximum number of keys to return")
	listCmd.Flags().BoolVar(&listBuckets, "buckets", false, "list buckets instead of objects")
	listCmd.Flags().BoolVarP(&listLongForm, "long", "l", false, "use long listing format")
}

func runList(cmd *cobra.Command, args []string) error {
	// Create client
	c, err := getClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// List buckets
	if listBuckets {
		result, err := c.ListBuckets(ctx)
		if err != nil {
			return fmt.Errorf("failed to list buckets: %w", err)
		}

		if len(result.Buckets) == 0 {
			fmt.Println("No buckets found")
			return nil
		}

		fmt.Printf("Found %d bucket(s):\n", len(result.Buckets))
		for _, bucket := range result.Buckets {
			if listLongForm {
				fmt.Printf("  %s (created: %s)\n", bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("  %s\n", bucket.Name)
			}
		}
		return nil
	}

	// List objects
	if len(args) == 0 {
		return fmt.Errorf("bucket name is required (or use --buckets to list buckets)")
	}

	bucket := args[0]

	opts := &client.ListOptions{
		Prefix:  listPrefix,
		MaxKeys: listMaxKeys,
	}

	result, err := c.ListObjects(ctx, bucket, opts)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	if len(result.Objects) == 0 {
		fmt.Println("No objects found")
		return nil
	}

	fmt.Printf("Found %d object(s) in bucket '%s':\n", len(result.Objects), bucket)

	// Calculate column widths for long format
	if listLongForm {
		var maxKeyLen int
		for _, obj := range result.Objects {
			if len(obj.Key) > maxKeyLen {
				maxKeyLen = len(obj.Key)
			}
		}

		// Print header
		fmt.Printf("  %-*s  %12s  %s\n", maxKeyLen, "KEY", "SIZE", "LAST MODIFIED")
		fmt.Println(strings.Repeat("-", maxKeyLen+12+25+6))

		// Print objects
		for _, obj := range result.Objects {
			fmt.Printf("  %-*s  %12s  %s\n",
				maxKeyLen,
				obj.Key,
				formatSize(obj.Size),
				obj.LastModified.Format("2006-01-02 15:04:05"))
		}

		// Print summary
		var totalSize int64
		for _, obj := range result.Objects {
			totalSize += obj.Size
		}
		fmt.Printf("\nTotal: %d objects, %s\n", len(result.Objects), formatSize(totalSize))
	} else {
		for _, obj := range result.Objects {
			fmt.Printf("  %s\n", obj.Key)
		}
	}

	return nil
}
