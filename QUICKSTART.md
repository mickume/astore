# Zot Artifact Store - Quick Start Guide

## Running the Development Server

### Step 1: Build the Binary

```bash
make build
```

This creates `bin/zot-artifact-store` (71MB static binary).

### Step 2: Start the Server

```bash
./bin/zot-artifact-store --config config/config-minimal.yaml
```

### Expected Output

```
Zot Artifact Store v0.1.0-dev
Starting...
Loading configuration from: config/config-minimal.yaml
{"level":"info","message":"Initializing extension registry"}
{"level":"info","extension":"s3api","message":"registered extension"}
{"level":"info","extension":"rbac","message":"registered extension"}
{"level":"info","extension":"supplychain","message":"registered extension"}
{"level":"info","extension":"metrics","message":"registered extension"}
{"level":"info","message":"Initializing storage"}
{"level":"info","address":"0.0.0.0","port":"8080","message":"Starting Zot Artifact Store"}
{"level":"info","extensions":4,"message":"Extensions registered"}
```

✅ **Server is now running on http://localhost:8080**

### Step 3: Test the Server

```bash
# Test the API
curl http://localhost:8080/v2/

# Expected response: {}
```

## Troubleshooting

### Issue 1: Port Already in Use

**Error:**
```
{"level":"error","error":"listen tcp 0.0.0.0:8080: bind: address already in use"}
```

**Solution:**
```bash
# Find and kill the process using port 8080
lsof -ti:8080 | xargs kill -9

# Or kill all zot-artifact-store processes
pkill -f zot-artifact-store

# Then try again
./bin/zot-artifact-store --config config/config-minimal.yaml
```

### Issue 2: Cache Database Timeout

**Warning (non-fatal):**
```
{"level":"error","error":"timeout","dbPath":"/tmp/zot-artifacts/cache.db","message":"unable to create cache db"}
```

**Solution:**
```bash
# Remove stale lock files
rm -rf /tmp/zot-artifacts/cache.db*

# Or disable deduplication in config (already done in config-minimal.yaml)
storage:
  dedupe: false
  gc: false
```

### Issue 3: Permission Denied on /tmp/zot-artifacts

**Error:**
```
permission denied
```

**Solution:**
```bash
# Ensure directory is writable
mkdir -p /tmp/zot-artifacts
chmod 755 /tmp/zot-artifacts

# Or use a different directory in config
storage:
  rootDirectory: /path/to/your/directory
```

## Configuration Options

### Minimal Configuration (Development)

**File:** `config/config-minimal.yaml`

```yaml
# HTTP Server
http:
  address: 0.0.0.0
  port: "8080"

# Storage
storage:
  rootDirectory: /tmp/zot-artifacts
  dedupe: false  # Disabled for development
  gc: false      # Disabled for development

# Logging
log:
  level: info
```

### Custom Port

To use a different port:

```yaml
http:
  port: "9000"
```

### Custom Storage Directory

```yaml
storage:
  rootDirectory: /data/artifacts
```

### Debug Logging

```yaml
log:
  level: debug
```

## Running in Background

### Option 1: Background with Logs

```bash
./bin/zot-artifact-store --config config/config-minimal.yaml > /tmp/zot.log 2>&1 &
echo $! > /tmp/zot.pid

# View logs
tail -f /tmp/zot.log

# Stop server
kill $(cat /tmp/zot.pid)
```

### Option 2: Using nohup

```bash
nohup ./bin/zot-artifact-store --config config/config-minimal.yaml &

# View logs
tail -f nohup.out

# Stop server
pkill -f zot-artifact-store
```

## Testing the API

### OCI Registry Endpoints

```bash
# List repositories
curl http://localhost:8080/v2/_catalog

# Get manifest
curl http://localhost:8080/v2/<name>/manifests/<tag>
```

### S3-Compatible API (Future)

Once extensions are fully integrated:

```bash
# List buckets
curl http://localhost:8080/s3

# Upload object
curl -X PUT -T file.bin http://localhost:8080/s3/mybucket/myfile

# Download object
curl http://localhost:8080/s3/mybucket/myfile -o downloaded.bin
```

### Health Checks (Future)

```bash
# Health check
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/health/ready

# Liveness probe
curl http://localhost:8080/health/live
```

### Metrics (Future)

```bash
# Prometheus metrics
curl http://localhost:8080/metrics
```

## Stopping the Server

### Graceful Shutdown

```bash
# Send SIGTERM for graceful shutdown
pkill -TERM -f zot-artifact-store

# Or use Ctrl+C if running in foreground
```

### Force Stop

```bash
# Send SIGKILL to force stop
pkill -9 -f zot-artifact-store
```

## Next Steps

1. **Explore Configuration:**
   - See `config/README.md` for all configuration options
   - Try `config/config-production.yaml` for production settings

2. **Read Documentation:**
   - `docs/IMPLEMENTATION_STATUS.md` - Overall project status
   - `docs/S3_API.md` - S3 API reference
   - `docs/PHASE*_COMPLETE.md` - Detailed phase documentation

3. **Test Client SDKs:**
   - Go: `pkg/client/`
   - Python: `pkg/client-python/`
   - JavaScript: `pkg/client-js/`

4. **Deploy:**
   - Container: `make podman-build`
   - Kubernetes: See `deployments/kubernetes/`

## Common Use Cases

### Development Testing

```bash
# Quick start with minimal config
./bin/zot-artifact-store --config config/config-minimal.yaml
```

### Production Deployment

```bash
# Use production config with TLS, RBAC, cloud storage
./bin/zot-artifact-store --config config/config-production.yaml
```

### Custom Configuration

```bash
# Create your own config
cp config/config.yaml.example my-config.yaml
# Edit my-config.yaml
./bin/zot-artifact-store --config my-config.yaml
```

## FAQ

**Q: Why does it say "lint extension is disabled"?**
A: This is a warning from the base Zot registry. The lint extension is not needed for artifact store functionality.

**Q: Can I use this without configuration file?**
A: Yes, the server will use defaults (port 8080, /tmp/zot-artifacts storage).

**Q: How do I enable RBAC?**
A: Configure the RBAC extension in your config file (see config/config.yaml.example for details). Note: Full extension configuration loading is planned for a future update.

**Q: Is HTTPS supported?**
A: Yes, configure TLS in the http section of your config file.

## Getting Help

- **Issues:** Check the troubleshooting section above
- **Documentation:** See `docs/` directory
- **Configuration:** See `config/README.md`
- **Examples:** See `examples/` directory

---

**Quick Reference:**

```bash
# Build
make build

# Run
./bin/zot-artifact-store --config config/config-minimal.yaml

# Test
curl http://localhost:8080/v2/

# Stop
pkill -f zot-artifact-store
```

✅ You're ready to go!
