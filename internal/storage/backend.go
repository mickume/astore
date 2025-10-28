package storage

import (
	"context"
	"io"
)

// Backend defines the interface for storage backends
type Backend interface {
	// WriteObject writes an object to storage
	// Returns the number of bytes written and any error
	WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error)

	// ReadObject reads an object from storage
	// Returns a ReadCloser for streaming the object content
	ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// ReadObjectRange reads a byte range from an object
	// offset is the starting byte position, length is the number of bytes to read
	ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error)

	// DeleteObject deletes an object from storage
	DeleteObject(ctx context.Context, bucket, key string) error

	// ObjectExists checks if an object exists in storage
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)

	// CreateBucket creates a new bucket/container
	CreateBucket(ctx context.Context, bucket string) error

	// DeleteBucket deletes a bucket/container (must be empty)
	DeleteBucket(ctx context.Context, bucket string) error

	// BucketExists checks if a bucket exists
	BucketExists(ctx context.Context, bucket string) (bool, error)

	// GetObjectSize returns the size of an object in bytes
	GetObjectSize(ctx context.Context, bucket, key string) (int64, error)

	// GetObjectHash returns the SHA256 hash of an object
	GetObjectHash(ctx context.Context, bucket, key string) (string, error)

	// Name returns the backend name (filesystem, s3, gcs, azure)
	Name() string

	// HealthCheck performs a health check on the backend
	HealthCheck(ctx context.Context) error
}

// BackendConfig contains common configuration for storage backends
type BackendConfig struct {
	// Type is the backend type (filesystem, s3, gcs, azure)
	Type string

	// Filesystem configuration
	RootDirectory string

	// S3 configuration
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3UseSSL          bool

	// GCS configuration
	GCSBucket          string
	GCSCredentialsFile string
	GCSProjectID       string

	// Azure configuration
	AzureAccountName   string
	AzureAccountKey    string
	AzureContainerName string
	AzureEndpoint      string

	// Common options
	EnableChecksum bool // Enable SHA256 integrity verification
	MaxRetries     int  // Maximum number of retries for failed operations
	RetryDelay     int  // Delay between retries in milliseconds
}

// NewBackend creates a new storage backend based on configuration
func NewBackend(config *BackendConfig) (Backend, error) {
	switch config.Type {
	case "filesystem", "":
		return NewFileSystemBackend(config.RootDirectory, config.EnableChecksum)
	case "s3":
		return NewS3Backend(config)
	case "gcs":
		return NewGCSBackend(config)
	case "azure":
		return NewAzureBackend(config)
	default:
		return nil, &AppError{
			Code:    "INVALID_BACKEND",
			Message: "unsupported storage backend type: " + config.Type,
		}
	}
}

// AppError represents a storage backend error
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
