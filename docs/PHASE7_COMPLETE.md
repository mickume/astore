# Phase 7: Go Client SDK - COMPLETE ✅

## Overview

Phase 7 implements a comprehensive Go SDK for the Zot Artifact Store, providing a type-safe, idiomatic Go client library for artifact management with support for authentication, progress tracking, multipart uploads, and supply chain security operations.

**Completion Date:** 2025-10-28

## Implementation Summary

### Components Delivered

1. **Client Foundation** (`pkg/client/client.go`)
2. **Core Operations** (`pkg/client/operations.go`)
3. **Supply Chain Integration** (`pkg/client/supplychain.go`)
4. **Comprehensive Tests** (43 tests passing)

## Features

### 1. Client Foundation

**Client Structure:**

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    token      string
    userAgent  string
}

type Config struct {
    BaseURL            string       // Artifact store endpoint
    Token              string       // Bearer authentication token
    HTTPClient         *http.Client // Custom HTTP client (optional)
    Timeout            time.Duration // Request timeout (default: 30s)
    InsecureSkipVerify bool         // Skip TLS verification (testing only)
    UserAgent          string       // Custom User-Agent header
}
```

**Client Creation:**

```go
config := &client.Config{
    BaseURL: "https://artifacts.example.com",
    Token:   "your-bearer-token",
    Timeout: 60 * time.Second,
}

c, err := client.NewClient(config)
if err != nil {
    log.Fatal(err)
}

// Update token dynamically
c.SetToken("new-token")
```

**Features:**
- Configurable HTTP client with custom timeouts
- Automatic bearer token authentication
- Custom User-Agent support
- TLS configuration (including insecure mode for testing)
- Automatic error handling and classification
- Context-aware requests for cancellation

### 2. Core Artifact Operations

**Upload:**

```go
// Simple upload
data := bytes.NewReader(artifactData)
err := c.Upload(ctx, "releases", "app-1.0.0.tar.gz", data, int64(len(artifactData)), nil)

// Upload with metadata and progress tracking
opts := &client.UploadOptions{
    ContentType: "application/gzip",
    Metadata: map[string]string{
        "version": "1.0.0",
        "author":  "ci-system",
    },
    ProgressCallback: func(bytesTransferred int64) {
        fmt.Printf("Uploaded: %d bytes\n", bytesTransferred)
    },
}

err := c.Upload(ctx, "releases", "app-1.0.0.tar.gz", data, size, opts)
```

**Download:**

```go
// Simple download
var buffer bytes.Buffer
err := c.Download(ctx, "releases", "app-1.0.0.tar.gz", &buffer, nil)

// Download with range request and progress tracking
opts := &client.DownloadOptions{
    Range: "bytes=0-1023", // First 1KB
    ProgressCallback: func(bytesTransferred int64) {
        fmt.Printf("Downloaded: %d bytes\n", bytesTransferred)
    },
}

err := c.Download(ctx, "releases", "app-1.0.0.tar.gz", file, opts)
```

**List Objects:**

```go
// List all objects in bucket
result, err := c.ListObjects(ctx, "releases", nil)
for _, obj := range result.Objects {
    fmt.Printf("%s (%d bytes)\n", obj.Key, obj.Size)
}

// List with prefix filter
opts := &client.ListOptions{
    Prefix:  "app/",
    MaxKeys: 100,
}
result, err := c.ListObjects(ctx, "releases", opts)
```

**Object Metadata:**

```go
obj, err := c.GetObjectMetadata(ctx, "releases", "app-1.0.0.tar.gz")
fmt.Printf("Size: %d bytes\n", obj.Size)
fmt.Printf("Type: %s\n", obj.ContentType)
fmt.Printf("ETag: %s\n", obj.ETag)
fmt.Printf("Version: %s\n", obj.Metadata["Version"])
```

**Delete Object:**

```go
err := c.DeleteObject(ctx, "releases", "app-1.0.0.tar.gz")
```

**Copy Object:**

```go
err := c.CopyObject(ctx, "releases", "app-1.0.0.tar.gz", "archive", "app-1.0.0-backup.tar.gz")
```

### 3. Bucket Management

**Create Bucket:**

```go
err := c.CreateBucket(ctx, "my-new-bucket")
```

**List Buckets:**

```go
result, err := c.ListBuckets(ctx)
for _, bucket := range result.Buckets {
    fmt.Printf("%s (created: %s)\n", bucket.Name, bucket.CreationDate)
}
```

**Delete Bucket:**

```go
err := c.DeleteBucket(ctx, "old-bucket")
```

### 4. Multipart Upload

For large files (>5MB recommended):

```go
// Initiate multipart upload
upload, err := c.InitiateMultipartUpload(ctx, "releases", "large-app.tar.gz", &client.UploadOptions{
    ContentType: "application/gzip",
    Metadata: map[string]string{
        "size": "500MB",
    },
})

// Upload parts
parts := []client.CompletedPart{}
partSize := int64(5 * 1024 * 1024) // 5MB parts

for partNumber := 1; partNumber <= totalParts; partNumber++ {
    partData := getPartData(partNumber) // Your function to get part data
    etag, err := upload.UploadPart(ctx, partNumber, partData, partSize)
    if err != nil {
        upload.Abort(ctx) // Abort on error
        return err
    }

    parts = append(parts, client.CompletedPart{
        PartNumber: partNumber,
        ETag:       etag,
    })
}

// Complete multipart upload
err = upload.Complete(ctx, parts)

// Or abort if needed
err = upload.Abort(ctx)
```

### 5. Supply Chain Operations

**Sign Artifact:**

```go
signature, err := c.SignArtifact(ctx, "releases", "app-1.0.0.tar.gz", privateKeyPEM)
fmt.Printf("Signed with ID: %s\n", signature.ID)
fmt.Printf("Algorithm: %s\n", signature.Algorithm)
```

**Verify Signatures:**

```go
result, err := c.VerifySignatures(ctx, "releases", "app-1.0.0.tar.gz", []string{publicKeyPEM})
if result.Valid {
    fmt.Println("All signatures valid!")
} else {
    fmt.Printf("Verification failed: %s\n", result.Message)
}
```

**Get Signatures:**

```go
signatures, err := c.GetSignatures(ctx, "releases", "app-1.0.0.tar.gz")
for _, sig := range signatures {
    fmt.Printf("Signature: %s (signed by %s)\n", sig.ID, sig.SignedBy)
}
```

**Attach SBOM:**

```go
sbomContent := `{
    "spdxVersion": "SPDX-2.3",
    "packages": [...]
}`

sbom, err := c.AttachSBOM(ctx, "releases", "app-1.0.0.tar.gz", "spdx", sbomContent)
fmt.Printf("SBOM attached: %s\n", sbom.ID)
```

**Get SBOM:**

```go
sbom, err := c.GetSBOM(ctx, "releases", "app-1.0.0.tar.gz")
fmt.Printf("Format: %s\n", sbom.Format)
fmt.Printf("Content: %s\n", sbom.Content)
```

**Add Attestation:**

```go
attData := map[string]interface{}{
    "buildId":    "12345",
    "status":     "success",
    "duration":   "5m30s",
    "testsPassed": 142,
}

att, err := c.AddAttestation(ctx, "releases", "app-1.0.0.tar.gz", "build", attData)
fmt.Printf("Attestation added: %s\n", att.ID)
```

**Get Attestations:**

```go
attestations, err := c.GetAttestations(ctx, "releases", "app-1.0.0.tar.gz")
for _, att := range attestations {
    fmt.Printf("Type: %s, ID: %s\n", att.Type, att.ID)
    fmt.Printf("Data: %v\n", att.Data)
}
```

### 6. Error Handling

The SDK provides comprehensive error handling with typed errors:

```go
import "github.com/candlekeep/zot-artifact-store/internal/errors"

err := c.Upload(ctx, "bucket", "key", data, size, nil)
if err != nil {
    // Check error type
    if errors.IsNotFound(err) {
        fmt.Println("Bucket not found")
    } else if errors.IsUnauthorized(err) {
        fmt.Println("Authentication failed")
    } else if errors.IsServiceUnavailable(err) {
        fmt.Println("Service temporarily unavailable, retry later")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

**Error Types:**
- `ErrorCodeBadRequest` (400) - Invalid request
- `ErrorCodeUnauthorized` (401) - Authentication required
- `ErrorCodeForbidden` (403) - Permission denied
- `ErrorCodeNotFound` (404) - Resource not found
- `ErrorCodeConflict` (409) - Resource conflict
- `ErrorCodeServiceUnavailable` (503) - Service unavailable
- `ErrorCodeInternal` (500) - Internal server error

### 7. Progress Tracking

Track upload and download progress:

```go
var totalBytes int64
progressCallback := func(bytes int64) {
    totalBytes = bytes
    percentage := float64(bytes) / float64(totalSize) * 100
    fmt.Printf("\rProgress: %.2f%% (%d/%d bytes)", percentage, bytes, totalSize)
}

opts := &client.UploadOptions{
    ProgressCallback: progressCallback,
}

err := c.Upload(ctx, bucket, key, data, size, opts)
```

### 8. Context Support

All operations support context for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := c.Upload(ctx, bucket, key, data, size, nil)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())

go func() {
    // Cancel on user interrupt
    <-interruptSignal
    cancel()
}()

err := c.Download(ctx, bucket, key, writer, nil)
```

## Testing

### Test Coverage

```
=== RUN   TestNewClient
=== RUN   TestNewClient/Create_client_with_valid_config                 ✅
=== RUN   TestNewClient/Create_client_with_missing_base_URL             ✅
=== RUN   TestNewClient/Create_client_with_invalid_base_URL             ✅
=== RUN   TestNewClient/Create_client_with_custom_HTTP_client           ✅
=== RUN   TestNewClient/Create_client_with_insecure_TLS                 ✅
--- PASS: TestNewClient (5/5)

=== RUN   TestSetToken
=== RUN   TestSetToken/Update_authentication_token                      ✅
--- PASS: TestSetToken (1/1)

=== RUN   TestHTTPErrorHandling
=== RUN   TestHTTPErrorHandling/400_Bad_Request                         ✅
=== RUN   TestHTTPErrorHandling/401_Unauthorized                        ✅
=== RUN   TestHTTPErrorHandling/403_Forbidden                           ✅
=== RUN   TestHTTPErrorHandling/404_Not_Found                           ✅
=== RUN   TestHTTPErrorHandling/409_Conflict                            ✅
=== RUN   TestHTTPErrorHandling/500_Internal_Server_Error               ✅
=== RUN   TestHTTPErrorHandling/503_Service_Unavailable                 ✅
=== RUN   TestHTTPErrorHandling/200_OK                                  ✅
--- PASS: TestHTTPErrorHandling (8/8)

=== RUN   TestAuthenticationHeader
=== RUN   TestAuthenticationHeader/Request_includes_bearer_token        ✅
=== RUN   TestAuthenticationHeader/Request_without_token_has_no_auth_header ✅
--- PASS: TestAuthenticationHeader (2/2)

=== RUN   TestUserAgent
=== RUN   TestUserAgent/Request_includes_default_user_agent             ✅
=== RUN   TestUserAgent/Request_includes_custom_user_agent              ✅
--- PASS: TestUserAgent (2/2)

=== RUN   TestBucketOperations
=== RUN   TestBucketOperations/Create_bucket                            ✅
=== RUN   TestBucketOperations/Delete_bucket                            ✅
=== RUN   TestBucketOperations/List_buckets                             ✅
--- PASS: TestBucketOperations (3/3)

=== RUN   TestObjectOperations
=== RUN   TestObjectOperations/Upload_object                            ✅
=== RUN   TestObjectOperations/Upload_with_metadata                     ✅
=== RUN   TestObjectOperations/Download_object                          ✅
=== RUN   TestObjectOperations/Download_with_range                      ✅
=== RUN   TestObjectOperations/Get_object_metadata                      ✅
=== RUN   TestObjectOperations/Delete_object                            ✅
=== RUN   TestObjectOperations/List_objects                             ✅
=== RUN   TestObjectOperations/List_objects_with_prefix                 ✅
=== RUN   TestObjectOperations/Copy_object                              ✅
--- PASS: TestObjectOperations (9/9)

=== RUN   TestMultipartUpload
=== RUN   TestMultipartUpload/Initiate_multipart_upload                 ✅
=== RUN   TestMultipartUpload/Upload_part                               ✅
=== RUN   TestMultipartUpload/Complete_multipart_upload                 ✅
=== RUN   TestMultipartUpload/Abort_multipart_upload                    ✅
--- PASS: TestMultipartUpload (4/4)

=== RUN   TestProgressCallbacks
=== RUN   TestProgressCallbacks/Upload_progress_callback                ✅
=== RUN   TestProgressCallbacks/Download_progress_callback              ✅
--- PASS: TestProgressCallbacks (2/2)

=== RUN   TestSupplyChainOperations
=== RUN   TestSupplyChainOperations/Sign_artifact                       ✅
=== RUN   TestSupplyChainOperations/Get_signatures                      ✅
=== RUN   TestSupplyChainOperations/Verify_signatures                   ✅
=== RUN   TestSupplyChainOperations/Attach_SBOM                         ✅
=== RUN   TestSupplyChainOperations/Get_SBOM                            ✅
=== RUN   TestSupplyChainOperations/Add_attestation                     ✅
=== RUN   TestSupplyChainOperations/Get_attestations                    ✅
--- PASS: TestSupplyChainOperations (7/7)

PASS
ok  	github.com/candlekeep/zot-artifact-store/pkg/client	0.198s
```

**Total Tests:** 43/43 passing

### Test Scenarios

- ✅ Client creation and configuration
- ✅ Authentication (bearer tokens)
- ✅ HTTP error handling (400, 401, 403, 404, 409, 500, 503)
- ✅ User-Agent headers
- ✅ Bucket operations (create, delete, list)
- ✅ Object operations (upload, download, delete, list)
- ✅ Custom metadata
- ✅ Range requests
- ✅ Multipart uploads
- ✅ Progress callbacks
- ✅ Supply chain operations (sign, verify, SBOM, attestations)
- ✅ Copy operations

## Files Added/Modified

### New Files (6)

- `pkg/client/client.go` - Client foundation and HTTP handling (280 lines)
- `pkg/client/operations.go` - Core artifact operations (350 lines)
- `pkg/client/supplychain.go` - Supply chain integration (200 lines)
- `pkg/client/client_test.go` - Client tests (220 lines)
- `pkg/client/operations_test.go` - Operations tests (400 lines)
- `pkg/client/supplychain_test.go` - Supply chain tests (220 lines)

**Total:** ~1,670 lines of production code + tests

## Usage Examples

### Complete Upload/Download Workflow

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/candlekeep/zot-artifact-store/pkg/client"
)

func main() {
    // Create client
    c, err := client.NewClient(&client.Config{
        BaseURL: "https://artifacts.example.com",
        Token:   os.Getenv("ARTIFACT_STORE_TOKEN"),
        Timeout: 60 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Read artifact file
    data, err := os.ReadFile("myapp-1.0.0.tar.gz")
    if err != nil {
        log.Fatal(err)
    }

    // Upload with progress tracking
    fmt.Println("Uploading artifact...")
    err = c.Upload(ctx, "releases", "myapp-1.0.0.tar.gz",
        bytes.NewReader(data),
        int64(len(data)),
        &client.UploadOptions{
            ContentType: "application/gzip",
            Metadata: map[string]string{
                "version": "1.0.0",
                "commit":  "abc123",
            },
            ProgressCallback: func(bytes int64) {
                pct := float64(bytes) / float64(len(data)) * 100
                fmt.Printf("\rProgress: %.1f%%", pct)
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nUpload complete!")

    // Sign the artifact
    privateKey := os.Getenv("SIGNING_KEY")
    sig, err := c.SignArtifact(ctx, "releases", "myapp-1.0.0.tar.gz", privateKey)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Artifact signed: %s\n", sig.ID)

    // Download and verify
    var buffer bytes.Buffer
    err = c.Download(ctx, "releases", "myapp-1.0.0.tar.gz", &buffer, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Verify signature
    publicKey := os.Getenv("VERIFY_KEY")
    result, err := c.VerifySignatures(ctx, "releases", "myapp-1.0.0.tar.gz", []string{publicKey})
    if err != nil {
        log.Fatal(err)
    }

    if result.Valid {
        fmt.Println("Signature verification passed!")
    } else {
        log.Fatalf("Signature verification failed: %s", result.Message)
    }
}
```

### CI/CD Integration

```go
// Upload build artifact
func uploadBuildArtifact(buildID string, artifactPath string) error {
    c, _ := client.NewClient(&client.Config{
        BaseURL: os.Getenv("ARTIFACT_STORE_URL"),
        Token:   os.Getenv("ARTIFACT_STORE_TOKEN"),
    })

    data, _ := os.ReadFile(artifactPath)
    ctx := context.Background()

    // Upload
    err := c.Upload(ctx, "builds", fmt.Sprintf("build-%s.tar.gz", buildID),
        bytes.NewReader(data), int64(len(data)),
        &client.UploadOptions{
            Metadata: map[string]string{
                "build-id": buildID,
                "commit":   os.Getenv("GIT_COMMIT"),
                "branch":   os.Getenv("GIT_BRANCH"),
            },
        },
    )
    if err != nil {
        return err
    }

    // Add build attestation
    attestData := map[string]interface{}{
        "buildId":      buildID,
        "status":       "success",
        "testsPassed":  142,
        "testsFailed":  0,
        "coverage":     "85.3%",
        "duration":     "5m30s",
    }

    _, err = c.AddAttestation(ctx, "builds", fmt.Sprintf("build-%s.tar.gz", buildID),
        "build", attestData)

    return err
}
```

## Best Practices

### 1. Context Management

Always use context with timeout or cancellation:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := c.Upload(ctx, bucket, key, data, size, nil)
```

### 2. Error Handling

Check for specific error types:

```go
err := c.Download(ctx, bucket, key, writer, nil)
if err != nil {
    if errors.IsNotFound(err) {
        // Handle not found
    } else if errors.IsServiceUnavailable(err) {
        // Retry later
    } else {
        // General error handling
    }
}
```

### 3. Large File Uploads

Use multipart upload for files >5MB:

```go
if fileSize > 5*1024*1024 { // >5MB
    upload, _ := c.InitiateMultipartUpload(ctx, bucket, key, opts)
    // Upload in parts
    // Complete upload
} else {
    c.Upload(ctx, bucket, key, data, size, opts)
}
```

### 4. Progress Tracking

Provide user feedback for long operations:

```go
opts := &client.UploadOptions{
    ProgressCallback: func(bytes int64) {
        fmt.Printf("\rUploaded: %d MB", bytes/1024/1024)
    },
}
```

### 5. Token Management

Update tokens without recreating client:

```go
// Initial token
c, _ := client.NewClient(&client.Config{
    BaseURL: url,
    Token:   initialToken,
})

// Token refresh
newToken := refreshToken()
c.SetToken(newToken)
```

## Integration Benefits

### For Applications
- Type-safe Go API
- Idiomatic Go patterns
- Comprehensive error handling
- Progress tracking built-in
- Context-aware operations

### For CI/CD
- Easy integration with build pipelines
- Attestation support for build metadata
- SBOM attachment for compliance
- Signature verification for security

### For Testing
- Mockable HTTP client
- Configurable timeouts
- Insecure mode for local testing

## Known Limitations

1. **In-Memory Buffering**: Multipart uploads buffer parts in memory for retries
2. **No Streaming Signatures**: Signature operations require full content in memory
3. **No Concurrent Upload**: Multipart parts are uploaded sequentially
4. **Limited Retry Logic**: Retries handled by server-side, not client-side

## Future Enhancements

### Phase 7.1: Advanced Features

1. **Concurrent Multipart Upload**
   - Upload parts in parallel
   - Configurable concurrency level
   - Progress aggregation

2. **Streaming Operations**
   - Stream-based signing/verification
   - Reduced memory footprint
   - Large file support

3. **Built-in Retry Logic**
   - Exponential backoff
   - Configurable retry policy
   - Automatic retry for transient errors

### Phase 7.2: Performance

1. **Connection Pooling**
   - Optimized HTTP connection reuse
   - Configurable pool size
   - Keep-alive tuning

2. **Compression**
   - Automatic gzip compression
   - Configurable compression level
   - Bandwidth optimization

3. **Caching**
   - Metadata caching
   - ETags for conditional requests
   - Reduced network overhead

## Conclusion

Phase 7 successfully delivers a production-ready Go SDK:

- ✅ **Complete API Coverage**: All S3 and supply chain operations
- ✅ **Type-Safe**: Strongly typed Go API
- ✅ **Well-Tested**: 43/43 tests passing
- ✅ **Production Ready**: Error handling, timeouts, authentication
- ✅ **Developer Friendly**: Idiomatic Go, comprehensive documentation
- ✅ **Supply Chain Support**: Full integration with signing, SBOM, attestations

The Zot Artifact Store Go SDK provides a robust, type-safe client library for artifact management in Go applications.

---

**Status:** ✅ COMPLETE
**Date:** 2025-10-28
**Tests:** 43/43 passing
**Lines of Code:** ~1,670 (production + tests)
**Next Phase:** Phase 8 (Python Client SDK) or Phase 10 (CLI Tool)
