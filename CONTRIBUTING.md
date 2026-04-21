# Contributing to Velo Deploy

Thank you for your interest in contributing to Velo Deploy!

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

Before submitting a bug report:

1. Check existing [issues](https://github.com/antojsh/velo-deploy/issues) to avoid duplicates
2. Verify the bug exists in the latest version
3. Include:
   - Your OS and Linux distribution
   - Go version (`go version`)
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant log output

### Suggesting Features

We welcome feature suggestions! Please:

1. Search existing issues first
2. Describe the use case and motivation
3. Explain how it should work
4. Consider backward compatibility

### Pull Requests

#### Process

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/your-username/velo-deploy.git
   cd velo-deploy
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes**
5. **Test** your changes:
   ```bash
   go test ./...
   go vet ./...
   go fmt ./...
   ```
6. **Commit** with clear messages:
   ```bash
   git commit -m "Add: brief description of changes"
   ```
7. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
8. Open a **Pull Request**

#### Guidelines

- Follow Go standard conventions
- Run `go fmt` on all code
- Ensure `go vet` passes without warnings
- Write tests for new functionality
- Update documentation for user-facing changes
- Keep commits atomic and focused
- Write clear commit messages

#### Commit Message Format

```
<type>: <short description>

<longer description if needed>

<optional: related issue #>
```

**Types:**
- `Add:` New feature
- `Fix:` Bug fix
- `Update:` Update existing functionality
- `Refactor:` Code refactoring
- `Docs:` Documentation changes
- `Test:` Test additions/changes

### Development Setup

```bash
# Install Go 1.22 or later
# Ubuntu/Debian:
sudo apt update && sudo apt install golang-go

# Verify installation
go version

# Clone the repo
git clone https://github.com/antojsh/velo-deploy.git
cd velo-deploy

# Build
go build -o velo-deploy ./cmd/deploy

# Run tests
go test ./...

# Run with sample config (for TUI testing)
./velo-deploy
```

### Testing

Run tests before submitting:

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Code Review Process

1. All submissions require review via PR
2. Reviewers may request changes
3. Once approved, maintainers will merge
4. Be responsive to feedback

## Areas for Contribution

- [ ] Static site rebuild functionality
- [ ] Additional static site generators support
- [ ] Health check endpoints
- [ ] Metrics/monitoring integration
- [ ] Database migration tooling
- [ ] Multi-server support
- [ ] CI/CD integration examples

## Questions?

Open an issue for general questions or discussions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
