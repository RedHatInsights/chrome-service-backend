# E2E API Tests

End-to-end tests for the Chrome Service Backend API. These tests verify the happy path functionality of all major API endpoints by making real HTTP requests to a running instance of the service.

## Quick Start

```bash
# Terminal 1: Start infrastructure (PostgreSQL, Kafka, Unleash)
make infra

# Terminal 2: Start the API server
make dev

# Terminal 3: Run E2E tests
make test-e2e
```

## Test Coverage

| Endpoint | Tests | File |
|----------|-------|------|
| `GET/PATCH /user` | 7 | `identity_test.go` |
| `GET/POST /favorite-pages` | 3 | `favoritePage_test.go` |
| `GET/POST /last-visited` | 3 | `lastVisited_test.go` |
| `GET/POST /recently-used-workspaces` | 6 | `recentlyUsedWorkspaces_test.go` |
| `GET/PATCH /self-report` | 2 | `selfReport_test.go` |

## Configuration

Tests are configured via environment variables. Defaults target a local instance.

| Variable | Default | Description |
|----------|---------|-------------|
| `E2E_BASE_URL` | `http://localhost:8000` | API base URL (no trailing slash) |
| `E2E_USER_ID` | `test-user-123` | User ID for x-rh-identity header |
| `E2E_ACCOUNT_ID` | `123456` | Account ID |
| `E2E_ORG_ID` | `654321` | Organization ID |
| `E2E_USERNAME` | `testuser` | Username |

You can also create an `e2e/.env` file (see `.env.example`).

## Running Tests

```bash
# All tests
make test-e2e

# Specific test file
cd e2e && go test -v ./identity_test.go ./utils.go ./config.go ./main_test.go

# Specific test
cd e2e && go test -v -run TestGetUserIdentity

# Against a different environment
E2E_BASE_URL=https://console.stage.redhat.com make test-e2e
```

## Adding New Tests

1. Create `e2e/newEndpoint_test.go`
2. Use the `TestClient` from `utils.go` for requests
3. Follow existing patterns for assertions
4. Update this README with coverage

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
}
```

## Troubleshooting

**Connection refused**: Make sure `make infra` and `make dev` are running.

**401 Unauthorized**: Check x-rh-identity header generation in `utils.go`. Verify environment variables match a valid test user.

**500 Internal Server Error**: Check server logs. Verify database migration ran (`make migrate`).
