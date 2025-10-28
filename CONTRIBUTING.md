# Contributing to Zot Artifact Store

Thank you for your interest in contributing to Zot Artifact Store!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Podman (preferred) or Docker
- Make
- Git

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/candlekeep/zot-artifact-store.git
cd zot-artifact-store

# Run development setup
./scripts/dev-setup.sh

# Verify build
make build

# Run tests
make test
```

## Development Workflow

### Test-Driven Development (TDD)

This project strictly follows TDD principles:

1. **Write tests first** - Before implementing any feature, write tests that define expected behavior
2. **Run tests** - Verify tests fail initially (red)
3. **Implement feature** - Write minimal code to make tests pass (green)
4. **Refactor** - Improve code while keeping tests passing
5. **Coverage** - Maintain 90% code coverage target

### Test Patterns

Use the Given-When-Then pattern for all tests:

```go
func TestFeature(t *testing.T) {
    t.Run("Description of behavior", func(t *testing.T) {
        // Given: Setup test preconditions
        logger := test.NewTestLogger(t)
        ext := NewExtension()

        // When: Execute the action being tested
        result := ext.DoSomething()

        // Then: Verify the results
        test.AssertEqual(t, expected, result, "result should match expected")
    })
}
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` and `golangci-lint`
- Write clear, self-documenting code
- Add comments for complex logic
- Keep functions small and focused

### Commit Messages

Use conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Example:
```
feat(s3api): implement multipart upload support

Add support for S3-compatible multipart uploads with:
- Upload initiation endpoint
- Part upload handling
- Upload completion and abort

Closes #42
```

### Pull Request Process

1. **Create a branch** - Use descriptive names: `feat/s3-multipart-upload`
2. **Write tests** - Ensure tests cover new functionality
3. **Implement feature** - Follow TDD workflow
4. **Run checks** - `make test lint`
5. **Update docs** - Document new features or changes
6. **Create PR** - Use PR template and link related issues
7. **Code review** - Address reviewer feedback
8. **Merge** - Squash commits if requested

### PR Checklist

- [ ] Tests written and passing
- [ ] Code coverage maintained (90%+)
- [ ] Linter passes
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow convention
- [ ] Branch is up to date with main

## Project Structure

### Adding New Features

1. **Extensions**: Add new extensions in `internal/extensions/`
2. **APIs**: Add API handlers in `internal/api/`
3. **Models**: Add data models in `internal/models/`
4. **Tests**: Add tests alongside code (e.g., `feature_test.go`)

### Phase-Based Development

The project follows a 12-phase development plan:

- **Phase 1**: Foundation (✅ Complete)
- **Phase 2**: Core S3 API
- **Phase 3**: RBAC
- **Phase 4**: Supply chain security
- **Phase 5**: Storage backends
- **Phase 6**: Metrics
- **Phase 7-9**: Client libraries
- **Phase 10**: CLI tool
- **Phase 11-12**: Testing and operator

See [tasks.md](.kiro/specs/zot-artifact-store/tasks.md) for detailed phase breakdown.

## Testing

### Unit Tests

```bash
# Run all unit tests
make test-unit

# Run specific test
go test -v ./internal/extensions/s3api/...

# Run with coverage
go test -cover ./...
```

### Integration Tests

```bash
# Run integration tests
make test-integration
```

### E2E Tests

```bash
# Run end-to-end tests
make test-e2e
```

## Building and Running

### Local Development

```bash
# Build binary
make build

# Run locally
make run

# Or with custom config
./bin/zot-artifact-store --config config/config.yaml
```

### Container Development

```bash
# Build container
make podman-build

# Run container
make podman-run

# View logs
podman logs -f zot-artifact-store
```

## Documentation

### Adding Documentation

- **API docs**: Update OpenAPI specs in `api/openapi/`
- **User docs**: Add guides in `docs/`
- **Code docs**: Use godoc comments
- **README**: Update for significant changes

### Generating Docs

```bash
# Generate API documentation
make docs

# Generate godoc
godoc -http=:6060
```

## Getting Help

- Check existing documentation
- Search existing issues
- Ask in discussions
- Create an issue for bugs or feature requests

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow community guidelines

## License

By contributing, you agree that your contributions will be licensed under the project's license.
