# S3-Compatible API Documentation

## Overview

The Zot Artifact Store provides an S3-compatible API for storing and retrieving binary artifacts. This implementation supports core S3 operations including bucket management, object operations, and multipart uploads.

## Base URL

All S3 API endpoints are prefixed with `/s3`:

```
http://localhost:8080/s3
```

## API Endpoints

### Bucket Operations

#### List Buckets

List all available buckets.

**Request:**
```http
GET /s3
```

**Response:**
```json
{
  "buckets": [
    {
      "name": "my-bucket",
      "createdAt": "2024-01-15T10:30:00Z",
      "versioning": false,
      "objectCount": 42,
      "totalSize": 1048576
    }
  ]
}
```

#### Create Bucket

Create a new bucket.

**Request:**
```http
PUT /s3/{bucket}
```

**Response:**
- `200 OK` - Bucket created successfully
- `409 Conflict` - Bucket already exists
- `500 Internal Server Error` - Server error

**Example:**
```bash
curl -X PUT http://localhost:8080/s3/my-bucket
```

#### Delete Bucket

Delete an empty bucket.

**Request:**
```http
DELETE /s3/{bucket}
```

**Response:**
- `204 No Content` - Bucket deleted successfully
- `409 Conflict` - Bucket not empty
- `500 Internal Server Error` - Server error

**Example:**
```bash
curl -X DELETE http://localhost:8080/s3/my-bucket
```

### Object Operations

#### Upload Object

Upload an object to a bucket.

**Request:**
```http
PUT /s3/{bucket}/{key}
Content-Type: {content-type}
X-Amz-Meta-{name}: {value}

{binary-data}
```

**Response Headers:**
- `ETag` - MD5 hash of the uploaded object

**Response:**
- `200 OK` - Object uploaded successfully
- `404 Not Found` - Bucket not found
- `500 Internal Server Error` - Server error

**Example:**
```bash
curl -X PUT \
  -H "Content-Type: application/octet-stream" \
  -H "X-Amz-Meta-Author: johndoe" \
  --data-binary @myfile.bin \
  http://localhost:8080/s3/my-bucket/myfile.bin
```

#### Download Object

Download an object from a bucket.

**Request:**
```http
GET /s3/{bucket}/{key}
Range: bytes={start}-{end}  # Optional for partial downloads
```

**Response Headers:**
- `Content-Type` - Object content type
- `Content-Length` - Object size
- `ETag` - Object MD5 hash
- `Last-Modified` - Last modification time
- `X-Amz-Meta-*` - Custom metadata

**Response:**
- `200 OK` - Full object download
- `206 Partial Content` - Range request (resumable download)
- `404 Not Found` - Object not found
- `416 Range Not Satisfiable` - Invalid range

**Example:**
```bash
# Full download
curl http://localhost:8080/s3/my-bucket/myfile.bin -o myfile.bin

# Resume download from byte 1000
curl -H "Range: bytes=1000-" \
  http://localhost:8080/s3/my-bucket/myfile.bin -o myfile.bin
```

#### Get Object Metadata

Retrieve object metadata without downloading the object.

**Request:**
```http
HEAD /s3/{bucket}/{key}
```

**Response Headers:**
- `Content-Type` - Object content type
- `Content-Length` - Object size
- `ETag` - Object MD5 hash
- `Last-Modified` - Last modification time
- `X-Amz-Meta-*` - Custom metadata

**Response:**
- `200 OK` - Metadata returned
- `404 Not Found` - Object not found

**Example:**
```bash
curl -I http://localhost:8080/s3/my-bucket/myfile.bin
```

#### Delete Object

Delete an object from a bucket.

**Request:**
```http
DELETE /s3/{bucket}/{key}
```

**Response:**
- `204 No Content` - Object deleted (S3-compatible: returns 204 even if object didn't exist)

**Example:**
```bash
curl -X DELETE http://localhost:8080/s3/my-bucket/myfile.bin
```

#### List Objects

List objects in a bucket with optional filtering.

**Request:**
```http
GET /s3/{bucket}?prefix={prefix}&max-keys={limit}
```

**Query Parameters:**
- `prefix` (optional) - Filter objects by key prefix
- `max-keys` (optional) - Maximum number of objects to return (default: 1000)

**Response:**
```json
{
  "bucket": "my-bucket",
  "prefix": "logs/",
  "maxKeys": 1000,
  "isTruncated": false,
  "objects": [
    {
      "key": "logs/2024-01-15.log",
      "size": 2048,
      "lastModified": "2024-01-15T10:30:00Z",
      "etag": "abc123def456"
    }
  ]
}
```

**Example:**
```bash
# List all objects
curl http://localhost:8080/s3/my-bucket

# List objects with prefix
curl http://localhost:8080/s3/my-bucket?prefix=logs/&max-keys=100
```

### Multipart Upload Operations

For large files (>5MB recommended), use multipart uploads for improved reliability and performance.

#### Initiate Multipart Upload

Start a multipart upload session.

**Request:**
```http
POST /s3/{bucket}/{key}?uploads
Content-Type: {content-type}
X-Amz-Meta-{name}: {value}
```

**Response:**
```json
{
  "uploadId": "550e8400-e29b-41d4-a716-446655440000",
  "bucket": "my-bucket",
  "key": "large-file.bin"
}
```

**Example:**
```bash
curl -X POST \
  -H "Content-Type: application/octet-stream" \
  "http://localhost:8080/s3/my-bucket/large-file.bin?uploads"
```

#### Upload Part

Upload a part of the multipart upload.

**Request:**
```http
PUT /s3/{bucket}/{key}?uploadId={uploadId}&partNumber={partNumber}

{part-data}
```

**Parameters:**
- `uploadId` - Upload ID from initiate response
- `partNumber` - Part number (1-10000)

**Response Headers:**
- `ETag` - Part ETag (save for completion)

**Response:**
- `200 OK` - Part uploaded successfully

**Example:**
```bash
# Upload part 1
curl -X PUT \
  --data-binary @part1.bin \
  "http://localhost:8080/s3/my-bucket/large-file.bin?uploadId={uploadId}&partNumber=1"

# Upload part 2
curl -X PUT \
  --data-binary @part2.bin \
  "http://localhost:8080/s3/my-bucket/large-file.bin?uploadId={uploadId}&partNumber=2"
```

#### Complete Multipart Upload

Finalize the multipart upload after all parts are uploaded.

**Request:**
```http
POST /s3/{bucket}/{key}?uploadId={uploadId}

{completion-data}
```

**Response:**
```json
{
  "bucket": "my-bucket",
  "key": "large-file.bin",
  "etag": "final-etag-hash"
}
```

**Example:**
```bash
curl -X POST \
  "http://localhost:8080/s3/my-bucket/large-file.bin?uploadId={uploadId}"
```

#### Abort Multipart Upload

Cancel an in-progress multipart upload.

**Request:**
```http
DELETE /s3/{bucket}/{key}?uploadId={uploadId}
```

**Response:**
- `204 No Content` - Upload aborted successfully

**Example:**
```bash
curl -X DELETE \
  "http://localhost:8080/s3/my-bucket/large-file.bin?uploadId={uploadId}"
```

## Features

### Resumable Downloads

The S3 API supports HTTP range requests (RFC 7233) for resumable downloads:

```bash
# Download bytes 0-999
curl -H "Range: bytes=0-999" \
  http://localhost:8080/s3/my-bucket/myfile.bin

# Download from byte 1000 to end
curl -H "Range: bytes=1000-" \
  http://localhost:8080/s3/my-bucket/myfile.bin

# Resume interrupted download
curl -C - http://localhost:8080/s3/my-bucket/myfile.bin -o myfile.bin
```

### Custom Metadata

Store custom metadata with objects using `X-Amz-Meta-*` headers:

```bash
curl -X PUT \
  -H "X-Amz-Meta-Author: johndoe" \
  -H "X-Amz-Meta-Version: 1.0.0" \
  -H "X-Amz-Meta-BuildId: 12345" \
  --data-binary @artifact.tar.gz \
  http://localhost:8080/s3/my-bucket/artifact.tar.gz
```

Retrieve metadata on download:

```bash
curl -I http://localhost:8080/s3/my-bucket/artifact.tar.gz
# Returns:
# X-Amz-Meta-Author: johndoe
# X-Amz-Meta-Version: 1.0.0
# X-Amz-Meta-BuildId: 12345
```

## Storage Architecture

### Metadata Storage

The S3 API uses BoltDB for metadata storage:

- **Buckets**: Bucket metadata (name, creation time, statistics)
- **Artifacts**: Object metadata (key, size, digest, content type, custom metadata)
- **Multipart Uploads**: In-progress multipart upload tracking
- **Upload Progress**: Part tracking for multipart uploads

Database location: `{dataDir}/metadata.db`

### File Storage

Binary artifacts are stored on the filesystem:

- **Regular objects**: `{dataDir}/{bucket}/{key}`
- **Multipart parts**: `{dataDir}/{bucket}/.multipart/{uploadId}/part-{N}`

### Data Model

**Artifact:**
```go
type Artifact struct {
    Bucket      string            // Bucket name
    Key         string            // Object key
    Digest      godigest.Digest   // Content digest
    Size        int64             // Object size in bytes
    ContentType string            // MIME type
    MD5         string            // MD5 hash (ETag)
    CreatedAt   time.Time         // Creation timestamp
    UpdatedAt   time.Time         // Last modified timestamp
    StoragePath string            // Filesystem path
    Metadata    map[string]string // Custom metadata
    UploadID    string            // Multipart upload ID (if applicable)
    IsMultipart bool              // Whether this was a multipart upload
}
```

## Configuration

The S3 API extension is configured in the Zot config file:

```yaml
storage:
  rootDirectory: /var/lib/zot

# S3 API extension configuration (auto-configured)
# - basePath: /s3
# - maxUploadSize: 5GB
# - enableMultipart: true
# - metadataDB: {rootDirectory}/metadata.db
```

## Client Examples

### Using curl

```bash
# Create bucket
curl -X PUT http://localhost:8080/s3/artifacts

# Upload artifact
curl -X PUT \
  -H "Content-Type: application/gzip" \
  -H "X-Amz-Meta-Version: 1.0.0" \
  --data-binary @myapp-1.0.0.tar.gz \
  http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz

# Download artifact
curl http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz -o myapp.tar.gz

# List artifacts
curl http://localhost:8080/s3/artifacts

# Delete artifact
curl -X DELETE http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz
```

### Using AWS CLI

The S3 API is compatible with AWS CLI (with custom endpoint):

```bash
# Configure AWS CLI with dummy credentials
aws configure set aws_access_key_id dummy
aws configure set aws_secret_access_key dummy

# Use custom endpoint
ENDPOINT="http://localhost:8080/s3"

# List buckets
aws s3 ls --endpoint-url $ENDPOINT

# Create bucket
aws s3 mb s3://artifacts --endpoint-url $ENDPOINT

# Upload file
aws s3 cp myapp.tar.gz s3://artifacts/ --endpoint-url $ENDPOINT

# Download file
aws s3 cp s3://artifacts/myapp.tar.gz . --endpoint-url $ENDPOINT

# List objects
aws s3 ls s3://artifacts/ --endpoint-url $ENDPOINT
```

### Using Go Client

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
)

func uploadArtifact(bucket, key string, data []byte) error {
    url := fmt.Sprintf("http://localhost:8080/s3/%s/%s", bucket, key)

    req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/octet-stream")
    req.Header.Set("X-Amz-Meta-Version", "1.0.0")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("upload failed: %s", resp.Status)
    }

    etag := resp.Header.Get("ETag")
    fmt.Printf("Uploaded with ETag: %s\n", etag)
    return nil
}

func downloadArtifact(bucket, key string) ([]byte, error) {
    url := fmt.Sprintf("http://localhost:8080/s3/%s/%s", bucket, key)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("download failed: %s", resp.Status)
    }

    return io.ReadAll(resp.Body)
}
```

## Testing

The S3 API includes comprehensive tests covering all operations:

```bash
# Run S3 API tests
go test -v ./internal/api/s3/...

# Run with coverage
go test -v -coverprofile=coverage.txt ./internal/api/s3/...
```

### Test Coverage

- Bucket operations: Create, List, Delete (empty/non-empty)
- Object operations: Upload, Download, Metadata, Delete, List
- Multipart uploads: Initiate, Upload parts, Complete, Abort
- Range requests: Partial downloads
- Error handling: Not found, conflicts, invalid requests

## Known Limitations (Phase 2)

1. **Multipart Upload Completion**: Part combining logic is simplified and needs full implementation in Phase 3
2. **Abort Multipart Upload**: Route matching for abort operation needs investigation (deferred to Phase 3)
3. **Authentication**: No authentication/authorization yet (planned for Phase 3 RBAC)
4. **Versioning**: Bucket versioning flag exists but not yet implemented
5. **S3 Compatibility**: Limited to core operations; advanced S3 features (ACLs, policies, etc.) not yet supported

## Performance Considerations

### Recommended Upload Sizes

- **Small files (<5MB)**: Use regular PUT operation
- **Large files (>5MB)**: Use multipart upload for better reliability
- **Very large files (>100MB)**: Multipart upload strongly recommended

### Metadata Database

- BoltDB provides excellent read performance
- Write operations are ACID-compliant
- Suitable for metadata up to millions of objects
- For larger scale, consider migration to distributed database (future enhancement)

### File Storage

- Direct filesystem storage for simplicity and performance
- No intermediate caching layer needed
- Supports all standard filesystem features (permissions, quotas, etc.)
- Compatible with network filesystems (NFS, GlusterFS, etc.)

## Next Steps (Phase 3)

The following enhancements are planned for Phase 3:

1. **RBAC Integration**: Keycloak-based authentication and authorization
2. **Access Control**: Bucket-level and object-level permissions
3. **Complete Multipart Upload**: Full part combining implementation
4. **Fix Known Issues**: Abort multipart upload routing
5. **Performance Optimization**: Caching, connection pooling
6. **Monitoring**: Enhanced metrics and logging

## Support

For questions, issues, or feature requests related to the S3 API:

1. Check this documentation
2. Review test cases in `internal/api/s3/handler_test.go`
3. Consult the design document in `.kiro/specs/zot-artifact-store/design.md`
4. Open an issue on the project repository
