# Contributing to Stargazer

Thank you for your interest in contributing to Stargazer! We're excited to have you join our community of developers working to make Kubernetes troubleshooting easier for everyone.

This document provides guidelines and instructions for contributing to the project. By participating in this project, you agree to abide by our Code of Conduct.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Commit Message Format](#commit-message-format)
- [Branch Naming Conventions](#branch-naming-conventions)
- [Documentation Requirements](#documentation-requirements)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)
- [Questions and Support](#questions-and-support)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please treat others with respect and consideration. We expect all contributors to:

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

Unacceptable behavior includes harassment, trolling, insulting comments, and any conduct that would be inappropriate in a professional setting.

## Getting Started

Before you begin contributing, please:

1. Read this document thoroughly
2. Review the [README.md](./README.md) to understand the project structure and goals
3. Check existing [Issues](../../issues) to see if your bug or feature has already been reported
4. Fork the repository to your own GitHub account
5. Clone your fork locally

## Development Setup

### Prerequisites

Ensure you have the following installed:

- **Go 1.21 or higher** - [Install Go](https://golang.org/doc/install)
- **Node.js 16 or higher** - [Install Node.js](https://nodejs.org/)
- **npm or yarn** - Comes with Node.js
- **Wails CLI v2** - Required for desktop app development
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **kubectl** - For testing with real Kubernetes clusters
- **golangci-lint** (optional but recommended) - For code linting
  ```bash
  # macOS
  brew install golangci-lint

  # Linux
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```
- **air** (optional) - For hot-reloading during development
  ```bash
  go install github.com/air-verse/air@latest
  ```

### Initial Setup

1. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/stargazer.git
   cd stargazer
   ```

2. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/maplecitymadman/stargazer.git
   ```

3. **Install Go dependencies:**
   ```bash
   make deps
   # or manually
   go mod download
   go mod tidy
   ```

4. **Install frontend dependencies:**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

5. **Verify your setup:**
   ```bash
   # Build CLI
   make build

   # Run tests
   make test

   # Build desktop app (optional)
   make build-gui
   ```

### Ensure PATH is Configured

Make sure `$(go env GOPATH)/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add this to your `.bashrc`, `.zshrc`, or equivalent shell configuration file.

## Development Workflow

### Running the Application

**CLI Development:**
```bash
# Run CLI directly
make run

# Or with auto-reload
make dev
```

**Desktop App Development:**
```bash
# Run desktop app in development mode with hot-reload
make dev-gui
```

**Frontend Only:**
```bash
cd frontend
npm run dev
```

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write clean, readable code
   - Follow the code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes:**
   ```bash
   # Run all tests
   make test

   # Run tests with coverage
   make test-coverage

   # Run linter
   make lint

   # Format code
   make fmt
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

5. **Keep your branch up to date:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

6. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Style Guidelines

### Go Code Style

We follow standard Go conventions and best practices:

1. **Formatting:**
   - Use `gofmt` for code formatting (run `make fmt`)
   - All Go code must be formatted before committing
   - Line length should be reasonable (typically under 120 characters)

2. **Naming Conventions:**
   - Use camelCase for variable and function names
   - Use PascalCase for exported functions and types
   - Use descriptive names (avoid single-letter variables except in short loops)
   - Package names should be lowercase, single-word

3. **Code Organization:**
   - Group related functionality together
   - Keep functions focused and single-purpose
   - Use interfaces to define behavior
   - Minimize package dependencies

4. **Error Handling:**
   - Always check and handle errors
   - Return errors rather than panicking
   - Wrap errors with context using `fmt.Errorf` with `%w`
   - Use descriptive error messages

5. **Comments:**
   - Add godoc comments for all exported functions and types
   - Comment format: `// FunctionName does something specific`
   - Explain "why" not "what" in inline comments
   - Keep comments up to date with code changes

6. **Imports:**
   - Group imports: standard library, external packages, internal packages
   - Use goimports or organize manually

Example:
```go
// GetPodStatus retrieves the current status of a pod in the specified namespace.
// It returns an error if the pod cannot be found or if there's a connection issue.
func GetPodStatus(ctx context.Context, namespace, name string) (*PodStatus, error) {
    if namespace == "" {
        return nil, fmt.Errorf("namespace cannot be empty")
    }

    pod, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, name, err)
    }

    return convertPodStatus(pod), nil
}
```

### Frontend Code Style

1. **TypeScript:**
   - Use TypeScript for all new code
   - Define proper types and interfaces
   - Avoid `any` type unless absolutely necessary

2. **React Components:**
   - Use functional components with hooks
   - Keep components focused and reusable
   - Use meaningful component and prop names

3. **Formatting:**
   - Run `npm run lint` before committing
   - Follow Next.js and React best practices

### Linting

Run the linter before submitting your PR:

```bash
# Go linting
make lint

# Frontend linting
cd frontend && npm run lint
```

If `golangci-lint` reports issues, fix them before submitting. Common issues include:
- Unused variables or imports
- Error handling violations
- Code complexity warnings
- Inefficient code patterns

## Testing Requirements

All contributions must include appropriate tests:

### Go Tests

1. **Unit Tests:**
   - Write tests for all new functions and methods
   - Test both success and failure cases
   - Use table-driven tests for multiple scenarios
   - Aim for at least 70% code coverage

2. **Test Location:**
   - Place tests in `*_test.go` files alongside the code
   - Use package-level test files for integration tests

3. **Running Tests:**
   ```bash
   # Run all tests
   make test

   # Run tests with coverage
   make test-coverage

   # Run specific package tests
   go test -v ./internal/k8s/...
   ```

4. **Test Example:**
   ```go
   func TestGetPodStatus(t *testing.T) {
       tests := []struct {
           name      string
           namespace string
           podName   string
           wantErr   bool
       }{
           {"valid pod", "default", "test-pod", false},
           {"empty namespace", "", "test-pod", true},
           {"empty pod name", "default", "", true},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               _, err := GetPodStatus(context.Background(), tt.namespace, tt.podName)
               if (err != nil) != tt.wantErr {
                   t.Errorf("GetPodStatus() error = %v, wantErr %v", err, tt.wantErr)
               }
           })
       }
   }
   ```

### Frontend Tests

- Add tests for complex components and utilities
- Use Jest and React Testing Library
- Test user interactions and edge cases

### Test Coverage

- Aim for at least 70% code coverage for new code
- View coverage report: `make test-coverage` (opens `coverage.html`)
- Critical paths should have higher coverage

## Pull Request Process

### Before Submitting

1. Ensure your code passes all tests: `make test`
2. Run the linter and fix any issues: `make lint`
3. Format your code: `make fmt`
4. Update documentation if needed
5. Verify your branch is up to date with `main`
6. Test the CLI and desktop app builds locally

### Submitting a Pull Request

1. **Push your branch** to your fork on GitHub

2. **Create a Pull Request** from your fork to the main repository

3. **Fill out the PR template** with:
   - Clear description of changes
   - Link to related issues (use "Closes #123" or "Fixes #123")
   - Screenshots or GIFs for UI changes
   - Testing steps and results
   - Any breaking changes or migration notes

4. **PR Title Format:**
   - Use conventional commit format: `type: description`
   - Examples: `feat: add pod logs filtering`, `fix: resolve memory leak in scanner`

5. **Checklist:**
   - [ ] Code follows project style guidelines
   - [ ] All tests pass locally
   - [ ] New tests added for new functionality
   - [ ] Documentation updated (if applicable)
   - [ ] Commit messages follow convention
   - [ ] No merge conflicts with main branch

### Review Process

- Maintainers will review your PR within a few days
- Address any feedback or requested changes
- Keep discussions respectful and constructive
- Once approved, a maintainer will merge your PR

### After Merge

- Delete your feature branch
- Pull the latest changes from upstream
- Your contribution will be included in the next release

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

Must be one of:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Code style changes (formatting, missing semi-colons, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Changes to build process, tooling, dependencies, etc.
- **ci**: Changes to CI/CD configuration

### Scope (Optional)

The scope should specify the area of the codebase:

- `cli`: CLI-related changes
- `gui`: Desktop app changes
- `k8s`: Kubernetes client changes
- `api`: API server changes
- `config`: Configuration management
- `frontend`: Frontend/UI changes
- `docs`: Documentation

### Examples

```
feat(k8s): add support for StatefulSet scanning

- Implement StatefulSet discovery
- Add tests for StatefulSet status checks
- Update API to expose StatefulSet information

Closes #42
```

```
fix(cli): resolve panic when kubeconfig is missing

Check for kubeconfig existence before attempting to load.
Add helpful error message pointing to setup documentation.

Fixes #128
```

```
docs: update installation instructions for Windows

Add PowerShell installation steps and troubleshooting section.
```

```
chore(deps): update kubernetes client-go to v0.35.0
```

### Subject Guidelines

- Use imperative mood ("add feature" not "added feature")
- Don't capitalize first letter
- No period at the end
- Keep under 72 characters

## Branch Naming Conventions

Use descriptive branch names that indicate the type of work:

### Format

```
<type>/<short-description>
```

### Types

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions or updates
- `chore/` - Maintenance tasks

### Examples

```
feature/pod-logs-filtering
fix/memory-leak-scanner
docs/contributing-guide
refactor/k8s-client-interface
test/add-deployment-tests
chore/update-dependencies
```

### Guidelines

- Use lowercase with hyphens
- Be descriptive but concise
- Include issue number if applicable: `fix/128-kubeconfig-panic`

## Documentation Requirements

Good documentation is essential for project maintainability:

### Code Documentation

1. **Godoc Comments:**
   - All exported functions, types, and constants must have godoc comments
   - Start with the name of the item being documented
   - Be clear and concise

2. **Complex Logic:**
   - Add comments explaining non-obvious logic
   - Document algorithm choices and trade-offs
   - Explain "why" not just "what"

### User Documentation

When adding features, update:

1. **README.md** - For major features or changes to setup/usage
2. **Command Help** - Update CLI help text for new commands
3. **API Documentation** - Document new API endpoints or changes

### Pull Request Documentation

Include in your PR description:

- What problem does this solve?
- How does it work?
- Any breaking changes?
- Migration steps (if applicable)
- Screenshots/GIFs for UI changes

## Reporting Bugs

Found a bug? Help us fix it!

### Before Reporting

1. Check if the bug has already been reported in [Issues](../../issues)
2. Verify you're using the latest version
3. Test with a minimal reproduction case

### Bug Report Template

Create a new issue with the following information:

**Title:** Brief, descriptive title

**Description:**
- What happened?
- What did you expect to happen?
- Steps to reproduce
- Environment details (OS, Go version, Kubernetes version)
- Error messages or logs
- Screenshots (if applicable)

**Example:**

```
**Bug:** CLI crashes when scanning namespace with no pods

**Steps to Reproduce:**
1. Create empty namespace: `kubectl create namespace empty`
2. Run: `stargazer scan --namespace empty`
3. See crash with panic

**Expected:** Should report "No issues found" or "No pods in namespace"

**Environment:**
- OS: macOS 13.5
- Stargazer: v0.1.0-dev
- Go: 1.21.0
- Kubernetes: v1.28.0

**Logs:**
```
panic: runtime error: index out of range
...
```
```

## Feature Requests

Have an idea for a new feature?

### Before Requesting

1. Check existing [Issues](../../issues) to see if it's been suggested
2. Consider if it fits the project's goals and scope
3. Think about implementation complexity and maintenance

### Feature Request Template

Create a new issue with:

**Title:** Clear, concise feature description

**Description:**
- What problem does this solve?
- Who would benefit from this feature?
- Proposed solution or implementation ideas
- Alternative solutions considered
- Additional context or screenshots

**Example:**

```
**Feature:** Add support for filtering issues by severity

**Problem:**
Users need to focus on critical issues first but the UI shows all issues
with equal priority.

**Solution:**
Add severity levels (Critical, Warning, Info) to issues and allow filtering
in the UI. Could use color coding (red, yellow, blue) for visual distinction.

**Benefits:**
- Helps users prioritize troubleshooting efforts
- Reduces noise from low-priority issues
- Aligns with industry best practices

**Alternatives:**
- Sort by severity (less flexible)
- Add tags/labels (more complex)
```

## Questions and Support

Need help or have questions?

- **Documentation:** Check the [README.md](./README.md) first
- **Discussions:** Use [GitHub Discussions](../../discussions) for questions and community interaction
- **Issues:** For bugs and feature requests only
- **Chat:** Join our community chat (if available)

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Wails Documentation](https://wails.io/docs/introduction)
- [Next.js Documentation](https://nextjs.org/docs)
- [Kubernetes client-go](https://github.com/kubernetes/client-go)
- [Conventional Commits](https://www.conventionalcommits.org/)

## License

By contributing to Stargazer, you agree that your contributions will be licensed under the same [MIT License](./LICENSE) that covers the project.

---

Thank you for contributing to Stargazer! Your efforts help make Kubernetes troubleshooting easier for everyone. If you have any questions about contributing, feel free to ask in the discussions or open an issue.

Happy coding!
