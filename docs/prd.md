# Product requirements

Build a lightweight, scalable microservice for handling large binary file uploads and downloads with enterprise-grade reliability features including resumable transfers and integrity verification.

The service will support S3-like bucket organization and include client libraries for Go, Python, and JavaScript, plus a CLI tool based on the golang libary.

The service supports binary files but also upload/download etc of the following artefacts:
- OCI container images
- packages like rpm
- Helm charts (via helm push)
- SBOMs (Software Bill of Materials)
- Signatures and attestations
- Terraform modules

Management of artefacts like e.g. containers, rpms or helm charts is done via their native APIs or protocols. The language libraries and the CLI tool do not need to support these artefacts, only simple binaries.

Implement RBAC for access to all the stored artefacts. If the implementation uses another codebase as starting point, investigate if this codebase already has RBAC and can support Keycloak.

The service should be support "software-supply-chain security" as its core value proposition: artefact signing, implicit SBOM support, attestations etc. This must be supported in collaboration with popular CI/CD tools lke Tekton or GitHub Actions.

## Key Features
- HTTP-based file upload/download with chunked transfer support
- SHA256 file integrity verification  
- Resumable uploads using HTTP 206 Partial Content
- S3-compatible bucket/location organization
- Pluggable storage backends (Local, S3, Google, Azure)
- Bearer token authentication with optional anonymous downloads
- Prometheus metrics integration
- Cross-platform client libraries and CLI tool
- public health-check API endpoint for Kubernetes readyness-probes
- supports and uses OpenTelemetry

## Technical Architecture

### Technology Stack
- Language: Go (Golang)
- Metadata Store: BoltDB (embedded key-value store)
- Storage Backends: Pluggable architecture supporting:
  - Local filesystem (default)
  - Amazon S3
  - Google Cloud Storage
  - Azure
- Keycloak for authentication
- use popular golang libraries only:
    - Echo framework
    - logrus


### Coding requiremnts
- keep it as simple as possible, don't over-engineer the solution.
- tests from the beginning, to avoid major code changes late in the process.
- prioritize security and safe coding practices.

## API Operations

### Core Object Operations
- **PUT Object** - Upload files to buckets. Supports multipart uploads for large files and allows setting metadata, storage classes, and access permissions.
- **GET Object** - Download files from bucket. Supports range requests for partial downloads and conditional requests based on ETags or modification dates.
- **DELETE Object** - Remove individual objects from buckets. Can be batched for efficiency when deleting multiple objects.
- **HEAD Object** - Retrieve object metadata without downloading the actual content. Useful for checking file existence, size, and properties.

### Bucket Management
- **CREATE Bucket** - Create new buckets with specified regions and configuration options like versioning and encryption.
- **LIST Objects** - Enumerate objects in a bucket with filtering options by prefix, delimiter, and pagination support for large datasets.
- **DELETE Bucket** - Remove empty buckets. Recursevly delete content if not empty.

### Access Control & Security
- **PUT/GET Bucket Policy** - Manage bucket-level permissions using JSON policy documents for fine-grained access control.
- **Generate Presigned URLs** - Create temporary URLs that allow limited-time access to private objects without exposing credentials.

### Advanced Features
- **Copy Object** - Duplicate objects within or between buckets without downloading/uploading, useful for backups and data organization.

Investigate and propose other relevant and useful API endpoints.


## Zot Registry Analysis

### Overview
Zot is a production-ready, vendor-neutral OCI image registry that stores images in OCI image format and follows the OCI distribution specification. It's written in Go (~174k lines of code across 342 files) and designed as a modern alternative to Docker Distribution.

### Architecture Complexity Assessment

**Core Components:**
- **API Layer** (`pkg/api/`) - HTTP server, authentication, authorization, routing
- **Storage Layer** (`pkg/storage/`) - Pluggable storage backends (local filesystem, S3)
- **Extensions** (`pkg/extensions/`) - Modular extension system for additional features
- **Metadata** (`pkg/meta/`) - Database abstraction layer (BoltDB, DynamoDB, Redis)
- **Scheduler** (`pkg/scheduler/`) - Background task management

**Key Strengths for Artifact Store:**
- **Pluggable Storage**: Already supports S3 backend with local caching
- **Extension Architecture**: Well-designed plugin system for adding custom functionality
- **OCI Compliance**: Full OCI distribution spec compliance
- **Built-in Features**: Authentication, authorization, garbage collection, deduplication
- **Cloud Native**: Designed for containerized deployments

**Complexity Analysis:**

**Low Complexity (1-3 months):**
- Basic artifact storage using existing OCI blob storage
- Leverage existing S3 backend and authentication
- Simple metadata extensions for artifact types
- Basic REST API for artifact operations

**Medium Complexity (3-6 months):**
- Custom artifact type validation and processing
- Advanced metadata indexing and search
- Integration with existing CI/CD pipelines
- Custom UI for artifact management

**High Complexity (6+ months):**
- Complex artifact relationship modeling
- Advanced security scanning integration
- Multi-tenant isolation
- Custom storage optimization for specific artifact types

### Recommended Approach

**Phase 1 (Low Complexity):**
1. Fork Zot and create artifact-specific extensions
2. Implement artifact type registry and validation
3. Add artifact-specific metadata handling
4. Create basic REST API endpoints for artifact operations

**Phase 2 (Medium Complexity):**
1. Develop custom UI for artifact browsing
2. Add advanced search and filtering capabilities
3. Implement artifact lifecycle management
4. Add integration APIs for CI/CD tools

**Advantages:**
- Mature, production-tested codebase
- Strong OCI compliance and cloud storage support
- Extensible architecture reduces custom development
- Active community and good documentation
- Built-in security and performance features

**Considerations:**
- Go expertise required for deep customizations
- Need to maintain compatibility with upstream changes
- Some artifact-specific features may require significant extension development## 
Harbor Registry Analysis

### Overview
Harbor is a CNCF-hosted, enterprise-grade container registry that extends Docker Distribution with security, identity, and management features. It's written in Go (~224k lines of code across 1,562 files) and designed as a comprehensive registry solution with a full web UI and enterprise features.

### Architecture Complexity Assessment

**Core Components:**
- **Core Service** (`src/core/`) - Main API server with Beego framework
- **Portal** (`src/portal/`) - Angular-based web UI (separate frontend)
- **Job Service** (`src/jobservice/`) - Background job processing system
- **Registry Control** (`src/registryctl/`) - Registry management service
- **Controllers** (`src/controller/`) - Business logic layer with 30+ controllers
- **Database Layer** - PostgreSQL with comprehensive ORM

**Key Strengths for Artifact Store:**
- **Enterprise Features**: RBAC, LDAP/OIDC integration, audit logging
- **Multi-Artifact Support**: Already handles containers, Helm charts, OCI artifacts
- **Comprehensive UI**: Full-featured web interface for management
- **Replication**: Multi-registry synchronization capabilities
- **Security**: Built-in vulnerability scanning, image signing support
- **Job System**: Robust background processing for cleanup, scanning, replication

**Complexity Analysis:**

**Low Complexity (2-4 months):**
- Extend existing artifact processors for new types
- Add custom artifact metadata fields
- Leverage existing storage and authentication
- Use existing REST API patterns

**Medium Complexity (4-8 months):**
- Custom artifact lifecycle policies
- Specialized artifact processing workflows
- Integration with external artifact tools
- Custom UI components for artifact types

**High Complexity (8+ months):**
- Major architectural changes for non-OCI artifacts
- Custom storage backends for specialized artifacts
- Advanced artifact relationship modeling
- Multi-tenant artifact isolation

### Recommended Approach

**Phase 1 (Medium Complexity):**
1. Extend Harbor's artifact processor system
2. Add custom artifact type definitions
3. Implement artifact-specific metadata handling
4. Create custom UI components for new artifact types

**Phase 2 (High Complexity):**
1. Develop specialized storage adapters if needed
2. Add advanced artifact workflow management
3. Implement custom replication policies
4. Add artifact-specific security policies

**Advantages:**
- Enterprise-ready with comprehensive features
- Strong security and compliance capabilities
- Excellent web UI and user experience
- Robust job system for background processing
- Active CNCF community and enterprise support
- Multi-artifact support already built-in

**Considerations:**
- Much larger and more complex codebase
- Requires PostgreSQL database
- Heavier resource requirements
- More complex deployment and maintenance
- Beego framework may be less familiar than modern Go frameworks

## Zot vs Harbor Comparison

### Codebase Complexity
- **Zot**: ~174k lines, 342 files - Focused, lightweight
- **Harbor**: ~224k lines, 1,562 files - Comprehensive, enterprise-grade

### Architecture Philosophy
- **Zot**: Minimal, extensible registry focused on OCI compliance
- **Harbor**: Full-featured enterprise registry with comprehensive management

### Development Effort for Artifact Store

| Aspect | Zot | Harbor |
|--------|-----|--------|
| **Learning Curve** | Low - Simple, focused codebase | High - Complex enterprise architecture |
| **Time to MVP** | 1-3 months | 2-4 months |
| **Customization Effort** | Medium - Extension system | High - Complex controller system |
| **Maintenance Overhead** | Low - Minimal dependencies | High - Multiple services, database |
| **Enterprise Features** | Basic - Need custom development | Comprehensive - Built-in |
| **UI Development** | High - Build from scratch | Low - Extend existing Angular app |
| **Storage Flexibility** | High - Simple pluggable backends | Medium - More complex but powerful |
| **Deployment Complexity** | Low - Single binary | High - Multi-service architecture |

### Recommendation

**Choose Zot if:**
- You need a lightweight, focused solution
- You want to build custom artifact-specific features
- You prefer minimal dependencies and simple deployment
- You have a small team and want faster development cycles
- You need maximum flexibility in storage and processing

**Choose Harbor if:**
- You need enterprise features (RBAC, audit, compliance)
- You want a comprehensive web UI out of the box
- You have existing Harbor expertise or infrastructure
- You need multi-registry replication and management
- You're building for large enterprise environments

**For your artifact store requirements, Zot appears to be the better choice** due to its:
- Simpler architecture allowing faster development
- Better alignment with your "keep it simple" philosophy
- Lower operational overhead
- More flexible extension system for custom artifact types
- Easier integration with your preferred tech stack (Echo, logrus)

