# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### What's Implemented
- Core S3-compatible API (13 endpoints, 17/17 tests passing)
- RBAC with Keycloak integration (7/7 tests passing)
- Supply chain security (signing, SBOM, attestations, 11/11 tests passing)
- Multi-cloud storage backends (16/16 tests passing)
- Client SDKs for Go, Python, JavaScript (all tests passing)
- CLI tool with full command set
- Observability (Prometheus, OpenTelemetry, health checks)

### What's Missing
- Kubernetes operator Go controller implementation (only CRD YAML exists)
- Full integration testing and production validation
- Some extension packages have build failures
- Complete OpenShift deployment workflow

## [0.1.0-dev] - 2025-10-28

**Status:** Core functionality implemented, not production-ready

### Implemented Features

#### Added
- Initial project structure with Go modules
- Zot v1.4.3 integration as base registry
- Extension framework for modular architecture
  - Extension interface and registry
  - S3 API extension (stub)
  - RBAC extension (stub)
  - Supply chain security extension (stub)
  - Enhanced metrics extension (stub)
- Testing infrastructure
  - TDD test helpers and utilities
  - Mock implementations for storage
  - Given-When-Then test patterns
  - Integration and E2E test structure
- Container deployment support
  - Containerfile with Red Hat UBI 9 base images
  - Multi-stage build for optimized images
  - Podman build scripts
  - Container compose configuration
- Development tooling
  - Makefile with common tasks
  - Development setup scripts
  - Build automation
- Kubernetes/OpenShift support
  - ZotArtifactStore Custom Resource Definition (CRD)
  - Sample CR configurations (minimal and full)
  - Operator documentation and structure
- Documentation
  - Comprehensive README
  - Contributing guidelines
  - Getting started guide
  - Operator deployment guide

#### Technical Details
- **Go Version**: 1.25.3
- **Zot Version**: v1.4.3
- **Base Images**: Red Hat UBI 9 (ubi9/go-toolset, ubi9/ubi-minimal)
- **Build**: Static binary with CGO_ENABLED=0
- **Extensions Registered**: 4 (s3api, rbac, supplychain, metrics)
- **Test Coverage**: 27.3% (foundation code)

#### Infrastructure
- Extension registry with setup/shutdown lifecycle
- Graceful shutdown handling
- Structured logging with zerolog
- Configuration management
- Build system with Podman support

### Dependencies
- zotregistry.io/zot v1.4.3
- github.com/gorilla/mux (via Zot)
- github.com/rs/zerolog (via Zot)
- Various supporting libraries (see go.mod)

### Build Artifacts
- Binary: `bin/zot-artifact-store`
- Container Image: `zot-artifact-store:0.1.0-dev`
- CRD: `zotartifactstores.artifacts.zot.io/v1alpha1`

## Future Phases

### Phase 2 - Core S3 API (Next)
- S3-compatible REST API
- Resumable uploads
- Bucket operations
- Artifact metadata management

### Phase 3 - RBAC
- Keycloak integration
- Bearer token authentication
- Policy engine
- Audit logging

### Phase 4 - Supply Chain Security
- Cosign and Notary signing
- SPDX and CycloneDX SBOM support
- Attestation management
- Signature verification

### Phase 5 - Storage Backends
- Local filesystem implementation
- S3 storage integration
- Azure Blob Storage support
- Google Cloud Storage support

### Phase 6 - Metrics and Observability
- Prometheus metrics
- OpenTelemetry tracing
- Health check endpoints
- OpenShift monitoring integration

### Phase 7-9 - Client Libraries
- Go SDK
- Python SDK
- JavaScript/Node.js SDK

### Phase 10 - CLI Tool
- Cobra-based CLI
- Upload/download commands
- Sign/verify operations
- Configuration management

### Phase 11 - Testing and Reliability
- Comprehensive test coverage (90%+)
- Performance testing
- Chaos testing
- Security testing

### Phase 12 - Operator and Production
- Kubernetes operator implementation
- Operator SDK integration
- Lifecycle management
- Production hardening

## Development Notes

### Known Issues
- Server requires valid storage configuration to start
- Some Zot features disabled (lint extension)
- Operator not yet implemented (Phase 12)

### Breaking Changes
None - initial development version

### Deprecations
None - initial development version

[Unreleased]: https://github.com/candlekeep/zot-artifact-store/compare/v0.1.0-dev...HEAD
[0.1.0-dev]: https://github.com/candlekeep/zot-artifact-store/releases/tag/v0.1.0-dev
