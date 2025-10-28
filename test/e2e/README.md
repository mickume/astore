# End-to-End Tests

E2E tests verify complete user workflows from start to finish.

## Structure

- Test complete scenarios as a user would experience them
- Use real server instances (started in containers)
- Test full API flows
- Verify cross-cutting concerns (auth, metrics, logging)

## Running E2E Tests

```bash
# Run all E2E tests
make test-e2e

# Run specific E2E test
go test -v -run TestE2E ./test/e2e/...
```

## Test Scenarios

E2E tests will cover:

1. **Artifact Upload and Download**
   - Upload binary artifact via S3 API
   - Download artifact
   - Verify integrity

2. **Authentication and Authorization**
   - Login with Keycloak
   - Access control verification
   - Token refresh

3. **Supply Chain Security**
   - Sign artifact
   - Attach SBOM
   - Add attestation
   - Verify signature

4. **Complete CI/CD Workflow**
   - Build artifact
   - Sign and attest
   - Upload to registry
   - Download in deployment
   - Verify provenance

## Implementation

E2E tests will be implemented in Phase 11-12 after all core features are complete.
