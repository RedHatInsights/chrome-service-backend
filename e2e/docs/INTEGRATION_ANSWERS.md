# E2E Integration: Answers to Key Questions

This document addresses the three critical questions for integrating E2E tests into the Konflux CI pipeline.

## Question 1: Pipeline Integration

**How do we integrate with the existing build pipeline?**

### Current Pipeline

Your existing pipeline uses `docker-build-run-unit-tests-dynamic-env.yaml` from konflux-pipelines:
- Builds Docker image
- Runs unit tests with PostgreSQL/Unleash sidecars
- Pushes image to Quay

### Integration Options

#### Option A: Extend Existing Pipeline with Post-Build Script (Easiest)

Add E2E tests as a post-build step in the existing pipeline:

**Pros**:
- Minimal changes to existing pipeline
- Reuses existing infrastructure
- Quick to implement

**Cons**:
- Tests run in build container, not against deployed service
- Requires service to be running in same pod

**Implementation**:
Add to `.tekton/chrome-service-pull-request.yaml`:

```yaml
spec:
  params:
    - name: post-build-script
      value: |
        #!/bin/bash
        set -e
        
        # Start service in background
        /app/chrome-service-backend &
        SERVICE_PID=$!
        
        # Wait for service to be ready
        for i in {1..30}; do
          if curl -f http://localhost:8000/health; then
            break
          fi
          sleep 2
        done
        
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

**Note**: This approach has limitations - it tests against a local instance, not the deployed service with real Clowder configuration.

#### Option B: Custom Pipeline with Deployment Step (Recommended)

Create a custom pipeline that:
1. Builds image
2. Runs unit tests
3. **Deploys to ephemeral environment**
4. **Runs E2E tests against deployment**

**Pros**:
- Tests against real deployment with Clowder
- Validates actual deployment configuration
- Most realistic testing environment

**Cons**:
- Requires custom pipeline definition
- More complex setup
- Longer pipeline execution time

**Implementation**:
See `.tekton/tasks/e2e-tests.yaml` and `KONFLUX_INTEGRATION.md` for complete pipeline.

#### Option C: Separate Post-Deployment Job

Run E2E tests as a separate pipeline that triggers after deployment:

**Pros**:
- Decouples testing from build
- Can run independently
- Easier to debug

**Cons**:
- Requires orchestration between pipelines
- PR status not tied to E2E results

### Recommendation

**For MVP**: Start with **Option A** (post-build script) to validate tests work in CI environment.

**For Production**: Move to **Option B** (custom pipeline) to test against real deployments.

---

## Question 2: Spinning Up API Instance

**How do we "spin up" an instance of the API with our code changes?**

### Approach 1: Use Existing Deployment Mechanism

Your ClowdApp (`deploy/clowdapp.yml`) already defines how to deploy chrome-service. Two deployment targets:

#### A. Ephemeral Environment (Per-PR)

**If using Bonfire**:

```bash
# Deploy PR image to ephemeral namespace
bonfire deploy chrome-service \
  --source=appsre \
  --ref-env insights-stage \
  --set-image-tag on-pr-<git-revision> \
  --namespace ephemeral-<pr-number>

# Get service URL
ROUTE=$(oc get route chrome-service-api -n ephemeral-<pr-number> -o jsonpath='{.spec.host}')
APP_URL="https://${ROUTE}"
```

**Pros**:
- Isolated environment per PR
- Full Clowder integration (DB, Kafka, Unleash)
- Realistic testing environment

**Cons**:
- Requires Bonfire setup
- Resource intensive
- Slower pipeline

#### B. Stage Environment

Use the existing stage deployment:

```bash
# PR image is deployed to stage automatically
# Or manually trigger deployment

# Test against stage
APP_URL="https://console.stage.redhat.com"
```

**Pros**:
- Simple - no additional infrastructure
- Tests against real stage environment
- Faster than ephemeral

**Cons**:
- Shared environment (concurrent PRs conflict)
- Can't test breaking changes safely
- Requires deployment to stage first

### Approach 2: In-Pod Service

Run the service in the same pod as the tests:

```yaml
# In pipeline task
steps:
  - name: start-service
    image: $(params.output-image)
    script: |
      # Start service with test database
      /app/chrome-service-backend &
      
      # Wait for ready
      while ! curl -f http://localhost:8000/health; do sleep 1; done
```

**Pros**:
- Fast - no external deployment
- Self-contained
- Simple to implement

**Cons**:
- Doesn't test real Clowder configuration
- Requires sidecar PostgreSQL/Unleash/Kafka
- Not representative of production

### Connectivity to Stage Resources

**Challenge**: Tests need to connect to stage resources (DB, Kafka, Unleash).

**Solutions**:

1. **Ephemeral with Clowder**: Clowder automatically provisions resources
   ```yaml
   # ClowdApp spec already defines dependencies
   database:
     name: chrome-service
   featureFlags: true
   kafkaTopics: [...]
   ```

2. **Connect to Stage Services**: Configure service to use stage resources
   ```yaml
   env:
     - name: CLOWDER_ENABLED
       value: "true"
     # Clowder injects DB/Kafka/Unleash configs
   ```

3. **Mock External Dependencies**: For in-pod approach
   ```yaml
   # Use test doubles
   - PostgreSQL sidecar (already in pipeline)
   - Unleash sidecar (already in pipeline)
   - Mock Kafka (optional)
   ```

### Recommendation

**For MVP**: Use **In-Pod Service** (Approach 2) - reuse existing PostgreSQL/Unleash sidecars.

**For Production**: Use **Ephemeral Environment** (Approach 1A) via Bonfire for realistic testing.

---

## Question 3: Executing Tests Against API

**How do we execute the tests against the API instance?**

### Execution Environment

Tests run in a Go container with network access to the API.

### Implementation Options

#### Option 1: Direct Test Execution (Simplest)

Run tests directly from the workspace:

```yaml
steps:
  - name: run-e2e-tests
    image: golang:1.26
    workingDir: $(workspaces.source.path)
    env:
      - name: E2E_BASE_URL
        value: $(params.APP_URL)
      - name: E2E_USER_ID
        value: $(params.TEST_USER_ID)
      # ... other credentials
    script: |
      cd e2e
      go mod download
      go test -v ./...
```

**Pros**:
- Simple
- Direct access to source code
- Easy to debug

**Cons**:
- Requires Go toolchain
- Rebuilds test binary each run

#### Option 2: Pre-Built Test Binary

Build test binary during image build:

```dockerfile
# In Dockerfile
RUN cd e2e && go test -c -o /go/bin/chrome-e2e-tests
```

```yaml
# In task
steps:
  - name: run-e2e-tests
    image: $(params.output-image)
    script: |
      /usr/bin/chrome-e2e-tests -test.v
```

**Pros**:
- Faster execution (pre-compiled)
- Smaller runtime image possible
- Tests bundled with service

**Cons**:
- Larger service image
- Test dependencies in production image
- Less flexible

#### Option 3: Separate Test Container

Build dedicated test container:

```dockerfile
# Dockerfile.e2e
FROM golang:1.26
WORKDIR /tests
COPY e2e/ .
RUN go mod download
CMD ["go", "test", "-v", "./..."]
```

**Pros**:
- Clean separation
- Can version test image separately
- Reusable across pipelines

**Cons**:
- Additional image to build/maintain
- More complex pipeline

### Network Access

**Scenario A: In-Pod Service**
```yaml
# Service and tests in same pod
- name: run-tests
  env:
    - name: E2E_BASE_URL
      value: "http://localhost:8000"
```

**Scenario B: Deployed Service**
```yaml
# Tests access via route/ingress
- name: run-tests
  env:
    - name: E2E_BASE_URL
      value: "https://chrome-service-api-ephemeral-123.apps.cluster.example.com"
```

**Scenario C: Service in Different Namespace**
```yaml
# Tests access via service DNS
- name: run-tests
  env:
    - name: E2E_BASE_URL
      value: "http://chrome-service-api.ephemeral-namespace.svc.cluster.local:8000"
```

### Authentication

Tests need valid `x-rh-identity` header:

```yaml
# From secret
- name: E2E_USER_ID
  valueFrom:
    secretKeyRef:
      name: chrome-e2e-test-credentials
      key: user-id

# Or from parameter (less secure)
- name: E2E_USER_ID
  value: "test-user-123"
```

### Recommendation

**For MVP**: Use **Option 1** (Direct Test Execution) with in-pod service.

**For Production**: Use **Option 1** with deployed service (ephemeral or stage).

---

## Complete Implementation Plan

### Phase 1: MVP (Quick Win)

**Goal**: Get E2E tests running in CI as fast as possible.

1. **Add post-build script** to existing pipeline
2. **Run service in background** in test step
3. **Execute tests against localhost**
4. **Use hardcoded test credentials**

**Timeline**: 1-2 days

```yaml
# .tekton/chrome-service-pull-request.yaml
spec:
  params:
    - name: post-build-script
      value: |
        #!/bin/bash
        set -e
        
        # Start dependencies (already running as sidecars)
        # PostgreSQL on localhost:5432
        # Unleash on localhost:4242
        
        # Run migrations
        export PGSQL_HOSTNAME=localhost
        export PGSQL_USER=chrome
        export PGSQL_PASSWORD=chrome
        export PGSQL_DATABASE=db
        export PGSQL_PORT=5432
        export UNLEASH_API_TOKEN=default:development.unleash-insecure-api-token
        
        /go/bin/chrome-migrate
        
        # Start service in background
        /go/bin/chrome-service-backend &
        SERVICE_PID=$!
        
        # Wait for health check
        for i in {1..30}; do
          if curl -f http://localhost:8000/health 2>/dev/null; then
            echo "Service is ready"
            break
          fi
          sleep 2
        done
        
        # Run E2E tests
        export E2E_BASE_URL="http://localhost:8000"
        export E2E_USER_ID="test-user-123"
        export E2E_ACCOUNT_ID="123456"
        export E2E_ORG_ID="654321"
        export E2E_USERNAME="testuser"
        
        cd e2e
        go mod download
        go test -v ./...
        
        # Cleanup
        kill $SERVICE_PID
```

### Phase 2: Production (Full Integration)

**Goal**: Test against realistic environment.

1. **Deploy to ephemeral** using Bonfire
2. **Run E2E test task** against ephemeral
3. **Use secret-based credentials**
4. **Report results to PR**

**Timeline**: 1-2 weeks

See `KONFLUX_INTEGRATION.md` for complete implementation.

---

## Decision Matrix

| Aspect | MVP Approach | Production Approach |
|--------|-------------|-------------------|
| **Pipeline** | Extend existing | Custom pipeline |
| **API Instance** | In-pod service | Ephemeral deployment |
| **Dependencies** | Existing sidecars | Clowder-provisioned |
| **Test Execution** | Direct from source | Tekton task |
| **Credentials** | Hardcoded | Kubernetes secret |
| **Environment** | Localhost | Stage-like ephemeral |
| **Timeline** | 1-2 days | 1-2 weeks |

## Recommendations

1. **Start with MVP**: Get tests running quickly to validate approach
2. **Iterate to Production**: Add proper deployment once MVP works
3. **Document learnings**: Update this guide with real deployment URLs, credentials, etc.
4. **Monitor and adjust**: Track test reliability, execution time, resource usage

## Action Items

- [ ] Decide: MVP or Production approach first?
- [ ] Confirm: Can we use existing PostgreSQL/Unleash sidecars?
- [ ] Setup: Create test user in stage environment
- [ ] Create: Kubernetes secret for test credentials
- [ ] Test: Run MVP implementation in test PR
- [ ] Measure: Execution time and resource usage
- [ ] Iterate: Move to production approach if needed
