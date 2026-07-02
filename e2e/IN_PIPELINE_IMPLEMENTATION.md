# In-Pipeline E2E Implementation Guide

## Overview

This guide provides step-by-step instructions for integrating E2E tests into the existing Konflux pipeline using the in-pipeline sidecar approach (Option 2).

**Goal**: Run E2E tests in the same build pipeline, reusing existing PostgreSQL and Unleash sidecars.

**Timeline**: 1-2 days to implement and validate

## Why This Approach?

✅ **Fast to implement** - Extends existing pipeline structure  
✅ **Reuses infrastructure** - PostgreSQL and Unleash sidecars already running  
✅ **No deployment needed** - Tests run immediately after build  
✅ **Quick feedback** - Results available in ~5 minutes  
✅ **Isolated** - Each PR gets its own test environment  

## Current Pipeline Analysis

Your existing pipeline already has:
- PostgreSQL sidecar at `localhost:5432`
- Unleash sidecar at `localhost:4242`
- Unit tests running successfully with these sidecars
- `unit-tests-script` parameter that runs migrations and tests

## Implementation Steps

### Step 1: Update Pipeline Definition

We'll extend the existing `unit-tests-script` to include E2E tests.

**File**: `.tekton/chrome-service-pull-request.yaml`

Find the `unit-tests-script` parameter (around line 57-72) and replace it with:

```yaml
- name: unit-tests-script
  value: |
    #!/bin/bash
    set -ex

    export GOTOOLCHAIN=auto
    export PGSQL_USER="$(params.PGSQL_USER)"
    export PGSQL_PASSWORD="$(params.PGSQL_PASSWORD)"
    export PGSQL_HOSTNAME="$(params.PGSQL_HOSTNAME)"
    export PGSQL_PORT="$(params.PGSQL_PORT)"
    export PGSQL_DATABASE="$(params.PGSQL_DATABASE)"
    export UNLEASH_API_TOKEN="$(params.UNLEASH_API_TOKEN)"
    export UNLEASH_ADMIN_TOKEN="$(params.UNLEASH_ADMIN_TOKEN)"

    # Run database migration
    echo "========================================="
    echo "Running database migrations..."
    echo "========================================="
    make migrate

    # Run unit tests
    echo "========================================="
    echo "Running unit tests..."
    echo "========================================="
    make test

    # Build service binary for E2E tests
    echo "========================================="
    echo "Building service binary..."
    echo "========================================="
    CGO_ENABLED=1 go build -o /tmp/chrome-service-backend

    # Start service in background
    echo "========================================="
    echo "Starting chrome-service for E2E tests..."
    echo "========================================="
    /tmp/chrome-service-backend &
    SERVICE_PID=$!
    echo "Service started with PID: $SERVICE_PID"

    # Wait for service to be ready (with timeout)
    echo "Waiting for service to be ready..."
    MAX_WAIT=60
    COUNTER=0
    while [ $COUNTER -lt $MAX_WAIT ]; do
      if curl -sf http://localhost:8000/health > /dev/null 2>&1; then
        echo "✓ Service is ready!"
        break
      fi
      echo "  Waiting... ($COUNTER/$MAX_WAIT)"
      sleep 2
      COUNTER=$((COUNTER + 2))
    done

    if [ $COUNTER -ge $MAX_WAIT ]; then
      echo "✗ Service failed to start within ${MAX_WAIT}s"
      kill $SERVICE_PID 2>/dev/null || true
      exit 1
    fi

    # Verify service is responding
    curl -v http://localhost:8000/health

    # Run E2E tests
    echo "========================================="
    echo "Running E2E tests..."
    echo "========================================="
    export E2E_BASE_URL="http://localhost:8000"
    export E2E_USER_ID="e2e-test-user-$(date +%s)"
    export E2E_ACCOUNT_ID="e2e-test-account"
    export E2E_ORG_ID="e2e-test-org"
    export E2E_USERNAME="e2e-bot"

    cd e2e
    go mod download
    
    # Run tests with timeout
    if timeout 300 go test -v ./...; then
      echo "========================================="
      echo "✓ E2E tests PASSED"
      echo "========================================="
      E2E_EXIT_CODE=0
    else
      echo "========================================="
      echo "✗ E2E tests FAILED"
      echo "========================================="
      E2E_EXIT_CODE=1
    fi

    # Cleanup - kill service
    echo "Cleaning up service (PID: $SERVICE_PID)..."
    kill $SERVICE_PID 2>/dev/null || true
    sleep 2
    kill -9 $SERVICE_PID 2>/dev/null || true

    # Exit with E2E test result
    exit $E2E_EXIT_CODE
```

### Step 2: Apply Same Changes to Push Pipeline

**File**: `.tekton/chrome-service-push.yaml`

Apply the same changes to the `unit-tests-script` parameter (around line 54-69).

### Step 3: Add Required Environment Variables

The service needs some additional environment variables to run. Add these to the `env-vars` parameter:

```yaml
- name: env-vars
  value: |
    PGSQL_USER=chrome
    PGSQL_PASSWORD=chrome
    PGSQL_HOSTNAME=localhost
    PGSQL_PORT=5432
    PGSQL_DATABASE=db
    UNLEASH_API_TOKEN=default:development.unleash-insecure-api-token
    UNLEASH_ADMIN_TOKEN=*:*.unleash-insecure-api-token
    LOG_LEVEL=info
    CLOWDER_ENABLED=false
    TEMPLATES_WD=./widget-dashboard-defaults
```

### Step 4: Validation Steps

Before committing, validate the changes:

#### 4.1 Local Testing

Test the script locally to ensure it works:

```bash
# Set up environment
export PGSQL_USER=chrome
export PGSQL_PASSWORD=chrome
export PGSQL_HOSTNAME=localhost
export PGSQL_PORT=5432
export PGSQL_DATABASE=db
export UNLEASH_API_TOKEN=default:development.unleash-insecure-api-token
export UNLEASH_ADMIN_TOKEN='*:*.unleash-insecure-api-token'

# Make sure local infrastructure is running
make infra

# Run the combined script
make migrate
make test

# Build and start service
go build -o /tmp/chrome-service-backend
/tmp/chrome-service-backend &
SERVICE_PID=$!

# Wait for ready
sleep 5
curl http://localhost:8000/health

# Run E2E tests
export E2E_BASE_URL="http://localhost:8000"
export E2E_USER_ID="test-user-123"
export E2E_ACCOUNT_ID="123456"
export E2E_ORG_ID="654321"
export E2E_USERNAME="testuser"

cd e2e
go test -v ./...

# Cleanup
kill $SERVICE_PID
```

#### 4.2 Test PR

Create a test branch and PR to validate in CI:

```bash
git checkout -b test/e2e-in-pipeline
git add .tekton/
git commit -m "feat: Add E2E tests to CI pipeline"
git push origin test/e2e-in-pipeline
```

Monitor the pipeline run:
- Check unit tests still pass
- Verify service starts successfully
- Confirm E2E tests run and pass
- Check cleanup happens properly

### Step 5: Monitoring and Debugging

#### Check Pipeline Logs

```bash
# List recent pipeline runs
tkn pipelinerun list -n hcc-platex-services-tenant

# View logs for specific run
tkn pipelinerun logs <pipelinerun-name> -n hcc-platex-services-tenant -f

# Filter for E2E test section
tkn pipelinerun logs <pipelinerun-name> -n hcc-platex-services-tenant | grep -A 50 "Running E2E tests"
```

#### Common Issues and Solutions

**Issue 1: Service won't start**

```bash
# Check the logs for error messages
# Common causes:
# - Missing environment variables
# - Database not ready
# - Port already in use
```

**Solution**:
- Increase wait timeout
- Add more detailed logging
- Check database connection

**Issue 2: Tests timeout**

```bash
# E2E tests take too long
```

**Solution**:
- Increase test timeout from 300s to 600s
- Check if service is slow to respond
- Verify network connectivity

**Issue 3: Service doesn't stop cleanly**

```bash
# Service process doesn't terminate
```

**Solution**:
- Use `kill -9` as fallback
- Add `set +e` before kill commands
- Wait briefly after kill

### Step 6: Fine-Tuning

Once basic implementation works, optimize:

#### 6.1 Parallel Execution

Tests run in serial by default. For faster execution:

```bash
# Run tests in parallel
go test -v -parallel 4 ./...
```

#### 6.2 Better Logging

Add structured output for easier debugging:

```bash
echo "========================================="
echo "E2E Test Results"
echo "========================================="
go test -v ./... 2>&1 | tee /tmp/e2e-results.log

# Count pass/fail
PASSED=$(grep -c "^--- PASS:" /tmp/e2e-results.log || echo "0")
FAILED=$(grep -c "^--- FAIL:" /tmp/e2e-results.log || echo "0")
echo "Tests passed: $PASSED"
echo "Tests failed: $FAILED"
```

#### 6.3 Conditional E2E Execution

Run E2E tests only on certain conditions:

```bash
# Only run E2E on main or release branches
if [[ "$(params.target_branch)" == "main" ]] || [[ "$(params.target_branch)" == release/* ]]; then
  # Run E2E tests
fi
```

## Complete Example

Here's the complete updated `unit-tests-script` with all optimizations:

```yaml
- name: unit-tests-script
  value: |
    #!/bin/bash
    set -e  # Exit on error, but we'll handle cleanup

    # Color output for better readability
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color

    echo_header() {
      echo ""
      echo "========================================="
      echo "$1"
      echo "========================================="
    }

    # Set environment variables
    export GOTOOLCHAIN=auto
    export PGSQL_USER="$(params.PGSQL_USER)"
    export PGSQL_PASSWORD="$(params.PGSQL_PASSWORD)"
    export PGSQL_HOSTNAME="$(params.PGSQL_HOSTNAME)"
    export PGSQL_PORT="$(params.PGSQL_PORT)"
    export PGSQL_DATABASE="$(params.PGSQL_DATABASE)"
    export UNLEASH_API_TOKEN="$(params.UNLEASH_API_TOKEN)"
    export UNLEASH_ADMIN_TOKEN="$(params.UNLEASH_ADMIN_TOKEN)"
    export LOG_LEVEL=info
    export CLOWDER_ENABLED=false
    export TEMPLATES_WD=./widget-dashboard-defaults

    # Track service PID for cleanup
    SERVICE_PID=""
    cleanup() {
      if [ -n "$SERVICE_PID" ]; then
        echo_header "Cleaning up service (PID: $SERVICE_PID)"
        kill $SERVICE_PID 2>/dev/null || true
        sleep 2
        kill -9 $SERVICE_PID 2>/dev/null || true
      fi
    }
    trap cleanup EXIT

    # Run database migration
    echo_header "Running database migrations"
    make migrate

    # Run unit tests
    echo_header "Running unit tests"
    make test

    # Build service binary
    echo_header "Building service binary for E2E tests"
    CGO_ENABLED=1 go build -o /tmp/chrome-service-backend

    # Start service
    echo_header "Starting chrome-service"
    /tmp/chrome-service-backend > /tmp/service.log 2>&1 &
    SERVICE_PID=$!
    echo "Service started with PID: $SERVICE_PID"

    # Wait for service to be ready
    echo "Waiting for service to be ready..."
    MAX_WAIT=60
    COUNTER=0
    while [ $COUNTER -lt $MAX_WAIT ]; do
      if curl -sf http://localhost:8000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Service is ready!${NC}"
        break
      fi
      echo "  Waiting... ($COUNTER/$MAX_WAIT)"
      sleep 2
      COUNTER=$((COUNTER + 2))
    done

    if [ $COUNTER -ge $MAX_WAIT ]; then
      echo -e "${RED}✗ Service failed to start within ${MAX_WAIT}s${NC}"
      echo "Service logs:"
      cat /tmp/service.log
      exit 1
    fi

    # Verify health endpoint
    echo "Verifying health endpoint..."
    curl -v http://localhost:8000/health || {
      echo -e "${RED}Health check failed${NC}"
      cat /tmp/service.log
      exit 1
    }

    # Run E2E tests
    echo_header "Running E2E tests"
    export E2E_BASE_URL="http://localhost:8000"
    export E2E_USER_ID="e2e-test-user-$(date +%s)"
    export E2E_ACCOUNT_ID="e2e-test-account"
    export E2E_ORG_ID="e2e-test-org"
    export E2E_USERNAME="e2e-bot"

    cd e2e
    go mod download

    # Run tests with timeout and capture results
    E2E_EXIT_CODE=0
    if timeout 300 go test -v -parallel 2 ./... 2>&1 | tee /tmp/e2e-results.log; then
      echo ""
      echo -e "${GREEN}=========================================${NC}"
      echo -e "${GREEN}✓ E2E tests PASSED${NC}"
      echo -e "${GREEN}=========================================${NC}"
    else
      E2E_EXIT_CODE=$?
      echo ""
      echo -e "${RED}=========================================${NC}"
      echo -e "${RED}✗ E2E tests FAILED (exit code: $E2E_EXIT_CODE)${NC}"
      echo -e "${RED}=========================================${NC}"
    fi

    # Print summary
    PASSED=$(grep -c "^--- PASS:" /tmp/e2e-results.log 2>/dev/null || echo "0")
    FAILED=$(grep -c "^--- FAIL:" /tmp/e2e-results.log 2>/dev/null || echo "0")
    echo ""
    echo "E2E Test Summary:"
    echo "  Passed: $PASSED"
    echo "  Failed: $FAILED"

    # Cleanup happens via trap
    exit $E2E_EXIT_CODE
```

## Success Criteria

✅ Unit tests still pass  
✅ Service starts successfully  
✅ Health endpoint responds  
✅ E2E tests run and pass  
✅ Service stops cleanly  
✅ Pipeline completes in <10 minutes  
✅ No resource leaks  

## Next Steps After Implementation

1. **Monitor first few PRs** - Watch for any issues
2. **Measure performance** - Track pipeline execution time
3. **Document learnings** - Update team wiki
4. **Plan Bonfire migration** - When ready for ephemeral environments

## Rollback Plan

If issues arise, rollback is simple:

```bash
# Revert to previous pipeline version
git revert <commit-hash>
git push origin main
```

Or temporarily disable E2E tests:

```yaml
# Comment out E2E section in unit-tests-script
# Run only unit tests
make test
# exit 0  # Skip E2E for now
```

## Estimated Timeline

- **Day 1 Morning**: Update pipeline files
- **Day 1 Afternoon**: Test locally, create test PR
- **Day 2 Morning**: Monitor test PR, fix issues
- **Day 2 Afternoon**: Merge and monitor production PRs

## Questions or Issues?

See `e2e/README.md` for troubleshooting guide or reach out to the team.
