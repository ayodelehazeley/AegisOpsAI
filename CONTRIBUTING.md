# Contributing to AegisOpsAI

Thank you for your interest in contributing to AegisOpsAI! This document provides guidelines and information to help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Module Development Guide](#module-development-guide)
- [Testing Requirements](#testing-requirements)

## Code of Conduct

By participating in this project, you agree to act respectfully and professionally toward all other contributors.

## Getting Started

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/<your-user>/aegisops.git
   cd aegisops
   ```
3. **Create a branch** for your work:
   ```bash
   git checkout -b feat/my-feature
   ```

## Development Setup

### Prerequisites

| Tool             | Version |
|------------------|---------|
| Go               | 1.23+   |
| Docker & Compose | Latest  |
| golangci-lint    | 1.61+   |
| goose            | 3.x     |

### Running Locally

```bash
# Start infrastructure
docker compose up -d postgres nats

# Apply migrations
make migrate-up

# Run the application
make run
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires running Postgres & NATS)
make test-integration

# Linting
make lint
```

## Code Standards

- **Language**: Go 1.23, idiomatic style.
- **Formatting**: `gofmt` / `goimports` — CI enforces this.
- **Linting**: Must pass `golangci-lint run ./...`.
- **Documentation**: All exported types and functions must have doc comments.
- **Error Handling**: Always wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Logging**: Use `log/slog` structured logging — no `fmt.Println` or `log.Printf`.
- **Configuration**: Use Viper; all config values must have sensible defaults.
- **Package Layout**: Application code lives under `internal/`. Each module is a separate package.

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type       | Description                          |
|------------|--------------------------------------|
| `feat`     | A new feature                        |
| `fix`      | A bug fix                            |
| `docs`     | Documentation changes                |
| `refactor` | Code changes that neither fix nor add|
| `test`     | Adding or updating tests             |
| `chore`    | Build, CI, tooling changes           |
| `perf`     | Performance improvements             |

### Examples

```
feat(incidents): add severity escalation logic
fix(auth): handle expired JWT tokens gracefully
docs(readme): add architecture diagram
test(events): add integration tests for ingestion pipeline
chore(ci): upgrade golangci-lint to v1.61
```

## Pull Request Process

1. **Ensure** all tests pass and lint is clean.
2. **Update** documentation if your change affects public APIs or behavior.
3. **Open** a PR against the `main` branch.
4. **Fill in** the PR template with a description of changes, related issues, and testing done.
5. **Request review** from at least one maintainer.
6. **Squash merge** is preferred for clean history.

### PR Checklist

- [ ] Tests added / updated
- [ ] Documentation updated
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] Commit messages follow conventional commit format

## Module Development Guide

AegisOpsAI is a **modular monolith**. Each feature area is an independent module under `internal/`.

### Module Structure

```
internal/mymodule/
├── handler.go       # HTTP handlers (Chi routes)
├── service.go       # Business logic
├── repository.go    # Database access (pgx)
├── models.go        # Domain types
├── events.go        # NATS event publishing / subscribing
└── mymodule_test.go # Tests
```

### Module Guidelines

1. **Define interfaces** for your service and repository layers.
2. **Accept dependencies** via constructor injection.
3. **Register routes** by implementing a `RegisterRoutes(r chi.Router)` method.
4. **Publish domain events** via NATS JetStream for cross-module communication.
5. **Never import** another module's internal types — use shared types in `internal/common/`.

## Testing Requirements

- **Unit tests** are required for all business logic in service and repository layers.
- **Table-driven tests** are preferred for functions with multiple input/output cases.
- **Integration tests** should use the `//go:build integration` build tag.
- **Test coverage** should not decrease — aim for ≥ 80% on new code.
- **Mocks** should be generated via interfaces, not concrete types.

```go
// Example table-driven test
func TestSeverityValidation(t *testing.T) {
    tests := []struct {
        name     string
        severity string
        wantErr  bool
    }{
        {"valid critical", "critical", false},
        {"valid high", "high", false},
        {"invalid", "unknown", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateSeverity(tt.severity)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateSeverity(%q) error = %v, wantErr %v", tt.severity, err, tt.wantErr)
            }
        })
    }
}
```

---

Thank you for contributing to AegisOpsAI! If you have questions, open a discussion on GitHub.
