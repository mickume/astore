package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileSystemBackend implements Backend for local filesystem storage
type FileSystemBackend struct {
	rootDir        string
	enableChecksum bool
	mu             sync.RWMutex
}

// NewFileSystemBackend creates a new filesystem storage backend
func NewFileSystemBackend(rootDir string, enableChecksum bool) (*FileSystemBackend, error) {
	if rootDir == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "root directory cannot be empty",
		}
	}

	// Create root directory if it doesn't exist
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, &AppError{
			Code:    "INIT_ERROR",
			Message: "failed to create root directory",
			Err:     err,
		}
	}

	return &FileSystemBackend{
		rootDir:        rootDir,
		enableChecksum: enableChecksum,
	}, nil
}

// Name returns the backend name
func (fs *FileSystemBackend) Name() string {
	return "filesystem"
}

// WriteObject writes an object to the filesystem
func (fs *FileSystemBackend) WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	objectPath := fs.getObjectPath(bucket, key)

	// Create directory structure
	dir := filepath.Dir(objectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to create directory",
			Err:     err,
		}
	}

	// Create temporary file for atomic write
	tempFile := objectPath + ".tmp"
	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to create file",
			Err:     err,
		}
	}

	var written int64
	var hash string

	if fs.enableChecksum {
		// Calculate SHA256 while writing
		hasher := sha256.New()
		teeReader := io.TeeReader(reader, hasher)
		written, err = io.Copy(file, teeReader)
		if err == nil {
			hash = hex.EncodeToString(hasher.Sum(nil))
		}
	} else {
		written, err = io.Copy(file, reader)
	}

	file.Close()

	if err != nil {
		os.Remove(tempFile)
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to write object",
			Err:     err,
		}
	}

	// Atomic rename
	if err := os.Rename(tempFile, objectPath); err != nil {
		os.Remove(tempFile)
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to finalize object",
			Err:     err,
		}
	}

	// Store checksum if enabled
	if fs.enableChecksum && hash != "" {
		checksumPath := objectPath + ".sha256"
		if err := os.WriteFile(checksumPath, []byte(hash), 0644); err != nil {
			// Log warning but don't fail the operation
			fmt.Printf("warning: failed to write checksum file: %v\n", err)
		}
	}

	return written, nil
}

// ReadObject reads an object from the filesystem
func (fs *FileSystemBackend) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	objectPath := fs.getObjectPath(bucket, key)

	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to open object",
			Err:     err,
		}
	}

	// Verify checksum if enabled
	if fs.enableChecksum {
		if err := fs.verifyChecksum(objectPath); err != nil {
			file.Close()
			return nil, err
		}
	}

	return file, nil
}

// ReadObjectRange reads a byte range from an object
func (fs *FileSystemBackend) ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	objectPath := fs.getObjectPath(bucket, key)

	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to open object",
			Err:     err,
		}
	}

	// Seek to offset
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		file.Close()
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to seek to offset",
			Err:     err,
		}
	}

	// Wrap in a limited reader
	limitedReader := &limitedReadCloser{
		reader: io.LimitReader(file, length),
		closer: file,
	}

	return limitedReader, nil
}

// DeleteObject deletes an object from the filesystem
func (fs *FileSystemBackend) DeleteObject(ctx context.Context, bucket, key string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	objectPath := fs.getObjectPath(bucket, key)

	if err := os.Remove(objectPath); err != nil {
		if os.IsNotExist(err) {
			return &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to delete object",
			Err:     err,
		}
	}

	// Remove checksum file if it exists
	checksumPath := objectPath + ".sha256"
	os.Remove(checksumPath) // Ignore errors

	return nil
}

// ObjectExists checks if an object exists
func (fs *FileSystemBackend) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	objectPath := fs.getObjectPath(bucket, key)
	_, err := os.Stat(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
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

// CreateBucket creates a new bucket directory
func (fs *FileSystemBackend) CreateBucket(ctx context.Context, bucket string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	bucketPath := filepath.Join(fs.rootDir, bucket)

	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		return &AppError{
			Code:    "CREATE_ERROR",
			Message: "failed to create bucket",
			Err:     err,
		}
	}

	return nil
}

// DeleteBucket deletes a bucket directory (must be empty)
func (fs *FileSystemBackend) DeleteBucket(ctx context.Context, bucket string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	bucketPath := filepath.Join(fs.rootDir, bucket)

	// Check if bucket is empty
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppError{
				Code:    "NOT_FOUND",
				Message: "bucket not found",
				Err:     err,
			}
		}
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to read bucket",
			Err:     err,
		}
	}

	if len(entries) > 0 {
		return &AppError{
			Code:    "BUCKET_NOT_EMPTY",
			Message: "bucket is not empty",
		}
	}

	if err := os.Remove(bucketPath); err != nil {
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to delete bucket",
			Err:     err,
		}
	}

	return nil
}

// BucketExists checks if a bucket exists
func (fs *FileSystemBackend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	bucketPath := filepath.Join(fs.rootDir, bucket)
	info, err := os.Stat(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check bucket existence",
			Err:     err,
		}
	}

	return info.IsDir(), nil
}

// GetObjectSize returns the size of an object
func (fs *FileSystemBackend) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	objectPath := fs.getObjectPath(bucket, key)
	info, err := os.Stat(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
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

	return info.Size(), nil
}

// GetObjectHash returns the SHA256 hash of an object
func (fs *FileSystemBackend) GetObjectHash(ctx context.Context, bucket, key string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	objectPath := fs.getObjectPath(bucket, key)

	// Check if checksum file exists
	checksumPath := objectPath + ".sha256"
	if data, err := os.ReadFile(checksumPath); err == nil {
		return string(data), nil
	}

	// Calculate checksum if not stored
	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", &AppError{
				Code:    "NOT_FOUND",
				Message: "object not found",
				Err:     err,
			}
		}
		return "", &AppError{
			Code:    "READ_ERROR",
			Message: "failed to open object",
			Err:     err,
		}
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", &AppError{
			Code:    "HASH_ERROR",
			Message: "failed to calculate hash",
			Err:     err,
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HealthCheck performs a health check
func (fs *FileSystemBackend) HealthCheck(ctx context.Context) error {
	// Check if root directory is accessible
	if _, err := os.Stat(fs.rootDir); err != nil {
		return &AppError{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "root directory not accessible",
			Err:     err,
		}
	}

	// Try to create a test file
	testPath := filepath.Join(fs.rootDir, ".healthcheck")
	if err := os.WriteFile(testPath, []byte("ok"), 0644); err != nil {
		return &AppError{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "cannot write to root directory",
			Err:     err,
		}
	}

	// Clean up test file
	os.Remove(testPath)

	return nil
}

// Helper methods

func (fs *FileSystemBackend) getObjectPath(bucket, key string) string {
	return filepath.Join(fs.rootDir, bucket, key)
}

func (fs *FileSystemBackend) verifyChecksum(objectPath string) error {
	checksumPath := objectPath + ".sha256"
	storedHash, err := os.ReadFile(checksumPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Checksum file doesn't exist, skip verification
			return nil
		}
		return &AppError{
			Code:    "CHECKSUM_ERROR",
			Message: "failed to read checksum file",
			Err:     err,
		}
	}

	// Calculate current hash
	file, err := os.Open(objectPath)
	if err != nil {
		return &AppError{
			Code:    "READ_ERROR",
			Message: "failed to open object for verification",
			Err:     err,
		}
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return &AppError{
			Code:    "CHECKSUM_ERROR",
			Message: "failed to calculate hash",
			Err:     err,
		}
	}

	calculatedHash := hex.EncodeToString(hasher.Sum(nil))
	if calculatedHash != string(storedHash) {
		return &AppError{
			Code:    "CHECKSUM_MISMATCH",
			Message: fmt.Sprintf("checksum mismatch: expected %s, got %s", storedHash, calculatedHash),
		}
	}

	return nil
}

// limitedReadCloser wraps a limited reader with a closer
type limitedReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (lrc *limitedReadCloser) Read(p []byte) (n int, err error) {
	return lrc.reader.Read(p)
}

func (lrc *limitedReadCloser) Close() error {
	return lrc.closer.Close()
}
