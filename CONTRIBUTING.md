# Contributing to OpenVPN Client

Thank you for your interest in contributing to OpenVPN Client! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Commit Messages](#commit-messages)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Please:

- Be respectful and considerate in all interactions
- Accept constructive criticism gracefully
- Focus on what is best for the community and project
- Show empathy towards other community members

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openvpn-client.git
   cd openvpn-client
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/tldr-it-stepankutaj/openvpn-client.git
   ```
4. **Keep your fork synced**:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Make
- [OpenVPN Manager](https://github.com/tldr-it-stepankutaj/openvpn-mng) running locally (for integration testing)

### Setup Steps

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build all binaries**:
   ```bash
   make build
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Run linter** (requires golangci-lint):
   ```bash
   make lint
   ```

### Project Structure

```
openvpn-client/
├── cmd/
│   ├── login/          # auth-user-pass-verify binary
│   ├── connect/        # client-connect binary
│   ├── disconnect/     # client-disconnect binary
│   └── firewall/       # firewall rules generator binary
├── internal/
│   ├── api/            # API client for OpenVPN Manager
│   ├── config/         # Configuration handling
│   ├── firewall/       # Firewall rule generators (nftables, iptables)
│   ├── logger/         # Structured logging
│   └── utils/          # Utility functions
├── samples/            # Sample configurations
└── help/               # Documentation
```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check if the issue already exists. When creating a bug report, include:

- **Clear title** describing the issue
- **Steps to reproduce** the behavior
- **Expected behavior** vs **actual behavior**
- **Environment details** (OS, Go version, OpenVPN version)
- **Relevant logs** or error messages (JSON log output)
- **Configuration** (sanitized, without tokens/passwords)

Use the GitHub issue tracker: [Create a new issue](https://github.com/tldr-it-stepankutaj/openvpn-client/issues/new)

### Suggesting Features

Feature requests are welcome! Please:

- Check if the feature has already been requested
- Provide a clear description of the feature
- Explain the use case and benefits
- Consider if it aligns with the project's goals

### Contributing Code

1. **Find an issue** to work on, or create one for discussion
2. **Comment on the issue** to let others know you're working on it
3. **Create a feature branch** from `main`
4. **Make your changes** following our coding standards
5. **Write/update tests** for your changes
6. **Submit a pull request**

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests**:
   ```bash
   make test
   ```

3. **Run linter**:
   ```bash
   make vet
   ```

4. **Format code**:
   ```bash
   make fmt
   ```

5. **Build successfully**:
   ```bash
   make build
   ```

6. **Test on Linux** (if possible):
   ```bash
   make build-linux
   ```

### Creating the Pull Request

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR on GitHub** with:
   - Clear, descriptive title
   - Reference to related issue(s) using `Fixes #123` or `Relates to #123`
   - Description of changes made
   - Any breaking changes noted

3. **PR Template**:
   ```markdown
   ## Description
   Brief description of changes

   ## Related Issue
   Fixes #(issue number)

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-reviewed the code
   - [ ] Added/updated comments for complex logic
   - [ ] Updated documentation if needed
   - [ ] Added tests if applicable
   - [ ] All tests pass locally
   - [ ] Tested with OpenVPN Manager API
   ```

### Review Process

- PRs require at least one approval before merging
- Address all review comments
- Keep the PR focused and reasonably sized
- Be patient - reviews may take time

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://go.dev/doc/effective_go).

### Project-Specific Guidelines

#### Naming Conventions

- **Files**: lowercase with underscores (`api_client.go`)
- **Packages**: lowercase, single word (`api`, `config`, `firewall`)
- **Interfaces**: descriptive names (`Firewall`)
- **Structs**: PascalCase (`Client`, `Config`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported

#### Error Handling

```go
// Good - wrap errors with context
if err != nil {
    return fmt.Errorf("failed to validate user: %w", err)
}

// Good - use specific error messages
return fmt.Errorf("user not found: %s", username)
```

#### Logging

Use structured logging with `log/slog`:

```go
log.Info("user authenticated",
    "username", username,
    "user_id", userID,
)

log.Error("authentication failed",
    "username", username,
    "error", err,
)
```

#### Context Usage

Pass context to functions that make API calls:

```go
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
    // ...
}
```

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style (formatting, etc.) |
| `refactor` | Code refactoring |
| `test` | Adding/updating tests |
| `chore` | Maintenance tasks |

### Scopes

| Scope | Description |
|-------|-------------|
| `api` | API client changes |
| `config` | Configuration handling |
| `firewall` | Firewall rule generation |
| `login` | Login command |
| `connect` | Connect command |
| `disconnect` | Disconnect command |
| `deps` | Dependencies |

### Examples

```bash
feat(firewall): add support for ipset

fix(api): handle connection timeout properly

docs(readme): add troubleshooting section

refactor(config): extract validation logic

chore(deps): update gopkg.in/yaml.v3 to v3.0.1
```

### Guidelines

- Use imperative mood: "add" not "added" or "adds"
- Keep subject line under 72 characters
- Reference issues in the footer: `Fixes #123`

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/api/...

# Run with coverage
make test-coverage
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple cases
- Mock HTTP responses for API tests
- Test both success and error paths

```go
func TestClient_ValidateVpnUser(t *testing.T) {
    tests := []struct {
        name     string
        username string
        password string
        wantErr  bool
    }{
        {
            name:     "valid credentials",
            username: "testuser",
            password: "testpass",
            wantErr:  false,
        },
        {
            name:     "invalid credentials",
            username: "baduser",
            password: "badpass",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Code Comments

- Comment exported functions and types
- Explain "why", not "what"
- Keep comments up to date with code changes

### README Updates

Update README.md when:
- Adding new features
- Changing configuration options
- Modifying CLI arguments
- Updating dependencies

### Sample Files

Update sample files in `samples/` when:
- Changing configuration format
- Adding new firewall features
- Modifying OpenVPN integration

## Questions?

If you have questions about contributing:

1. Check existing [issues](https://github.com/tldr-it-stepankutaj/openvpn-client/issues)
2. Open a new issue with your question
3. Tag it with the `question` label

## Related Projects

- [OpenVPN Manager](https://github.com/tldr-it-stepankutaj/openvpn-mng) - Web-based administration for VPN users, networks, and groups

Thank you for contributing to OpenVPN Client!
