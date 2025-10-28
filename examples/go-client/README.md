# Go Client SDK Example

This example demonstrates how to use the Zot Artifact Store Go SDK.

## Prerequisites

- Go 1.21 or higher
- Running Zot Artifact Store instance

## Running the Example

1. Set environment variables (optional):
   ```bash
   export ARTIFACT_STORE_URL="http://localhost:8080"
   export ARTIFACT_STORE_TOKEN="your-bearer-token"
   ```

2. Run the example:
   ```bash
   go run main.go
   ```

## What This Example Demonstrates

- Creating a bucket
- Uploading artifacts with metadata and progress tracking
- Getting object metadata
- Listing objects in a bucket
- Downloading artifacts with progress tracking
- Adding build attestations
- Retrieving attestations

## Expected Output

```
Creating bucket 'examples'...

Uploading artifact...
Upload progress: 100.0%
Upload complete!

Getting object metadata...
  Size: 33 bytes
  Content-Type: text/plain
  ETag: ...
  Description: Example artifact
  Version: 1.0.0

Listing objects in 'examples' bucket...
  - example-artifact.txt (33 bytes)

Downloading artifact...
Download progress: 33 bytes
Downloaded content: This is example artifact content

Adding build attestation...
Attestation added: att-...

Getting attestations...
  Type: build, ID: att-...
  Data: map[buildId:example-001 duration:1m30s status:success testsFailed:0 testsPassed:10]

✅ Example completed successfully!
```

## More Examples

See the [Phase 7 Complete Documentation](../../docs/PHASE7_COMPLETE.md) for more usage examples including:
- Multipart uploads for large files
- Signing and verifying artifacts
- SBOM attachment
- Error handling
- CI/CD integration
