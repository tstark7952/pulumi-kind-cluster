# Contributing to Kubernetes Local Cluster with Pulumi

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/pulumi-kind-cluster.git
   cd pulumi-kind-cluster
   ```
3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/tstark7952/pulumi-kind-cluster.git
   ```

## Development Setup

### Prerequisites
- macOS (Apple Silicon or Intel)
- Homebrew
- Go 1.24+
- Pulumi CLI
- Lima
- Kind
- kubectl

### Install Dependencies
```bash
brew install pulumi go lima kind kubectl
go mod download
```

### Set Up Pulumi
```bash
echo "your-test-passphrase" > .pulumi-passphrase
chmod 600 .pulumi-passphrase
export PULUMI_CONFIG_PASSPHRASE_FILE="$(pwd)/.pulumi-passphrase"
pulumi stack init dev
```

## Making Changes

### Branch Naming Convention
Use descriptive branch names:
- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/modifications

Example:
```bash
git checkout -b feat/add-worker-nodes
```

### Commit Messages
Follow conventional commits format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

Example:
```
feat(cluster): add support for custom CNI plugins

Implemented configuration option to allow users to specify
alternative CNI plugins beyond Calico.

Closes #123
```

## Submitting Changes

### Before Submitting
1. **Test your changes:**
   ```bash
   pulumi up
   # Verify cluster works
   pulumi destroy
   ```

2. **Update documentation** if needed

3. **Run linting:**
   ```bash
   go fmt ./...
   go vet ./...
   ```

4. **Ensure no secrets** are committed:
   ```bash
   git diff --cached | grep -i "passphrase\|secret\|key"
   ```

### Pull Request Process

1. **Push to your fork:**
   ```bash
   git push origin feat/your-feature
   ```

2. **Open a Pull Request** on GitHub

3. **Fill out the PR template** completely

4. **Link related issues** using "Closes #123"

5. **Respond to review feedback** promptly

### PR Review Criteria
- [ ] Code follows project style and conventions
- [ ] Changes are well-tested
- [ ] Documentation is updated
- [ ] No merge conflicts with main
- [ ] Health checks pass after deployment
- [ ] No security vulnerabilities introduced

## Coding Standards

### Go Code Style
- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for complex logic
- Handle errors explicitly
- Use meaningful variable names

### Pulumi Best Practices
- Use resource dependencies appropriately
- Implement proper deletion handlers
- Make operations idempotent
- Add descriptive resource names
- Use configuration for customizable values

### Documentation Style
- Use clear, concise language
- Include code examples
- Update README for user-facing changes
- Add inline comments for complex code
- Keep documentation in sync with code

## Testing

### Manual Testing Checklist
Test your changes with:
- [ ] Fresh installation (`pulumi up` from scratch)
- [ ] Upgrade scenario (existing cluster)
- [ ] Destruction (`pulumi destroy`)
- [ ] Multiple runs (idempotency)
- [ ] Different configurations (CPU, memory, disk)
- [ ] Health checks pass
- [ ] Both Apple Silicon and Intel (if possible)

### Testing Commands
```bash
# Deploy
pulumi up -y

# Verify health
kubectl get nodes
kubectl get pods -A

# Check specific features
limactl list
docker context ls
kind get clusters

# Clean up
pulumi destroy -y
```

### Edge Cases to Test
- VM already exists
- Cluster already exists
- Partial deployment recovery
- Network interruptions
- Insufficient resources

## Areas for Contribution

We welcome contributions in these areas:

### High Priority
- 🐛 Bug fixes
- 📖 Documentation improvements
- 🧪 Test coverage
- 🔒 Security enhancements

### Feature Ideas
- Additional CNI support (Flannel, Cilium)
- Windows/Linux support (via different VMs)
- Custom node configurations
- Monitoring/observability stack
- Helm chart deployment examples
- Multi-cluster support
- Resource optimization

### Documentation Needs
- Video tutorials
- Troubleshooting guide expansion
- Architecture diagrams
- Use case examples
- Performance tuning guide

## Questions?

- Open an issue for discussion
- Check existing issues and PRs
- Review documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🎉
