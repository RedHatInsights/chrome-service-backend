# E2E API Tests

End-to-end tests for the Chrome Service Backend API. These tests verify the "happy path" functionality of all major API endpoints by making real HTTP requests to a running instance of the service.

## Test Coverage

The E2E test suite covers the following endpoints:

### Identity (`/api/chrome-service/v1/user`)
- `GET /user` - Get user identity
- `POST /user/update-ui-preview` - Update UI preview preference
- `POST /user/mark-preview-seen` - Mark preview as seen
- `POST /user/update-active-workspace` - Update active workspace
- `POST /user/visited-bundles` - Add visited bundle
- `GET /user/visited-bundles` - Get visited bundles
- `GET /user/intercom` - Get Intercom hash

### Favorite Pages (`/api/chrome-service/v1/favorite-pages`)
- `GET /favorite-pages` - Get favorite pages (with various query parameters)
- `POST /favorite-pages` - Add/remove favorite page

### Last Visited (`/api/chrome-service/v1/last-visited`)
- `GET /last-visited` - Get last visited pages
- `POST /last-visited` - Store last visited pages

### Recently Used Workspaces (`/api/chrome-service/v1/recently-used-workspaces`)
- `GET /recently-used-workspaces` - Get recently used workspaces
- `POST /recently-used-workspaces` - Save recently used workspaces
- Validation tests for workspace payloads

### Self Report (`/api/chrome-service/v1/self-report`)
- `GET /self-report` - Get user self report
- `PATCH /self-report` - Update user self report

## Configuration

The tests are configured via environment variables. You can set these in your shell or create a `.env` file in the `e2e` directory.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `E2E_BASE_URL` | Base URL of the API to test | `http://localhost:8000` |
| `E2E_USER_ID` | Test user's ID for x-rh-identity header | `test-user-123` |
| `E2E_ACCOUNT_ID` | Test user's account ID | `123456` |
| `E2E_ORG_ID` | Test user's organization ID | `654321` |
| `E2E_USERNAME` | Test user's username | `testuser` |

### Example Configurations

#### Local Development
```bash
E2E_BASE_URL=http://localhost:8000
E2E_USER_ID=test-user-123
E2E_ACCOUNT_ID=123456
E2E_ORG_ID=654321
E2E_USERNAME=testuser
```

#### Stage Environment
```bash
E2E_BASE_URL=https://console.stage.redhat.com
E2E_USER_ID=<your-stage-user-id>
E2E_ACCOUNT_ID=<your-stage-account-id>
E2E_ORG_ID=<your-stage-org-id>
E2E_USERNAME=<your-stage-username>
```

#### Production Environment
```bash
E2E_BASE_URL=https://console.redhat.com
E2E_USER_ID=<your-prod-user-id>
E2E_ACCOUNT_ID=<your-prod-account-id>
E2E_ORG_ID=<your-prod-org-id>
E2E_USERNAME=<your-prod-username>
```

## Running the Tests

### Prerequisites

1. The target environment must be running and accessible
2. You must have valid user credentials for the environment
3. The user identity specified in the environment variables must have access to the API

### Local Execution

1. Copy the example environment file:
   ```bash
   cd e2e
   cp .env.example .env
   ```

2. Edit `.env` with your configuration

3. Run the tests:
   ```bash
   # From the repository root
   make test-e2e
   
   # Or from the e2e directory
   cd e2e
   go test -v ./...
   ```

### Running Against Different Environments

You can override environment variables at runtime:

```bash
# Test against local server
E2E_BASE_URL=http://localhost:8000 make test-e2e

# Test against stage
E2E_BASE_URL=https://console.stage.redhat.com \
E2E_USER_ID=<user-id> \
E2E_ACCOUNT_ID=<account-id> \
E2E_ORG_ID=<org-id> \
E2E_USERNAME=<username> \
make test-e2e
```

### Running Specific Tests

```bash
# Run only identity tests
cd e2e
go test -v -run TestGetUserIdentity

# Run only workspace tests
cd e2e
go test -v -run TestRecentlyUsedWorkspaces
```

## CI/CD Integration

### GitHub Actions

The tests can be integrated into GitHub Actions workflows:

```yaml
- name: Run E2E Tests
  env:
    E2E_BASE_URL: https://console.stage.redhat.com
    E2E_USER_ID: ${{ secrets.E2E_USER_ID }}
    E2E_ACCOUNT_ID: ${{ secrets.E2E_ACCOUNT_ID }}
    E2E_ORG_ID: ${{ secrets.E2E_ORG_ID }}
    E2E_USERNAME: ${{ secrets.E2E_USERNAME }}
  run: make test-e2e
```

### Konflux/Tekton

For Konflux CI, create a task that runs the E2E tests:

```yaml
- name: e2e-tests
  image: golang:1.26
  script: |
    #!/bin/bash
    export E2E_BASE_URL=https://console.stage.redhat.com
    export E2E_USER_ID=$(cat /secrets/e2e-user-id)
    export E2E_ACCOUNT_ID=$(cat /secrets/e2e-account-id)
    export E2E_ORG_ID=$(cat /secrets/e2e-org-id)
    export E2E_USERNAME=$(cat /secrets/e2e-username)
    make test-e2e
```

## Test Design

### Authentication

All tests use the `x-rh-identity` header for authentication. The header is automatically constructed from the environment variables and includes:
- User ID
- Account ID
- Organization ID
- Username
- Additional user metadata

### Test Client

The `TestClient` provides helper methods for:
- Creating authenticated requests
- Making HTTP requests with proper headers
- Asserting response status codes
- Unmarshaling JSON responses
- Error handling

### Test Organization

Each endpoint has its own test file:
- `identity_test.go` - User identity endpoints
- `favoritePage_test.go` - Favorite pages endpoints
- `lastVisited_test.go` - Last visited pages endpoints
- `recentlyUsedWorkspaces_test.go` - Recently used workspaces endpoints
- `selfReport_test.go` - Self report endpoints

### Test Data

Tests create and clean up their own data. Each test:
1. Creates test data via POST/PATCH requests
2. Verifies the data via GET requests
3. The backend handles cleanup via the user identity system

## Troubleshooting

### Tests fail with 401 Unauthorized

- Verify your `x-rh-identity` header is correctly formatted
- Check that the user credentials in the environment variables are valid
- Ensure the user has access to the target environment

### Tests fail with connection errors

- Verify the `E2E_BASE_URL` is correct and accessible
- Check network connectivity to the target environment
- Ensure the service is running and healthy

### Tests fail with 500 Internal Server Error

- Check the server logs for error details
- Verify the database is accessible and properly migrated
- Ensure all required services (Unleash, Kafka) are running

## Adding New Tests

To add tests for a new endpoint:

1. Create a new test file: `e2e/newEndpoint_test.go`
2. Import the necessary packages and define response types
3. Write test functions following the existing patterns
4. Use the `TestClient` helper methods for requests
5. Update this README with the new test coverage

Example:
```go
package e2e

import (
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewEndpoint(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET("/api/chrome-service/v1/new-endpoint")
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var response map[string]interface{}
	client.AssertJSONResponse(body, &response)
	
	// Add assertions
}
```
