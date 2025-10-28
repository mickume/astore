package client_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestBucketOperations(t *testing.T) {
	t.Run("Create bucket", func(t *testing.T) {
		// Given: Test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "PUT", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3/test-bucket", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Creating a bucket
		ctx := context.Background()
		err := c.CreateBucket(ctx, "test-bucket")

		// Then: Request succeeds
		test.AssertNoError(t, err, "create bucket")
	})

	t.Run("Delete bucket", func(t *testing.T) {
		// Given: Test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "DELETE", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3/test-bucket", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Deleting a bucket
		ctx := context.Background()
		err := c.DeleteBucket(ctx, "test-bucket")

		// Then: Request succeeds
		test.AssertNoError(t, err, "delete bucket")
	})

	t.Run("List buckets", func(t *testing.T) {
		// Given: Test server returning buckets
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"buckets": [{"name": "bucket1", "creationDate": "2024-01-01T00:00:00Z"}]}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Listing buckets
		ctx := context.Background()
		result, err := c.ListBuckets(ctx)

		// Then: Returns bucket list
		test.AssertNoError(t, err, "list buckets")
		test.AssertTrue(t, result != nil, "result should not be nil")
		test.AssertTrue(t, len(result.Buckets) == 1, "should have 1 bucket")
		test.AssertEqual(t, "bucket1", result.Buckets[0].Name, "bucket name")
	})
}

func TestObjectOperations(t *testing.T) {
	t.Run("Upload object", func(t *testing.T) {
		// Given: Test server and upload data
		uploadData := []byte("test artifact content")
		var receivedData []byte
		var receivedContentType string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "PUT", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3/test-bucket/test-key", r.URL.Path, "request path")
			receivedContentType = r.Header.Get("Content-Type")
			receivedData, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Uploading an object
		ctx := context.Background()
		opts := &client.UploadOptions{
			ContentType: "application/gzip",
		}
		err := c.Upload(ctx, "test-bucket", "test-key", bytes.NewReader(uploadData), int64(len(uploadData)), opts)

		// Then: Upload succeeds with correct data
		test.AssertNoError(t, err, "upload")
		test.AssertEqual(t, "application/gzip", receivedContentType, "content type")
		test.AssertEqual(t, string(uploadData), string(receivedData), "uploaded data")
	})

	t.Run("Upload with metadata", func(t *testing.T) {
		// Given: Test server and upload data with metadata
		uploadData := []byte("test content")
		var receivedMetadata map[string]string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMetadata = make(map[string]string)
			for key, values := range r.Header {
				if strings.HasPrefix(key, "X-Amz-Meta-") {
					metaKey := key[11:]
					if len(values) > 0 {
						receivedMetadata[metaKey] = values[0]
					}
				}
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Uploading with metadata
		ctx := context.Background()
		opts := &client.UploadOptions{
			Metadata: map[string]string{
				"version": "1.0",
				"author":  "test-user",
			},
		}
		err := c.Upload(ctx, "test-bucket", "test-key", bytes.NewReader(uploadData), int64(len(uploadData)), opts)

		// Then: Metadata is included in request
		test.AssertNoError(t, err, "upload")
		test.AssertEqual(t, "1.0", receivedMetadata["Version"], "version metadata")
		test.AssertEqual(t, "test-user", receivedMetadata["Author"], "author metadata")
	})

	t.Run("Download object", func(t *testing.T) {
		// Given: Test server with artifact data
		artifactData := []byte("downloaded artifact content")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write(artifactData)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Downloading an object
		ctx := context.Background()
		var buffer bytes.Buffer
		err := c.Download(ctx, "test-bucket", "test-key", &buffer, nil)

		// Then: Download succeeds with correct data
		test.AssertNoError(t, err, "download")
		test.AssertEqual(t, string(artifactData), buffer.String(), "downloaded data")
	})

	t.Run("Download with range", func(t *testing.T) {
		// Given: Test server with range support
		var receivedRange string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedRange = r.Header.Get("Range")
			w.WriteHeader(http.StatusPartialContent)
			w.Write([]byte("partial"))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Downloading with range
		ctx := context.Background()
		var buffer bytes.Buffer
		opts := &client.DownloadOptions{
			Range: "bytes=0-1023",
		}
		err := c.Download(ctx, "test-bucket", "test-key", &buffer, opts)

		// Then: Range header is set
		test.AssertNoError(t, err, "download")
		test.AssertEqual(t, "bytes=0-1023", receivedRange, "range header")
	})

	t.Run("Get object metadata", func(t *testing.T) {
		// Given: Test server with object metadata
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "HEAD", r.Method, "HTTP method")
			w.Header().Set("Content-Length", "1024")
			w.Header().Set("Content-Type", "application/gzip")
			w.Header().Set("ETag", "abc123")
			w.Header().Set("X-Amz-Meta-version", "1.0")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Getting object metadata
		ctx := context.Background()
		obj, err := c.GetObjectMetadata(ctx, "test-bucket", "test-key")

		// Then: Metadata is returned
		test.AssertNoError(t, err, "get metadata")
		test.AssertTrue(t, obj != nil, "object should not be nil")
		test.AssertEqual(t, int64(1024), obj.Size, "object size")
		test.AssertEqual(t, "application/gzip", obj.ContentType, "content type")
		test.AssertEqual(t, "abc123", obj.ETag, "etag")
		test.AssertEqual(t, "1.0", obj.Metadata["Version"], "custom metadata")
	})

	t.Run("Delete object", func(t *testing.T) {
		// Given: Test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "DELETE", r.Method, "HTTP method")
			test.AssertEqual(t, "/s3/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Deleting an object
		ctx := context.Background()
		err := c.DeleteObject(ctx, "test-bucket", "test-key")

		// Then: Delete succeeds
		test.AssertNoError(t, err, "delete object")
	})

	t.Run("List objects", func(t *testing.T) {
		// Given: Test server returning objects
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertTrue(t, strings.Contains(r.URL.Path, "/s3/test-bucket"), "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"contents": [{"key": "obj1", "size": 100}], "isTruncated": false}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Listing objects
		ctx := context.Background()
		result, err := c.ListObjects(ctx, "test-bucket", nil)

		// Then: Returns object list
		test.AssertNoError(t, err, "list objects")
		test.AssertTrue(t, result != nil, "result should not be nil")
		test.AssertTrue(t, len(result.Objects) == 1, "should have 1 object")
		test.AssertEqual(t, "obj1", result.Objects[0].Key, "object key")
		test.AssertEqual(t, int64(100), result.Objects[0].Size, "object size")
	})

	t.Run("List objects with prefix", func(t *testing.T) {
		// Given: Test server
		var receivedPrefix string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedPrefix = r.URL.Query().Get("prefix")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"contents": [], "isTruncated": false}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Listing with prefix
		ctx := context.Background()
		opts := &client.ListOptions{
			Prefix: "test/",
		}
		c.ListObjects(ctx, "test-bucket", opts)

		// Then: Prefix is included in query
		test.AssertEqual(t, "test/", receivedPrefix, "prefix parameter")
	})

	t.Run("Copy object", func(t *testing.T) {
		// Given: Test server
		var receivedCopySource string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "PUT", r.Method, "HTTP method")
			receivedCopySource = r.Header.Get("X-Amz-Copy-Source")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Copying an object
		ctx := context.Background()
		err := c.CopyObject(ctx, "source-bucket", "source-key", "dest-bucket", "dest-key")

		// Then: Copy succeeds with correct headers
		test.AssertNoError(t, err, "copy object")
		test.AssertEqual(t, "/source-bucket/source-key", receivedCopySource, "copy source header")
	})
}

func TestMultipartUpload(t *testing.T) {
	t.Run("Initiate multipart upload", func(t *testing.T) {
		// Given: Test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "POST", r.Method, "HTTP method")
			test.AssertTrue(t, strings.Contains(r.URL.RawQuery, "uploads"), "should have uploads parameter")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"uploadId": "test-upload-id"}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Initiating multipart upload
		ctx := context.Background()
		upload, err := c.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)

		// Then: Multipart upload is initiated
		test.AssertNoError(t, err, "initiate multipart upload")
		test.AssertTrue(t, upload != nil, "upload should not be nil")
		test.AssertEqual(t, "test-upload-id", upload.UploadID, "upload ID")
		test.AssertEqual(t, "test-bucket", upload.Bucket, "bucket")
		test.AssertEqual(t, "test-key", upload.Key, "key")
	})

	t.Run("Upload part", func(t *testing.T) {
		// Given: Test server and multipart upload
		partData := []byte("part content")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.RawQuery, "uploads") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"uploadId": "test-upload-id"}`))
			} else {
				test.AssertTrue(t, strings.Contains(r.URL.RawQuery, "uploadId=test-upload-id"), "should have uploadId")
				test.AssertTrue(t, strings.Contains(r.URL.RawQuery, "partNumber=1"), "should have partNumber")
				w.Header().Set("ETag", "part-etag-1")
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})
		ctx := context.Background()
		upload, _ := c.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)

		// When: Uploading a part
		etag, err := upload.UploadPart(ctx, 1, bytes.NewReader(partData), int64(len(partData)))

		// Then: Part upload succeeds
		test.AssertNoError(t, err, "upload part")
		test.AssertEqual(t, "part-etag-1", etag, "part etag")
	})

	t.Run("Complete multipart upload", func(t *testing.T) {
		// Given: Test server and multipart upload
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.RawQuery, "uploads") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"uploadId": "test-upload-id"}`))
			} else if r.Method == "POST" {
				test.AssertTrue(t, strings.Contains(r.URL.RawQuery, "uploadId=test-upload-id"), "should have uploadId")
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})
		ctx := context.Background()
		upload, _ := c.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)

		// When: Completing multipart upload
		parts := []client.CompletedPart{
			{PartNumber: 1, ETag: "etag1"},
			{PartNumber: 2, ETag: "etag2"},
		}
		err := upload.Complete(ctx, parts)

		// Then: Multipart upload completes
		test.AssertNoError(t, err, "complete multipart upload")
	})

	t.Run("Abort multipart upload", func(t *testing.T) {
		// Given: Test server and multipart upload
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.RawQuery, "uploads") && r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"uploadId": "test-upload-id"}`))
			} else if r.Method == "DELETE" {
				test.AssertTrue(t, strings.Contains(r.URL.RawQuery, "uploadId=test-upload-id"), "should have uploadId")
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})
		ctx := context.Background()
		upload, _ := c.InitiateMultipartUpload(ctx, "test-bucket", "test-key", nil)

		// When: Aborting multipart upload
		err := upload.Abort(ctx)

		// Then: Multipart upload is aborted
		test.AssertNoError(t, err, "abort multipart upload")
	})
}

func TestProgressCallbacks(t *testing.T) {
	t.Run("Upload progress callback", func(t *testing.T) {
		// Given: Test server and progress tracking
		uploadData := []byte("test content for progress")
		var progressCalled bool
		var totalBytes int64

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Uploading with progress callback
		ctx := context.Background()
		opts := &client.UploadOptions{
			ProgressCallback: func(bytes int64) {
				progressCalled = true
				totalBytes = bytes
			},
		}
		err := c.Upload(ctx, "test-bucket", "test-key", bytes.NewReader(uploadData), int64(len(uploadData)), opts)

		// Then: Progress callback is called
		test.AssertNoError(t, err, "upload")
		test.AssertTrue(t, progressCalled, "progress callback should be called")
		test.AssertTrue(t, totalBytes > 0, "total bytes should be tracked")
	})

	t.Run("Download progress callback", func(t *testing.T) {
		// Given: Test server and progress tracking
		downloadData := []byte("test download content")
		var progressCalled bool
		var totalBytes int64

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(downloadData)
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Downloading with progress callback
		ctx := context.Background()
		var buffer bytes.Buffer
		opts := &client.DownloadOptions{
			ProgressCallback: func(bytes int64) {
				progressCalled = true
				totalBytes = bytes
			},
		}
		err := c.Download(ctx, "test-bucket", "test-key", &buffer, opts)

		// Then: Progress callback is called
		test.AssertNoError(t, err, "download")
		test.AssertTrue(t, progressCalled, "progress callback should be called")
		test.AssertTrue(t, totalBytes > 0, "total bytes should be tracked")
	})
}
