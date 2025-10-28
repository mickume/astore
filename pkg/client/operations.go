package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
)

// CreateBucket creates a new bucket
func (c *Client) CreateBucket(ctx context.Context, bucket string) error {
	urlPath := fmt.Sprintf("/s3/%s", bucket)

	resp, err := c.doRequest(ctx, "PUT", urlPath, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteBucket deletes a bucket
func (c *Client) DeleteBucket(ctx context.Context, bucket string) error {
	urlPath := fmt.Sprintf("/s3/%s", bucket)

	resp, err := c.doRequest(ctx, "DELETE", urlPath, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ListBuckets lists all buckets
func (c *Client) ListBuckets(ctx context.Context) (*ListBucketsResult, error) {
	resp, err := c.doRequest(ctx, "GET", "/s3", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListBucketsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &result, nil
}

// Upload uploads an artifact to the specified bucket and key
func (c *Client) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, opts *UploadOptions) error {
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Wrap reader with progress tracking if callback provided
	reader := data
	if opts.ProgressCallback != nil {
		reader = &progressReader{
			reader:   data,
			callback: opts.ProgressCallback,
		}
	}

	// Prepare headers
	headers := make(map[string]string)
	if opts.ContentType != "" {
		headers["Content-Type"] = opts.ContentType
	} else {
		headers["Content-Type"] = "application/octet-stream"
	}

	if size > 0 {
		headers["Content-Length"] = strconv.FormatInt(size, 10)
	}

	// Add custom metadata headers
	for key, value := range opts.Metadata {
		headers["X-Amz-Meta-"+key] = value
	}

	urlPath := fmt.Sprintf("/s3/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "PUT", urlPath, reader, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Download downloads an artifact from the specified bucket and key
func (c *Client) Download(ctx context.Context, bucket, key string, writer io.Writer, opts *DownloadOptions) error {
	if opts == nil {
		opts = &DownloadOptions{}
	}

	// Prepare headers
	headers := make(map[string]string)
	if opts.Range != "" {
		headers["Range"] = opts.Range
	}

	urlPath := fmt.Sprintf("/s3/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "GET", urlPath, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Wrap writer with progress tracking if callback provided
	w := writer
	if opts.ProgressCallback != nil {
		w = &progressWriter{
			writer:   writer,
			callback: opts.ProgressCallback,
		}
	}

	// Copy response body to writer
	if _, err := io.Copy(w, resp.Body); err != nil {
		return errors.NewInternal("failed to download artifact: " + err.Error())
	}

	return nil
}

// GetObjectMetadata retrieves metadata for an object without downloading it
func (c *Client) GetObjectMetadata(ctx context.Context, bucket, key string) (*Object, error) {
	urlPath := fmt.Sprintf("/s3/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "HEAD", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse metadata from headers
	obj := &Object{
		Key:         key,
		ContentType: resp.Header.Get("Content-Type"),
		ETag:        resp.Header.Get("ETag"),
		Metadata:    make(map[string]string),
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			obj.Size = size
		}
	}

	// Extract custom metadata
	for headerKey, values := range resp.Header {
		if len(headerKey) > 11 && headerKey[:11] == "X-Amz-Meta-" {
			metaKey := headerKey[11:]
			if len(values) > 0 {
				obj.Metadata[metaKey] = values[0]
			}
		}
	}

	return obj, nil
}

// DeleteObject deletes an artifact
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	urlPath := fmt.Sprintf("/s3/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "DELETE", urlPath, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ListObjects lists objects in a bucket
func (c *Client) ListObjects(ctx context.Context, bucket string, opts *ListOptions) (*ListObjectsResult, error) {
	if opts == nil {
		opts = &ListOptions{MaxKeys: 1000}
	}

	// Build query parameters
	params := url.Values{}
	if opts.Prefix != "" {
		params.Set("prefix", opts.Prefix)
	}
	if opts.MaxKeys > 0 {
		params.Set("max-keys", strconv.Itoa(opts.MaxKeys))
	}
	if opts.Delimiter != "" {
		params.Set("delimiter", opts.Delimiter)
	}

	urlPath := fmt.Sprintf("/s3/%s?%s", bucket, params.Encode())

	resp, err := c.doRequest(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListObjectsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &result, nil
}

// CopyObject copies an object from one location to another
func (c *Client) CopyObject(ctx context.Context, sourceBucket, sourceKey, destBucket, destKey string) error {
	headers := map[string]string{
		"X-Amz-Copy-Source": fmt.Sprintf("/%s/%s", sourceBucket, sourceKey),
	}

	urlPath := fmt.Sprintf("/s3/%s/%s", destBucket, destKey)

	resp, err := c.doRequest(ctx, "PUT", urlPath, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// MultipartUpload represents a multipart upload session
type MultipartUpload struct {
	Bucket   string
	Key      string
	UploadID string
	client   *Client
}

// InitiateMultipartUpload starts a multipart upload session
func (c *Client) InitiateMultipartUpload(ctx context.Context, bucket, key string, opts *UploadOptions) (*MultipartUpload, error) {
	if opts == nil {
		opts = &UploadOptions{}
	}

	headers := make(map[string]string)
	if opts.ContentType != "" {
		headers["Content-Type"] = opts.ContentType
	}

	// Add custom metadata headers
	for metaKey, value := range opts.Metadata {
		headers["X-Amz-Meta-"+metaKey] = value
	}

	urlPath := fmt.Sprintf("/s3/%s/%s?uploads", bucket, key)

	resp, err := c.doRequest(ctx, "POST", urlPath, nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		UploadID string `json:"uploadId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &MultipartUpload{
		Bucket:   bucket,
		Key:      key,
		UploadID: result.UploadID,
		client:   c,
	}, nil
}

// UploadPart uploads a part for a multipart upload
func (mu *MultipartUpload) UploadPart(ctx context.Context, partNumber int, data io.Reader, size int64) (string, error) {
	headers := make(map[string]string)
	if size > 0 {
		headers["Content-Length"] = strconv.FormatInt(size, 10)
	}

	urlPath := fmt.Sprintf("/s3/%s/%s?uploadId=%s&partNumber=%d",
		mu.Bucket, mu.Key, mu.UploadID, partNumber)

	resp, err := mu.client.doRequest(ctx, "PUT", urlPath, data, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Get ETag from response
	etag := resp.Header.Get("ETag")
	return etag, nil
}

// CompletedPart represents a completed part in a multipart upload
type CompletedPart struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
}

// Complete completes a multipart upload
func (mu *MultipartUpload) Complete(ctx context.Context, parts []CompletedPart) error {
	// Prepare request body
	body := struct {
		Parts []CompletedPart `json:"parts"`
	}{
		Parts: parts,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return errors.NewInternal("failed to marshal request: " + err.Error())
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	urlPath := fmt.Sprintf("/s3/%s/%s?uploadId=%s", mu.Bucket, mu.Key, mu.UploadID)

	resp, err := mu.client.doRequest(ctx, "POST", urlPath, bytes.NewReader(bodyBytes), headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Abort aborts a multipart upload
func (mu *MultipartUpload) Abort(ctx context.Context) error {
	urlPath := fmt.Sprintf("/s3/%s/%s?uploadId=%s", mu.Bucket, mu.Key, mu.UploadID)

	resp, err := mu.client.doRequest(ctx, "DELETE", urlPath, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
