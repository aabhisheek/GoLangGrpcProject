# Contributing Guide

Thank you for considering contributing to the Voucher & Payment Service! This document provides guidelines for contributing to the project.

## Code of Conduct

Be respectful, collaborative, and professional in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone <your-fork-url>`
3. Create a feature branch: `git checkout -b feature/amazing-feature`
4. Make your changes
5. Test your changes
6. Commit and push
7. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- protoc (Protocol Buffers compiler)

### Setup Development Environment

**Linux/macOS:**
```bash
./scripts/setup.sh
```

**Windows:**
```cmd
scripts\setup.bat
```

**Manual Setup:**
```bash
# Install dependencies
go mod download

# Start infrastructure
docker-compose up -d

# Run migrations
go run cmd/migrate/main.go up

# Generate protobuf
make proto

# Run the service
make run
```

## Project Structure

```
.
├── api/                    # API definitions
│   ├── proto/              # Protocol Buffer files
│   └── generated/          # Generated gRPC code
├── cmd/                    # Application entry points
│   ├── server/             # Main server
│   └── migrate/            # Migration tool
├── internal/               # Private application code
│   ├── domain/             # Business entities
│   ├── repository/         # Data access
│   ├── service/            # Business logic
│   ├── grpc/               # gRPC handlers
│   ├── cache/              # Caching layer
│   ├── messaging/          # Message queue
│   ├── observability/      # Tracing & logging
│   └── worker/             # Worker pool
├── migrations/             # Database migrations
└── scripts/                # Helper scripts
```

## Coding Standards

### Go Style Guide

Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md):

1. **Naming Conventions**
   - Use camelCase for variables
   - Use PascalCase for exported types
   - Use descriptive names

```go
// Good
userID := "user-001"
type VoucherService struct {}

// Bad
uid := "user-001"
type voucherservice struct {}
```

2. **Error Handling**
   - Always handle errors
   - Wrap errors with context
   - Use custom error types

```go
// Good
voucher, err := repo.GetByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get voucher %s: %w", id, err)
}

// Bad
voucher, _ := repo.GetByID(ctx, id)
```

3. **Context**
   - Always pass context as first parameter
   - Don't store context in structs
   - Respect context cancellation

```go
// Good
func (s *Service) GetVoucher(ctx context.Context, id string) (*Voucher, error)

// Bad
func (s *Service) GetVoucher(id string) (*Voucher, error)
```

4. **Concurrency**
   - Use channels for communication
   - Protect shared state with mutexes
   - Always handle goroutine cleanup

```go
// Good
done := make(chan struct{})
go func() {
    defer close(done)
    // work
}()
<-done
```

### Code Formatting

- Use `gofmt` or `goimports` for formatting
- Run `go vet` for static analysis
- Use `golangci-lint` for comprehensive linting

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...

# Lint (if installed)
golangci-lint run
```

## Testing

### Writing Tests

1. **Unit Tests**
   - Test individual functions
   - Use table-driven tests
   - Mock external dependencies

```go
func TestVoucherService_GetVoucher(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        want    *Voucher
        wantErr bool
    }{
        {
            name: "valid voucher",
            id:   "voucher-001",
            want: &Voucher{ID: "voucher-001"},
            wantErr: false,
        },
        {
            name: "not found",
            id:   "invalid-id",
            want: nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

2. **Integration Tests**
   - Test component interactions
   - Use testcontainers if possible
   - Tag with `//go:build integration`

3. **Benchmark Tests**
   - Test performance-critical code
   - Use `testing.B`
   - Analyze results

```go
func BenchmarkGenerateVoucherCode(b *testing.B) {
    for i := 0; i < b.N; i++ {
        generateVoucherCode()
    }
}
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test -v ./internal/service/...

# With coverage
go test -cover ./...

# Benchmarks
make benchmark

# Race detection
go test -race ./...
```

## Pull Request Process

1. **Before Submitting**
   - Update documentation if needed
   - Add tests for new features
   - Ensure all tests pass
   - Format code with `gofmt`
   - Update CHANGELOG.md

2. **PR Title Format**
   ```
   [TYPE] Short description
   
   Types: feat, fix, docs, style, refactor, test, chore
   ```

   Examples:
   - `[feat] Add wallet freeze functionality`
   - `[fix] Resolve race condition in worker pool`
   - `[docs] Update API documentation`

3. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No new warnings
   ```

4. **Review Process**
   - Address reviewer comments
   - Keep discussion professional
   - Be open to feedback
   - Update PR as needed

## Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Tests
- `chore`: Maintenance

**Examples:**
```
feat(voucher): add search by brand

Implement brand-based search with caching support.

Closes #123
```

```
fix(wallet): prevent negative balance

Add validation to ensure balance cannot go below zero.

Fixes #456
```

## Documentation

### Code Documentation

- Document all exported types and functions
- Use complete sentences
- Include examples for complex functionality

```go
// VoucherService handles voucher business logic including
// search, retrieval, and purchase operations. It integrates
// with caching and message queue for optimal performance.
type VoucherService struct {
    // unexported fields
}

// SearchVouchers searches for available vouchers based on the
// provided criteria. Results are cached for 5 minutes.
//
// Example:
//   vouchers, count, err := service.SearchVouchers(ctx, params)
func (s *VoucherService) SearchVouchers(ctx context.Context, params SearchParams) ([]*Voucher, int32, error) {
    // implementation
}
```

### API Documentation

- Update API_GUIDE.md for API changes
- Include request/response examples
- Document error scenarios

### Architecture Documentation

- Update ARCHITECTURE.md for design changes
- Document decision rationale
- Include diagrams if helpful

## Performance Considerations

1. **Avoid Premature Optimization**
   - Measure first
   - Optimize hot paths
   - Use profiling tools

2. **Database**
   - Use connection pooling
   - Optimize queries
   - Add appropriate indexes
   - Use prepared statements

3. **Caching**
   - Cache frequently accessed data
   - Set appropriate TTLs
   - Invalidate on writes

4. **Concurrency**
   - Use worker pools
   - Avoid goroutine leaks
   - Handle backpressure

## Security Guidelines

1. **Input Validation**
   - Validate all inputs
   - Sanitize user data
   - Use parameterized queries

2. **Authentication/Authorization**
   - Implement proper auth
   - Validate tokens
   - Check permissions

3. **Sensitive Data**
   - Never log sensitive info
   - Encrypt in transit and at rest
   - Follow data protection regulations

4. **Dependencies**
   - Keep dependencies updated
   - Review security advisories
   - Use `go mod tidy`

## Release Process

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. Create GitHub release with notes

## Getting Help

- Check existing documentation
- Search closed issues
- Ask in discussions
- Contact maintainers

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing! 🎉

