# Quick E2E Implementation - Copy/Paste Guide

## What This Does

Adds E2E tests to your existing CI pipeline by extending the `unit-tests-script` to:
1. Build the service binary
2. Start it in the background (using existing PostgreSQL/Unleash sidecars)
3. Run E2E tests against localhost:8000
4. Clean up and report results

## Implementation (5 minutes)

### Step 1: Update Pull Request Pipeline

**File**: `.tekton/chrome-service-pull-request.yaml`

Find this section (around line 57):
```yaml
- name: unit-tests-script
  value: |
    #!/bin/bash
    set -ex

    export GOTOOLCHAIN=auto
    # ... existing exports ...
    
    make migrate
    make test
```

**Replace with**:
```yaml
- name: unit-tests-script
  value: |
    #!/bin/bash
    set -e

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

    # Run migrations and unit tests
    make migrate
    make test

    # Build service binary
    echo "Building service for E2E tests..."
    CGO_ENABLED=1 go build -o /tmp/chrome-service-backend

    # Start service in background
    echo "Starting service..."
    /tmp/chrome-service-backend > /tmp/service.log 2>&1 &
    SERVICE_PID=$!

    # Wait for ready
    echo "Waiting for service..."
    for i in {1..30}; do
      if curl -sf http://localhost:8000/health > /dev/null 2>&1; then
        echo "Service ready!"
        break
      fi
      sleep 2
    done

    # Run E2E tests
    echo "Running E2E tests..."
    export E2E_BASE_URL="http://localhost:8000"
    export E2E_USER_ID="e2e-test-user-$(date +%s)"
    export E2E_ACCOUNT_ID="e2e-test-account"
    export E2E_ORG_ID="e2e-test-org"
    export E2E_USERNAME="e2e-bot"

    cd e2e
    go mod download
    E2E_EXIT_CODE=0
    timeout 300 go test -v ./... || E2E_EXIT_CODE=$?

    # Cleanup
    kill $SERVICE_PID 2>/dev/null || true
    exit $E2E_EXIT_CODE
```

### Step 2: Update Push Pipeline

**File**: `.tekton/chrome-service-push.yaml`

Apply the **same change** to the `unit-tests-script` parameter (around line 54).

### Step 3: Test

Create a test PR:
```bash
git checkout -b test/add-e2e-to-pipeline
git add .tekton/
git commit -m "feat: Add E2E tests to CI pipeline"
git push origin test/add-e2e-to-pipeline
```

Monitor the pipeline run in the GitHub PR or via:
```bash
tkn pipelinerun list -n hcc-platex-services-tenant
```

## What to Expect

**Pipeline logs will show**:
```
Building service for E2E tests...
Starting service...
Waiting for service...
Service ready!
Running E2E tests...
=== RUN   TestGetUserIdentity
--- PASS: TestGetUserIdentity (0.12s)
=== RUN   TestUpdateUserPreview
--- PASS: TestUpdateUserPreview (0.23s)
...
PASS
```

**Total pipeline time**: ~5-8 minutes (was ~3-4 minutes without E2E)

## If Something Fails

### Service won't start
Check service logs in pipeline output. Common issues:
- Database connection (should work - using existing sidecar)
- Missing environment variable

### E2E tests fail
Check which test failed. Common issues:
- Authentication (check x-rh-identity header)
- Timing (service not fully ready)

### Timeout
Increase timeout from 300s to 600s:
```bash
timeout 600 go test -v ./...
```

## Rollback

If issues, revert the commit:
```bash
git revert <commit-hash>
git push
```

Or temporarily comment out E2E section:
```yaml
# cd e2e
# go mod download
# timeout 300 go test -v ./...
```

## That's It!

E2E tests now run on every PR automatically. No infrastructure changes needed.
