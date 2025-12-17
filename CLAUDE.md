# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific package
go test ./domain/...
go test ./usecase/...
go test ./infrastructure/...
go test ./interface/...

# Build application
go build -o freee-oauth-app .

# Run application (requires environment variables)
export FREEE_CLIENT_ID="your-client-id"
export FREEE_CLIENT_SECRET="your-client-secret"
go run main.go
```

## Architecture

This project follows Clean Architecture / DDD principles with strict dependency direction (inner layers have no dependencies on outer layers).

### Layer Structure

```
domain/          → Entities and repository interfaces (no external dependencies)
usecase/         → Application business logic (depends only on domain)
infrastructure/  → External service implementations (implements domain interfaces)
interface/       → HTTP handlers (depends on usecase)
main.go          → Dependency injection and composition root
```

### Key Interfaces (defined in domain/)

- `TokenRepository`: Token persistence interface (implemented by `FileTokenRepository`)
- `OAuthProvider`: OAuth flow interface (implemented by `FreeeOAuthProvider`)

### Dependency Injection

All dependencies are wired in `main.go` through `initializeApp()`. Infrastructure implementations are injected into use cases via constructor injection.

## Development Approach

This codebase was developed using TDD (Test-Driven Development). When adding features:
1. Write failing tests first
2. Implement minimum code to pass tests
3. Refactor while keeping tests green
