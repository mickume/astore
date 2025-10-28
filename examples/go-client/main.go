package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
)

func main() {
	// Create client
	c, err := client.NewClient(&client.Config{
		BaseURL: getEnv("ARTIFACT_STORE_URL", "http://localhost:8080"),
		Token:   os.Getenv("ARTIFACT_STORE_TOKEN"),
		Timeout: 60 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example: Create bucket
	fmt.Println("Creating bucket 'examples'...")
	if err := c.CreateBucket(ctx, "examples"); err != nil {
		log.Printf("Note: %v (bucket may already exist)\n", err)
	}

	// Example: Upload artifact
	artifactData := []byte("This is example artifact content")
	fmt.Println("\nUploading artifact...")

	err = c.Upload(ctx, "examples", "example-artifact.txt",
		bytes.NewReader(artifactData),
		int64(len(artifactData)),
		&client.UploadOptions{
			ContentType: "text/plain",
			Metadata: map[string]string{
				"description": "Example artifact",
				"version":     "1.0.0",
			},
			ProgressCallback: func(bytesTransferred int64) {
				pct := float64(bytesTransferred) / float64(len(artifactData)) * 100
				fmt.Printf("\rUpload progress: %.1f%%", pct)
			},
		},
	)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	fmt.Println("\nUpload complete!")

	// Example: Get object metadata
	fmt.Println("\nGetting object metadata...")
	obj, err := c.GetObjectMetadata(ctx, "examples", "example-artifact.txt")
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}

	fmt.Printf("  Size: %d bytes\n", obj.Size)
	fmt.Printf("  Content-Type: %s\n", obj.ContentType)
	fmt.Printf("  ETag: %s\n", obj.ETag)
	for key, value := range obj.Metadata {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Example: List objects
	fmt.Println("\nListing objects in 'examples' bucket...")
	result, err := c.ListObjects(ctx, "examples", nil)
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}

	for _, obj := range result.Objects {
		fmt.Printf("  - %s (%d bytes)\n", obj.Key, obj.Size)
	}

	// Example: Download artifact
	fmt.Println("\nDownloading artifact...")
	var downloadBuffer bytes.Buffer
	err = c.Download(ctx, "examples", "example-artifact.txt", &downloadBuffer, &client.DownloadOptions{
		ProgressCallback: func(bytesTransferred int64) {
			fmt.Printf("\rDownload progress: %d bytes", bytesTransferred)
		},
	})
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Printf("\nDownloaded content: %s\n", downloadBuffer.String())

	// Example: Add attestation
	fmt.Println("\nAdding build attestation...")
	attestData := map[string]interface{}{
		"buildId":     "example-001",
		"status":      "success",
		"testsPassed": 10,
		"testsFailed": 0,
		"duration":    "1m30s",
	}

	att, err := c.AddAttestation(ctx, "examples", "example-artifact.txt", "build", attestData)
	if err != nil {
		log.Fatalf("Failed to add attestation: %v", err)
	}
	fmt.Printf("Attestation added: %s\n", att.ID)

	// Example: Get attestations
	fmt.Println("\nGetting attestations...")
	attestations, err := c.GetAttestations(ctx, "examples", "example-artifact.txt")
	if err != nil {
		log.Fatalf("Failed to get attestations: %v", err)
	}

	for _, att := range attestations {
		fmt.Printf("  Type: %s, ID: %s\n", att.Type, att.ID)
		fmt.Printf("  Data: %v\n", att.Data)
	}

	fmt.Println("\n✅ Example completed successfully!")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
