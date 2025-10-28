package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
)

// Client is the Zot Artifact Store Go SDK client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	userAgent  string
}

// Config contains configuration options for the client
type Config struct {
	// BaseURL is the artifact store endpoint (e.g., "https://artifacts.example.com")
	BaseURL string

	// Token is the bearer authentication token (optional)
	Token string

	// HTTPClient allows providing a custom HTTP client (optional)
	HTTPClient *http.Client

	// Timeout specifies the request timeout (default: 30s)
	Timeout time.Duration

	// InsecureSkipVerify skips TLS certificate verification (for testing only)
	InsecureSkipVerify bool

	// UserAgent sets the User-Agent header (optional)
	UserAgent string
}

// NewClient creates a new Zot Artifact Store client
func NewClient(config *Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, errors.NewBadRequest("baseURL is required")
	}

	// Parse and validate base URL
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, errors.NewBadRequest("invalid baseURL: " + err.Error())
	}

	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.NewBadRequest("baseURL must use http or https scheme")
	}

	// Create HTTP client if not provided
	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}

		if config.InsecureSkipVerify {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		httpClient = &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "zot-artifact-store-go-client/1.0"
	}

	return &Client{
		baseURL:    baseURL.String(),
		httpClient: httpClient,
		token:      config.Token,
		userAgent:  userAgent,
	}, nil
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// doRequest performs an HTTP request with authentication and error handling
func (c *Client) doRequest(ctx context.Context, method, urlPath string, body io.Reader, headers map[string]string) (*http.Response, error) {
	// Construct full URL
	fullURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, errors.NewInternal("failed to parse base URL: " + err.Error())
	}

	// Parse urlPath to separate path and query
	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, errors.NewInternal("failed to parse URL path: " + err.Error())
	}

	fullURL.Path = path.Join(fullURL.Path, parsedPath.Path)
	fullURL.RawQuery = parsedPath.RawQuery

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, errors.NewInternal("failed to create request: " + err.Error())
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewServiceUnavailable("request failed: " + err.Error())
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// Try to parse error response
		var errResp struct {
			Error string `json:"error,omitempty"`
			Code  string `json:"code,omitempty"`
		}

		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &errResp)

		message := errResp.Error
		if message == "" {
			message = string(body)
		}
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}

		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, errors.NewBadRequest(message)
		case http.StatusUnauthorized:
			return nil, errors.NewUnauthorized(message)
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorCodeForbidden, message)
		case http.StatusNotFound:
			return nil, errors.NewNotFound(message)
		case http.StatusConflict:
			return nil, errors.New(errors.ErrorCodeConflict, message)
		case http.StatusTooManyRequests:
			return nil, errors.New(errors.ErrorCodeServiceUnavailable, message)
		case http.StatusInternalServerError:
			return nil, errors.NewInternal(message)
		case http.StatusServiceUnavailable:
			return nil, errors.NewServiceUnavailable(message)
		default:
			return nil, errors.NewInternal(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, message))
		}
	}

	return resp, nil
}

// Bucket represents a bucket in the artifact store
type Bucket struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
}

// Object represents an object/artifact in the artifact store
type Object struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"lastModified"`
	ETag         string            `json:"etag,omitempty"`
	ContentType  string            `json:"contentType,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UploadOptions contains options for uploading artifacts
type UploadOptions struct {
	// ContentType specifies the MIME type of the artifact
	ContentType string

	// Metadata contains custom key-value metadata
	Metadata map[string]string

	// ProgressCallback is called periodically during upload with bytes transferred
	ProgressCallback func(bytesTransferred int64)
}

// DownloadOptions contains options for downloading artifacts
type DownloadOptions struct {
	// Range specifies a byte range to download (e.g., "bytes=0-1023")
	Range string

	// ProgressCallback is called periodically during download with bytes transferred
	ProgressCallback func(bytesTransferred int64)
}

// ListOptions contains options for listing objects
type ListOptions struct {
	// Prefix filters objects by key prefix
	Prefix string

	// MaxKeys limits the number of objects returned (default: 1000)
	MaxKeys int

	// Delimiter is used for hierarchical listing
	Delimiter string
}

// ListBucketsResult contains the result of listing buckets
type ListBucketsResult struct {
	Buckets []Bucket `json:"buckets"`
}

// ListObjectsResult contains the result of listing objects
type ListObjectsResult struct {
	Objects        []Object `json:"contents"`
	CommonPrefixes []string `json:"commonPrefixes,omitempty"`
	IsTruncated    bool     `json:"isTruncated"`
	NextMarker     string   `json:"nextMarker,omitempty"`
}

// progressReader wraps an io.Reader and calls a callback with progress
type progressReader struct {
	reader   io.Reader
	callback func(int64)
	total    int64
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.total += int64(n)
		if pr.callback != nil {
			pr.callback(pr.total)
		}
	}
	return n, err
}

// progressWriter wraps an io.Writer and calls a callback with progress
type progressWriter struct {
	writer   io.Writer
	callback func(int64)
	total    int64
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if n > 0 {
		pw.total += int64(n)
		if pw.callback != nil {
			pw.callback(pw.total)
		}
	}
	return n, err
}
