package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSBackend implements Backend for Google Cloud Storage
type GCSBackend struct {
	client         *storage.Client
	bucketName     string
	enableChecksum bool
}

// NewGCSBackend creates a new Google Cloud Storage backend
func NewGCSBackend(config *BackendConfig) (*GCSBackend, error) {
	if config.GCSBucket == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "GCS bucket name is required",
		}
	}

	var opts []option.ClientOption

	// Add credentials file if provided
	if config.GCSCredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(config.GCSCredentialsFile))
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, &AppError{
			Code:    "INIT_ERROR",
			Message: "failed to create GCS client",
			Err:     err,
		}
	}

	return &GCSBackend{
		client:         client,
		bucketName:     config.GCSBucket,
		enableChecksum: config.EnableChecksum,
	}, nil
}

// Name returns the backend name
func (gcs *GCSBackend) Name() string {
	return "gcs"
}

// WriteObject writes an object to GCS
func (gcs *GCSBackend) WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	writer := obj.NewWriter(ctx)

	var written int64
	var hash string
	var err error

	if gcs.enableChecksum {
		// Calculate SHA256 while writing
		hasher := sha256.New()
		teeReader := io.TeeReader(reader, hasher)

		written, err = io.Copy(writer, teeReader)
		if err != nil {
			writer.Close()
			return 0, &AppError{
				Code:    "WRITE_ERROR",
				Message: "failed to write object to GCS",
				Err:     err,
			}
		}

		hash = hex.EncodeToString(hasher.Sum(nil))

		// Set metadata with SHA256
		writer.Metadata = map[string]string{
			"sha256": hash,
		}
	} else {
		written, err = io.Copy(writer, reader)
		if err != nil {
			writer.Close()
			return 0, &AppError{
				Code:    "WRITE_ERROR",
				Message: "failed to write object to GCS",
				Err:     err,
			}
		}
	}

	if err := writer.Close(); err != nil {
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to finalize object write",
			Err:     err,
		}
	}

	return written, nil
}

// ReadObject reads an object from GCS
func (gcs *GCSBackend) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to read object from GCS",
			Err:     err,
		}
	}

	// Verify checksum if enabled
	if gcs.enableChecksum {
		attrs, err := obj.Attrs(ctx)
		if err == nil && attrs.Metadata != nil {
			if expectedHash, ok := attrs.Metadata["sha256"]; ok {
				return &checksumVerifyingReader{
					reader:      reader,
					expectedSum: expectedHash,
				}, nil
			}
		}
	}

	return reader, nil
}

// ReadObjectRange reads a byte range from an object in GCS
func (gcs *GCSBackend) ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	reader, err := obj.NewRangeReader(ctx, offset, length)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to read object range from GCS",
			Err:     err,
		}
	}

	return reader, nil
}

// DeleteObject deletes an object from GCS
func (gcs *GCSBackend) DeleteObject(ctx context.Context, bucket, key string) error {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to delete object from GCS",
			Err:     err,
		}
	}

	return nil
}

// ObjectExists checks if an object exists in GCS
func (gcs *GCSBackend) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check object existence",
			Err:     err,
		}
	}

	return true, nil
}

// CreateBucket creates a new GCS bucket (no-op for prefix-based buckets)
func (gcs *GCSBackend) CreateBucket(ctx context.Context, bucket string) error {
	// In GCS backend, we use prefixes within a single bucket
	// So this is a no-op
	return nil
}

// DeleteBucket deletes a GCS bucket (no-op for prefix-based buckets)
func (gcs *GCSBackend) DeleteBucket(ctx context.Context, bucket string) error {
	// In GCS backend, we use prefixes within a single bucket
	// So this is a no-op
	return nil
}

// BucketExists checks if a bucket exists
func (gcs *GCSBackend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	// Check if the main GCS bucket exists
	gcsBucket := gcs.client.Bucket(gcs.bucketName)
	_, err := gcsBucket.Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return false, nil
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check bucket existence",
			Err:     err,
		}
	}

	return true, nil
}

// GetObjectSize returns the size of an object in GCS
func (gcs *GCSBackend) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return 0, &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return 0, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get object size",
			Err:     err,
		}
	}

	return attrs.Size, nil
}

// GetObjectHash returns the SHA256 hash of an object
func (gcs *GCSBackend) GetObjectHash(ctx context.Context, bucket, key string) (string, error) {
	objectKey := gcs.getObjectKey(bucket, key)
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectKey)

	// Check if hash is stored in metadata
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return "", &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return "", &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get object metadata",
			Err:     err,
		}
	}

	// Check metadata for SHA256
	if attrs.Metadata != nil {
		if hash, ok := attrs.Metadata["sha256"]; ok {
			return hash, nil
		}
	}

	// If not in metadata, download and calculate
	reader, err := gcs.ReadObject(ctx, bucket, key)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", &AppError{
			Code:    "HASH_ERROR",
			Message: "failed to calculate hash",
			Err:     err,
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HealthCheck performs a health check on GCS
func (gcs *GCSBackend) HealthCheck(ctx context.Context) error {
	// Try to list objects (with max 1 result)
	gcsBucket := gcs.client.Bucket(gcs.bucketName)
	query := &storage.Query{
		Prefix:    "",
		Delimiter: "",
	}

	it := gcsBucket.Objects(ctx, query)
	_, err := it.Next()
	if err != nil && err != iterator.Done {
		return &AppError{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "GCS health check failed",
			Err:     err,
		}
	}

	return nil
}

// Helper methods

// getObjectKey returns the full GCS object key with bucket prefix
func (gcs *GCSBackend) getObjectKey(bucket, key string) string {
	// Use prefixes to organize objects within the GCS bucket
	return fmt.Sprintf("%s/%s", bucket, key)
}
