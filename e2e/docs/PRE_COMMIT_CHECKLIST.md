# Pre-Commit Checklist

Before committing the E2E test suite to the repository, verify:

## ✅ Core Test Suite

- [ ] All test files compile: `cd e2e && go mod tidy && go build ./...`
- [ ] Tests can be built: `cd e2e && go test -c`
- [ ] No syntax errors: `go vet ./...`
- [ ] Go module is valid: `go mod verify`

## ✅ Documentation

- [ ] All .md files are present and readable
- [ ] No TODO/FIXME placeholders left in docs
- [ ] Links between documents work (especially in START_HERE.md)
- [ ] Code examples in docs are correct

## ✅ Configuration

- [ ] `.env.example` has all required variables
- [ ] `.gitignore` excludes `.env` and test artifacts
- [ ] No secrets or credentials in any files
- [ ] No hardcoded URLs/IPs (except localhost and examples)

## ✅ CI Integration (if applying patch)

- [ ] Patch file syntax is correct
- [ ] Pipeline YAML is valid
- [ ] No tabs in YAML (spaces only)
- [ ] Environment variables are properly quoted

## ✅ Repository Cleanliness

- [ ] No compiled binaries (*.test files)
- [ ] No temporary files
- [ ] No personal .env files
- [ ] Only intended files staged for commit

## Quick Validation Commands

```bash
# Check what will be committed
git status
git diff --cached

# Validate Go code
cd e2e
go mod tidy
go vet ./...
go test -c
cd ..

# Check for secrets
git diff --cached | grep -i "password\|secret\|token" | grep -v "example\|template"

# Check file count
git status --short | wc -l
```

## Recommended Commit Strategy

### Option 1: Single Commit (Simpler)
```bash
git add e2e/ .tekton/ Makefile AGENTS.md
git commit -m "feat: Add E2E API test suite with Konflux CI integration

- Complete E2E test suite (839 lines) covering all major endpoints
- Two integration paths: in-pipeline (immediate) and Bonfire (future)
- Comprehensive documentation and implementation guides
- Ready-to-apply patch for in-pipeline integration

Test coverage:
- Identity endpoints (7 scenarios)
- Favorite pages (3 scenarios)
- Last visited pages (3 scenarios)
- Recently used workspaces (6+ scenarios)
- Self report (2 scenarios)

See e2e/START_HERE.md for implementation guide."
```

### Option 2: Two Commits (Cleaner History)

**Commit 1: Test Suite**
```bash
git add e2e/ Makefile AGENTS.md
git commit -m "feat: Add E2E API test suite

- Complete test suite with 25+ scenarios
- Environment-agnostic configuration
- Comprehensive documentation
- See e2e/START_HERE.md for details"
```

**Commit 2: CI Integration**
```bash
git add .tekton/
git commit -m "feat: Add Konflux CI integration for E2E tests

- In-pipeline integration (ready to apply)
- Bonfire/ephemeral strategy (future)
- Reusable Tekton task
- See .tekton/QUICK_IMPLEMENTATION.md for details"
```

## Before Pushing

- [ ] Create feature branch: `git checkout -b feat/e2e-test-suite`
- [ ] Push to fork first if testing: `git push origin feat/e2e-test-suite`
- [ ] Review on GitHub before creating PR
- [ ] Ensure PR description references implementation docs

## PR Description Template

```markdown
## Summary
Adds comprehensive E2E API test suite with Konflux CI integration strategy.

## What's Included
- ✅ Complete E2E test suite (839 lines, 25+ test scenarios)
- ✅ In-pipeline integration (ready to implement)
- ✅ Bonfire/ephemeral strategy (future goal)
- ✅ Comprehensive documentation

## Test Coverage
- Identity endpoints
- Favorite pages
- Last visited pages
- Recently used workspaces  
- Self report

## Implementation
Two paths available:
1. **In-Pipeline** (pragmatic): Apply `.tekton/e2e-integration.patch` - see `.tekton/QUICK_IMPLEMENTATION.md`
2. **Bonfire** (long-term): Full strategy in `e2e/KONFLUX_INTEGRATION.md`

## Getting Started
See `e2e/START_HERE.md` for navigation and quick start guide.

## Testing
- [ ] Tests compile successfully
- [ ] Local execution verified
- [ ] Documentation reviewed
```

## Common Issues to Check

### 1. Line Endings
```bash
# Ensure consistent line endings (especially in scripts)
git config core.autocrlf input
```

### 2. File Permissions
```bash
# No executable Go files
find e2e -name "*.go" -perm +111 -ls
```

### 3. Large Files
```bash
# Check for accidentally large files
find e2e .tekton -type f -size +100k
```

### 4. Sensitive Data
```bash
# Double-check for secrets
grep -r "pk_" e2e/ .tekton/ || echo "No private keys found"
grep -r "sk_" e2e/ .tekton/ || echo "No secret keys found"
```

## After Committing

1. **Don't push immediately** - Review the commit first
2. **Check commit message**: `git log -1 --pretty=full`
3. **Review changes**: `git show HEAD`
4. **If wrong, amend**: `git commit --amend`
5. **When ready**: `git push origin feat/e2e-test-suite`

## Rollback if Needed

```bash
# Before pushing - reset to previous commit
git reset --soft HEAD~1

# After pushing - revert the commit
git revert HEAD
git push origin feat/e2e-test-suite
```
