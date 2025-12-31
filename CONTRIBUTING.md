# Contributing to FinOps Enforcer

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce**
- **Expected vs actual behavior**
- **Environment details** (Kubernetes version, Go version, etc.)
- **Logs and error messages**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use case**: Why is this needed?
- **Proposed solution**: What should it do?
- **Alternatives considered**: What other approaches did you consider?

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** with clear, documented code
3. **Add tests** for new functionality
4. **Ensure tests pass**: `make test`
5. **Update documentation** if needed
6. **Follow code style** (run `make fmt`)
7. **Submit pull request** with clear description

#### Pull Request Guidelines

- Keep changes focused and atomic
- Write clear commit messages
- Reference related issues
- Ensure CI passes
- Request review from maintainers

## Development Setup

### Prerequisites

```bash
# Required
go 1.21+
docker
kubectl
kind (for local testing)

# Optional
helm
golangci-lint
```

### Local Development

```bash
# Clone repository
git clone https://github.com/yourusername/finops-enforcer.git
cd finops-enforcer

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally (requires kubeconfig)
make run
```

### Local Kubernetes Testing

```bash
# Create local cluster
make kind-create

# Install OpenCost (required dependency)
helm repo add opencost https://opencost.github.io/opencost-helm-chart
helm install opencost opencost/opencost \
  --namespace opencost \
  --create-namespace

# Build and load image
make kind-load

# Deploy
make deploy

# Test
kubectl apply -f config/samples/
make logs
```

## Code Style

### Go Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting: `make fmt`
- Run linter: `golangci-lint run`
- Write meaningful comments for exported functions
- Keep functions focused and testable

### Commit Messages

Format:
```
<type>: <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Build/tooling changes

Example:
```
feat: add traffic threshold condition

Implement traffic-based idle detection using Prometheus metrics.
Resources with zero traffic for the idle window are considered idle.

Closes #42
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run specific package
go test ./pkg/policy/...

# With coverage
make coverage
```

### Integration Tests

```bash
# Requires kind cluster
make test-integration
```

### Manual Testing

```bash
# Deploy to test cluster
make deploy

# Apply test policy
kubectl apply -f config/samples/test-policy-dryrun.yaml

# Watch logs
make logs

# Verify behavior
kubectl get enforcementpolicies -n finops-system -o wide
```

## Documentation

### What to Document

- **New features**: Update README.md and relevant docs
- **Configuration changes**: Update POLICIES.md
- **Breaking changes**: Update DESIGN.md and CHANGELOG.md
- **Operational changes**: Update RUNBOOK.md

### Documentation Standards

- Use clear, concise language
- Provide examples
- Link between related docs
- Keep up to date with code

## Project Structure

```
.
├── api/              # CRD definitions
├── cmd/              # Entry points (controller, CLI)
├── pkg/              # Core libraries
│   ├── controller/   # Kubernetes controller
│   ├── policy/       # Policy engine
│   ├── cost/         # OpenCost integration
│   ├── enforcement/  # Action execution
│   ├── metrics/      # Prometheus metrics
│   └── notifications/# Slack integration
├── config/           # Kubernetes manifests
├── deploy/           # Deployment configs (Helm)
├── docs/             # Documentation
└── test/             # Test utilities
```

## Review Process

### What Reviewers Look For

1. **Correctness**: Does it work as intended?
2. **Tests**: Are there adequate tests?
3. **Documentation**: Is it documented?
4. **Code quality**: Is it maintainable?
5. **Safety**: Could it cause issues in production?

### Getting Your PR Merged

- Address review feedback promptly
- Keep PR scope focused
- Rebase on main if needed
- Ensure CI is green

## Release Process

(For maintainers)

1. Update version in `Chart.yaml`, `go.mod`
2. Update `CHANGELOG.md`
3. Create release branch: `release/v0.x.x`
4. Tag release: `git tag v0.x.x`
5. Build and push Docker image
6. Publish Helm chart
7. Create GitHub release with notes

## Questions?

- Open an issue for questions
- Join discussions on GitHub
- Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to FinOps Enforcer! 🎉
