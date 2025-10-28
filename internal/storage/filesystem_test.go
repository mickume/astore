package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestFileSystemBackend(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "fs-backend-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := storage.NewFileSystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("failed to create filesystem backend: %v", err)
	}

	ctx := context.Background()
	bucket := "test-bucket"
	key := "test-object.txt"
	content := []byte("test content")

	t.Run("Backend name", func(t *testing.T) {
		// When: Getting backend name
		name := backend.Name()

		// Then: Returns filesystem
		test.AssertEqual(t, "filesystem", name, "backend name")
	})

	t.Run("Create bucket", func(t *testing.T) {
		// When: Creating a bucket
		err := backend.CreateBucket(ctx, bucket)

		// Then: Succeeds without error
		test.AssertNoError(t, err, "create bucket")
	})

	t.Run("Bucket exists after creation", func(t *testing.T) {
		// Given: Bucket was created
		// When: Checking if bucket exists
		exists, err := backend.BucketExists(ctx, bucket)

		// Then: Returns true
		test.AssertNoError(t, err, "bucket exists check")
		test.AssertTrue(t, exists, "bucket should exist")
	})

	t.Run("Write object", func(t *testing.T) {
		// When: Writing an object
		reader := bytes.NewReader(content)
		written, err := backend.WriteObject(ctx, bucket, key, reader, int64(len(content)))

		// Then: Succeeds and returns correct size
		test.AssertNoError(t, err, "write object")
		test.AssertEqual(t, int64(len(content)), written, "bytes written")
	})

	t.Run("Object exists after write", func(t *testing.T) {
		// Given: Object was written
		// When: Checking if object exists
		exists, err := backend.ObjectExists(ctx, bucket, key)

		// Then: Returns true
		test.AssertNoError(t, err, "object exists check")
		test.AssertTrue(t, exists, "object should exist")
	})

	t.Run("Get object size", func(t *testing.T) {
		// Given: Object was written
		// When: Getting object size
		size, err := backend.GetObjectSize(ctx, bucket, key)

		// Then: Returns correct size
		test.AssertNoError(t, err, "get object size")
		test.AssertEqual(t, int64(len(content)), size, "object size")
	})

	t.Run("Get object hash", func(t *testing.T) {
		// Given: Object was written with checksum enabled
		// When: Getting object hash
		hash, err := backend.GetObjectHash(ctx, bucket, key)

		// Then: Returns valid SHA256 hash
		test.AssertNoError(t, err, "get object hash")
		test.AssertTrue(t, len(hash) == 64, "hash should be SHA256 (64 chars)")
	})

	t.Run("Read object", func(t *testing.T) {
		// Given: Object was written
		// When: Reading the object
		reader, err := backend.ReadObject(ctx, bucket, key)
		test.AssertNoError(t, err, "read object")
		defer reader.Close()

		// Then: Content matches what was written
		data, err := io.ReadAll(reader)
		test.AssertNoError(t, err, "read object content")
		test.AssertEqual(t, string(content), string(data), "object content")
	})

	t.Run("Read object range", func(t *testing.T) {
		// Given: Object was written
		// When: Reading a range of the object
		offset := int64(5)
		length := int64(4)
		reader, err := backend.ReadObjectRange(ctx, bucket, key, offset, length)
		test.AssertNoError(t, err, "read object range")
		defer reader.Close()

		// Then: Content matches the requested range
		data, err := io.ReadAll(reader)
		test.AssertNoError(t, err, "read range content")
		expected := content[offset : offset+length]
		test.AssertEqual(t, string(expected), string(data), "range content")
	})

	t.Run("Delete object", func(t *testing.T) {
		// Given: Object exists
		// When: Deleting the object
		err := backend.DeleteObject(ctx, bucket, key)

		// Then: Succeeds without error
		test.AssertNoError(t, err, "delete object")
	})

	t.Run("Object does not exist after deletion", func(t *testing.T) {
		// Given: Object was deleted
		// When: Checking if object exists
		exists, err := backend.ObjectExists(ctx, bucket, key)

		// Then: Returns false
		test.AssertNoError(t, err, "object exists check after deletion")
		test.AssertFalse(t, exists, "object should not exist")
	})

	t.Run("Delete bucket", func(t *testing.T) {
		// Given: Bucket is empty
		// When: Deleting the bucket
		err := backend.DeleteBucket(ctx, bucket)

		// Then: Succeeds without error
		test.AssertNoError(t, err, "delete bucket")
	})

	t.Run("Health check", func(t *testing.T) {
		// When: Running health check
		err := backend.HealthCheck(ctx)

		// Then: Succeeds without error
		test.AssertNoError(t, err, "health check")
	})
}

func TestFileSystemBackendErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs-backend-error-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := storage.NewFileSystemBackend(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create filesystem backend: %v", err)
	}

	ctx := context.Background()

	t.Run("Read non-existent object", func(t *testing.T) {
		// When: Reading an object that doesn't exist
		_, err := backend.ReadObject(ctx, "bucket", "nonexistent")

		// Then: Returns error
		test.AssertError(t, err, "should return error for non-existent object")
	})

	t.Run("Delete non-existent object", func(t *testing.T) {
		// When: Deleting an object that doesn't exist
		err := backend.DeleteObject(ctx, "bucket", "nonexistent")

		// Then: Returns error
		test.AssertError(t, err, "should return error for non-existent object")
	})

	t.Run("Object does not exist", func(t *testing.T) {
		// When: Checking if non-existent object exists
		exists, err := backend.ObjectExists(ctx, "bucket", "nonexistent")

		// Then: Returns false without error
		test.AssertNoError(t, err, "exists check should not error")
		test.AssertFalse(t, exists, "object should not exist")
	})
}
