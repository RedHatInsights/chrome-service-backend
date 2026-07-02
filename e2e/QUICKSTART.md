# E2E Tests - Quick Start Guide

## 5-Minute Setup

### 1. Configure Environment
```bash
cd e2e
cp .env.example .env
```

Edit `.env`:
```bash
E2E_BASE_URL=https://console.stage.redhat.com
E2E_USER_ID=<your-user-id>
E2E_ACCOUNT_ID=<your-account-id>
E2E_ORG_ID=<your-org-id>
E2E_USERNAME=<your-username>
```

### 2. Run Tests
```bash
# From repository root
make test-e2e

# Or from e2e directory
cd e2e
go test -v ./...
```

## Common Commands

```bash
# Run all E2E tests
make test-e2e

# Run specific test
cd e2e && go test -v -run TestGetUserIdentity

# Run tests with environment override
E2E_BASE_URL=http://localhost:8000 make test-e2e

# Run tests against stage
E2E_BASE_URL=https://console.stage.redhat.com make test-e2e
```

## Test Against Local Server

```bash
# Terminal 1: Start server
make dev

# Terminal 2: Run E2E tests
E2E_BASE_URL=http://localhost:8000 \
E2E_USER_ID=test-user-123 \
E2E_ACCOUNT_ID=123456 \
E2E_ORG_ID=654321 \
E2E_USERNAME=testuser \
make test-e2e
```

## Troubleshooting

### "401 Unauthorized"
→ Check your user credentials in `.env`

### "Connection refused"
→ Verify E2E_BASE_URL is correct and server is running

### "Test timeout"
→ Check network connectivity to target environment

## What Gets Tested

✅ User Identity (GET, PATCH, POST operations)  
✅ Favorite Pages (GET, POST with various params)  
✅ Last Visited Pages (GET, POST)  
✅ Recently Used Workspaces (GET, POST, validation)  
✅ Self Report (GET, PATCH)  

## Next Steps

- Read [README.md](README.md) for detailed documentation
- Read [KONFLUX.md](KONFLUX.md) for CI/CD integration
- Read [SUMMARY.md](SUMMARY.md) for project overview
