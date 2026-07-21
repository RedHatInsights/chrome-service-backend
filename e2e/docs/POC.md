# E2E Tests Proof of Concept

This document provides step-by-step instructions for manually testing the E2E test suite before integrating into the CI pipeline.

## Prerequisites

1. Access to the stage environment (`https://console.stage.redhat.com`)
2. Valid test user credentials for stage
3. Go 1.26+ installed locally
4. Network access to stage environment

## Step 1: Manual Local Testing

### 1.1 Configure Test Credentials

```bash
cd e2e
cp .env.example .env
```

Edit `.env`:
```bash
E2E_BASE_URL=https://console.stage.redhat.com
E2E_USER_ID=<your-test-user-id>
E2E_ACCOUNT_ID=<your-account-id>
E2E_ORG_ID=<your-org-id>
E2E_USERNAME=<your-username>
```

### 1.2 Run Tests Locally

```bash
# Install dependencies
go mod download

# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestGetUserIdentity
```

**Expected output**:
```
=== RUN   TestGetUserIdentity
--- PASS: TestGetUserIdentity (0.23s)
=== RUN   TestUpdateUserPreview
--- PASS: TestUpdateUserPreview (0.45s)
...
PASS
ok      github.com/RedHatInsights/chrome-service-backend/e2e    3.142s
```

## Step 2: Test Against Ephemeral Environment

### 2.1 Deploy to Ephemeral Namespace

If you have access to deploy ephemeral environments:

```bash
# Using Bonfire
bonfire deploy chrome-service \
  --source=appsre \
  --ref-env insights-stage \
  --set-image-tag <your-pr-image-tag> \
  --namespace <your-namespace>

# Get the route URL
ROUTE_URL=$(oc get route chrome-service-api -n <your-namespace> -o jsonpath='{.spec.host}')
echo "Service available at: https://${ROUTE_URL}"
```

### 2.2 Run Tests Against Ephemeral

```bash
cd e2e

E2E_BASE_URL=https://${ROUTE_URL} \
E2E_USER_ID=<test-user-id> \
E2E_ACCOUNT_ID=<test-account-id> \
E2E_ORG_ID=<test-org-id> \
E2E_USERNAME=<test-username> \
go test -v ./...
```

## Step 3: Test the Tekton Task

### 3.1 Apply the Task Definition

```bash
# From repository root
kubectl apply -f .tekton/tasks/e2e-tests.yaml -n hcc-platex-services-tenant
```

### 3.2 Create Test Workspace

```bash
kubectl create -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: chrome-e2e-poc-workspace
  namespace: hcc-platex-services-tenant
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
```

### 3.3 Clone Source to Workspace

You'll need to populate the workspace with the source code. One way:

```bash
# Create a pod to clone the repo
kubectl run -n hcc-platex-services-tenant clone-source \
  --image=alpine/git \
  --restart=Never \
  --overrides='
{
  "spec": {
    "containers": [{
      "name": "clone-source",
      "image": "alpine/git",
      "command": ["sh", "-c"],
      "args": ["git clone https://github.com/RedHatInsights/chrome-service-backend.git /workspace/source && sleep 3600"],
      "volumeMounts": [{
        "name": "source",
        "mountPath": "/workspace/source"
      }]
    }],
    "volumes": [{
      "name": "source",
      "persistentVolumeClaim": {
        "claimName": "chrome-e2e-poc-workspace"
      }
    }]
  }
}'

# Wait for clone to complete
kubectl wait --for=condition=ready pod/clone-source -n hcc-platex-services-tenant --timeout=60s
```

### 3.4 Run the E2E Test Task

```bash
tkn task start chrome-service-e2e-tests \
  -n hcc-platex-services-tenant \
  --param APP_URL=https://console.stage.redhat.com \
  --param TEST_USER_ID=<your-test-user-id> \
  --param TEST_ACCOUNT_ID=<your-account-id> \
  --param TEST_ORG_ID=<your-org-id> \
  --param TEST_USERNAME=<your-username> \
  --workspace name=source,claimName=chrome-e2e-poc-workspace \
  --showlog
```

**Expected output**:
```
[verify-connectivity] Verifying connectivity to https://console.stage.redhat.com...
[verify-connectivity] ✓ Host console.stage.redhat.com is reachable
[verify-connectivity] ✓ Health endpoint is accessible
[run-e2e-tests] =========================================
[run-e2e-tests] Chrome Service E2E Tests
[run-e2e-tests] =========================================
[run-e2e-tests] Target URL: https://console.stage.redhat.com
...
[run-e2e-tests] ✓ E2E Tests PASSED
[summary] Status: passed
[summary] Tests run: 21
```

### 3.5 Cleanup

```bash
# Delete the clone pod
kubectl delete pod clone-source -n hcc-platex-services-tenant

# Delete the PVC (optional)
kubectl delete pvc chrome-e2e-poc-workspace -n hcc-platex-services-tenant
```

## Step 4: Integration Testing

### 4.1 Create a Test PR

Create a branch with a small change:

```bash
git checkout -b test/e2e-integration
echo "# E2E Test Integration" >> e2e/README.md
git add e2e/README.md
git commit -m "test: E2E integration POC"
git push origin test/e2e-integration
```

### 4.2 Modify PR Pipeline (Temporary Test)

Create a test version of the PR pipeline with E2E tests:

```yaml
# .tekton/chrome-service-pull-request-e2e.yaml
# Copy from chrome-service-pull-request.yaml and add:

spec:
  params:
    # ... existing params ...
    
    # Add after unit-tests-script
    - name: post-build-script
      value: |
        #!/bin/bash
        set -e
        
        echo "Running E2E tests against stage..."
        
        export E2E_BASE_URL="https://console.stage.redhat.com"
        export E2E_USER_ID="<test-user-id>"
        export E2E_ACCOUNT_ID="<test-account-id>"
        export E2E_ORG_ID="<test-org-id>"
        export E2E_USERNAME="<test-username>"
        
        cd e2e
        go mod download
        go test -v ./...
```

### 4.3 Monitor Pipeline Run

```bash
# List pipeline runs
tkn pipelinerun list -n hcc-platex-services-tenant

# Watch specific run
tkn pipelinerun logs <pipeline-run-name> -n hcc-platex-services-tenant -f
```

## Step 5: Validation Checklist

- [ ] Tests run successfully locally against stage
- [ ] Tests run successfully from Tekton task
- [ ] Tests can authenticate with test user
- [ ] Tests can reach all endpoints
- [ ] Test failures are properly reported
- [ ] Retry logic works as expected
- [ ] Pipeline integration works end-to-end

## Common Issues and Solutions

### Issue: "connection refused" errors

**Diagnosis**:
```bash
# Test connectivity manually
curl -v https://console.stage.redhat.com/health
```

**Solutions**:
- Verify URL is correct
- Check if service is deployed and running
- Verify network policies allow access
- Check if there's a VPN requirement

### Issue: "401 Unauthorized" errors

**Diagnosis**:
```bash
# Test auth header manually
cd e2e
go run -e 'package main; import "testing"; func main() { config := GetConfig(); client := NewTestClient(&testing.T{}, config); println(client.CreateIdentityHeader()) }'
```

**Solutions**:
- Verify test user credentials are correct
- Check user exists in target environment
- Verify user has required permissions
- Check x-rh-identity header format

### Issue: Flaky tests

**Symptoms**: Tests pass sometimes, fail other times

**Solutions**:
- Increase retry count and delay
- Check for race conditions in test data
- Verify backend handles concurrent requests
- Look for timing dependencies in tests

### Issue: Tests timeout in CI

**Solutions**:
- Increase `TEST_TIMEOUT` parameter
- Check if service is slow to start
- Verify network latency between CI and service
- Consider running tests in same cluster as service

## Next Steps After POC

1. **Create dedicated test user**:
   - Request service account or dedicated test user in stage
   - Document user creation process
   - Store credentials in Kubernetes secret

2. **Automate integration**:
   - Create custom pipeline or extend existing one
   - Add E2E test task to pipeline
   - Configure credentials from secrets

3. **Set up monitoring**:
   - Configure test result reporting
   - Set up alerts for failures
   - Create dashboard for test metrics

4. **Document process**:
   - Update team documentation
   - Create runbook for test failures
   - Document how to add new tests

## Questions to Answer

Before full integration, clarify:

1. **Test User**: How do we create/manage test user in stage?
2. **Deployment**: Do we test against ephemeral or stage environment?
3. **Pipeline**: Which pipeline should run E2E tests (PR, push, both)?
4. **Failures**: What happens when E2E tests fail (block merge, warning)?
5. **Credentials**: How are test credentials managed and rotated?
