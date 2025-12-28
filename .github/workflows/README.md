# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated testing, code coverage, and continuous integration.

## Workflows

### Tests and Coverage (`test.yml`)

This workflow runs on every push to main/master/develop branches and on all pull requests.

**Jobs:**
- **test**: Runs all unit tests with race detection and generates code coverage
- **lint**: Runs golangci-lint for code quality checks
- **build**: Builds the binary to ensure compilation succeeds

**Features:**
- ✅ Runs all unit tests with `go test`
- ✅ Race condition detection
- ✅ Code coverage measurement
- ✅ Uploads coverage to Codecov
- ✅ Archives coverage reports as artifacts

## Setup Instructions

### Basic Setup (Works Immediately)

The `test.yml` workflow will work out of the box for:
- Running tests on PRs and pushes
- Generating coverage reports
- Linting code
- Building the binary

Simply commit the workflow files and they will run automatically.

### Codecov Integration (Optional)

To get coverage badges and detailed reports at [Codecov](https://codecov.io):

1. **Sign up at Codecov**
   - Go to https://codecov.io
   - Sign in with your GitHub account

2. **Add your repository**
   - Codecov will automatically detect your repository after the first workflow run
   - For public repositories, no token is required
   - For private repositories, continue to step 3

3. **Add Codecov Token (Private Repos Only)**
   - Go to your repository settings on Codecov
   - Copy your Codecov upload token
   - Add the token as a GitHub secret:
     - Go to your repository → Settings → Secrets and variables → Actions
     - Click "New repository secret"
     - Name: `CODECOV_TOKEN`
     - Value: Your Codecov token

4. **View Coverage Reports**
   - After the first workflow run, visit https://codecov.io/gh/devskill-org/miners-scheduler
   - View detailed coverage reports, trends, and file-by-file analysis
   - The badge in README.md will automatically update

**Badge:**
```markdown
[![Coverage](https://codecov.io/gh/devskill-org/miners-scheduler/branch/main/graph/badge.svg)](https://codecov.io/gh/devskill-org/miners-scheduler)
```

### Alternative: Simple Test Status Badge

If you don't want to use Codecov, you can use just the test status badge:

```markdown
[![Tests](https://github.com/devskill-org/miners-scheduler/actions/workflows/test.yml/badge.svg)](https://github.com/devskill-org/miners-scheduler/actions/workflows/test.yml)
```

## Viewing Coverage Reports

### In Codecov
1. Visit https://codecov.io/gh/devskill-org/miners-scheduler
2. View coverage percentage, trends, and file-by-file breakdown
3. See coverage changes in pull requests

### In GitHub Actions
1. Go to the Actions tab in your repository
2. Click on a workflow run
3. Scroll to "Artifacts" section
4. Download `code-coverage-report`

### Locally
Run the coverage report locally:

```bash
# Generate coverage
make test-coverage

# Or manually
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Then open `coverage.html` in your browser.

## Workflow Triggers

### test.yml
- ✅ Push to `main`, `master`, or `develop` branches
- ✅ Pull requests targeting `main`, `master`, or `develop`

## Status Badges

Add these badges to your README.md:

```markdown
[![Tests](https://github.com/devskill-org/miners-scheduler/actions/workflows/test.yml/badge.svg)](https://github.com/devskill-org/miners-scheduler/actions/workflows/test.yml)
[![Coverage](https://codecov.io/gh/devskill-org/miners-scheduler/branch/main/graph/badge.svg)](https://codecov.io/gh/devskill-org/miners-scheduler)
[![Go Report Card](https://goreportcard.com/badge/github.com/devskill-org/miners-scheduler)](https://goreportcard.com/report/github.com/devskill-org/miners-scheduler)
```

## Troubleshooting

### Tests fail locally but pass in CI (or vice versa)
- Ensure Go version matches between local and CI (currently 1.25.1)
- Check for race conditions with `go test -race ./...`
- Clear test cache: `go clean -testcache`

### Coverage not uploading to Codecov
- Check that workflow ran successfully
- For private repos, verify `CODECOV_TOKEN` secret is set correctly
- Check Codecov dashboard for error messages
- Ensure repository is added to Codecov

### Lint failures
- Run locally: `golangci-lint run`
- Install: `brew install golangci-lint` (macOS) or see https://golangci-lint.run/usage/install/
- Fix issues or adjust `.golangci.yml` configuration

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Codecov Documentation](https://docs.codecov.com/)
- [golangci-lint Documentation](https://golangci-lint.run/)