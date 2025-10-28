package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

// AzureBackend implements Backend for Azure Blob Storage
type AzureBackend struct {
	client         *azblob.Client
	containerName  string
	enableChecksum bool
}

// NewAzureBackend creates a new Azure Blob Storage backend
func NewAzureBackend(config *BackendConfig) (*AzureBackend, error) {
	if config.AzureAccountName == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "Azure account name is required",
		}
	}

	if config.AzureContainerName == "" {
		return nil, &AppError{
			Code:    "INVALID_CONFIG",
			Message: "Azure container name is required",
		}
	}

	// Build service URL
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", config.AzureAccountName)
	if config.AzureEndpoint != "" {
		serviceURL = config.AzureEndpoint
	}

	// Create credential
	cred, err := azblob.NewSharedKeyCredential(config.AzureAccountName, config.AzureAccountKey)
	if err != nil {
		return nil, &AppError{
			Code:    "INIT_ERROR",
			Message: "failed to create Azure credentials",
			Err:     err,
		}
	}

	// Create client
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, &AppError{
			Code:    "INIT_ERROR",
			Message: "failed to create Azure client",
			Err:     err,
		}
	}

	return &AzureBackend{
		client:         client,
		containerName:  config.AzureContainerName,
		enableChecksum: config.EnableChecksum,
	}, nil
}

// Name returns the backend name
func (az *AzureBackend) Name() string {
	return "azure"
}

// WriteObject writes an object to Azure Blob Storage
func (az *AzureBackend) WriteObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) (int64, error) {
	blobName := az.getBlobName(bucket, key)

	var data []byte
	var hash string
	var err error

	if az.enableChecksum {
		// Read data and calculate hash
		data, err = io.ReadAll(reader)
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

		reader = bytes.NewReader(data)
		size = int64(len(data))
	}

	// Upload blob
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	uploadOptions := &azblob.UploadStreamOptions{}

	// Add SHA256 as metadata if enabled
	if az.enableChecksum && hash != "" {
		uploadOptions.Metadata = map[string]*string{
			"sha256": &hash,
		}
	}

	_, err = blobClient.UploadStream(ctx, reader, uploadOptions)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.ContainerNotFound) {
				return 0, &AppError{
					Code:    "NOT_FOUND",
					Message: "container not found",
					Err:     err,
				}
			}
		}
		return 0, &AppError{
			Code:    "WRITE_ERROR",
			Message: "failed to upload blob to Azure",
			Err:     err,
		}
	}

	return size, nil
}

// ReadObject reads an object from Azure Blob Storage
func (az *AzureBackend) ReadObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	downloadResponse, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return nil, &AppError{
					Code:    "NOT_FOUND",
					Message: "blob not found",
					Err:     err,
				}
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to download blob from Azure",
			Err:     err,
		}
	}

	body := downloadResponse.NewRetryReader(ctx, &azblob.RetryReaderOptions{})

	// Verify checksum if enabled
	if az.enableChecksum && downloadResponse.Metadata != nil {
		if expectedHash, ok := downloadResponse.Metadata["sha256"]; ok && expectedHash != nil {
			return &checksumVerifyingReader{
				reader:      body,
				expectedSum: *expectedHash,
			}, nil
		}
	}

	return body, nil
}

// ReadObjectRange reads a byte range from an object in Azure
func (az *AzureBackend) ReadObjectRange(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	downloadOptions := &blob.DownloadStreamOptions{
		Range: blob.HTTPRange{
			Offset: offset,
			Count:  length,
		},
	}

	downloadResponse, err := blobClient.DownloadStream(ctx, downloadOptions)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return nil, &AppError{
					Code:    "NOT_FOUND",
					Message: "blob not found",
					Err:     err,
				}
			}
		}
		return nil, &AppError{
			Code:    "READ_ERROR",
			Message: "failed to download blob range from Azure",
			Err:     err,
		}
	}

	return downloadResponse.NewRetryReader(ctx, &azblob.RetryReaderOptions{}), nil
}

// DeleteObject deletes an object from Azure Blob Storage
func (az *AzureBackend) DeleteObject(ctx context.Context, bucket, key string) error {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	_, err := blobClient.Delete(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return &AppError{
					Code:    "NOT_FOUND",
					Message: "blob not found",
					Err:     err,
				}
			}
		}
		return &AppError{
			Code:    "DELETE_ERROR",
			Message: "failed to delete blob from Azure",
			Err:     err,
		}
	}

	return nil
}

// ObjectExists checks if an object exists in Azure
func (az *AzureBackend) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return false, nil
			}
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check blob existence",
			Err:     err,
		}
	}

	return true, nil
}

// CreateBucket creates a new Azure container (no-op for prefix-based buckets)
func (az *AzureBackend) CreateBucket(ctx context.Context, bucket string) error {
	// In Azure backend, we use prefixes within a single container
	// So this is a no-op
	return nil
}

// DeleteBucket deletes an Azure container (no-op for prefix-based buckets)
func (az *AzureBackend) DeleteBucket(ctx context.Context, bucket string) error {
	// In Azure backend, we use prefixes within a single container
	// So this is a no-op
	return nil
}

// BucketExists checks if a container exists
func (az *AzureBackend) BucketExists(ctx context.Context, bucket string) (bool, error) {
	// Check if the main Azure container exists
	containerClient := az.client.ServiceClient().NewContainerClient(az.containerName)
	_, err := containerClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.ContainerNotFound) {
				return false, nil
			}
		}
		return false, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to check container existence",
			Err:     err,
		}
	}

	return true, nil
}

// GetObjectSize returns the size of an object in Azure
func (az *AzureBackend) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return 0, &AppError{
					Code:    "NOT_FOUND",
					Message: "blob not found",
					Err:     err,
				}
			}
		}
		return 0, &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get blob properties",
			Err:     err,
		}
	}

	if props.ContentLength == nil {
		return 0, &AppError{
			Code:    "STAT_ERROR",
			Message: "blob size not available",
		}
	}

	return *props.ContentLength, nil
}

// GetObjectHash returns the SHA256 hash of an object
func (az *AzureBackend) GetObjectHash(ctx context.Context, bucket, key string) (string, error) {
	blobName := az.getBlobName(bucket, key)
	blobClient := az.client.ServiceClient().NewContainerClient(az.containerName).NewBlockBlobClient(blobName)

	// Check if hash is stored in metadata
	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if bloberror.HasCode(err, bloberror.BlobNotFound) {
				return "", &AppError{
					Code:    "NOT_FOUND",
					Message: "blob not found",
					Err:     err,
				}
			}
		}
		return "", &AppError{
			Code:    "STAT_ERROR",
			Message: "failed to get blob metadata",
			Err:     err,
		}
	}

	// Check metadata for SHA256
	if props.Metadata != nil {
		if hash, ok := props.Metadata["sha256"]; ok && hash != nil {
			return *hash, nil
		}
	}

	// If not in metadata, download and calculate
	reader, err := az.ReadObject(ctx, bucket, key)
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

// HealthCheck performs a health check on Azure
func (az *AzureBackend) HealthCheck(ctx context.Context) error {
	// Try to get container properties
	containerClient := az.client.ServiceClient().NewContainerClient(az.containerName)
	_, err := containerClient.GetProperties(ctx, nil)
	if err != nil {
		return &AppError{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "Azure health check failed",
			Err:     err,
		}
	}

	return nil
}

// Helper methods

// getBlobName returns the full blob name with bucket prefix
func (az *AzureBackend) getBlobName(bucket, key string) string {
	// Use prefixes to organize blobs within the Azure container
	return fmt.Sprintf("%s/%s", bucket, key)
}
