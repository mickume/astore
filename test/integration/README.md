# Integration Tests

Integration tests verify that components work together correctly.

## Structure

- Tests should use real dependencies where possible (databases, storage, etc.)
- Use TestContainers for containerized dependencies (Keycloak, Postgres, etc.)
- Follow Given-When-Then pattern
- Clean up resources after tests

## Running Integration Tests

```bash
# Run all integration tests
make test-integration

# Run specific integration test
go test -v -run TestIntegration ./test/integration/...
```

## Test Patterns

### Given-When-Then Format

```go
func TestFeature(t *testing.T) {
    // Given: Setup test preconditions
    // ... setup code ...

    // When: Execute the action being tested
    // ... action code ...

    // Then: Verify the results
    // ... assertions ...
}
```

### Resource Cleanup

Always use `defer` or `t.Cleanup()` for resource cleanup:

```go
func TestWithContainer(t *testing.T) {
    container := startTestContainer(t)
    t.Cleanup(func() {
        container.Stop()
    })

    // ... test code ...
}
```

## TestContainers Usage

Integration tests should use TestContainers for external dependencies:

- Keycloak: For RBAC testing
- PostgreSQL/BoltDB: For metadata storage testing
- MinIO: For S3 storage backend testing
- Prometheus: For metrics testing

Example to be implemented in Phase 3+.
