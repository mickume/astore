# Phase 8: Python Client SDK - COMPLETE ✅

## Overview

Phase 8 implements a comprehensive Python SDK for the Zot Artifact Store, providing a Pythonic, easy-to-use client library for artifact management with support for authentication, progress tracking, multipart uploads, and supply chain security operations.

**Completion Date:** 2025-10-28

## Implementation Summary

### Components Delivered

1. **Client Foundation** (`astore_client/client.py`)
2. **Core Operations** (`astore_client/operations.py`)
3. **Supply Chain Integration** (`astore_client/supplychain.py`)
4. **Data Models** (`astore_client/models.py`)
5. **Exception Hierarchy** (`astore_client/exceptions.py`)
6. **Comprehensive Tests** (38 tests passing)
7. **Package Setup** (`setup.py`, `requirements.txt`)
8. **Documentation** (`README.md`)

## Features

### 1. Client Foundation

**Client Structure:**

```python
from astore_client import Client, Config

# Create configuration
config = Config(
    base_url="https://artifacts.example.com",
    token="your-bearer-token",
    timeout=60,                        # Request timeout in seconds
    insecure_skip_verify=False,        # Skip TLS verification (testing only)
    user_agent="my-app/1.0"           # Custom User-Agent
)

# Create client
client = Client(config)

# Update token dynamically
client.set_token("new-token")
```

**Features:**
- Configurable HTTP session with custom timeouts
- Automatic bearer token authentication
- Custom User-Agent support
- TLS configuration (including insecure mode for testing)
- Automatic error handling with typed exceptions
- Pythonic API design

### 2. Core Artifact Operations

**Upload:**

```python
import io

# Simple upload
data = io.BytesIO(b"artifact data")
client.upload("releases", "app-1.0.0.tar.gz", data, len(data))

# Upload with metadata and progress tracking
def progress_callback(bytes_transferred):
    print(f"Uploaded: {bytes_transferred} bytes")

with open("app.tar.gz", "rb") as f:
    client.upload(
        "releases",
        "app-1.0.0.tar.gz",
        f,
        os.path.getsize("app.tar.gz"),
        content_type="application/gzip",
        metadata={
            "version": "1.0.0",
            "author": "ci-system",
        },
        progress_callback=progress_callback
    )
```

**Download:**

```python
# Simple download
with open("downloaded.tar.gz", "wb") as f:
    client.download("releases", "app-1.0.0.tar.gz", f)

# Download with range request and progress tracking
def progress_callback(bytes_transferred):
    print(f"Downloaded: {bytes_transferred} bytes")

with open("partial.dat", "wb") as f:
    client.download(
        "releases",
        "app-1.0.0.tar.gz",
        f,
        byte_range="bytes=0-1023",  # First 1KB
        progress_callback=progress_callback
    )
```

**List Objects:**

```python
# List all objects in bucket
result = client.list_objects("releases")
for obj in result.objects:
    print(f"{obj.key} ({obj.size} bytes)")

# List with prefix filter
result = client.list_objects("releases", prefix="app/", max_keys=100)
```

**Object Metadata:**

```python
obj = client.get_object_metadata("releases", "app-1.0.0.tar.gz")
print(f"Size: {obj.size} bytes")
print(f"Type: {obj.content_type}")
print(f"ETag: {obj.etag}")
print(f"Version: {obj.metadata.get('version')}")
```

**Delete Object:**

```python
client.delete_object("releases", "app-1.0.0.tar.gz")
```

**Copy Object:**

```python
client.copy_object(
    "releases", "app-1.0.0.tar.gz",
    "archive", "app-1.0.0-backup.tar.gz"
)
```

### 3. Bucket Management

**Create Bucket:**

```python
client.create_bucket("my-new-bucket")
```

**List Buckets:**

```python
result = client.list_buckets()
for bucket in result.buckets:
    print(f"{bucket.name} (created: {bucket.creation_date})")
```

**Delete Bucket:**

```python
client.delete_bucket("old-bucket")
```

### 4. Multipart Upload

For large files (>5MB recommended):

```python
from astore_client.models import CompletedPart

# Initiate multipart upload
upload = client.initiate_multipart_upload(
    "releases",
    "large-app.tar.gz",
    content_type="application/gzip",
    metadata={"size": "500MB"}
)

# Upload parts
parts = []
part_size = 5 * 1024 * 1024  # 5MB parts

with open("large-app.tar.gz", "rb") as f:
    part_number = 1
    while True:
        data = f.read(part_size)
        if not data:
            break

        etag = client.upload_part(
            upload.bucket,
            upload.key,
            upload.upload_id,
            part_number,
            io.BytesIO(data),
            len(data)
        )

        parts.append(CompletedPart(part_number=part_number, etag=etag))
        part_number += 1

# Complete multipart upload
client.complete_multipart_upload(
    upload.bucket,
    upload.key,
    upload.upload_id,
    parts
)

# Or abort if needed
client.abort_multipart_upload(upload.bucket, upload.key, upload.upload_id)
```

### 5. Supply Chain Operations

**Sign Artifact:**

```python
with open("private.pem") as f:
    private_key = f.read()

signature = client.sign_artifact("releases", "app-1.0.0.tar.gz", private_key)
print(f"Signed with ID: {signature.id}")
print(f"Algorithm: {signature.algorithm}")
```

**Verify Signatures:**

```python
with open("public.pem") as f:
    public_key = f.read()

result = client.verify_signatures(
    "releases",
    "app-1.0.0.tar.gz",
    [public_key]
)

if result.valid:
    print("✓ All signatures valid!")
else:
    print(f"✗ Verification failed: {result.message}")
```

**Get Signatures:**

```python
signatures = client.get_signatures("releases", "app-1.0.0.tar.gz")
for sig in signatures:
    print(f"Signature: {sig.id} (signed by {sig.signed_by})")
```

**Attach SBOM:**

```python
sbom_content = '{"spdxVersion": "SPDX-2.3", "packages": [...]}'

sbom = client.attach_sbom(
    "releases",
    "app-1.0.0.tar.gz",
    "spdx",
    sbom_content
)
print(f"SBOM attached: {sbom.id}")
```

**Get SBOM:**

```python
sbom = client.get_sbom("releases", "app-1.0.0.tar.gz")
print(f"Format: {sbom.format}")
print(f"Content: {sbom.content}")
```

**Add Attestation:**

```python
attestation = client.add_attestation(
    "releases",
    "app-1.0.0.tar.gz",
    "build",
    {
        "buildId": "12345",
        "status": "success",
        "duration": "5m30s",
        "testsPassed": 142,
    }
)
print(f"Attestation added: {attestation.id}")
```

**Get Attestations:**

```python
attestations = client.get_attestations("releases", "app-1.0.0.tar.gz")
for att in attestations:
    print(f"Type: {att.type}, ID: {att.id}")
    print(f"Data: {att.data}")
```

### 6. Error Handling

The SDK provides a comprehensive exception hierarchy:

```python
from astore_client import (
    NotFoundError,
    UnauthorizedError,
    ForbiddenError,
    ConflictError,
    ArtifactStoreError
)

try:
    client.download("releases", "nonexistent.tar.gz", f)
except NotFoundError:
    print("Artifact not found")
except UnauthorizedError:
    print("Authentication failed")
except ForbiddenError:
    print("Permission denied")
except ConflictError:
    print("Resource conflict")
except ArtifactStoreError as e:
    print(f"Error: {e.message} (status: {e.status_code})")
```

**Exception Types:**
- `ArtifactStoreError` - Base exception for all errors
- `BadRequestError` (400) - Invalid request
- `UnauthorizedError` (401) - Authentication required
- `ForbiddenError` (403) - Permission denied
- `NotFoundError` (404) - Resource not found
- `ConflictError` (409) - Resource conflict
- `InternalServerError` (500) - Internal server error
- `ServiceUnavailableError` (503) - Service unavailable

### 7. Progress Tracking

Track upload and download progress:

```python
def progress_callback(bytes_transferred):
    percentage = (bytes_transferred / total_size) * 100
    print(f"\rProgress: {percentage:.1f}% ({bytes_transferred}/{total_size} bytes)", end="")

with open("large-file.tar.gz", "rb") as f:
    size = os.path.getsize("large-file.tar.gz")
    client.upload(
        "releases",
        "large-file.tar.gz",
        f,
        size,
        progress_callback=progress_callback
    )
```

## Testing

### Test Coverage

```
============================= test session starts ==============================
platform darwin -- Python 3.14.0, pytest-8.4.2, pluggy-1.6.0
collected 38 items

tests/test_client.py::TestClientConfiguration::test_create_client_with_valid_config PASSED
tests/test_client.py::TestClientConfiguration::test_create_client_with_missing_base_url PASSED
tests/test_client.py::TestClientConfiguration::test_create_client_with_token PASSED
tests/test_client.py::TestClientConfiguration::test_create_client_with_custom_timeout PASSED
tests/test_client.py::TestClientConfiguration::test_create_client_with_insecure_skip_verify PASSED
tests/test_client.py::TestClientConfiguration::test_create_client_with_custom_user_agent PASSED
tests/test_client.py::TestClientConfiguration::test_base_url_trailing_slash_removed PASSED
tests/test_client.py::TestSetToken::test_update_authentication_token PASSED
tests/test_client.py::TestURLBuilding::test_url_building_with_path PASSED
tests/test_client.py::TestURLBuilding::test_url_building_with_leading_slash PASSED

tests/test_operations.py::TestBucketOperations::test_create_bucket PASSED
tests/test_operations.py::TestBucketOperations::test_delete_bucket PASSED
tests/test_operations.py::TestBucketOperations::test_list_buckets PASSED
tests/test_operations.py::TestObjectOperations::test_upload_object PASSED
tests/test_operations.py::TestObjectOperations::test_upload_with_metadata PASSED
tests/test_operations.py::TestObjectOperations::test_download_object PASSED
tests/test_operations.py::TestObjectOperations::test_download_with_range PASSED
tests/test_operations.py::TestObjectOperations::test_get_object_metadata PASSED
tests/test_operations.py::TestObjectOperations::test_delete_object PASSED
tests/test_operations.py::TestObjectOperations::test_list_objects PASSED
tests/test_operations.py::TestObjectOperations::test_list_objects_with_prefix PASSED
tests/test_operations.py::TestObjectOperations::test_copy_object PASSED
tests/test_operations.py::TestMultipartUpload::test_initiate_multipart_upload PASSED
tests/test_operations.py::TestMultipartUpload::test_upload_part PASSED
tests/test_operations.py::TestMultipartUpload::test_complete_multipart_upload PASSED
tests/test_operations.py::TestMultipartUpload::test_abort_multipart_upload PASSED
tests/test_operations.py::TestErrorHandling::test_404_not_found PASSED
tests/test_operations.py::TestErrorHandling::test_409_conflict PASSED
tests/test_operations.py::TestProgressCallbacks::test_upload_progress_callback PASSED
tests/test_operations.py::TestProgressCallbacks::test_download_progress_callback PASSED

tests/test_supplychain.py::TestSupplyChainOperations::test_sign_artifact PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_get_signatures PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_verify_signatures PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_verify_signatures_failure PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_attach_sbom PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_get_sbom PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_add_attestation PASSED
tests/test_supplychain.py::TestSupplyChainOperations::test_get_attestations PASSED

============================== 38 passed in 0.04s ==============================
```

**Total Tests:** 38/38 passing

### Test Scenarios

- ✅ Client creation and configuration (7 tests)
- ✅ Token management (1 test)
- ✅ URL building (2 tests)
- ✅ Bucket operations (3 tests)
- ✅ Object operations (9 tests)
- ✅ Multipart uploads (4 tests)
- ✅ Error handling (2 tests)
- ✅ Progress callbacks (2 tests)
- ✅ Supply chain operations (8 tests)

## Files Added

### Package Structure (14 files)

```
pkg/client-python/
├── astore_client/
│   ├── __init__.py                 # Package initialization (50 lines)
│   ├── client.py                   # Client foundation (400 lines)
│   ├── operations.py               # Core operations (270 lines)
│   ├── supplychain.py              # Supply chain ops (150 lines)
│   ├── models.py                   # Data models (100 lines)
│   └── exceptions.py               # Exception hierarchy (90 lines)
├── tests/
│   ├── __init__.py                 # Test package init
│   ├── conftest.py                 # Pytest fixtures (25 lines)
│   ├── test_client.py              # Client tests (120 lines)
│   ├── test_operations.py          # Operations tests (350 lines)
│   └── test_supplychain.py         # Supply chain tests (180 lines)
├── setup.py                        # Package setup (60 lines)
├── requirements.txt                # Dependencies
├── requirements-dev.txt            # Development dependencies
└── README.md                       # Package documentation (400 lines)
```

**Total:** ~2,200 lines of production code + tests + documentation

## Usage Examples

### Complete Upload/Download Workflow

```python
import os
from astore_client import Client, Config

# Create client
config = Config(
    base_url="https://artifacts.example.com",
    token=os.getenv("ARTIFACT_STORE_TOKEN")
)
client = Client(config)

# Upload artifact
print("Uploading artifact...")
with open("myapp-1.0.0.tar.gz", "rb") as f:
    size = os.path.getsize("myapp-1.0.0.tar.gz")

    def progress(bytes_transferred):
        pct = (bytes_transferred / size) * 100
        print(f"\rProgress: {pct:.1f}%", end="")

    client.upload(
        "releases",
        "myapp-1.0.0.tar.gz",
        f,
        size,
        content_type="application/gzip",
        metadata={"version": "1.0.0", "commit": "abc123"},
        progress_callback=progress
    )

print("\nUpload complete!")

# Sign the artifact
with open("private.pem") as f:
    private_key = f.read()

sig = client.sign_artifact("releases", "myapp-1.0.0.tar.gz", private_key)
print(f"Artifact signed: {sig.id}")

# Download and verify
with open("downloaded.tar.gz", "wb") as f:
    client.download("releases", "myapp-1.0.0.tar.gz", f)

# Verify signature
with open("public.pem") as f:
    public_key = f.read()

result = client.verify_signatures(
    "releases",
    "myapp-1.0.0.tar.gz",
    [public_key]
)

if result.valid:
    print("✓ Signature verification passed!")
else:
    print(f"✗ Verification failed: {result.message}")
```

### CI/CD Integration

```python
#!/usr/bin/env python3
"""Upload build artifact to artifact store"""

import os
import sys
from astore_client import Client, Config

def upload_build_artifact(build_id, artifact_path):
    """Upload build artifact with metadata"""
    config = Config(
        base_url=os.getenv("ARTIFACT_STORE_URL"),
        token=os.getenv("ARTIFACT_STORE_TOKEN")
    )
    client = Client(config)

    # Upload artifact
    with open(artifact_path, "rb") as f:
        size = os.path.getsize(artifact_path)

        client.upload(
            "builds",
            f"build-{build_id}.tar.gz",
            f,
            size,
            metadata={
                "build-id": build_id,
                "commit": os.getenv("GIT_COMMIT"),
                "branch": os.getenv("GIT_BRANCH"),
            }
        )

    # Add build attestation
    client.add_attestation(
        "builds",
        f"build-{build_id}.tar.gz",
        "build",
        {
            "buildId": build_id,
            "status": "success",
            "testsPassed": 142,
            "testsFailed": 0,
            "coverage": "85.3%",
            "duration": "5m30s",
        }
    )

    print(f"✓ Build artifact uploaded successfully")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: upload.py <build-id> <artifact-path>")
        sys.exit(1)

    upload_build_artifact(sys.argv[1], sys.argv[2])
```

## Installation

### From PyPI (when published):

```bash
pip install astore-client
```

### From source:

```bash
cd pkg/client-python
pip install -e .
```

### Development installation:

```bash
cd pkg/client-python
pip install -e .[dev]
```

## Best Practices

### 1. Use Context Managers

Always use context managers for file operations:

```python
with open("artifact.tar.gz", "rb") as f:
    client.upload("bucket", "key", f, os.path.getsize("artifact.tar.gz"))
```

### 2. Error Handling

Check for specific error types:

```python
from astore_client import NotFoundError, UnauthorizedError

try:
    client.download("bucket", "key", f)
except NotFoundError:
    print("Artifact not found")
except UnauthorizedError:
    print("Authentication failed")
```

### 3. Large File Uploads

Use multipart upload for files >5MB:

```python
if file_size > 5 * 1024 * 1024:  # >5MB
    # Use multipart upload
    upload = client.initiate_multipart_upload(bucket, key)
    # Upload parts...
    client.complete_multipart_upload(bucket, key, upload.upload_id, parts)
else:
    # Regular upload
    client.upload(bucket, key, data, size)
```

### 4. Progress Tracking

Provide user feedback for long operations:

```python
def progress_callback(bytes_transferred):
    print(f"\rUploaded: {bytes_transferred / 1024 / 1024:.1f} MB", end="")

client.upload(bucket, key, data, size, progress_callback=progress_callback)
```

### 5. Environment Variables

Use environment variables for configuration in CI/CD:

```python
import os
from astore_client import Client, Config

config = Config(
    base_url=os.getenv("ASTORE_SERVER"),
    token=os.getenv("ASTORE_TOKEN")
)
client = Client(config)
```

## Integration Benefits

### For Python Applications
- Pythonic API design
- Type hints for IDE support
- Comprehensive exception hierarchy
- Progress tracking built-in
- Context manager support

### For CI/CD
- Easy integration with Python build scripts
- Attestation support for build metadata
- SBOM attachment for compliance
- Signature verification for security

### For Data Science
- Seamless artifact management
- Version tracking with metadata
- Large file support (multipart)
- Integration with Jupyter notebooks

## Known Limitations

1. **In-Memory Buffering**: Parts are buffered in memory during upload
2. **No Async Support**: Synchronous API only (no asyncio)
3. **No Concurrent Upload**: Multipart parts uploaded sequentially
4. **Limited Retry Logic**: No automatic retry on transient failures

## Future Enhancements

### Phase 8.1: Advanced Features

1. **Async Support**
   - Asyncio-based client
   - Concurrent operations
   - Non-blocking I/O

2. **Streaming Operations**
   - Stream-based upload/download
   - Reduced memory footprint
   - Generator-based iteration

3. **Built-in Retry Logic**
   - Exponential backoff
   - Configurable retry policy
   - Automatic retry for transient errors

### Phase 8.2: Performance

1. **Concurrent Uploads**
   - Parallel part uploads
   - Thread pool executor
   - Progress aggregation

2. **Connection Pooling**
   - Optimized HTTP connection reuse
   - Configurable pool size
   - Keep-alive tuning

3. **Caching**
   - Metadata caching
   - ETags for conditional requests
   - Reduced network overhead

## Conclusion

Phase 8 successfully delivers a production-ready Python SDK:

- ✅ **Complete API Coverage**: All S3 and supply chain operations
- ✅ **Pythonic Design**: Idiomatic Python API with type hints
- ✅ **Well-Tested**: 38/38 tests passing
- ✅ **Production Ready**: Error handling, timeouts, authentication
- ✅ **Developer Friendly**: Comprehensive documentation and examples
- ✅ **Supply Chain Support**: Full integration with signing, SBOM, attestations

The Zot Artifact Store Python SDK provides a robust, Pythonic client library for artifact management in Python applications.

---

**Status:** ✅ COMPLETE
**Date:** 2025-10-28
**Tests:** 38/38 passing
**Lines of Code:** ~2,200 (production + tests + docs)
**Python Version:** 3.8+
**Dependencies:** requests >= 2.25.0, urllib3 >= 1.26.0
**Next Phase:** Phase 9 (JavaScript Client SDK)
