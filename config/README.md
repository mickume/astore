# Zot Artifact Store Configuration

This directory contains example configuration files for different deployment scenarios.

## Configuration Files

### 1. `config.yaml.example`
**Complete reference configuration with all options documented**

This is the master configuration reference showing:
- All available configuration options
- Detailed comments explaining each setting
- Multiple deployment scenarios
- Cloud storage backend examples (S3, GCS, Azure)
- Extension configuration details

Use this as a reference when customizing your configuration.

### 2. `config-minimal.yaml`
**Quick start configuration for development**

Minimal configuration for:
- Local development and testing
- Getting started quickly
- Learning the basics

Features:
- Local filesystem storage (`/tmp/zot-artifacts`)
- No authentication (RBAC disabled)
- All other extensions enabled with defaults
- Suitable for development only

**Usage:**
```bash
./bin/zot-artifact-store --config config/config-minimal.yaml
```

### 3. `config-production.yaml`
**Production-ready secure configuration**

Enterprise configuration for:
- Production deployments
- High security environments
- Cloud-native architectures

Features:
- TLS encryption enabled
- RBAC with Keycloak authentication
- Cloud storage backend (S3)
- Strict security policies (no anonymous access)
- Required signatures and SBOMs
- Full observability (Prometheus, tracing, health checks)
- Comprehensive audit logging

**Usage:**
```bash
# Set environment variables for secrets
export KEYCLOAK_CLIENT_SECRET=your-secret

# Run with production config
./bin/zot-artifact-store --config config/config-production.yaml
```

## Configuration Structure

The configuration is organized into these main sections:

### 1. HTTP Server (`http`)
- Listen address and port
- TLS/HTTPS configuration
- CORS settings
- Authentication (optional)

### 2. Storage (`storage`)
- Local filesystem or cloud backend
- Deduplication and garbage collection
- Storage drivers (S3, GCS, Azure)

### 3. Logging (`log`)
- Log levels
- Output files
- Audit logging

### 4. Extensions (`extensions`)
Four extensions provide additional functionality:

#### a. S3 API Extension (`s3api`)
S3-compatible REST API for binary artifact storage
- Bucket and object operations
- Multipart uploads
- Pre-signed URLs
- Custom metadata

#### b. RBAC Extension (`rbac`)
Authentication and authorization
- Keycloak integration (JWT tokens)
- Policy-based access control
- Audit logging
- Anonymous access control

#### c. Supply Chain Extension (`supplychain`)
Security and compliance features
- Artifact signing (RSA, Cosign, Notary)
- SBOM support (SPDX, CycloneDX)
- Build attestations
- Provenance tracking

#### d. Metrics Extension (`metrics`)
Observability and monitoring
- Prometheus metrics
- OpenTelemetry tracing
- Health check endpoints
- Performance monitoring

## Creating Your Configuration

### Step 1: Choose a Starting Point

**For development:**
```bash
cp config-minimal.yaml config.yaml
```

**For production:**
```bash
cp config-production.yaml config.yaml
```

**For custom configuration:**
```bash
cp config.yaml.example config.yaml
```

### Step 2: Customize Settings

Edit `config.yaml` to match your environment:

1. **Update paths:**
   - `storage.rootDirectory` - Where to store artifacts
   - `log.output` - Where to write logs
   - TLS certificate paths (if using HTTPS)

2. **Configure storage backend:**
   - Use local filesystem for development
   - Use S3/GCS/Azure for production

3. **Enable/disable extensions:**
   - Set `enabled: true/false` for each extension
   - Configure extension-specific settings

4. **Set authentication:**
   - Configure RBAC with Keycloak for production
   - Or use htpasswd for simple auth
   - Or disable for development

5. **Configure observability:**
   - Enable Prometheus metrics
   - Set up OpenTelemetry tracing endpoint
   - Configure health check paths

### Step 3: Secure Secrets

**Never commit secrets to version control!**

Use environment variables for sensitive data:
```yaml
keycloak:
  clientSecret: ${KEYCLOAK_CLIENT_SECRET}

storageDriver:
  accesskey: ${AWS_ACCESS_KEY_ID}
  secretkey: ${AWS_SECRET_ACCESS_KEY}
```

## Configuration Validation

To validate your configuration:

```bash
# Test configuration (dry run)
./bin/zot-artifact-store --config config.yaml --dry-run

# Start with verbose logging to see configuration
./bin/zot-artifact-store --config config.yaml --log-level debug
```

## Environment Variables

You can override configuration with environment variables:

```bash
# Server settings
export ZOT_HTTP_ADDRESS=0.0.0.0
export ZOT_HTTP_PORT=8080

# Storage
export ZOT_STORAGE_ROOT=/data/artifacts

# Keycloak
export KEYCLOAK_URL=https://keycloak.example.com
export KEYCLOAK_REALM=production
export KEYCLOAK_CLIENT_SECRET=secret

# Cloud storage
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=AKIAxxxx
export AWS_SECRET_ACCESS_KEY=secret
export S3_BUCKET=artifacts
```

## Deployment Scenarios

### Local Development
```yaml
storage:
  rootDirectory: /tmp/zot-artifacts
extensions:
  rbac:
    enabled: false
```

### Docker Container
```yaml
storage:
  rootDirectory: /zot/data
log:
  output: /zot/logs/zot-artifact-store.log
```

### Kubernetes/OpenShift
```yaml
storage:
  storageDriver:
    name: s3
    region: us-east-1
    bucket: k8s-artifacts
extensions:
  metrics:
    enabled: true
    health:
      readinessPath: /health/ready
      livenessPath: /health/live
```

### High Availability Setup
```yaml
storage:
  storageDriver:
    name: s3  # Shared storage
extensions:
  rbac:
    enabled: true
  metrics:
    enabled: true
    tracing:
      enabled: true
```

## Troubleshooting

### Configuration Not Loading
- Check file permissions (should be readable by the process)
- Validate YAML syntax (use a YAML linter)
- Check logs for parsing errors

### Storage Errors
- Verify `rootDirectory` exists and is writable
- For cloud storage, verify credentials and permissions
- Check network connectivity to cloud endpoints

### Extension Errors
- Check extension is enabled in config
- Verify metadata database path is writable
- Check extension-specific requirements (e.g., Keycloak URL)

### Authentication Issues
- Verify Keycloak URL is accessible
- Check realm and client ID match
- Ensure client secret is correct
- Review audit logs for auth failures

## Best Practices

1. **Security:**
   - Always use TLS in production
   - Enable RBAC for access control
   - Require signatures and SBOMs for compliance
   - Enable audit logging
   - Use secrets management (not plaintext passwords)

2. **Reliability:**
   - Use cloud storage for high availability
   - Enable health checks for Kubernetes
   - Configure appropriate timeouts
   - Set up monitoring and alerting

3. **Performance:**
   - Enable storage deduplication
   - Configure garbage collection
   - Tune upload size limits
   - Use multipart uploads for large files

4. **Observability:**
   - Enable Prometheus metrics
   - Configure distributed tracing
   - Set up centralized logging
   - Monitor health check endpoints

## Further Reading

- [Zot Configuration Documentation](https://zotregistry.io/v2.0.0/admin-guide/admin-configuration/)
- [S3 API Reference](../docs/S3_API.md)
- [RBAC Configuration Guide](../docs/PHASE3_RBAC.md)
- [Supply Chain Security](../docs/PHASE4_COMPLETE.md)
- [Metrics & Observability](../docs/PHASE6_COMPLETE.md)
