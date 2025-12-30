# Setting Up Coverage Badge with Codecov

This guide will help you set up the Codecov coverage badge for your repository.

## Overview

Codecov provides free coverage tracking and badges for open-source projects. It integrates seamlessly with GitHub Actions and provides detailed coverage reports, trends, and PR comments.

## Setup Steps

### 1. Sign Up for Codecov

1. Go to https://codecov.io
2. Click "Sign up with GitHub"
3. Authorize Codecov to access your GitHub account
4. Grant Codecov access to your repositories

### 2. Add Your Repository

For **public repositories**:
- Codecov will automatically detect your repository after the first workflow run
- No additional configuration needed!
- The workflow will upload coverage data automatically

For **private repositories**:
- You need to add a Codecov token (see step 3)

### 3. Configure Token (Private Repos Only)

If your repository is private:

1. **Get Your Codecov Token**
   - Navigate to https://codecov.io/gh/devskill-org/ems
   - Click on "Settings" in the left sidebar
   - Copy the "Repository Upload Token"

2. **Add Token to GitHub Secrets**
   - Go to your GitHub repository
   - Navigate to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: Paste the token from Codecov
   - Click "Add secret"

### 4. Verify Setup

1. **Commit and push** your changes (if you haven't already)
2. **Wait for workflow** to complete (check the Actions tab)
3. **Visit Codecov** dashboard at https://codecov.io/gh/devskill-org/ems
4. **Check your README** - the badge should now display the coverage percentage

## Badge Configuration

The coverage badge is already configured in your README.md:

```markdown
[![Coverage](https://codecov.io/gh/devskill-org/ems/branch/main/graph/badge.svg)](https://codecov.io/gh/devskill-org/ems)
```

If you use a different default branch (e.g., `master`), update the badge URL:

```markdown
[![Coverage](https://codecov.io/gh/devskill-org/ems/branch/master/graph/badge.svg)](https://codecov.io/gh/devskill-org/ems)
```

## Codecov Features

Once set up, Codecov provides:

### Coverage Dashboard
- Overall coverage percentage
- Coverage trends over time
- File-by-file coverage breakdown
- Uncovered lines highlighted

### Pull Request Comments
- Automatic comments on PRs showing coverage changes
- Coverage diff (increase/decrease)
- New uncovered lines highlighted
- Pass/fail status based on coverage thresholds

### Coverage Graphs
- Coverage history graphs
- Sunburst and tree visualizations
- Compare coverage across branches

### Notifications
- Slack/Email notifications for coverage changes
- Customizable coverage thresholds
- CI status checks

## Badge Customization

You can customize your badge appearance:

### Default Badge
```markdown
[![Coverage](https://codecov.io/gh/devskill-org/ems/branch/main/graph/badge.svg)](https://codecov.io/gh/devskill-org/ems)
```

### Flat Style Badge
```markdown
[![Coverage](https://codecov.io/gh/devskill-org/ems/branch/main/graph/badge.svg?style=flat)](https://codecov.io/gh/devskill-org/ems)
```

### Flat-Square Style Badge
```markdown
[![Coverage](https://codecov.io/gh/devskill-org/ems/branch/main/graph/badge.svg?style=flat-square)](https://codecov.io/gh/devskill-org/ems)
```

## Troubleshooting

### Badge Shows "unknown"

**Possible causes:**
- Workflow hasn't run yet
- Coverage upload failed
- Repository not found on Codecov

**Solutions:**
1. Check that the workflow completed successfully in the Actions tab
2. Verify the Codecov upload step didn't fail
3. For private repos, ensure `CODECOV_TOKEN` is set correctly
4. Visit Codecov dashboard to check for error messages

### Coverage Not Uploading

**Possible causes:**
- Token is incorrect (private repos)
- Network issues during upload
- Coverage file not generated

**Solutions:**
1. Check the workflow logs for the "Upload coverage to Codecov" step
2. Verify `coverage.out` file is being generated
3. For private repos, double-check the `CODECOV_TOKEN` secret
4. Try re-running the workflow

### Badge Not Updating

**Possible causes:**
- Browser cache
- Badge URL pointing to wrong branch
- Codecov processing delay

**Solutions:**
1. Clear your browser cache or try incognito mode
2. Verify the branch name in the badge URL matches your default branch
3. Wait a few minutes - Codecov may take time to process
4. Add a version parameter to force refresh: `?v=1`

### Private Repo: "401 Unauthorized"

**Solution:**
- Ensure `CODECOV_TOKEN` secret is set in GitHub repository settings
- Verify the token is correct (copy from Codecov settings)
- Token must have the correct permissions

## Configuration Options

### Codecov YAML Configuration (Optional)

You can add a `codecov.yml` file to your repository root for advanced configuration:

```yaml
# codecov.yml
coverage:
  status:
    project:
      default:
        target: 70%        # Target coverage
        threshold: 2%      # Allow 2% drop
    patch:
      default:
        target: 80%        # New code should have 80% coverage

comment:
  layout: "reach, diff, flags, files"
  behavior: default

ignore:
  - "test_data/**"
  - "**/*_test.go"
```

### Coverage Thresholds

Set minimum coverage requirements:

```yaml
coverage:
  status:
    project:
      default:
        target: auto
        threshold: 1%
        if_ci_failed: error
```

## Best Practices

1. **Set Realistic Targets**
   - Start with current coverage as baseline
   - Gradually increase target over time
   - Don't set unrealistic 100% targets

2. **Review Coverage Reports**
   - Check coverage in PRs before merging
   - Focus on covering critical paths
   - Use coverage to find untested code

3. **Monitor Trends**
   - Watch for coverage decreases
   - Celebrate coverage improvements
   - Use graphs to track progress

4. **Ignore Appropriate Files**
   - Test files don't need coverage
   - Generated code can be ignored
   - Vendor directories should be excluded

5. **Educate Team**
   - Share coverage reports with team
   - Discuss coverage in code reviews
   - Make coverage part of your workflow

## Additional Resources

- [Codecov Documentation](https://docs.codecov.com/)
- [Codecov GitHub Action](https://github.com/codecov/codecov-action)
- [Coverage Best Practices](https://docs.codecov.com/docs/common-recipe-list)
- [Codecov YAML Reference](https://docs.codecov.com/docs/codecov-yaml)

## Quick Setup Checklist

- [ ] Sign up at codecov.io with GitHub account
- [ ] Repository automatically detected after first workflow run
- [ ] For private repos: Add `CODECOV_TOKEN` to GitHub secrets
- [ ] Push changes and wait for workflow to complete
- [ ] Verify badge appears in README
- [ ] Visit Codecov dashboard to see detailed reports
- [ ] (Optional) Add codecov.yml for custom configuration
- [ ] (Optional) Set up coverage thresholds and notifications

---

**That's it!** Your coverage badge is now set up and will update automatically with each push. 🎉

For questions or issues, check the [Codecov Support](https://docs.codecov.com/) or open an issue in your repository.
