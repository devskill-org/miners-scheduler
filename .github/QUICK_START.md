# Quick Start Guide: CI/CD for Miners Scheduler

This guide will get your CI/CD pipeline running in less than 5 minutes.

## ✅ What's Already Working

Your repository now has:
- ✨ Automated tests on every PR
- 🔍 Code quality checks (linting)
- 🏗️ Build verification
- 📊 Coverage tracking

**No configuration needed** - these work out of the box!

## 🚀 Quick Setup (Optional Coverage Badge)

Want a coverage badge in your README showing the exact percentage?

### Codecov Setup (2-3 minutes)

1. Visit https://codecov.io and sign in with GitHub
2. Your repository will be automatically detected after the first workflow run
3. For private repos: Add `CODECOV_TOKEN` secret (get it from Codecov settings)
4. Done! Badge updates automatically

**That's it!** 🎉

For detailed instructions, see `.github/SETUP_COVERAGE_BADGE.md`

## 📖 Detailed Guides

- **Coverage badge setup**: `.github/SETUP_COVERAGE_BADGE.md`
- **CI/CD overview**: `.github/CI_CD_SETUP.md`
- **Contributing guide**: `CONTRIBUTING.md`
- **Workflow documentation**: `.github/workflows/README.md`

## 🧪 Test Locally Before Pushing

```bash
# Run all checks (same as CI)
make check test build

# Or individually
make fmt      # Format code
make vet      # Run go vet
make lint     # Run linter
make test     # Run tests with coverage
make build    # Build binary
```

## 👀 View Results

1. Go to your repo → **Actions** tab
2. Click any workflow run
3. See test results, coverage, and build status
4. Download artifacts if needed

## 🎯 What Triggers CI?

- ✅ Every push to `main`, `master`, or `develop`
- ✅ Every pull request to these branches

## 📊 Current Test Status

Check your badges at the top of README.md:
- **Tests**: Shows if tests are passing ✅
- **Coverage**: Shows test coverage percentage (via Codecov)
- **Go Report**: Overall code quality score
- **Go Version**: Go version used
- **License**: Project license

## ❓ Need Help?

- Tests failing? Check the Actions logs
- Badge not working? See `.github/SETUP_COVERAGE_BADGE.md`
- Contributing? Read `CONTRIBUTING.md`
- Questions? Open an issue

## 🎉 You're All Set!

Your CI/CD pipeline is ready. Every PR will now be:
- ✅ Tested automatically
- ✅ Linted for quality
- ✅ Verified to build

Happy coding! 🚀