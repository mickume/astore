# Zot Artifact Store Python Client

A Python client library for interacting with the Zot Artifact Store, providing support for artifact management, supply chain security, and RBAC.

## Installation

```bash
pip install astore-client
```

Or install from source:

```bash
cd pkg/client-python
pip install -e .
```

## Quick Start

```python
from astore_client import Client, Config

# Create client
config = Config(
    base_url="https://artifacts.example.com",
    token="your-bearer-token"
)
client = Client(config)

# Upload an artifact
with open("myapp.tar.gz", "rb") as f:
    client.upload(
        "releases",
        "myapp-1.0.0.tar.gz",
        f,
        os.path.getsize("myapp.tar.gz"),
        content_type="application/gzip",
        metadata={"version": "1.0.0", "env": "production"}
    )

# Download an artifact
with open("downloaded.tar.gz", "wb") as f:
    client.download("releases", "myapp-1.0.0.tar.gz", f)

# List artifacts
result = client.list_objects("releases", prefix="myapp/")
for obj in result.objects:
    print(f"{obj.key} ({obj.size} bytes)")
```

## Features

- **Artifact Management**: Upload, download, list, and delete artifacts
- **Bucket Operations**: Create, list, and delete buckets
- **Multipart Upload**: Support for large file uploads
- **Supply Chain Security**: Signing, verification, SBOM, and attestations
- **Progress Tracking**: Monitor upload/download progress
- **Custom Metadata**: Attach custom metadata to artifacts
- **Range Requests**: Download specific byte ranges
- **Authentication**: Bearer token authentication
- **Error Handling**: Comprehensive exception hierarchy

## Configuration

Create a client with configuration:

```python
from astore_client import Client, Config

config = Config(
    base_url="https://artifacts.example.com",
    token="your-token",                    # Optional
    timeout=60,                            # Request timeout in seconds
    insecure_skip_verify=False,            # Skip TLS verification (testing only)
    user_agent="my-app/1.0"               # Custom User-Agent
)

client = Client(config)
```

### Environment Variables

You can also use environment variables:

```bash
export ASTORE_SERVER=https://artifacts.example.com
export ASTORE_TOKEN=your-token
```

```python
import os
from astore_client import Client, Config

config = Config(
    base_url=os.getenv("ASTORE_SERVER"),
    token=os.getenv("ASTORE_TOKEN")
)
client = Client(config)
```

## Usage Examples

### Upload with Progress Tracking

```python
import os

def progress_callback(bytes_transferred):
    print(f"Uploaded: {bytes_transferred} bytes")

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

### Download with Range Request

```python
# Download first 1KB
with open("partial.dat", "wb") as f:
    client.download(
        "releases",
        "myapp-1.0.0.tar.gz",
        f,
        byte_range="bytes=0-1023"
    )
```

### Multipart Upload

```python
# For large files (>5MB recommended)
upload = client.initiate_multipart_upload(
    "releases",
    "large-app.tar.gz",
    content_type="application/gzip",
    metadata={"version": "2.0.0"}
)

parts = []
part_size = 5 * 1024 * 1024  # 5MB

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

# Complete upload
client.complete_multipart_upload(
    upload.bucket,
    upload.key,
    upload.upload_id,
    parts
)
```

### Supply Chain Operations

```python
# Sign an artifact
private_key = open("private.pem").read()
signature = client.sign_artifact("releases", "myapp-1.0.0.tar.gz", private_key)
print(f"Signed: {signature.id}")

# Verify signatures
public_key = open("public.pem").read()
result = client.verify_signatures(
    "releases",
    "myapp-1.0.0.tar.gz",
    [public_key]
)

if result.valid:
    print("✓ Signature verification passed")
else:
    print(f"✗ Verification failed: {result.message}")

# Attach SBOM
sbom_content = open("sbom.json").read()
sbom = client.attach_sbom(
    "releases",
    "myapp-1.0.0.tar.gz",
    "spdx",
    sbom_content
)

# Add build attestation
attestation = client.add_attestation(
    "releases",
    "myapp-1.0.0.tar.gz",
    "build",
    {
        "buildId": "12345",
        "status": "success",
        "tests": 142,
        "coverage": "85.3%"
    }
)
```

## Error Handling

The client provides a comprehensive exception hierarchy:

```python
from astore_client import (
    NotFoundError,
    UnauthorizedError,
    ForbiddenError,
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
except ArtifactStoreError as e:
    print(f"Error: {e.message} (status: {e.status_code})")
```

## Testing

Run tests:

```bash
cd pkg/client-python
pytest
```

Run tests with coverage:

```bash
pytest --cov=astore_client --cov-report=html
```

## API Reference

### Client Configuration

- `Config(base_url, token=None, timeout=60, insecure_skip_verify=False, user_agent="astore-python/1.0.0")`

### Bucket Operations

- `client.create_bucket(bucket)` - Create a bucket
- `client.delete_bucket(bucket)` - Delete a bucket
- `client.list_buckets()` - List all buckets

### Object Operations

- `client.upload(bucket, key, data, size, content_type, metadata, progress_callback)` - Upload artifact
- `client.download(bucket, key, writer, byte_range, progress_callback)` - Download artifact
- `client.get_object_metadata(bucket, key)` - Get artifact metadata
- `client.delete_object(bucket, key)` - Delete artifact
- `client.list_objects(bucket, prefix, max_keys)` - List artifacts
- `client.copy_object(source_bucket, source_key, dest_bucket, dest_key)` - Copy artifact

### Multipart Upload

- `client.initiate_multipart_upload(bucket, key, content_type, metadata)` - Start multipart upload
- `client.upload_part(bucket, key, upload_id, part_number, data, size)` - Upload part
- `client.complete_multipart_upload(bucket, key, upload_id, parts)` - Complete upload
- `client.abort_multipart_upload(bucket, key, upload_id)` - Abort upload

### Supply Chain Operations

- `client.sign_artifact(bucket, key, private_key)` - Sign artifact
- `client.get_signatures(bucket, key)` - Get signatures
- `client.verify_signatures(bucket, key, public_keys)` - Verify signatures
- `client.attach_sbom(bucket, key, format, content)` - Attach SBOM
- `client.get_sbom(bucket, key)` - Get SBOM
- `client.add_attestation(bucket, key, type, data)` - Add attestation
- `client.get_attestations(bucket, key)` - Get attestations

## Requirements

- Python 3.8+
- requests >= 2.25.0
- urllib3 >= 1.26.0

## License

[To be determined]

## See Also

- [Zot Artifact Store Documentation](../../docs/)
- [Go Client SDK](../client/)
- [CLI Tool](../../cmd/astore-cli/)
