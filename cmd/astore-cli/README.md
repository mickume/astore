# astore - Zot Artifact Store CLI

A command-line interface for managing artifacts in the Zot Artifact Store.

## Installation

Build from source:

```bash
go build -o astore ./cmd/astore-cli
```

Or use the Makefile:

```bash
make build-cli
```

## Quick Start

### 1. Initialize Configuration

```bash
astore config init
```

This will create a configuration file at `~/.astore.yaml` with your server URL and optional authentication token.

### 2. Upload an Artifact

```bash
astore upload myapp.tar.gz releases/myapp-1.0.0.tar.gz
```

### 3. List Artifacts

```bash
astore list releases
```

### 4. Download an Artifact

```bash
astore download releases/myapp-1.0.0.tar.gz
```

## Configuration

### Config File

Create `~/.astore.yaml`:

```yaml
# Server URL (required)
server: https://artifacts.example.com

# Authentication token (optional)
token: your-bearer-token-here

# Request timeout in seconds (default: 60)
timeout: 60

# Skip TLS certificate verification (default: false)
insecure: false
```

### Environment Variables

All configuration options can be set via environment variables with the `ASTORE_` prefix:

```bash
export ASTORE_SERVER=https://artifacts.example.com
export ASTORE_TOKEN=your-token
export ASTORE_TIMEOUT=120
export ASTORE_INSECURE=false
```

### Command-Line Flags

Flags override both config file and environment variables:

```bash
astore --server https://artifacts.example.com --token my-token list releases
```

## Commands

### config

Manage CLI configuration.

```bash
# Initialize configuration
astore config init

# Show current configuration
astore config show
```

### upload

Upload an artifact to the store.

```bash
# Basic upload
astore upload local-file.tar.gz bucket/remote-file.tar.gz

# Upload with metadata
astore upload --metadata version=1.0.0 --metadata env=prod app.tar.gz releases/app-1.0.0.tar.gz

# Upload with content type
astore upload --content-type application/gzip app.tar.gz releases/app-1.0.0.tar.gz

# Verbose mode with progress
astore upload -v app.tar.gz releases/app-1.0.0.tar.gz
```

### download

Download an artifact from the store.

```bash
# Download to current directory
astore download releases/app-1.0.0.tar.gz

# Download to specific location
astore download releases/app-1.0.0.tar.gz ./app.tar.gz

# Download with output flag
astore download --output ./app.tar.gz releases/app-1.0.0.tar.gz

# Download specific byte range
astore download --range bytes=0-1023 releases/app-1.0.0.tar.gz

# Verbose mode with progress
astore download -v releases/app-1.0.0.tar.gz
```

### list

List buckets or objects.

```bash
# List all buckets
astore list --buckets

# List objects in a bucket
astore list releases

# List with long format (shows size and date)
astore list --long releases

# List with prefix filter
astore list releases --prefix app/

# Limit number of results
astore list releases --max-keys 100
```

### info

Get detailed information about an artifact.

```bash
# Get artifact info
astore info releases/app-1.0.0.tar.gz
```

Output:
```
Artifact: releases/app-1.0.0.tar.gz
Size:         45.3 MB (47500000 bytes)
Content-Type: application/gzip
ETag:         "abc123..."
Last Modified: 2024-01-15 10:30:00 UTC

Metadata:
  version: 1.0.0
  env: production
```

### delete

Delete an artifact or bucket.

```bash
# Delete an artifact (with confirmation)
astore delete releases/app-1.0.0.tar.gz

# Delete without confirmation
astore delete --force releases/app-1.0.0.tar.gz

# Delete a bucket
astore delete --bucket releases

# Delete bucket without confirmation
astore delete --force --bucket releases
```

## Global Flags

All commands support these global flags:

- `--server <url>` - Artifact store server URL
- `--token <token>` - Authentication bearer token
- `--timeout <seconds>` - Request timeout (default: 60)
- `--insecure` - Skip TLS certificate verification
- `--config <file>` - Custom config file path
- `-v, --verbose` - Enable verbose output with progress tracking
- `--version` - Show version information
- `-h, --help` - Show help for any command

## Examples

### CI/CD Pipeline

```bash
#!/bin/bash
# Build and upload artifact in CI

# Set server URL from environment
export ASTORE_SERVER="https://artifacts.example.com"
export ASTORE_TOKEN="${CI_ARTIFACT_TOKEN}"

# Build application
make build

# Upload with build metadata
astore upload \
  --metadata build-id="${CI_BUILD_ID}" \
  --metadata commit="${CI_COMMIT_SHA}" \
  --metadata branch="${CI_BRANCH}" \
  dist/app-${VERSION}.tar.gz \
  releases/app-${VERSION}.tar.gz

echo "Artifact uploaded successfully"
```

### Download and Extract

```bash
# Download and extract artifact
astore download releases/app-1.0.0.tar.gz -o app.tar.gz
tar -xzf app.tar.gz
./app/install.sh
```

### List Recent Artifacts

```bash
# List artifacts with details
astore list --long releases

# Filter by prefix
astore list releases --prefix app/v1.
```

### Batch Operations

```bash
# Upload multiple files
for file in dist/*.tar.gz; do
  astore upload "$file" "releases/$(basename $file)"
done

# Download multiple artifacts
for artifact in $(astore list releases | grep "app-"); do
  astore download "releases/$artifact"
done
```

## Error Handling

The CLI provides clear error messages and appropriate exit codes:

- Exit code 0: Success
- Exit code 1: Error occurred

```bash
# Check upload success in scripts
if astore upload app.tar.gz releases/app.tar.gz; then
  echo "Upload successful"
else
  echo "Upload failed"
  exit 1
fi
```

## Debugging

Enable verbose mode to see detailed progress and debug information:

```bash
astore -v upload app.tar.gz releases/app.tar.gz
```

Output:
```
Using config file: /home/user/.astore.yaml
Uploading app.tar.gz to releases/app.tar.gz (45.3 MB)
Upload progress: 100.0% (45.3 MB / 45.3 MB)
✓ Uploaded app.tar.gz to releases/app.tar.gz
```

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
astore completion bash > /etc/bash_completion.d/astore

# Zsh
astore completion zsh > "${fpath[1]}/_astore"

# Fish
astore completion fish > ~/.config/fish/completions/astore.fish

# PowerShell
astore completion powershell > astore.ps1
```

## Security

### Token Storage

Store your authentication token securely:

1. Use environment variables (recommended for CI/CD)
2. Use config file with restricted permissions (chmod 600)
3. Use command-line flags (least secure, visible in process list)

### TLS Verification

Always verify TLS certificates in production. Only use `--insecure` for local testing:

```bash
# Development (local testing only)
astore --insecure --server http://localhost:8080 list releases

# Production (always verify TLS)
astore --server https://artifacts.example.com list releases
```

## Version

```bash
astore --version
```

Output:
```
astore version 1.0.0
```

## See Also

- [Phase 10 Documentation](../../docs/PHASE10_COMPLETE.md)
- [Go Client SDK](../../pkg/client/)
- [Zot Artifact Store Documentation](../../docs/)
