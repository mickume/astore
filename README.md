# Zot Artifact Store

An extension of the Zot OCI registry for storing binary artifacts with enterprise-grade supply chain security features.

## Overview

Zot Artifact Store extends the Zot OCI registry to provide:

- **S3-Compatible API** for binary artifact storage
- **Enterprise RBAC** with Keycloak integration
- **Supply Chain Security** with signing, SBOMs, and attestations
- **Multi-cloud Storage** supporting local filesystem, S3, Azure Blob, and Google Cloud Storage
- **Client Libraries** for Go, Python, and JavaScript
- **OpenShift Native** deployment with operator support

## Project Status

**Last Updated:** 2025-10-29

### Implementation Summary

This project has significant functionality implemented but is **not production-ready**. Key areas completed:

✅ **Core Functionality Implemented:**
- S3-compatible API with 13 endpoints (17/17 tests passing)
- RBAC with Keycloak JWT auth and policy engine (7/7 tests passing)
- Supply chain security: signing, SBOM, attestations (11/11 tests passing)
- Multi-cloud storage backends: Filesystem, S3, Azure, GCS (16/16 tests passing)
- Client SDKs: Go (43 tests), Python (38 tests), JavaScript (32 tests)
- CLI tool with full command set (3/3 tests passing)
- Metrics: Prometheus, OpenTelemetry, health checks (14/14 tests passing)

🟡 **Partial or Design-Only:**
- Kubernetes Operator (CRD defined, Go controller NOT implemented)
- Full OpenShift integration (designed but not fully tested)
- Extension integration (some build failures in extension packages)

❌ **Not Implemented:**
- Production-grade operator with reconciliation logic
- Complete end-to-end integration testing
- Full CI/CD pipeline validation

**Estimated Completion:** ~75% of core functionality, ~50% of production readiness

### Implemented Features

#### Core S3-Compatible API ✅
- Complete S3 API with 13 endpoints (17/17 tests passing)
- Bucket operations: create, list, delete
- Object operations: upload, download, metadata, delete, list
- Multipart upload support for large files
- Resumable downloads with HTTP range requests (RFC 7233)
- Custom metadata with X-Amz-Meta-* headers
- BoltDB metadata storage layer
- SHA256 integrity verification

#### RBAC & Authentication ✅
- JWT token validation with Keycloak OIDC/OAuth2
- Policy-based authorization engine (7/7 tests passing)
- Policy management API (CRUD operations)
- Comprehensive audit logging
- Wildcard and pattern matching for resources
- Anonymous access support (configurable)

#### Supply Chain Security ✅
- Cryptographic signing and verification (RSA-2048/SHA-256)
- SBOM support (SPDX, CycloneDX formats)
- Attestations (build, test, scan, provenance)
- Complete API with 8 endpoints (11/11 tests passing)

#### Storage Backends ✅
- Filesystem backend with atomic operations (16/16 tests passing)
- S3 backend (AWS, MinIO compatible)
- Azure Blob Storage backend
- Google Cloud Storage backend
- Retry mechanisms and circuit breakers

#### Client Libraries ✅
- **Go SDK:** Complete implementation (43/43 tests passing)
- **Python SDK:** Full-featured with type hints (38/38 tests passing)
- **JavaScript/TypeScript SDK:** Browser + Node.js (32/32 tests passing)
- **CLI Tool:** Cobra-based with all commands (3/3 tests passing)

#### Observability ✅
- Prometheus metrics (13 metrics, 14/14 tests passing)
- OpenTelemetry distributed tracing
- Health check endpoints (/health, /health/ready, /health/live)

### Known Limitations

#### Kubernetes Operator ❌
- CRD definitions exist in YAML
- Go controller implementation is **NOT implemented** (controllers/ directory is empty)
- Operator-based deployment is design-only

#### Integration Issues 🟡
- Some extension packages have build failures
- Full end-to-end integration not fully tested
- OpenShift-specific features designed but not production-validated

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Client Layer (CLI, SDKs, S3 Tools)                     │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  API Gateway (HTTP Router, Auth, RBAC, Metrics)         │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  Zot Core + Custom Extensions                           │
│  ├── S3 API Extension                                   │
│  ├── RBAC Extension (Keycloak)                          │
│  ├── Supply Chain Security Extension                    │
│  └── Enhanced Metrics Extension                         │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  Storage Backends (Local FS, S3, Azure, GCP)            │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- Podman or Docker
- Make

### Building

```bash
make build
```

### Running

```bash
./bin/zot-artifact-store --config config/config.yaml
```

### Using the S3 API

```bash
# Create a bucket
curl -X PUT http://localhost:8080/s3/artifacts

# Upload an artifact
curl -X PUT \
  -H "Content-Type: application/gzip" \
  -H "X-Amz-Meta-Version: 1.0.0" \
  --data-binary @myapp-1.0.0.tar.gz \
  http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz

# Download an artifact
curl http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz -o myapp.tar.gz

# List artifacts in bucket
curl http://localhost:8080/s3/artifacts

# Delete an artifact
curl -X DELETE http://localhost:8080/s3/artifacts/myapp-1.0.0.tar.gz
```

See [S3 API Documentation](docs/S3_API.md) for complete API reference.

## Development

### Project Structure

```
.
├── cmd/                    # Main applications
│   └── zot-artifact-store/ # Main server application
├── pkg/                    # Public libraries
│   └── client/             # Client SDKs (Go, Python, JS)
├── internal/               # Internal packages
│   ├── extensions/         # Zot extensions
│   │   ├── s3api/         # S3-compatible API
│   │   ├── rbac/          # RBAC with Keycloak
│   │   ├── supplychain/   # Signing, SBOM, attestations
│   │   └── metrics/       # Enhanced metrics
│   ├── storage/           # Storage backend integrations
│   ├── auth/              # Authentication
│   ├── api/               # API handlers
│   └── models/            # Data models
├── config/                # Configuration files
├── test/                  # Test suites
├── deployments/           # Deployment configurations
│   ├── container/         # Container deployment
│   └── operator/          # Kubernetes operator
├── api/                   # OpenAPI specifications
└── scripts/               # Build and utility scripts
```

### Testing

```bash
# Run all tests
make test

# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Check coverage
make coverage
```

## Documentation

### User Documentation
- [Getting Started](docs/GETTING_STARTED.md)
- [S3 API Reference](docs/S3_API.md) - Complete S3-compatible API documentation with examples

### Planning Documentation
- [Product Requirements](docs/prd.md)
- [Detailed Requirements](.kiro/specs/zot-artifact-store/requirements.md)
- [Design Document](.kiro/specs/zot-artifact-store/design.md)
- [Implementation Tasks](.kiro/specs/zot-artifact-store/tasks.md)

### Implementation Status
- [Phase 1: Foundation](docs/PHASE1_COMPLETE.md) - Extension framework and infrastructure
- [Phase 2: S3 API](docs/PHASE2_COMPLETE.md) - S3-compatible API implementation

## Deployment

### Container Deployment (Development)

```bash
podman build -t zot-artifact-store:latest -f deployments/container/Containerfile .
podman run -p 8080:8080 -v $(pwd)/config:/config zot-artifact-store:latest
```

### OpenShift Deployment (Production)

```bash
# Deploy operator
kubectl apply -f deployments/operator/config/crd/
kubectl apply -f deployments/operator/deploy/

# Create instance
kubectl apply -f deployments/operator/config/samples/
```

## Contributing

This project follows Test-Driven Development (TDD) practices:

1. Write tests first
2. Implement features to pass tests
3. Maintain 90% code coverage
4. Use AI-friendly test patterns (Given-When-Then)

## License

[To be determined]

## Support

For questions and issues, please check the project documentation or open an issue.
