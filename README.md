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

✅ **Phase 1 Complete: Foundation** - Extension framework, testing infrastructure, and deployment setup ready
🚧 **Phase 2 In Progress: Core S3 API** - S3-compatible API implementation

### Completed Features (Phase 1)
- ✅ Go project structure with Zot v1.4.3 integration
- ✅ Extension framework for modular features
- ✅ Four core extensions (stubs): S3 API, RBAC, Supply Chain, Metrics
- ✅ Testing infrastructure with TDD patterns
- ✅ Containerfile with Red Hat UBI base images
- ✅ Podman build scripts and development tools
- ✅ ZotArtifactStore CRD for Kubernetes operator
- ✅ Comprehensive project documentation

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

- [Product Requirements](docs/prd.md)
- [Detailed Requirements](.kiro/specs/zot-artifact-store/requirements.md)
- [Design Document](.kiro/specs/zot-artifact-store/design.md)
- [Implementation Tasks](.kiro/specs/zot-artifact-store/tasks.md)

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
