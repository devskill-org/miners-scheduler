# Contributing to Miners Scheduler

Thank you for your interest in contributing to Miners Scheduler! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Running Tests](#running-tests)
- [Code Quality](#code-quality)
- [CI/CD Pipeline](#cicd-pipeline)
- [Pull Request Process](#pull-request-process)
- [Code Coverage](#code-coverage)
- [Coding Standards](#coding-standards)

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/miners-scheduler.git
   cd miners-scheduler
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/devskill-org/miners-scheduler.git
   ```

## Development Setup

### Prerequisites

- Go 1.25.1 or later
- Make (optional, but recommended)
- golangci-lint (for linting)

### Install Dependencies

```bash
# Using Make
make deps

# Or manually
go mod download
go mod tidy
```

### Install Development Tools

```bash
# Using Make
make setup

# Or manually
go install golang.org/x/lint/golint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Running Tests

### Run All Tests

```bash
# Using Make
make test

# Or manually
go test -v -race -coverprofile=coverage.out ./...
```

### Run Tests with Coverage Report

```bash
# Using Make
make test-coverage

# Or manually
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Tests

```bash
# Run tests in a specific package
go test -v ./scheduler/

# Run a specific test
go test -v -run TestSchedulerRun ./scheduler/
```

## Code Quality

### Format Code

```bash
# Using Make
make fmt

# Or manually
gofmt -s -w .
go mod tidy
```

### Run Linters

```bash
# Using Make
make lint

# Or manually
golangci-lint run
```

### Run Vet

```bash
# Using Make
make vet

# Or manually
go vet ./...
```

### Run All Checks

```bash
make check
```

## CI/CD Pipeline

Our CI/CD pipeline automatically runs on:
- Every push to `main`, `master`, or `develop` branches
- Every pull request targeting these branches

### What the CI Pipeline Does

1. **Test Job**:
   - Runs all unit tests
   - Enables race detection
   - Generates code coverage reports
   - Uploads coverage to Codecov (if configured)
   - Creates coverage badge
   - Archives coverage reports

2. **Lint Job**:
   - Runs golangci-lint with comprehensive checks
   - Ensures code quality standards

3. **Build Job**:
   - Compiles the binary
   - Verifies successful build
   - Archives build artifacts

### Viewing CI Results

1. Navigate to the "Actions" tab in the GitHub repository
2. Click on the workflow run you want to inspect
3. Review job results and logs
4. Download artifacts if needed

### Local CI Simulation

Before pushing, you can simulate CI checks locally:

```bash
# Run the same checks as CI
make check test build

# Or step by step
make fmt      # Format code
make vet      # Run go vet
make lint     # Run golangci-lint
make test     # Run tests with coverage
make build    # Build binary
```

## Pull Request Process

### Before Submitting

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit them:
   ```bash
   git add .
   git commit -m "Description of your changes"
   ```

3. **Run all checks locally**:
   ```bash
   make check test
   ```

4. **Update tests** if you've added functionality

5. **Update documentation** if needed

6. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Submitting the PR

1. Push your branch:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a Pull Request on GitHub

3. Fill out the PR template with:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if applicable)

4. Wait for CI checks to pass (all three jobs must succeed)

5. Request review from maintainers

### PR Requirements

- ✅ All CI checks must pass
- ✅ Code coverage should not decrease
- ✅ All tests must pass
- ✅ Code must be formatted (`make fmt`)
- ✅ No linting errors
- ✅ Documentation updated if needed
- ✅ At least one approving review

## Code Coverage

We aim to maintain high code coverage (>70%). 

### Coverage Guidelines

- New features should include tests
- Bug fixes should include regression tests
- Critical paths should have >90% coverage
- Coverage should not decrease with new PRs

### Viewing Coverage Locally

```bash
# Generate and view coverage
make test-coverage

# Open coverage.html in your browser
```

### Coverage in CI

- Coverage is automatically calculated on every PR
- Coverage badge is updated on main branch
- Coverage reports are archived as artifacts

## Coding Standards

### Go Best Practices

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly

### Code Organization

```
miners-scheduler/
├── entsoe/          # ENTSO-E API client
├── meteo/           # Weather/meteorological data
├── miners/          # Miner control logic
├── mpc/             # Model Predictive Control
├── scheduler/       # Main scheduler logic
├── sigenergy/       # Energy signature analysis
├── sun/             # Solar calculations
├── utils/           # Utility functions
└── test_data/       # Test fixtures
```

### Testing Standards

- Write table-driven tests where appropriate
- Use descriptive test names: `TestFunctionName_Scenario_ExpectedResult`
- Mock external dependencies
- Test error paths, not just happy paths
- Use `t.Run()` for subtests

Example:
```go
func TestScheduler_Run_WithHighPrice_StopsMiners(t *testing.T) {
    // Arrange
    scheduler := NewScheduler(...)
    
    // Act
    err := scheduler.Run()
    
    // Assert
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Commit Messages

Follow conventional commits format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `style`: Formatting changes
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

Example:
```
feat(scheduler): add thermal protection for miners

Implements automatic mode switching when fan speeds exceed 70%
to prevent overheating. Recovers to standard mode when temps normalize.

Closes #123
```

## Getting Help

- Open an issue for bugs or feature requests
- Use discussions for questions
- Check existing issues and PRs before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing! 🎉