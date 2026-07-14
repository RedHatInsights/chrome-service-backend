# Konflux E2E Integration

This document explains how the E2E tests are integrated into Konflux CI/CD and how to troubleshoot issues.

**⚠️ Important**: This is the first chrome-service repository with in-pipeline E2E API tests running in Konflux!

## Architecture Overview

### Components

The E2E integration consists of two main parts:

1. **IntegrationTestScenario** (in `konflux-release-data` repo)
   - Location: `tenants-config/cluster/stone-prd-rh01/tenants/hcc-platex-services-tenant/chrome-service.bonfire-tekton.yaml`
   - What it does: Tells Konflux WHEN to run tests (after builds)
   - Points to: The task definition in this repository

2. **Task Definition** (in this repo)
   - Location: `.tekton/integration-test-scenarios/tasks/chrome-service-e2e-integration-test.yaml`
   - What it does: Defines HOW to run the tests (deployment + execution + cleanup)
   - Contains: 6 steps (parse-snapshot, deploy-with-bonfire, wait-for-deployment, clone-repository, run-e2e-tests, cleanup)

### Execution Flow

```
PR/Push to chrome-service-backend
           ↓
Konflux builds container image
           ↓
Build succeeds, creates Snapshot
           ↓
IntegrationTestScenario triggers
           ↓
┌──────────────────────────────────┐
│ chrome-service-e2e-integration   │
│ Task Execution                   │
└──────────────────────────────────┘
           ↓
┌─ Step 1: parse-snapshot ─────────┐
│ Extract chrome-service image     │
│ from Konflux Snapshot JSON       │
└──────────────────────────────────┘
           ↓
┌─ Step 2: deploy-with-bonfire ────┐
│ bonfire deploy chrome-service    │
│ - source: appsre                 │
│ - ref-env: insights-stage        │
│ - namespace: ephemeral-<random>  │
│ - includes all dependencies      │
└──────────────────────────────────┘
           ↓
┌─ Step 3: wait-for-deployment ────┐
│ Wait for ClowdApp reconciliation │
│ Wait for deployment available    │
│ Get route URL                    │
│ Verify health endpoint           │
└──────────────────────────────────┘
           ↓
┌─ Step 4: clone-repository ───────┐
│ git clone chrome-service-backend │
│ Get E2E test code                │
└──────────────────────────────────┘
           ↓
┌─ Step 5: run-e2e-tests ──────────┐
│ cd e2e                           │
│ go mod download                  │
│ go test -v ./...                 │
│ Report results                   │
└──────────────────────────────────┘
           ↓
┌─ Step 6: cleanup ────────────────┐
│ bonfire namespace release        │
│ Remove ephemeral namespace       │
└──────────────────────────────────┘
           ↓
Results reported to PR/Dashboard
```

## Dependencies

### Bonfire Deployment Includes

When Bonfire deploys chrome-service, it automatically includes:

- **PostgreSQL**: Database for chrome-service
- **Kafka**: Message broker
- **Unleash**: Feature flags service
- **ClowdApp Controller**: Manages the deployment lifecycle

### Environment Variables in CI

The task sets these environment variables for the E2E tests:

```yaml
E2E_BASE_URL: <route-from-deployed-service>
E2E_USER_ID: "e2e-integration-test-user"
E2E_ACCOUNT_ID: "e2e-integration-account"
E2E_ORG_ID: "e2e-integration-org"
E2E_USERNAME: "e2e-integration-bot"
```

## Troubleshooting

### Common Failure Scenarios

#### 1. Snapshot Parsing Fails

**Symptom**: Task fails in `parse-snapshot` step

**Cause**: Snapshot JSON doesn't contain chrome-service component

**Debug**:
```bash
# Check Tekton logs for the snapshot JSON
# Look for: echo '$(params.SNAPSHOT)' output
```

**Fix**: Verify the build actually succeeded and created a snapshot

---

#### 2. Bonfire Deployment Fails

**Symptom**: Task fails in `deploy-with-bonfire` step

**Possible Causes**:
- Bonfire service is down
- Namespace quota exceeded
- App-interface configuration issues
- bonfire-token secret missing/invalid

**Debug**:
```bash
# Check Tekton logs for bonfire output
# Look for errors in: bonfire deploy chrome-service
```

**Fix**:
- Check bonfire service status
- Verify bonfire-token secret exists in namespace
- Check app-interface for chrome-service configuration

---

#### 3. Deployment Never Becomes Ready

**Symptom**: Task times out in `wait-for-deployment` step

**Possible Causes**:
- Image pull failures
- ClowdApp dependencies not available
- Database migration failures
- Resource constraints

**Debug**:
```bash
# From logs, get the ephemeral namespace name
NAMESPACE="ephemeral-abc123"

# Check ClowdApp status
oc get clowdapp chrome-service -n $NAMESPACE -o yaml

# Check deployment status
oc get deployment chrome-service-api -n $NAMESPACE

# Check pod logs
oc logs -l app=chrome-service -n $NAMESPACE --tail=100

# Check events
oc get events -n $NAMESPACE --sort-by='.lastTimestamp'
```

**Fix**:
- Check if the built image is pullable
- Verify database migrations run successfully
- Check resource limits in ClowdApp definition

---

#### 4. Health Endpoint Not Responding

**Symptom**: Task fails with "Health check failed after N attempts"

**Possible Causes**:
- Service crashed after starting
- Route not configured correctly
- Health endpoint is broken

**Debug**:
```bash
# Check if pods are running
oc get pods -n $NAMESPACE -l app=chrome-service

# Check pod logs for errors
oc logs <pod-name> -n $NAMESPACE

# Try hitting health endpoint manually
curl -v https://<route>/health
```

**Fix**:
- Fix crashes in application startup
- Verify /health endpoint implementation
- Check route configuration in ClowdApp

---

#### 5. E2E Tests Fail

**Symptom**: Deployment succeeds but tests fail in `run-e2e-tests` step

**Possible Causes**:
- API behavior changed
- Test assumptions invalid in ephemeral environment
- Database state issues
- Authentication problems

**Debug**:
```bash
# Check test output in Tekton logs
# Look for specific test failure messages
# Example: "TestGetUserIdentity FAILED"
```

**Fix**:
- Run tests locally against stage to verify they still work
- Check if API changes broke test assumptions
- Verify test user identity is valid
- Update tests if API behavior intentionally changed

---

#### 6. Cleanup Fails

**Symptom**: Task succeeds but cleanup step has warnings

**Impact**: Low - orphaned namespaces will be garbage collected

**Cause**: Bonfire cleanup failures (usually transient)

**Debug**:
```bash
# Check cleanup logs
# Look for: bonfire namespace release output
```

**Fix**: Usually self-healing; namespaces have TTL and auto-expire

## Modifying the Integration

### Changing When Tests Run

Edit the IntegrationTestScenario in konflux-release-data:

```yaml
# Current: Tests are optional
metadata:
  labels:
    test.appstudio.openshift.io/optional: "true"

# To make required (blocks merges on failure):
metadata:
  labels:
    test.appstudio.openshift.io/optional: "false"
```

### Changing What Gets Deployed

Edit the `deploy-with-bonfire` step in the task definition:

```yaml
# Current deployment command:
bonfire deploy chrome-service \
  --source=appsre \
  --ref-env insights-stage \
  --set-image-tag ${IMAGE##*:} \
  --namespace $NAMESPACE

# To deploy additional components:
bonfire deploy chrome-service \
  --source=appsre \
  --ref-env insights-stage \
  --set-image-tag ${IMAGE##*:} \
  --namespace $NAMESPACE \
  --component another-service
```

### Changing Test Parameters

Edit the `run-e2e-tests` step environment variables:

```yaml
env:
  - name: E2E_BASE_URL
    value: ""  # Populated from deployed route
  - name: E2E_USER_ID
    value: "e2e-integration-test-user"
  # Add more as needed
```

### Changing Timeouts

```yaml
# Deployment timeout (default 600s)
spec:
  steps:
    - name: wait-for-deployment
      script: |
        oc wait --for=condition=ReconciliationSuccessful \
          clowdapp/chrome-service \
          -n $NAMESPACE \
          --timeout=600s  # <-- Change this

# Test execution timeout (default 600s)
    - name: run-e2e-tests
      image: golang:1.26
      timeout: 600s  # <-- Change this
```

## Updating the Integration

### When You Need to Update Task Definition

Edit `.tekton/integration-test-scenarios/tasks/chrome-service-e2e-integration-test.yaml` in this repository.

**Common reasons**:
- Add new deployment dependencies
- Change test execution parameters
- Update timeout values
- Add debugging steps
- Change cleanup behavior

**Process**:
1. Edit the task file
2. Test locally if possible (limited - some steps need Konflux)
3. Create PR to chrome-service-backend
4. Tests will run with the NEW task definition
5. Merge once validated

### When You Need to Update IntegrationTestScenario

Edit `chrome-service.bonfire-tekton.yaml` in konflux-release-data repository.

**Common reasons**:
- Change from optional to required
- Point to different branch (testing)
- Change to different task entirely

**Process**:
1. Create MR in konflux-release-data
2. Get appropriate approvals (CODEOWNERS)
3. Merge MR
4. Changes apply to cluster automatically
5. Next chrome-service build uses new configuration

## Tips for Development

### Testing Task Changes Locally

Limited local testing is possible:

```bash
# Validate YAML syntax
yamllint .tekton/integration-test-scenarios/tasks/chrome-service-e2e-integration-test.yaml

# Validate against Tekton schema
tkn task validate -f .tekton/integration-test-scenarios/tasks/chrome-service-e2e-integration-test.yaml
```

### Debugging in Konflux

1. **View live logs**: Konflux dashboard → Pipelines → Select run → View steps
2. **Re-trigger tests**: Comment `/retest` on PR
3. **Check ClowdApp**: Look at Tekton logs for namespace name, then inspect with `oc`

### Best Practices

- ✅ Keep task steps focused and single-purpose
- ✅ Add logging at each major step
- ✅ Use meaningful error messages
- ✅ Set appropriate timeouts
- ✅ Clean up resources even on failure (use `finally` blocks if needed)
- ✅ Test changes in a fork before merging to upstream

## Performance Metrics

Typical execution times:

| Step | Duration |
|------|----------|
| parse-snapshot | ~5s |
| deploy-with-bonfire | ~2-3min |
| wait-for-deployment | ~1-2min |
| clone-repository | ~10s |
| run-e2e-tests | ~10-15s |
| cleanup | ~30s |
| **Total** | **~5-8min** |

## Related Documentation

- [E2E Test README](./README.md) - Running tests locally
- [AGENTS.md](../AGENTS.md) - Overall testing strategy
- [Bonfire Documentation](https://github.com/RedHatInsights/bonfire) - Deployment tool
- [ClowdApp Documentation](https://github.com/RedHatInsights/clowder) - Cloud-native app platform

## Getting Help

- **Konflux Issues**: Ask in #konflux-users Slack channel
- **Bonfire Issues**: Ask in #insights-qe Slack channel  
- **Test Failures**: File issue in chrome-service-backend GitHub
- **Task Definition Issues**: Review with team, create PR with fixes
