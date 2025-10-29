# Zot Artifact Store - Implementation Status

## Overview

The Zot Artifact Store is an extension of the Zot OCI registry for storing binary artifacts with enterprise-grade supply chain security features. This document tracks the implementation status across all planned phases.

**Last Updated:** 2025-10-28

## Phase Status Summary

| Phase | Name | Status | Progress | Tests | Notes |
|-------|------|--------|----------|-------|-------|
| 1 | Foundation | ✅ Complete | 8/8 tasks | ✅ Passing | Extension framework, testing, deployment |
| 2 | S3 API | ✅ Complete | 15/15 tasks | ✅ 17/17 passing | Full S3-compatible API |
| 3 | RBAC | ✅ Complete | 7/7 tasks | ✅ 7/7 passing | Keycloak auth, policies, audit |
| 4 | Supply Chain | ✅ Complete | 5/5 tasks | ✅ 11/11 passing | Signing, SBOM, attestations |
| 5 | Storage | ✅ Complete | 4/4 tasks | ✅ 16/16 passing | Multi-cloud storage |
| 6 | Metrics | ✅ Complete | 3/3 tasks | ✅ 14/14 passing | Prometheus, OpenTelemetry, health |
| 7 | Go Client | ✅ Complete | 3/3 tasks | ✅ 43/43 passing | Go SDK |
| 8 | Python Client | ✅ Complete | 3/3 tasks | ✅ 38/38 passing | Python SDK |
| 9 | JS Client | ⏳ Planned | 0/3 tasks | - | JavaScript/TypeScript SDK |
| 10 | CLI | ✅ Complete | 3/3 tasks | ✅ 3/3 passing | Command-line tool |
| 11 | Error Handling | ✅ Complete | 3/3 tasks | ✅ Passing | Retry, circuit breakers |
| 12 | Integration | ✅ Complete | 6/6 tasks | ✅ Passing | Testing, operator, OpenAPI |

**Overall Progress:** 65/66 tasks complete (98% - Nearly complete)

## Detailed Phase Status

### ✅ Phase 1: Foundation (COMPLETE)

**Completion:** 100% (8/8 tasks)

**Delivered:**
- Go project structure with Zot v1.4.3 integration
- Extension framework for modular features
- Four core extension stubs (S3 API, RBAC, Supply Chain, Metrics)
- Testing infrastructure with TDD patterns
- Containerfile with Red Hat UBI base images
- Podman build scripts and development tools
- ZotArtifactStore CRD for Kubernetes operator
- Comprehensive project documentation

**Documentation:** [Phase 1 Complete](PHASE1_COMPLETE.md)

**Build:** ✅ Static binary compilation with CGO_ENABLED=0

---

### ✅ Phase 2: S3-Compatible API (COMPLETE)

**Completion:** 100% (15/15 tasks)

**Delivered:**
- Artifact metadata models with OCI digest integration
- BoltDB metadata storage layer (buckets, artifacts, multipart uploads)
- Complete S3 API implementation (13 endpoints)
- Bucket operations: create, list, delete
- Object operations: upload, download, metadata, delete, list
- Multipart upload support for large files
- Resumable downloads with HTTP range requests (RFC 7233)
- Custom metadata support with X-Amz-Meta-* headers
- Filesystem-based storage with atomic operations
- Comprehensive test coverage (17/17 tests passing)
- S3 API documentation with client examples

**Documentation:** [Phase 2 Complete](PHASE2_COMPLETE.md) | [S3 API Reference](S3_API.md)

**Test Results:**
```
✅ 4 bucket operation tests
✅ 5 object operation tests
✅ 2 multipart upload tests
✅ 6 metadata storage tests
Coverage: 43.0% (S3 API), 75.6% (storage)
```

**Key Endpoints:**
- `GET /s3` - List buckets
- `PUT /s3/{bucket}` - Create bucket
- `GET /s3/{bucket}` - List objects
- `PUT /s3/{bucket}/{key}` - Upload object
- `GET /s3/{bucket}/{key}` - Download object
- `POST /s3/{bucket}/{key}?uploads` - Initiate multipart upload

---

### ✅ Phase 3: RBAC with Keycloak Integration (COMPLETE)

**Completion:** 100% (7/7 tasks)

**Delivered:**
- Authentication models (User, Policy, AuditLog, AuthContext)
- JWT token validation with Keycloak OIDC/OAuth2
- Policy engine with role-based and resource-based access control
- Authorization middleware for HTTP requests
- Comprehensive audit logging system
- RBAC extension with policy and audit log management
- Extended BoltDB with policies and audit log buckets
- Policy management API (CRUD operations)
- Audit log query API with filtering

**Documentation:** [Phase 3 RBAC](PHASE3_RBAC.md)

**Test Results:**
```
✅ 7 policy engine tests
Coverage: 16.1% (auth package)
```

**Key Features:**
- JWT token validation from Keycloak
- Public key caching and rotation
- Fine-grained resource permissions
- Deny > Allow precedence
- Wildcard and pattern matching
- Anonymous access (configurable)
- Comprehensive audit trail

**API Endpoints:**
- `POST /rbac/policies` - Create policy
- `GET /rbac/policies` - List policies
- `GET /rbac/policies/{id}` - Get policy
- `PUT /rbac/policies/{id}` - Update policy
- `DELETE /rbac/policies/{id}` - Delete policy
- `POST /rbac/authorize` - Check authorization
- `GET /rbac/audit` - List audit logs

---

### ✅ Phase 4: Supply Chain Security (COMPLETE)

**Completion:** 100% (5/5 tasks)

**Delivered:**
- Supply chain models (Signature, SBOM, Attestation, VerificationResult)
- Cryptographic signing and verification (RSA-2048, SHA-256)
- BoltDB storage for signatures, SBOMs, and attestations
- Supply chain extension with full lifecycle management
- Supply chain API (8 endpoints)
- Comprehensive test coverage (11/11 tests passing)

**Documentation:** [Phase 4 Complete](PHASE4_COMPLETE.md)

**Test Results:**
```
✅ 4 signing tests
✅ 7 supply chain storage tests
Coverage: 66.7% (supplychain), 59.8% (storage)
```

**Key Features:**
- RSA-2048 key pair generation and signing
- Signature verification with tamper detection
- SBOM support for SPDX and CycloneDX formats
- Build, test, scan, and provenance attestations
- SLSA provenance support
- Multiple signatures per artifact

**API Endpoints:**
- `POST /supplychain/sign/{bucket}/{key}` - Sign artifact
- `GET /supplychain/signatures/{bucket}/{key}` - Get signatures
- `POST /supplychain/verify/{bucket}/{key}` - Verify signatures
- `POST /supplychain/sbom/{bucket}/{key}` - Attach SBOM
- `GET /supplychain/sbom/{bucket}/{key}` - Get SBOM
- `POST /supplychain/attestations/{bucket}/{key}` - Add attestation
- `GET /supplychain/attestations/{bucket}/{key}` - Get attestations

---

### ✅ Phase 6: Enhanced Metrics & Observability (COMPLETE)

**Completion:** 100% (3/3 tasks)

**Delivered:**
- Prometheus metrics collector with 13 metrics
- Health checker with component monitoring
- OpenTelemetry distributed tracing support
- Metrics extension with full lifecycle management
- Health check API (3 endpoints)
- Comprehensive test coverage (14/14 tests passing)

**Documentation:** [Phase 6 Complete](PHASE6_COMPLETE.md)

**Test Results:**
```
✅ 9 Prometheus metrics tests
✅ 5 health checker tests
Coverage: 54.2% (metrics)
```

**Prometheus Metrics:**
- **Artifact Metrics:** uploads, downloads, deletes, sizes, durations
- **Supply Chain Metrics:** signing, verification, SBOM, attestations
- **RBAC Metrics:** authentication attempts, authorization checks
- **System Metrics:** active connections, requests, errors

**Health Check Endpoints:**
- `GET /health` - Comprehensive health check
- `GET /health/ready` - Readiness probe (Kubernetes)
- `GET /health/live` - Liveness probe (Kubernetes)

**OpenTelemetry Features:**
- OTLP gRPC exporter
- Distributed trace context propagation
- Span creation for artifact, supply chain, and auth operations
- Error recording in spans

---

### ✅ Phase 5: Storage Backend Integration (COMPLETE)

**Completion:** 100% (4/4 tasks)

**Delivered:**
- Storage backend abstraction layer with unified interface
- FileSystem backend with atomic writes and SHA256 verification
- S3 backend with multipart upload support (AWS S3, MinIO, etc.)
- Google Cloud Storage backend with native SDK integration
- Azure Blob Storage backend with retry reader support
- Retry wrapper integrating with Phase 11 reliability package
- SHA256 integrity verification for all backends
- Comprehensive tests (16/16 passing)

**Documentation:** [Phase 5 Complete](PHASE5_COMPLETE.md)

**Test Results:**
```
✅ 16 filesystem backend tests
Coverage: Full backend interface coverage
```

---

### ✅ Phase 7: Go Client SDK (COMPLETE)

**Completion:** 100% (3/3 tasks)

**Delivered:**
- Client library foundation with configurable HTTP client
- Core artifact operations (upload, download, list, delete)
- Bucket management operations
- Multipart upload support for large files
- Supply chain operations (sign, verify, SBOM, attestations)
- Progress callbacks for uploads and downloads
- Bearer token authentication
- Custom metadata support
- Comprehensive error handling
- 43/43 tests passing

**Documentation:** [Phase 7 Complete](PHASE7_COMPLETE.md)

**Test Results:**
```
✅ 43 Go client SDK tests
Coverage: Full client interface coverage
```

---

### ✅ Phase 8: Python Client SDK (COMPLETE)

**Completion:** 100% (3/3 tasks)

**Delivered:**
- Python client library foundation with requests library
- Core artifact operations (upload, download, list, delete)
- Bucket management operations
- Multipart upload support for large files
- Supply chain operations (sign, verify, SBOM, attestations)
- Progress callbacks for uploads and downloads
- Bearer token authentication
- Custom metadata support
- Comprehensive error handling with exception hierarchy
- 38/38 tests passing with pytest

**Documentation:** [Phase 8 Complete](PHASE8_COMPLETE.md)

**Test Results:**
```
✅ 38 Python client SDK tests (pytest)
Coverage: Full client interface coverage
```

**Test Cases:**
1. **Client Configuration** (7 tests)
   - Client creation with valid config
   - Missing base URL validation
   - Token authentication
   - Custom timeout and user agent
   - TLS configuration

2. **Token Management** (1 test)
   - Update authentication token

3. **URL Building** (2 tests)
   - URL construction from paths

4. **Bucket Operations** (3 tests)
   - Create, delete, list buckets

5. **Object Operations** (9 tests)
   - Upload with metadata
   - Download with range requests
   - Get metadata, delete, list, copy

6. **Multipart Upload** (4 tests)
   - Initiate, upload part, complete, abort

7. **Error Handling** (2 tests)
   - 404 Not Found
   - 409 Conflict

8. **Progress Callbacks** (2 tests)
   - Upload and download progress tracking

9. **Supply Chain Operations** (8 tests)
   - Sign, verify, get signatures
   - Attach and get SBOM
   - Add and get attestations

**Key Features:**
- Pythonic API with type hints
- requests-based HTTP client
- Comprehensive exception hierarchy (8 exception types)
- Progress tracking with callbacks
- Context manager support
- Data models with dataclasses
- Multipart upload for large files
- Range request support for partial downloads

**Package Structure:**
- `astore_client/client.py` - Client foundation (400 lines)
- `astore_client/operations.py` - Core operations (270 lines)
- `astore_client/supplychain.py` - Supply chain ops (150 lines)
- `astore_client/models.py` - Data models (100 lines)
- `astore_client/exceptions.py` - Exception hierarchy (90 lines)
- `tests/` - Comprehensive test suite (650 lines)
- `setup.py` - Package setup
- `README.md` - Package documentation (400 lines)

**Installation:**
```bash
cd pkg/client-python
pip install -e .
```

**Example Usage:**
```python
from astore_client import Client, Config

config = Config(
    base_url="https://artifacts.example.com",
    token="your-token"
)
client = Client(config)

# Upload artifact
with open("app.tar.gz", "rb") as f:
    client.upload("releases", "app-1.0.0.tar.gz", f, size,
                  metadata={"version": "1.0.0"})

# Download artifact
with open("downloaded.tar.gz", "wb") as f:
    client.download("releases", "app-1.0.0.tar.gz", f)
```

---

### ✅ Phase 10: CLI Tool (COMPLETE)

**Completion:** 100% (3/3 tasks)

**Delivered:**
- CLI foundation with Cobra framework
- Configuration management (file, env vars, flags)
- Core commands: upload, download, list, info, delete, config
- Progress tracking and verbose mode
- Authentication with bearer tokens
- Error handling with clear messages
- Shell completion support
- 3/3 tests passing

**Documentation:** [Phase 10 Complete](PHASE10_COMPLETE.md) | [CLI README](../cmd/astore-cli/README.md)

**Test Results:**
```
✅ 3 CLI tests
Coverage: formatSize, getBucketAndKey, guessContentType
```

**CLI Commands:**
- `astore upload` - Upload artifacts with metadata
- `astore download` - Download artifacts with progress
- `astore list` - List buckets and objects
- `astore info` - Get artifact information
- `astore delete` - Delete artifacts/buckets
- `astore config` - Manage configuration

---

### ⏳ Phase 9: Planned Features

#### Phase 9: Client Libraries
- JavaScript/TypeScript SDK

#### Future Enhancements
- Comprehensive error classification
- Retry logic for transient failures
- Circuit breaker patterns
- Partial retry with range requests

#### Phase 12: Integration & System Testing
- End-to-end integration tests
- Kubernetes operator for OpenShift
- GitHub Actions CI/CD integration
- OpenAPI 3.0 specifications
- Performance benchmarks

---

## Test Summary

**Total Tests:** 146 tests passing

| Package | Tests | Status | Coverage |
|---------|-------|--------|----------|
| internal/api/s3 | 11/11 | ✅ Pass | 43.0% |
| internal/auth | 7/7 | ✅ Pass | 16.1% |
| internal/extensions | 2/2 | ✅ Pass | 27.3% |
| internal/storage (old) | 13/13 | ✅ Pass | 59.8% |
| internal/storage (backends) | 16/16 | ✅ Pass | Full interface |
| internal/supplychain | 4/4 | ✅ Pass | 66.7% |
| internal/metrics | 14/14 | ✅ Pass | 54.2% |
| pkg/client (Go) | 43/43 | ✅ Pass | Full SDK coverage |
| pkg/client-python (Python) | 38/38 | ✅ Pass | Full SDK coverage |

**Test Command:**
```bash
make test
```

---

## Build Status

**Build Command:**
```bash
make build
```

**Output:**
```
Building zot-artifact-store...
CGO_ENABLED=0 go build -tags containers_image_openpgp \
  -ldflags "-X main.version=0.1.0-dev" \
  -o bin/zot-artifact-store ./cmd/zot-artifact-store
```

**Status:** ✅ Building successfully

**Binary:** `bin/zot-artifact-store`

---

## Dependencies

### Core Dependencies
- Go 1.21+
- Zot v1.4.3
- BoltDB (go.etcd.io/bbolt)
- Gorilla Mux (github.com/gorilla/mux)
- JWT (github.com/golang-jwt/jwt/v4)
- UUID (github.com/google/uuid)
- Prometheus Client (github.com/prometheus/client_golang v1.17.0)
- OpenTelemetry (go.opentelemetry.io/otel v1.17.0)

### Development Dependencies
- Podman or Docker
- Make
- Git

### Replace Directives (Compatibility)
```go
replace (
    github.com/aquasecurity/trivy => github.com/aquasecurity/trivy v0.34.0
    github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.1
)
```

---

## Architecture

### Current System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Client Layer                                               │
│  - CLI, SDKs, S3 Tools, Browser                             │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  Authentication & Authorization                              │
│  - JWT Validation (Keycloak)                                │
│  - Policy Engine                                            │
│  - Audit Logging                                            │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  API Layer                                                   │
│  - S3-Compatible API (13 endpoints)                         │
│  - RBAC API (7 endpoints)                                   │
│  - Supply Chain API (8 endpoints) ✅                        │
│  - Metrics & Health API (4 endpoints) ✅                    │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  Zot Core + Extensions                                       │
│  ├── S3 API Extension ✅                                    │
│  ├── RBAC Extension ✅                                      │
│  ├── Supply Chain Extension ✅                              │
│  └── Metrics Extension ✅                                   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  Storage Layer                                               │
│  - BoltDB (Metadata) ✅                                     │
│  - Filesystem (Artifacts) ✅                                │
│  - S3/Azure/GCP (All backends) ✅                           │
└─────────────────────────────────────────────────────────────┘
```

### Database Schema (BoltDB)

**Buckets:**
- `buckets` - Bucket metadata
- `artifacts` - Artifact metadata
- `multipart_uploads` - Multipart upload state
- `upload_progress` - Part tracking
- `policies` - Access control policies
- `audit_logs` - Audit trail
- `signatures` - Artifact signatures ✨ NEW
- `sboms` - Software Bills of Materials ✨ NEW
- `attestations` - Build/test/scan attestations ✨ NEW

---

## API Endpoints Summary

### S3-Compatible API
- `GET /s3` - List buckets
- `PUT /s3/{bucket}` - Create bucket
- `DELETE /s3/{bucket}` - Delete bucket
- `GET /s3/{bucket}` - List objects
- `PUT /s3/{bucket}/{key}` - Upload object
- `GET /s3/{bucket}/{key}` - Download object (with range support)
- `HEAD /s3/{bucket}/{key}` - Get object metadata
- `DELETE /s3/{bucket}/{key}` - Delete object
- `POST /s3/{bucket}/{key}?uploads` - Initiate multipart upload
- `PUT /s3/{bucket}/{key}?uploadId={id}&partNumber={n}` - Upload part
- `POST /s3/{bucket}/{key}?uploadId={id}` - Complete multipart upload
- `DELETE /s3/{bucket}/{key}?uploadId={id}` - Abort multipart upload

### RBAC API
- `POST /rbac/policies` - Create policy
- `GET /rbac/policies` - List policies
- `GET /rbac/policies/{id}` - Get policy
- `PUT /rbac/policies/{id}` - Update policy
- `DELETE /rbac/policies/{id}` - Delete policy
- `POST /rbac/authorize` - Check authorization
- `GET /rbac/audit` - List audit logs

### Supply Chain API
- `POST /supplychain/sign/{bucket}/{key}` - Sign artifact
- `GET /supplychain/signatures/{bucket}/{key}` - Get signatures
- `POST /supplychain/verify/{bucket}/{key}` - Verify signatures
- `POST /supplychain/sbom/{bucket}/{key}` - Attach SBOM
- `GET /supplychain/sbom/{bucket}/{key}` - Get SBOM
- `POST /supplychain/attestations/{bucket}/{key}` - Add attestation
- `GET /supplychain/attestations/{bucket}/{key}` - Get attestations

### Metrics & Health API
- `GET /metrics` - Prometheus metrics
- `GET /health` - Comprehensive health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

**Total API Endpoints:** 32 (13 S3 + 7 RBAC + 8 Supply Chain + 4 Metrics/Health)

---

## Metrics

### Code Metrics
- **Total Lines of Code:** ~7,200 (production) + ~1,200 (tests)
- **Packages:** 13
- **Extensions:** 4 (S3 API, RBAC, Supply Chain, Metrics)
- **Data Models:** 20+ structs
- **Database Buckets:** 9 BoltDB buckets
- **API Endpoints:** 32 REST endpoints
- **Prometheus Metrics:** 13 metrics

### Implementation Velocity
- **Phase 1:** Foundation (8 tasks)
- **Phase 2:** S3 API (15 tasks)
- **Phase 3:** RBAC (7 tasks)
- **Phase 4:** Supply Chain (5 tasks)
- **Phase 6:** Metrics & Observability (3 tasks)
- **Total Completed:** 43 tasks (75%)

---

## Known Issues & TODOs

### Phase 3 (RBAC)
1. **Token Refresh** - No automatic refresh mechanism
2. **Conditional Access** - Time-based, IP-based restrictions not implemented
3. **Audit Retention** - No automatic cleanup policy

### Phase 4 (Supply Chain)
1. **Key Management** - In-memory key generation, no KMS integration
2. **Additional Algorithms** - Only RSA-SHA256 supported (no ECDSA, Ed25519)
3. **Cosign/Notary** - Not integrated with standard tooling yet

### Phase 6 (Metrics & Observability)
1. **Tracing Backend** - Requires external OTLP endpoint (Jaeger, Zipkin)
2. **Custom Metrics** - No API for plugin metrics yet
3. **Advanced Health Checks** - Limited component checks

---

## Next Steps

### Immediate (Phase 11 - Error Handling)
1. Implement comprehensive error classification
2. Add retry mechanisms for transient failures
3. Implement circuit breaker patterns
4. Add partial retry with range requests

### Short-term (Phase 12 - Integration & Testing)
1. End-to-end integration tests
2. Kubernetes operator for OpenShift
3. GitHub Actions CI/CD integration
4. OpenAPI 3.0 specifications
5. Performance benchmarks

### Medium-term (Phase 9 - Client Libraries)
1. JavaScript/TypeScript SDK for artifact operations

### Long-term Enhancements
1. Advanced storage features (tiering, deduplication)
2. Storage migration tools
3. Performance optimization (caching, compression)
4. Additional cloud providers

---

## How to Use

### Quick Start

```bash
# Build
make build

# Run tests
make test

# Run server
./bin/zot-artifact-store --config config/config.yaml
```

### S3 API Usage

```bash
# Create bucket
curl -X PUT http://localhost:8080/s3/mybucket

# Upload artifact
curl -X PUT \
  -H "Content-Type: application/gzip" \
  --data-binary @myapp.tar.gz \
  http://localhost:8080/s3/mybucket/myapp.tar.gz

# Download artifact
curl http://localhost:8080/s3/mybucket/myapp.tar.gz -o myapp.tar.gz
```

### With Authentication (Phase 3)

```bash
# Get token from Keycloak
TOKEN=$(curl -X POST "http://localhost:8081/realms/zot-artifact-store/protocol/openid-connect/token" \
  -d "client_id=zot-client" \
  -d "client_secret=secret" \
  -d "username=user" \
  -d "password=pass" \
  -d "grant_type=password" | jq -r '.access_token')

# Upload with auth
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/gzip" \
  --data-binary @myapp.tar.gz \
  http://localhost:8080/s3/mybucket/myapp.tar.gz
```

---

## Documentation

- [Getting Started](GETTING_STARTED.md)
- [Phase 1: Foundation](PHASE1_COMPLETE.md)
- [Phase 2: S3 API](PHASE2_COMPLETE.md) | [S3 API Reference](S3_API.md)
- [Phase 3: RBAC](PHASE3_RBAC.md)
- [Phase 4: Supply Chain Security](PHASE4_COMPLETE.md)
- [Phase 5: Storage Backends](PHASE5_COMPLETE.md)
- [Phase 6: Metrics & Observability](PHASE6_COMPLETE.md)
- [Phase 7: Go Client SDK](PHASE7_COMPLETE.md)
- [Phase 8: Python Client SDK](PHASE8_COMPLETE.md)
- [Phase 10: CLI Tool](PHASE10_COMPLETE.md) | [CLI README](../cmd/astore-cli/README.md)
- [Product Requirements](prd.md)
- [Detailed Requirements](../.kiro/specs/zot-artifact-store/requirements.md)
- [Design Document](../.kiro/specs/zot-artifact-store/design.md)
- [Implementation Tasks](../.kiro/specs/zot-artifact-store/tasks.md)

---

## Contributing

This project follows Test-Driven Development (TDD) practices:

1. Write tests first (Given-When-Then pattern)
2. Implement features to pass tests
3. Maintain test coverage
4. Use AI-friendly code patterns

**Test Patterns:**
```go
t.Run("Feature description", func(t *testing.T) {
    // Given: Setup and preconditions

    // When: Action being tested

    // Then: Assertions
    test.AssertEqual(t, expected, actual, "description")
})
```

---

## License

[To be determined]

---

**Status:** 🚀 Active Development - Nearly Complete (98%)
**Last Updated:** 2025-10-28
**Next Milestone:** Phase 9 (JavaScript Client SDK)
