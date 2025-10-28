package storage_test

import (
	"os"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestMetadataStore(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "metadata-test-*.db")
	test.AssertNoError(t, err, "creating temp file")
	dbPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(dbPath)

	t.Run("Create and get bucket", func(t *testing.T) {
		// Given: A metadata store
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		bucket := &models.Bucket{
			Name:       "test-bucket",
			Versioning: false,
		}

		// When: Creating a bucket
		err = store.CreateBucket(bucket)
		test.AssertNoError(t, err, "creating bucket")

		// Then: Bucket can be retrieved
		retrieved, err := store.GetBucket("test-bucket")
		test.AssertNoError(t, err, "getting bucket")
		test.AssertEqual(t, "test-bucket", retrieved.Name, "bucket name")
		test.AssertFalse(t, retrieved.CreatedAt.IsZero(), "created timestamp set")
	})

	t.Run("List buckets", func(t *testing.T) {
		// Given: A metadata store with multiple buckets
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		// When: Listing buckets
		buckets, err := store.ListBuckets()
		test.AssertNoError(t, err, "listing buckets")

		// Then: Buckets are returned
		test.AssertTrue(t, len(buckets) > 0, "buckets exist")
	})

	t.Run("Store and retrieve artifact", func(t *testing.T) {
		// Given: A metadata store
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		artifact := &models.Artifact{
			Bucket:      "test-bucket",
			Key:         "test-object",
			Size:        1024,
			ContentType: "application/octet-stream",
			StoragePath: "/path/to/object",
		}

		// When: Storing an artifact
		err = store.StoreArtifact(artifact)
		test.AssertNoError(t, err, "storing artifact")

		// Then: Artifact can be retrieved
		retrieved, err := store.GetArtifact("test-bucket", "test-object")
		test.AssertNoError(t, err, "getting artifact")
		test.AssertEqual(t, "test-object", retrieved.Key, "artifact key")
		test.AssertEqual(t, int64(1024), retrieved.Size, "artifact size")
	})

	t.Run("List artifacts with prefix", func(t *testing.T) {
		// Given: A metadata store with artifacts
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		// Add more artifacts
		for i := 0; i < 5; i++ {
			artifact := &models.Artifact{
				Bucket:      "test-bucket",
				Key:         "prefix/object" + string(rune('0'+i)),
				Size:        int64(i * 100),
				ContentType: "application/octet-stream",
			}
			store.StoreArtifact(artifact)
		}

		// When: Listing artifacts
		artifacts, err := store.ListArtifacts("test-bucket", "", 10)
		test.AssertNoError(t, err, "listing artifacts")

		// Then: Artifacts are returned
		test.AssertTrue(t, len(artifacts) > 0, "artifacts exist")
	})

	t.Run("Delete artifact", func(t *testing.T) {
		// Given: A metadata store with an artifact
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		// When: Deleting the artifact
		err = store.DeleteArtifact("test-bucket", "test-object")
		test.AssertNoError(t, err, "deleting artifact")

		// Then: Artifact cannot be retrieved
		_, err = store.GetArtifact("test-bucket", "test-object")
		test.AssertError(t, err, "artifact should not exist")
	})

	t.Run("Multipart upload lifecycle", func(t *testing.T) {
		// Given: A metadata store
		store, err := storage.NewMetadataStore(dbPath)
		test.AssertNoError(t, err, "creating metadata store")
		defer store.Close()

		upload := &models.MultipartUpload{
			UploadID:    "test-upload-123",
			Bucket:      "test-bucket",
			Key:         "large-file",
			ContentType: "application/octet-stream",
			Parts:       []models.MultipartPart{},
		}

		// When: Creating a multipart upload
		err = store.CreateMultipartUpload(upload)
		test.AssertNoError(t, err, "creating multipart upload")

		// Then: Upload can be retrieved
		retrieved, err := store.GetMultipartUpload("test-upload-123")
		test.AssertNoError(t, err, "getting multipart upload")
		test.AssertEqual(t, "test-upload-123", retrieved.UploadID, "upload ID")
		test.AssertEqual(t, "large-file", retrieved.Key, "object key")

		// When: Deleting the upload
		err = store.DeleteMultipartUpload("test-upload-123")
		test.AssertNoError(t, err, "deleting multipart upload")

		// Then: Upload cannot be retrieved
		_, err = store.GetMultipartUpload("test-upload-123")
		test.AssertError(t, err, "upload should not exist")
	})
}
