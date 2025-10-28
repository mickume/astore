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

✅ **Phase 1 Complete: Foundation** - Extension framework, testing infrastructure, and deployment setup
✅ **Phase 2 Complete: Core S3 API** - Full S3-compatible API with multipart uploads and resumable downloads
✅ **Phase 3 Complete: RBAC** - Keycloak integration, policy engine, and comprehensive audit logging
🚧 **Phase 4 In Progress: Supply Chain Security** - Artifact signing and verification, SBOM management

**Overall Progress:** 56% complete (32/57 foundation tasks) | [View Detailed Status](docs/IMPLEMENTATION_STATUS.md)

### Completed Features

#### Phase 1: Foundation
- ✅ Go project structure with Zot v1.4.3 integration
- ✅ Extension framework for modular features
- ✅ Four core extensions (stubs): S3 API, RBAC, Supply Chain, Metrics
- ✅ Testing infrastructure with TDD patterns
- ✅ Containerfile with Red Hat UBI base images
- ✅ Podman build scripts and development tools
- ✅ ZotArtifactStore CRD for Kubernetes operator
- ✅ Comprehensive project documentation

#### Phase 2: S3-Compatible API
- ✅ Artifact metadata models with OCI digest integration
- ✅ BoltDB metadata storage layer (buckets, artifacts, multipart uploads)
- ✅ Complete S3 API implementation (13 endpoints)
- ✅ Bucket operations: create, list, delete
- ✅ Object operations: upload, download, metadata, delete, list
- ✅ Multipart upload support for large files
- ✅ Resumable downloads with HTTP range requests (RFC 7233)
- ✅ Custom metadata support with X-Amz-Meta-* headers
- ✅ Filesystem-based storage with atomic operations
- ✅ Comprehensive test coverage (17/17 tests passing)
- ✅ S3 API documentation with client examples

#### Phase 3: RBAC with Keycloak Integration
- ✅ JWT token validation with Keycloak OIDC/OAuth2
- ✅ Policy-based authorization engine (resource + action based)
- ✅ Authentication and authorization HTTP middleware
- ✅ Comprehensive audit logging system
- ✅ Policy management API (create, read, update, delete policies)
- ✅ Audit log query API with filtering
- ✅ Admin role with full access
- ✅ Deny > Allow precedence for policies
- ✅ Wildcard and pattern matching for resources
- ✅ Anonymous access support (configurable for GET operations)
- ✅ Extended BoltDB with policies and audit logs
- ✅ Test coverage (7/7 policy tests passing)

#### Phase 4: Supply Chain Security (Partial)
- ✅ Supply chain models (Signature, SBOM, Attestation)
- ✅ Cryptographic signing and verification (RSA-SHA256)
- 🚧 SBOM storage and retrieval (pending)
- 🚧 Attestation management (pending)
- 🚧 Integration with S3 API workflow (pending)

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
