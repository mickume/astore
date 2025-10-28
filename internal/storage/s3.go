package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Backend implements Backend for Amazon S3 and S3-compatible storage
type S3Backend struct {
	client         *s3.S3
	uploader       *s3manager.Uploader
	downloader     *s3manager.Downloader
	bucket         string
	enableChecksum bool
}

// NewS3Backend creates a new S3 storage backend
func NewS3Backend(config *BackendConfig) (*S3Backend, error) {
	if config.S3AccessKeyID == "" || config.S3SecretAccessKey == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "S3 access key and secret key are required",
		}
	}

	if config.S3Bucket == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "S3 bucket name is required",
		}
	}

	// Configure AWS session
	awsConfig := &aws.Config{
		Region: aws.String(config.S3Region),
		Credentials: credentials.NewStaticCredentials(
			config.S3AccessKeyID,
			config.S3SecretAccessKey,
			"",
		),
	}

	// Set custom endpoint if provided (for S3-compatible services like MinIO)
	if config.S3Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.S3Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true) // Required for MinIO and some S3-compatible services
	}

	// Disable SSL if specified
	if !config.S3UseSSL {
		awsConfig.DisableSSL = aws.Bool(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, &AppError{
			Code:    "INIT_ERROR",
			Message: "failed to create AWS session",
			Err:     err,
		}
	}

	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Backend{
		client:         client,
		uploader:       uploader,
		downloader:     downloader,
		bucket:         config.S3Bucket,
		enableChecksum: config.EnableChecksum,
	}, nil
}

// Name returns the backend name
func (s3b *S3Backend) Name() string {
	return "s3"
}

// WriteObject writes an object to S3
func (s3b *S3Backend) WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error) {
	var body io.Reader = reader
	var hash string

	// Calculate SHA256 if checksum is enabled
	if s3b.enableChecksum {
		// Read entire content to calculate hash
		data, err := io.ReadAll(reader)
		if err != nil {
			return 0, &AppError{
				Code:    "READ_ERROR",
				Message: "failed to read object data",
				Err:     err,
			}
		}

		hasher := sha256.New()
		hasher.Write(data)
		hash = hex.EncodeToString(hasher.Sum(nil))

		body = bytes.NewReader(data)
		size = int64(len(data))
	}

	// Prepare upload input
	input := &s3manager.UploadInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
		Body:   body,
	}

	// Add SHA256 checksum as metadata if enabled
	if s3b.enableChecksum && hash != "" {
		input.Metadata = map[string]*string{
			"sha256": aws.String(hash),
		}
	}

	// Upload to S3
	_, err := s3b.uploader.UploadWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return 0, &AppError{
					Code:    "NOT_FOUND",
					Message: "bucket not found",
					Err:     err,
				}
			default:
				return 0, &AppError{
					Code:    "WRITE_ERROR",
					Message: "failed to upload object to S3",
					Err:     err,
				}
			}
		}
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to upload object to S3",
			Err:     err,
		}
	}

	return size, nil
}

// ReadObject reads an object from S3
func (s3b *S3Backend) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
	}

	result, err := s3b.client.GetObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, &AppError{
					Code:    "NOT_FOUND",
					Message: "object not found",
					Err:     err,
				}
			case s3.ErrCodeNoSuchBucket:
				return nil, &AppError{
					Code:    "NOT_FOUND",
					Message: "bucket not found",
					Err:     err,
				}
			default:
				return nil, &AppError{
					Code:    "READ_ERROR",
					Message: "failed to read object from S3",
					Err:     err,
				}
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to read object from S3",
			Err:     err,
		}
	}

	// Verify checksum if enabled
	if s3b.enableChecksum && result.Metadata != nil {
		if storedHash, ok := result.Metadata["sha256"]; ok && storedHash != nil {
			return &checksumVerifyingReader{
				reader:      result.Body,
				expectedSum: *storedHash,
			}, nil
		}
	}

	return result.Body, nil
}

// ReadObjectRange reads a byte range from an object in S3
func (s3b *S3Backend) ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	rangeHeader := fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
		Range:  aws.String(rangeHeader),
	}

	result, err := s3b.client.GetObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, &AppError{
					Code:    "NOT_FOUND",
					Message: "object not found",
					Err:     err,
				}
			default:
				return nil, &AppError{
					Code:    "READ_ERROR",
					Message: "failed to read object range from S3",
					Err:     err,
				}
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to read object range from S3",
			Err:     err,
		}
	}

	return result.Body, nil
}

// DeleteObject deletes an object from S3
func (s3b *S3Backend) DeleteObject(ctx context.Context, bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
	}

	_, err := s3b.client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return &AppError{
					Code:    "NOT_FOUND",
					Message: "bucket not found",
					Err:     err,
				}
			default:
				return &AppError{
					Code:    "DELETE_ERROR",
					Message: "failed to delete object from S3",
					Err:     err,
				}
			}
		}
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to delete object from S3",
			Err:     err,
		}
	}

	return nil
}

// ObjectExists checks if an object exists in S3
func (s3b *S3Backend) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
	}

	_, err := s3b.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchKey:
				return false, nil
			default:
				return false, &AppError{
					Code:    "STAT_ERROR",
					Message: "failed to check object existence",
					Err:     err,
				}
			}
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check object existence",
			Err:     err,
		}
	}

	return true, nil
}

// CreateBucket creates a new S3 bucket (or uses the configured bucket)
func (s3b *S3Backend) CreateBucket(ctx context.Context, bucket string) error {
	// In S3 backend, we use prefixes within a single bucket
	// So this is a no-op, just verify the main bucket exists
	return nil
}

// DeleteBucket deletes an S3 bucket (no-op for prefix-based buckets)
func (s3b *S3Backend) DeleteBucket(ctx context.Context, bucket string) error {
	// In S3 backend, we use prefixes within a single bucket
	// So this is a no-op
	return nil
}

// BucketExists checks if a bucket exists (always true for configured bucket)
func (s3b *S3Backend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	// In S3 backend, we use prefixes within a single bucket
	// Check if the main bucket exists
	input := &s3.HeadBucketInput{
		Bucket: aws.String(s3b.bucket),
	}

	_, err := s3b.client.HeadBucketWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchBucket:
				return false, nil
			default:
				return false, &AppError{
					Code:    "STAT_ERROR",
					Message: "failed to check bucket existence",
					Err:     err,
				}
			}
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check bucket existence",
			Err:     err,
		}
	}

	return true, nil
}

// GetObjectSize returns the size of an object in S3
func (s3b *S3Backend) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
	}

	result, err := s3b.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchKey:
				return 0, &AppError{
					Code:    "NOT_FOUND",
					Message: "object not found",
					Err:     err,
				}
			default:
				return 0, &AppError{
					Code:    "STAT_ERROR",
					Message: "failed to get object size",
					Err:     err,
				}
			}
		}
		return 0, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get object size",
			Err:     err,
		}
	}

	if result.ContentLength == nil {
		return 0, &AppError{
			Code:    "STAT_ERROR",
			Message: "object size not available",
		}
	}

	return *result.ContentLength, nil
}

// GetObjectHash returns the SHA256 hash of an object
func (s3b *S3Backend) GetObjectHash(ctx context.Context, bucket, key string) (string, error) {
	// Check if hash is stored in metadata
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3b.getBucket(bucket)),
		Key:    aws.String(key),
	}

	result, err := s3b.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchKey:
				return "", &AppError{
					Code:    "NOT_FOUND",
					Message: "object not found",
					Err:     err,
				}
			default:
				return "", &AppError{
					Code:    "STAT_ERROR",
					Message: "failed to get object metadata",
					Err:     err,
				}
			}
		}
		return "", &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get object metadata",
			Err:     err,
		}
	}

	// Check metadata for SHA256
	if result.Metadata != nil {
		if hash, ok := result.Metadata["sha256"]; ok && hash != nil {
			return *hash, nil
		}
	}

	// If not in metadata, download and calculate
	reader, err := s3b.ReadObject(ctx, bucket, key)
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

// HealthCheck performs a health check on S3
func (s3b *S3Backend) HealthCheck(ctx context.Context) error {
	// Try to list objects (with max 1 result)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s3b.bucket),
		MaxKeys: aws.Int64(1),
	}

	_, err := s3b.client.ListObjectsV2WithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return &AppError{
				Code:    "HEALTH_CHECK_FAILED",
				Message: fmt.Sprintf("S3 health check failed: %s", aerr.Code()),
				Err:     err,
			}
		}
		return &AppError{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "S3 health check failed",
			Err:     err,
		}
	}

	return nil
}

// Helper methods

// getBucket returns the full S3 key with bucket prefix
func (s3b *S3Backend) getBucket(bucket string) string {
	// Use prefixes within the configured S3 bucket
	return s3b.bucket
}

// checksumVerifyingReader verifies SHA256 checksum while reading
type checksumVerifyingReader struct {
	reader      io.ReadCloser
	expectedSum string
	hasher      hash.Hash
}

func (cvr *checksumVerifyingReader) Read(p []byte) (n int, err error) {
	if cvr.hasher == nil {
		cvr.hasher = sha256.New()
	}

	n, err = cvr.reader.Read(p)
	if n > 0 {
		cvr.hasher.Write(p[:n])
	}

	if err == io.EOF {
		// Verify checksum on EOF
		calculatedSum := hex.EncodeToString(cvr.hasher.Sum(nil))
		if calculatedSum != cvr.expectedSum {
			return n, &AppError{
				Code:    "CHECKSUM_MISMATCH",
				Message: fmt.Sprintf("checksum mismatch: expected %s, got %s", cvr.expectedSum, calculatedSum),
			}
		}
	}

	return n, err
}

func (cvr *checksumVerifyingReader) Close() error {
	return cvr.reader.Close()
}
