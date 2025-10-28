# Getting Started with Zot Artifact Store

This guide will help you get Zot Artifact Store up and running quickly.

## Installation

### Option 1: Run from Source

#### Prerequisites
- Go 1.21+
- Git

#### Steps

```bash
# Clone the repository
git clone https://github.com/candlekeep/zot-artifact-store.git
cd zot-artifact-store

# Build
make build

# Create config
cp config/config.yaml.example config/config.yaml

# Run
./bin/zot-artifact-store --config config/config.yaml
```

### Option 2: Run with Container

#### Prerequisites
- Podman or Docker

#### Steps

```bash
# Pull or build image
podman build -t zot-artifact-store:latest -f deployments/container/Containerfile .

# Create config directory
mkdir -p config data logs
cp config/config.yaml.example config/config.yaml

# Run
podman run -d \
  --name zot-artifact-store \
  -p 8080:8080 \
  -v $(pwd)/config:/zot/config:Z \
  -v $(pwd)/data:/zot/data:Z \
  zot-artifact-store:latest
```

### Option 3: Deploy on Kubernetes/OpenShift

#### Prerequisites
- kubectl or oc CLI
- Cluster access

#### Steps

```bash
# Install CRD
kubectl apply -f deployments/operator/config/crd/zotartifactstore_crd.yaml

# Create instance
kubectl apply -f deployments/operator/config/samples/zotartifactstore_minimal.yaml

# Check status
kubectl get zotartifactstore
```

## Configuration

### Basic Configuration

Edit `config/config.yaml`:

```yaml
http:
  address: 0.0.0.0
  port: "8080"

storage:
  rootDirectory: /zot/data
  dedupe: true
  gc: true

log:
  level: info
  output: /zot/logs/zot-artifact-store.log
```

### Storage Configuration

#### Local Filesystem

```yaml
storage:
  rootDirectory: /zot/data
```

#### S3

```yaml
storage:
  storageDriver:
    name: s3
    region: us-east-1
    bucket: my-artifacts
    secure: true
```

### Extension Configuration

Extensions are automatically registered. Configuration will be added in future phases.

## Verifying Installation

### Check Server Status

```bash
# If running locally
curl http://localhost:8080/v2/

# If running in container
curl http://localhost:8080/v2/

# If running in Kubernetes
kubectl port-forward svc/zot-minimal 8080:8080
curl http://localhost:8080/v2/
```

### View Logs

```bash
# Local
tail -f logs/zot-artifact-store.log

# Container
podman logs -f zot-artifact-store

# Kubernetes
kubectl logs -f deployment/zot-minimal
```

## Next Steps

### For Developers

1. Read the [Design Document](.kiro/specs/zot-artifact-store/design.md)
2. Review [Implementation Tasks](.kiro/specs/zot-artifact-store/tasks.md)
3. Check [Contributing Guide](../CONTRIBUTING.md)
4. Start with Phase 2 tasks

### For Users

1. Wait for Phase 2-4 completion for core features
2. Monitor project progress
3. Try experimental features as they're released

## Troubleshooting

### Server won't start

**Error**: "no storage config provided"

**Solution**: Ensure `config/config.yaml` has valid storage configuration

### Cannot connect to server

**Error**: Connection refused on port 8080

**Solution**:
- Check server is running: `ps aux | grep zot-artifact-store`
- Check port binding: `netstat -an | grep 8080`
- Check firewall rules

### Container won't build

**Error**: "exec: \"pkg-config\": executable file not found"

**Solution**: This is expected - build uses CGO_ENABLED=0 flag to avoid C dependencies

### Tests failing

**Error**: Various test failures

**Solution**:
```bash
# Clean and rebuild
make clean
go mod tidy
make build
make test
```

## Getting Help

- Review documentation in `docs/` directory
- Check project issues on GitHub
- Read design and requirements documents in `.kiro/specs/`

## What's Working Now (Phase 1)

✅ Basic server startup with Zot integration
✅ Extension framework loaded
✅ Four extensions registered (stubs):
- S3 API Extension
- RBAC Extension
- Supply Chain Security Extension
- Enhanced Metrics Extension

## What's Coming Next (Phase 2+)

🚧 S3-compatible API endpoints
🚧 Resumable uploads
🚧 Bucket management
🚧 RBAC with Keycloak
🚧 Artifact signing and SBOM support
🚧 Client libraries (Go, Python, JavaScript)
🚧 CLI tool
🚧 Full Kubernetes operator

## Development Workflow

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed development workflow.

Quick start:

```bash
# Setup development environment
./scripts/dev-setup.sh

# Make changes
# ... edit code ...

# Test
make test

# Build
make build

# Run
make run
```
