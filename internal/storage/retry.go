package storage

import (
	"bytes"
	"context"
	"io"
	"time"

	reliabilityErrors "github.com/candlekeep/zot-artifact-store/internal/errors"
	"github.com/candlekeep/zot-artifact-store/internal/reliability"
	"zotregistry.io/zot/pkg/log"
)

// RetryBackend wraps a Backend with retry capabilities
type RetryBackend struct {
	backend Backend
	retryer *reliability.Retryer
	logger  log.Logger
}

// NewRetryBackend creates a new backend with retry capabilities
func NewRetryBackend(backend Backend, config *BackendConfig, logger log.Logger) *RetryBackend {
	// Create retry policy based on configuration
	policy := &reliability.RetryPolicy{
		MaxAttempts:     config.MaxRetries,
		InitialDelay:    time.Duration(config.RetryDelay) * time.Millisecond,
		MaxDelay:        30 * time.Second,
		Multiplier:      2.0,
		RandomizeFactor: 0.2,
	}

	// Use default policy if not configured
	if config.MaxRetries == 0 {
		policy = reliability.DefaultRetryPolicy()
	}

	return &RetryBackend{
		backend: backend,
		retryer: reliability.NewRetryer(policy, logger),
		logger:  logger,
	}
}

// Name returns the backend name
func (rb *RetryBackend) Name() string {
	return rb.backend.Name() + "-with-retry"
}

// WriteObject writes an object with retry
func (rb *RetryBackend) WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error) {
	// For writes, we need to be able to re-read the data on retry
	// So we buffer it in memory for retries
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to read object data for retry",
			Err:     err,
		}
	}

	var written int64
	err = rb.retryer.Do(ctx, func(ctx context.Context) error {
		reader := bytes.NewReader(data)
		n, writeErr := rb.backend.WriteObject(ctx, bucket, key, reader, int64(len(data)))
		written = n
		return rb.wrapError(writeErr)
	})

	return written, err
}

// ReadObject reads an object with retry
func (rb *RetryBackend) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	var reader io.ReadCloser
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		r, readErr := rb.backend.ReadObject(ctx, bucket, key)
		reader = r
		return rb.wrapError(readErr)
	})

	return reader, err
}

// ReadObjectRange reads a byte range with retry
func (rb *RetryBackend) ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	var reader io.ReadCloser
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		r, readErr := rb.backend.ReadObjectRange(ctx, bucket, key, offset, length)
		reader = r
		return rb.wrapError(readErr)
	})

	return reader, err
}

// DeleteObject deletes an object with retry
func (rb *RetryBackend) DeleteObject(ctx context.Context, bucket, key string) error {
	return rb.retryer.Do(ctx, func(ctx context.Context) error {
		err := rb.backend.DeleteObject(ctx, bucket, key)
		return rb.wrapError(err)
	})
}

// ObjectExists checks if an object exists with retry
func (rb *RetryBackend) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	var exists bool
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		e, existsErr := rb.backend.ObjectExists(ctx, bucket, key)
		exists = e
		return rb.wrapError(existsErr)
	})

	return exists, err
}

// CreateBucket creates a bucket with retry
func (rb *RetryBackend) CreateBucket(ctx context.Context, bucket string) error {
	return rb.retryer.Do(ctx, func(ctx context.Context) error {
		err := rb.backend.CreateBucket(ctx, bucket)
		return rb.wrapError(err)
	})
}

// DeleteBucket deletes a bucket with retry
func (rb *RetryBackend) DeleteBucket(ctx context.Context, bucket string) error {
	return rb.retryer.Do(ctx, func(ctx context.Context) error {
		err := rb.backend.DeleteBucket(ctx, bucket)
		return rb.wrapError(err)
	})
}

// BucketExists checks if a bucket exists with retry
func (rb *RetryBackend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	var exists bool
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		e, existsErr := rb.backend.BucketExists(ctx, bucket)
		exists = e
		return rb.wrapError(existsErr)
	})

	return exists, err
}

// GetObjectSize gets object size with retry
func (rb *RetryBackend) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	var size int64
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		s, sizeErr := rb.backend.GetObjectSize(ctx, bucket, key)
		size = s
		return rb.wrapError(sizeErr)
	})

	return size, err
}

// GetObjectHash gets object hash with retry
func (rb *RetryBackend) GetObjectHash(ctx context.Context, bucket, key string) (string, error) {
	var hash string
	err := rb.retryer.Do(ctx, func(ctx context.Context) error {
		h, hashErr := rb.backend.GetObjectHash(ctx, bucket, key)
		hash = h
		return rb.wrapError(hashErr)
	})

	return hash, err
}

// HealthCheck performs health check with retry
func (rb *RetryBackend) HealthCheck(ctx context.Context) error {
	return rb.retryer.Do(ctx, func(ctx context.Context) error {
		err := rb.backend.HealthCheck(ctx)
		return rb.wrapError(err)
	})
}

// wrapError wraps storage errors to make them retryable based on error type
func (rb *RetryBackend) wrapError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if appErr, ok := err.(*AppError); ok {
		// Determine if error is retryable based on error code
		switch appErr.Code {
		case "NOT_FOUND", "BUCKET_NOT_EMPTY", "CHECKSUM_MISMATCH", "INVALID_CONFIG":
			// These are not retryable
			return reliabilityErrors.New(reliabilityErrors.ErrorCodeBadRequest, appErr.Message).WithDetail("original_code", appErr.Code)
		case "READ_ERROR", "WRITE_ERROR", "DELETE_ERROR", "STAT_ERROR", "HASH_ERROR", "HEALTH_CHECK_FAILED":
			// These are retryable (could be transient network issues)
			return reliabilityErrors.New(reliabilityErrors.ErrorCodeServiceUnavailable, appErr.Message).WithDetail("original_code", appErr.Code)
		default:
			// Default to retryable for unknown errors
			return reliabilityErrors.New(reliabilityErrors.ErrorCodeInternal, appErr.Message).WithDetail("original_code", appErr.Code)
		}
	}

	// Unknown error type, wrap as internal error (retryable)
	return reliabilityErrors.NewInternal(err.Error())
}
