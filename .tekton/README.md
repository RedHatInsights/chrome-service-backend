# Tekton Pipeline Configuration

This directory contains Tekton/Konflux CI pipeline definitions for chrome-service-backend.

## Files

### Pipeline Runs
- `chrome-service-pull-request.yaml` - Pipeline triggered on pull requests
- `chrome-service-push.yaml` - Pipeline triggered on pushes to main
- `chrome-service-sc-pull-request.yaml` - Service catalog PR pipeline
- `chrome-service-sc-push.yaml` - Service catalog push pipeline

### Tasks
- `tasks/e2e-tests.yaml` - E2E test task definition

## Current Pipeline

The project uses the `docker-build-run-unit-tests-dynamic-env` pipeline from [konflux-pipelines](https://github.com/RedHatInsights/konflux-pipelines).

**Current flow**:
1. Clone source code
2. Build Docker image
3. Run unit tests with PostgreSQL + Unleash sidecars
4. Push image to Quay

## E2E Test Integration

### E2E Test Task

The `chrome-service-e2e-tests` task runs the E2E test suite against a deployed instance.

**Parameters**:
- `APP_URL` - Base URL of the API (e.g., `https://console.stage.redhat.com`)
- `TEST_USER_ID` - Test user ID
- `TEST_ACCOUNT_ID` - Test account ID  
- `TEST_ORG_ID` - Test org ID
- `TEST_USERNAME` - Test username
- `MAX_RETRIES` - Retry attempts (default: 3)
- `RETRY_DELAY` - Delay between retries in seconds (default: 30)
- `TEST_TIMEOUT` - Overall timeout in seconds (default: 600)

**Results**:
- `test-status` - "passed" or "failed"
- `tests-run` - Number of tests executed

### Testing the E2E Task Locally

```bash
# Create a workspace PVC
kubectl create -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: chrome-e2e-test-workspace
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF

# Run the task
tkn task start chrome-service-e2e-tests \
  --param APP_URL=https://console.stage.redhat.com \
  --param TEST_USER_ID=e2e-test-user-001 \
  --param TEST_ACCOUNT_ID=e2e-test-account \
  --param TEST_ORG_ID=e2e-test-org \
  --param TEST_USERNAME=chrome-e2e-bot \
  --workspace name=source,claimName=chrome-e2e-test-workspace \
  --showlog
```

### Adding E2E Tests to Pipeline

**Option 1: Add to existing pipeline via script parameter**

Modify the PipelineRun to add an `e2e-tests-script` parameter:

```yaml
spec:
  params:
    # ... existing params ...
    
    - name: e2e-tests-script
      value: |
        #!/bin/bash
        set -e
        
        # Get deployed service URL (method depends on deployment strategy)
        # For Bonfire ephemeral:
        # APP_URL=$(oc get route chrome-service-api -n ${NAMESPACE} -o jsonpath='{.spec.host}')
        # APP_URL="https://${APP_URL}"
        
        # For stage:
        APP_URL="https://console.stage.redhat.com"
        
        export E2E_BASE_URL="${APP_URL}"
        export E2E_USER_ID="e2e-test-user-001"
        export E2E_ACCOUNT_ID="e2e-test-account"
        export E2E_ORG_ID="e2e-test-org"
        export E2E_USERNAME="chrome-e2e-bot"
        
        cd e2e
        go mod download
        go test -v ./...
```

**Option 2: Create custom pipeline**

See `../e2e/KONFLUX_INTEGRATION.md` for details on creating a custom pipeline that includes the E2E test task.

## Required Secrets

### Test User Credentials

Create a secret with test user credentials:

```bash
kubectl create secret generic chrome-e2e-test-credentials \
  --from-literal=user-id=e2e-test-user-001 \
  --from-literal=account-id=e2e-test-account \
  --from-literal=org-id=e2e-test-org \
  --from-literal=username=chrome-e2e-bot \
  -n hcc-platex-services-tenant
```

Reference in pipeline:

```yaml
- name: TEST_USER_ID
  valueFrom:
    secretKeyRef:
      name: chrome-e2e-test-credentials
      key: user-id
```

## Troubleshooting

### Task Not Found

If `chrome-service-e2e-tests` task is not found:

```bash
# Apply the task to your namespace
kubectl apply -f .tekton/tasks/e2e-tests.yaml -n hcc-platex-services-tenant
```

### Tests Can't Reach Service

- Verify the APP_URL is correct
- Check network policies allow egress
- Verify the service is deployed and healthy
- Check route/ingress configuration

### Authentication Failures

- Verify test user credentials are correct
- Check the test user exists in the target environment
- Verify the user has required permissions
- Check x-rh-identity header format

## References

- [Konflux Pipelines](https://github.com/RedHatInsights/konflux-pipelines)
- [Tekton Documentation](https://tekton.dev/docs/)
- [E2E Test Documentation](../e2e/README.md)
- [Konflux Integration Guide](../e2e/KONFLUX_INTEGRATION.md)
