# Konflux E2E Integration Strategy

## Overview

This document outlines the approach for integrating E2E tests into the existing Konflux CI pipeline for chrome-service-backend.

## Current Pipeline Analysis

### Existing Konflux Setup

**Pipeline**: `docker-build-run-unit-tests-dynamic-env.yaml`
- Builds Docker image
- Runs unit tests with PostgreSQL + Unleash sidecars
- Deploys via ClowdApp to ephemeral/stage environments

**Key Files**:
- `.tekton/chrome-service-pull-request.yaml` - PR pipeline
- `.tekton/chrome-service-push.yaml` - Main branch pipeline
- `deploy/clowdapp.yml` - ClowdApp deployment template

## Integration Strategy: Three Approaches

### Option 1: Post-Deployment E2E Tests (RECOMMENDED)

Run E2E tests **after** the service is deployed to an ephemeral/stage environment.

**Pros**:
- Tests against real deployed service with all Clowder dependencies
- Most realistic test environment (PostgreSQL, Kafka, Unleash, etc.)
- Can reuse existing stage infrastructure
- No need to mock Clowder dependencies

**Cons**:
- Requires waiting for deployment to complete
- Slightly longer pipeline execution time

**Implementation**: See "Recommended Implementation" section below

---

### Option 2: In-Pipeline Container-Based E2E

Run E2E tests in the build pipeline with containerized dependencies.

**Pros**:
- Faster feedback (no deployment wait)
- Self-contained test environment

**Cons**:
- Complex setup with PostgreSQL, Kafka, Unleash sidecars
- Doesn't test against real Clowder environment
- May not catch environment-specific issues

**Implementation**: Extend existing `unit-tests-script` to include E2E

---

### Option 3: Hybrid Approach

Run basic E2E tests in-pipeline, full E2E suite post-deployment.

**Pros**:
- Fast feedback for basic functionality
- Comprehensive testing against real environment

**Cons**:
- Most complex to maintain
- Duplicated test infrastructure

---

## Recommended Implementation (Option 1)

### Phase 1: Add E2E Test Task to Pipeline

Create a custom Tekton task that runs after deployment:

```yaml
# .tekton/tasks/e2e-tests.yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: chrome-service-e2e-tests
spec:
  description: Run E2E API tests against deployed chrome-service
  params:
    - name: APP_URL
      description: Base URL of the deployed application
      type: string
    - name: TEST_USER_ID
      description: Test user ID
      type: string
    - name: TEST_ACCOUNT_ID
      description: Test account ID
      type: string
    - name: TEST_ORG_ID
      description: Test org ID
      type: string
    - name: TEST_USERNAME
      description: Test username
      type: string
  workspaces:
    - name: source
      description: Workspace containing the source code
  steps:
    - name: run-e2e-tests
      image: golang:1.26
      workingDir: $(workspaces.source.path)
      env:
        - name: E2E_BASE_URL
          value: $(params.APP_URL)
        - name: E2E_USER_ID
          value: $(params.TEST_USER_ID)
        - name: E2E_ACCOUNT_ID
          value: $(params.TEST_ACCOUNT_ID)
        - name: E2E_ORG_ID
          value: $(params.TEST_ORG_ID)
        - name: E2E_USERNAME
          value: $(params.TEST_USERNAME)
      script: |
        #!/bin/bash
        set -e
        
        echo "Running E2E tests against ${E2E_BASE_URL}"
        
        # Install dependencies
        cd e2e
        go mod download
        
        # Run tests with retries (service might be starting up)
        MAX_RETRIES=5
        RETRY_DELAY=30
        
        for i in $(seq 1 $MAX_RETRIES); do
          echo "Attempt $i of $MAX_RETRIES"
          if go test -v ./...; then
            echo "E2E tests passed"
            exit 0
          fi
          
          if [ $i -lt $MAX_RETRIES ]; then
            echo "Tests failed, retrying in ${RETRY_DELAY}s..."
            sleep $RETRY_DELAY
          fi
        done
        
        echo "E2E tests failed after $MAX_RETRIES attempts"
        exit 1
```

### Phase 2: Determine Deployment URL

The challenge is getting the deployed service URL. Two approaches:

#### Approach A: Use Bonfire for Ephemeral Environment

If using Bonfire for ephemeral deploys:

```bash
# Get the route URL from the ephemeral namespace
APP_URL=$(oc get route chrome-service-api -n $NAMESPACE -o jsonpath='{.spec.host}')
APP_URL="https://${APP_URL}"
```

#### Approach B: Use Known Stage URL

For stage deployments:

```bash
APP_URL="https://console.stage.redhat.com"
```

### Phase 3: Create Test User/Identity

You'll need a dedicated test user in the stage environment:

```yaml
# Store credentials in a Kubernetes Secret
apiVersion: v1
kind: Secret
metadata:
  name: chrome-e2e-test-credentials
  namespace: hcc-platex-services-tenant
type: Opaque
stringData:
  user-id: "e2e-test-user-001"
  account-id: "e2e-test-account"
  org-id: "e2e-test-org"
  username: "chrome-e2e-bot"
```

### Phase 4: Update Pipeline to Include E2E Task

Modify `.tekton/chrome-service-pull-request.yaml`:

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  annotations:
    # ... existing annotations ...
    pipelinesascode.tekton.dev/pipeline: https://github.com/RedHatInsights/konflux-pipelines/blob/main/pipelines/docker-build-deploy-e2e.yaml
spec:
  params:
    # ... existing params ...
    
    # E2E test configuration
    - name: enable-e2e-tests
      value: "true"
    - name: e2e-test-target
      value: "ephemeral"  # or "stage"
    - name: e2e-test-user-id
      valueFrom:
        secretKeyRef:
          name: chrome-e2e-test-credentials
          key: user-id
    - name: e2e-test-account-id
      valueFrom:
        secretKeyRef:
          name: chrome-e2e-test-credentials
          key: account-id
    - name: e2e-test-org-id
      valueFrom:
        secretKeyRef:
          name: chrome-e2e-test-credentials
          key: org-id
    - name: e2e-test-username
      valueFrom:
        secretKeyRef:
          name: chrome-e2e-test-credentials
          key: username
    
    # E2E test script to run
    - name: e2e-tests-script
      value: |
        #!/bin/bash
        set -e
        
        # Wait for deployment to be ready
        echo "Waiting for deployment to be ready..."
        oc wait --for=condition=available --timeout=300s deployment/chrome-service-api -n ${NAMESPACE}
        
        # Get the route URL
        APP_URL=$(oc get route chrome-service-api -n ${NAMESPACE} -o jsonpath='{.spec.host}')
        APP_URL="https://${APP_URL}"
        
        echo "Testing against: ${APP_URL}"
        
        # Set E2E environment variables
        export E2E_BASE_URL="${APP_URL}"
        export E2E_USER_ID="$(params.e2e-test-user-id)"
        export E2E_ACCOUNT_ID="$(params.e2e-test-account-id)"
        export E2E_ORG_ID="$(params.e2e-test-org-id)"
        export E2E_USERNAME="$(params.e2e-test-username)"
        
        # Run E2E tests
        cd e2e
        go mod download
        go test -v ./...
```

### Phase 5: Alternative - Custom Pipeline Extension

Create a custom pipeline that extends the existing one:

```yaml
# .tekton/chrome-service-e2e-pipeline.yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: chrome-service-build-and-e2e
spec:
  params:
    - name: git-url
    - name: revision
    - name: output-image
    # ... other params ...
  
  workspaces:
    - name: workspace
    - name: git-auth
  
  tasks:
    # Reference the existing docker-build pipeline
    - name: build-and-unit-test
      taskRef:
        name: docker-build
      params:
        - name: git-url
          value: $(params.git-url)
        - name: revision
          value: $(params.revision)
        # ... other params ...
      workspaces:
        - name: workspace
          workspace: workspace
    
    # Deploy to ephemeral environment
    - name: deploy-ephemeral
      runAfter:
        - build-and-unit-test
      taskRef:
        name: bonfire-deploy
      params:
        - name: image
          value: $(params.output-image)
        - name: namespace
          value: $(params.ephemeral-namespace)
      workspaces:
        - name: workspace
          workspace: workspace
    
    # Run E2E tests
    - name: e2e-tests
      runAfter:
        - deploy-ephemeral
      taskRef:
        name: chrome-service-e2e-tests
      params:
        - name: APP_URL
          value: $(tasks.deploy-ephemeral.results.route-url)
        - name: TEST_USER_ID
          valueFrom:
            secretKeyRef:
              name: chrome-e2e-test-credentials
              key: user-id
        # ... other params ...
      workspaces:
        - name: source
          workspace: workspace
```

## Implementation Roadmap

### Step 1: Setup Test Infrastructure
- [ ] Create test user in stage environment
- [ ] Create Kubernetes secret with test credentials
- [ ] Verify test user can access stage chrome-service API

### Step 2: Create E2E Test Task
- [ ] Create `.tekton/tasks/e2e-tests.yaml`
- [ ] Test task locally using `tkn task start`

### Step 3: Determine Deployment Strategy
- [ ] Confirm if using Bonfire for ephemeral deployments
- [ ] Determine how to get deployed service URL
- [ ] Test URL retrieval method

### Step 4: Integrate into Pipeline
**Option A**: Use existing pipeline with custom script
- [ ] Add `e2e-tests-script` parameter to existing pipeline
- [ ] Add E2E test credentials to pipeline params

**Option B**: Create custom pipeline
- [ ] Create new pipeline definition
- [ ] Update PipelineRun to use new pipeline
- [ ] Test in PR

### Step 5: Validation
- [ ] Create test PR to trigger pipeline
- [ ] Verify E2E tests run after deployment
- [ ] Verify test results are captured
- [ ] Fix any connectivity/auth issues

### Step 6: Production Rollout
- [ ] Enable E2E tests on PR pipeline
- [ ] Enable E2E tests on push pipeline (optional)
- [ ] Document process for team
- [ ] Set up alerting for E2E test failures

## Key Questions to Answer

1. **Ephemeral Deployment**: 
   - Is Bonfire used for ephemeral environments in PRs?
   - What namespace pattern is used?
   - How do you get the route URL?

2. **Test User Setup**:
   - Can we create a dedicated service account/test user?
   - What permissions does it need?
   - Is it shared across PRs or unique per ephemeral environment?

3. **Pipeline Integration**:
   - Should E2E tests run on every PR or only on main?
   - What should happen if E2E tests fail? (Block merge vs. warning)
   - Should we run against stage or ephemeral?

4. **Stage Environment Access**:
   - Can the Konflux pipeline access stage environment?
   - Are there network policies blocking access?
   - Do we need VPN/proxy configuration?

## Troubleshooting Guide

### E2E Tests Can't Reach Deployed Service
**Symptom**: Connection timeout or refused  
**Solutions**:
- Verify route exists: `oc get routes -n $NAMESPACE`
- Check pod is running: `oc get pods -n $NAMESPACE`
- Verify network policies allow ingress
- Check if TLS certificates are valid

### Authentication Failures
**Symptom**: 401 Unauthorized  
**Solutions**:
- Verify test user credentials are correct
- Check x-rh-identity header is properly formatted
- Verify test user exists in the target environment
- Check if user has required permissions

### Tests Timeout Waiting for Deployment
**Symptom**: `oc wait` times out  
**Solutions**:
- Check deployment status: `oc get deployment chrome-service-api`
- Check pod logs: `oc logs -l app=chrome-service`
- Verify database migration completed
- Check if initContainers succeeded

### Database Connection Issues
**Symptom**: Service starts but can't connect to database  
**Solutions**:
- Verify Clowder config is correct
- Check database secret exists
- Verify service account has database permissions
- Check database is running in ClowdEnvironment

## Next Steps

1. **Discuss with team**:
   - Which approach to use (post-deploy vs in-pipeline)
   - Ephemeral vs stage testing
   - Test user creation process

2. **Proof of concept**:
   - Manually deploy to ephemeral namespace
   - Run E2E tests against it
   - Document any issues

3. **Pipeline integration**:
   - Create E2E test task
   - Add to pipeline
   - Test with PR

4. **Documentation**:
   - Update team docs
   - Create runbook for failures
   - Document test user management
