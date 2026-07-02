# Konflux CI Integration

This document describes how to integrate E2E tests into Konflux CI pipeline.

## Overview

The E2E tests can be run as part of the Konflux CI pipeline to verify API functionality against the stage environment after deployment.

## Tekton Task Example

Create a Tekton task that runs the E2E tests after the service is deployed to stage:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: chrome-service-e2e-tests
spec:
  description: Run E2E API tests against deployed service
  params:
    - name: BASE_URL
      description: Base URL of the deployed service
      type: string
      default: https://console.stage.redhat.com
  workspaces:
    - name: source
      description: Workspace containing the source code
  steps:
    - name: setup-go
      image: golang:1.26
      script: |
        #!/bin/bash
        cd $(workspaces.source.path)
        cd e2e
        go mod download
        
    - name: run-e2e-tests
      image: golang:1.26
      env:
        - name: E2E_BASE_URL
          value: $(params.BASE_URL)
        - name: E2E_USER_ID
          valueFrom:
            secretKeyRef:
              name: e2e-test-credentials
              key: user-id
        - name: E2E_ACCOUNT_ID
          valueFrom:
            secretKeyRef:
              name: e2e-test-credentials
              key: account-id
        - name: E2E_ORG_ID
          valueFrom:
            secretKeyRef:
              name: e2e-test-credentials
              key: org-id
        - name: E2E_USERNAME
          valueFrom:
            secretKeyRef:
              name: e2e-test-credentials
              key: username
      script: |
        #!/bin/bash
        set -e
        
        cd $(workspaces.source.path)
        
        echo "Running E2E tests against ${E2E_BASE_URL}"
        
        cd e2e
        go test -v ./... 2>&1 | tee test-results.log
        
        echo "E2E tests completed successfully"
```

## Pipeline Integration

Add the E2E test task to your Konflux pipeline:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: chrome-service-pipeline
spec:
  workspaces:
    - name: workspace
  tasks:
    # ... existing build and deploy tasks ...
    
    - name: e2e-tests
      taskRef:
        name: chrome-service-e2e-tests
      runAfter:
        - deploy-to-stage  # Run after deployment
      workspaces:
        - name: source
          workspace: workspace
      params:
        - name: BASE_URL
          value: https://console.stage.redhat.com
```

## Required Secrets

Create a Kubernetes secret with the test credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: e2e-test-credentials
type: Opaque
stringData:
  user-id: "<test-user-id>"
  account-id: "<test-account-id>"
  org-id: "<test-org-id>"
  username: "<test-username>"
```

## Creating the Secret

```bash
# Using kubectl
kubectl create secret generic e2e-test-credentials \
  --from-literal=user-id=<test-user-id> \
  --from-literal=account-id=<test-account-id> \
  --from-literal=org-id=<test-org-id> \
  --from-literal=username=<test-username> \
  -n <your-namespace>

# Or using oc (OpenShift CLI)
oc create secret generic e2e-test-credentials \
  --from-literal=user-id=<test-user-id> \
  --from-literal=account-id=<test-account-id> \
  --from-literal=org-id=<test-org-id> \
  --from-literal=username=<test-username>
```

## Test User Setup

For stage environment testing, you'll need:

1. A valid test user account in the stage environment
2. The user's ID, account ID, org ID, and username
3. The user must have appropriate permissions to access the API

You can create a dedicated test user or use an existing test account. Store the credentials securely in the Kubernetes secret.

## Running Tests Manually in Konflux

You can trigger the E2E tests manually using `tkn` CLI:

```bash
# Create a PipelineRun
tkn pipeline start chrome-service-pipeline \
  --workspace name=workspace,claimName=my-pvc \
  --showlog
```

## Monitoring Test Results

Test results can be monitored through:

1. Tekton Dashboard - View pipeline runs and task logs
2. OpenShift Console - Check pod logs and events
3. CI/CD dashboard - View test results and reports

## Troubleshooting

### Tests fail with authentication errors

- Verify the secret `e2e-test-credentials` exists in the correct namespace
- Check that the secret contains valid credentials
- Ensure the test user has access to the stage environment

### Tests fail with connection errors

- Verify the `BASE_URL` parameter is correct
- Check network policies allow outbound connections from the build namespace
- Ensure the stage environment is accessible from the cluster

### Tests timeout

- Increase the task timeout in the TaskRun spec
- Check if the stage environment is experiencing performance issues
- Consider running tests in smaller batches

## Best Practices

1. **Use dedicated test users** - Don't use production or personal accounts
2. **Rotate credentials** - Regularly update test user credentials
3. **Monitor test flakiness** - Track and investigate flaky tests
4. **Run on deployment** - Always run E2E tests after deploying to stage
5. **Alert on failures** - Configure notifications for E2E test failures
6. **Keep tests fast** - Optimize tests to run quickly for faster feedback

## Environment-Specific Configuration

You can run tests against different environments by parameterizing the BASE_URL:

```yaml
params:
  - name: BASE_URL
    value: $(params.ENVIRONMENT_URL)
```

Then pass different URLs for different environments:
- Stage: `https://console.stage.redhat.com`
- QA: `https://console.qa.redhat.com`
- Production: `https://console.redhat.com` (with appropriate safeguards)
