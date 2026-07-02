# Chrome Service Backend — Agent Guide

## Project Overview

Go backend service for the Red Hat Hybrid Cloud Console (HCC) Chrome UI framework. Provides user identity management, favorite pages, dashboard templates, last-visited tracking, recently-used workspaces, self-reporting, WebSocket-based real-time messaging, and static navigation/services configuration.

**Module**: `github.com/RedHatInsights/chrome-service-backend`

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go (see `go.mod` for version) |
| HTTP Router | [chi/v5](https://github.com/go-chi/chi) |
| ORM | [GORM](https://gorm.io/) (PostgreSQL in prod, SQLite in tests) |
| Messaging | [kafka-go](https://github.com/segmentio/kafka-go) (Segmentio) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Feature Flags | [Unleash client](https://github.com/Unleash/unleash-client-go) |
| Metrics | [Prometheus client_golang](https://github.com/prometheus/client_golang) |
| Logging | [logrus](https://github.com/sirupsen/logrus) |
| Config | [godotenv](https://github.com/joho/godotenv) + [Clowder](https://github.com/redhatinsights/app-common-go) |
| Testing | `testing` stdlib + [testify](https://github.com/stretchr/testify) |
| JSON Schema | [jsonschema/v5](https://github.com/santhosh-tekuri/jsonschema) + [gojsonschema](https://github.com/xeipuuv/gojsonschema) |

## Directory Structure

```
.
├── main.go                  # Entry point: chi router setup, middleware, route mounting
├── config/
│   └── config.go            # Configuration (env vars, Clowder, Kafka, Unleash, Intercom)
├── cmd/                     # CLI subcommands
│   ├── kafka/               # Kafka consumer runner
│   ├── migrate/             # Database migration runner
│   ├── search/              # Search index generator/publisher
│   ├── services/            # services-generated.json parser
│   ├── static/              # Static file server
│   ├── unleash/             # Unleash feature flag seeder
│   └── validate/            # JSON schema validator
├── rest/
│   ├── cloudevents/         # CloudEvents message formatting
│   ├── connectionhub/       # WebSocket connection hub (rooms, targets, connections)
│   ├── database/
│   │   └── db.go            # GORM DB init (PostgreSQL / SQLite), table migration
│   ├── featureflags/        # Unleash feature flag client wrapper
│   ├── kafka/
│   │   └── consumers.go     # Kafka consumer initialization
│   ├── logger/              # Custom chi request logger
│   ├── middleware/
│   │   ├── headers.go       # ParseHeaders: extracts x-rh-identity header
│   │   └── identity.go      # InjectUser: resolves/creates UserIdentity from header
│   ├── models/              # GORM models
│   │   ├── base.go          # BaseModel (ID, timestamps, soft delete)
│   │   ├── UserIdentity.go  # User identity with visited bundles, workspaces, etc.
│   │   ├── FavoritePage.go  # User favorite pages
│   │   ├── DashboardTemplate.go  # Widget dashboard templates with grid layouts
│   │   ├── SelfReport.go    # User self-report data
│   │   └── ProductsOfInterest.go # Product interest tracking
│   ├── routes/              # HTTP route handlers
│   │   ├── dashboardTemplate.go  # CRUD for dashboard templates
│   │   ├── favoritePage.go       # Get/set favorite pages
│   │   ├── identity.go           # User identity routes
│   │   ├── lastVisited.go        # Last visited pages tracking
│   │   ├── recentlyUsedWorkspaces.go # Recently used workspaces
│   │   ├── selfReport.go         # Self-report routes
│   │   └── websocket.go          # WebSocket upgrade endpoint
│   ├── service/             # Business logic layer
│   │   ├── identity.go           # User identity CRUD, Intercom hash, bundles
│   │   ├── dashboardTemplate.go  # Dashboard template operations
│   │   ├── baseLayoutLoader.go   # Load base dashboard layouts from YAML
│   │   ├── favoritePageService.go
│   │   ├── lastVisitedService.go
│   │   ├── selfReport.go
│   │   └── workspacesService.go
│   └── util/                # Shared utilities
│       ├── const.go              # Constants (context keys, header names)
│       ├── responses.go          # Generic list/item response structs
│       ├── xrh_header_parser.go  # X-RH-Identity header decoder
│       ├── user_identity_cache.go # In-memory user identity cache
│       ├── createChromeConfiguration.go # Chrome config file builder
│       └── testutils.go          # Shared test helpers
├── static/                  # Static navigation/services JSON files
├── spec/                    # OpenAPI spec
├── local/                   # Docker/Podman compose files for local infra
├── docs/                    # Documentation (see below)
├── e2e/                     # End-to-end API tests
│   ├── config.go            # E2E test configuration
│   ├── utils.go             # Test client and helper utilities
│   ├── identity_test.go     # User identity endpoint tests
│   ├── favoritePage_test.go # Favorite pages endpoint tests
│   ├── lastVisited_test.go  # Last visited pages endpoint tests
│   ├── recentlyUsedWorkspaces_test.go # Workspaces endpoint tests
│   ├── selfReport_test.go   # Self report endpoint tests
│   └── README.md            # E2E test documentation
└── Makefile                 # Build, test, run targets
```

## Architecture

### Request Flow

1. **main.go** initializes: godotenv → database → user cache → chrome config → feature flags → chi router
2. Public routes (no auth): `/health`, `/api/chrome-service/v1/static/*`, `/api/chrome-service/v1/spec/*`
3. Authenticated routes (`/api/chrome-service/v1/`):
   - `ParseHeaders` middleware: extracts and decodes `x-rh-identity` base64 header
   - `InjectUser` middleware: resolves `UserIdentity` from DB (with in-memory cache), injects into context
   - Route handlers in `rest/routes/` call service functions in `rest/service/`
   - Services interact with DB via global `database.DB` (GORM)
4. WebSocket route (`/wss/chrome-service/v1/ws`): gated by Unleash flag `chrome-service.websockets.enabled`
5. Kafka consumers feed messages to WebSocket connection hub for real-time broadcasting

### Key Patterns

- **Global DB**: `database.DB` is a package-level `*gorm.DB` variable initialized in `database.Init()`
- **Context-based auth**: User identity passed via `context.WithValue` using typed keys from `util/const.go`
- **Chi sub-routers**: Each resource gets a `MakeXxxRoutes(sub chi.Router)` function mounted in `main.go`
- **Service layer**: Route handlers delegate to `rest/service/` functions for business logic
- **GORM models**: All models embed `BaseModel` (ID, CreatedAt, UpdatedAt, DeletedAt for soft delete)
- **Config singleton**: `config.Get()` returns a pointer to the global `ChromeServiceConfig`
- **User identity cache**: In-memory cache (`util.UsersCache`) reduces DB queries; invalidated on updates

## Build & Run Commands

```bash
# Run server (seeds Unleash, generates search index, parses services)
make dev

# Run all tests with coverage
make test
# Equivalent to: go test -v ./... -coverprofile=c.out

# Run E2E tests against remote environment
make test-e2e
# Requires E2E_BASE_URL and other E2E_* env vars to be set

# Run database migration
make migrate

# Validate JSON schemas
make validate-schema

# Generate search index
make generate-search-index

# Parse services config
make parse-services

# Start local infrastructure (PostgreSQL, Kafka, Unleash)
make infra

# Tear down local infra
make clean-all

# Run security audit
make audit
```

## Testing Conventions

### Unit Tests

- **Framework**: stdlib `testing` + `testify/assert`
- **Database**: Tests use SQLite (set via `config.Get().Test = true` and `config.Get().DbName`)
- **TestMain pattern**: Each test package that needs DB access implements `TestMain(m *testing.M)`:
  1. Set `cfg.Test = true`
  2. Set `cfg.DashboardConfig.TemplatesWD = "../../"` (relative to test file location)
  3. Create timestamped SQLite DB name: `fmt.Sprintf("%d-services.db", time.Now().UnixNano())`
  4. Call `database.Init()` and `database.DB.AutoMigrate(...)` for needed models
  5. Run tests: `exitCode := m.Run()`
  6. Clean up: `os.Remove(dbName)`
  7. `os.Exit(exitCode)`
- **Test naming**: Descriptive strings in `t.Run("Should do X when Y", ...)`
- **Assertions**: Use `assert.Nil`, `assert.NotNil`, `assert.Equal`, `assert.True`, `assert.Contains`
- **Mock data**: Created directly via `database.DB.Create(...)` — no external mock frameworks
- **Test file locations**: Co-located with source files (`*_test.go` in same package)
- **Run specific test**: `go test -v -run TestFunctionName ./rest/service/`

### E2E Tests

- **Location**: `/e2e` directory
- **Purpose**: Happy path API testing against running service instances
- **Configuration**: Environment variables (`E2E_BASE_URL`, `E2E_USER_ID`, etc.)
- **Environments**: Can test against local, stage, or production
- **Authentication**: Uses `x-rh-identity` header generated from test user credentials
- **Run command**: `make test-e2e` or `cd e2e && go test -v ./...`
- **Documentation**: See `e2e/README.md` and `e2e/KONFLUX.md`
- **Coverage**: Identity, favorite pages, last-visited, workspaces, self-report endpoints

## Configuration

Environment variables loaded from `.env` file (see `.env.example`). Key variables:

| Variable | Purpose |
|----------|---------|
| `PGSQL_USER`, `PGSQL_PASSWORD`, `PGSQL_HOSTNAME`, `PGSQL_PORT`, `PGSQL_DATABASE` | PostgreSQL connection |
| `UNLEASH_API_TOKEN`, `UNLEASH_ADMIN_TOKEN` | Feature flag client/admin tokens |
| `LOG_LEVEL` | Logging level (info/debug/error) |
| `TEMPLATES_WD` | Working directory for dashboard template YAML files |
| `INTERCOM_*` | Intercom integration keys per product bundle |
| `RECENTLY_USED_WORKSPACES_MAX_SAVED` | Max recently-used workspaces stored (default: 10) |

In production, config is loaded from Clowder (`clowder.LoadedConfig`) instead of env vars.

## API Endpoints

All authenticated endpoints require `x-rh-identity` header (base64-encoded JSON with `identity.user.user_id`).

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (no auth) |
| GET | `/api/chrome-service/v1/static/*` | Static navigation files (no auth) |
| GET | `/api/chrome-service/v1/spec/*` | OpenAPI spec (no auth) |
| GET/POST | `/api/chrome-service/v1/last-visited` | Last visited pages |
| GET/POST | `/api/chrome-service/v1/recently-used-workspaces` | Recently used workspaces |
| GET/POST | `/api/chrome-service/v1/favorite-pages` | Favorite pages |
| GET/POST | `/api/chrome-service/v1/self-report` | Self-report data |
| GET/PATCH | `/api/chrome-service/v1/user` | User identity & preferences |
| POST | `/api/chrome-service/v1/emit-message` | Broadcast message via WebSocket |
| GET/POST/PATCH/DELETE | `/api/chrome-service/v1/dashboard-templates` | Dashboard template CRUD |
| WS | `/wss/chrome-service/v1/ws` | WebSocket connection (feature-flagged) |

## Documentation Index

| File | Topic |
|------|-------|
| `docs/cloud-services-config.md` | Navigation and All Services configuration |
| `docs/dashboard-layouts.md` | Widget dashboard layout system |
| `docs/feature-flags.md` | Feature flag setup and usage |
| `docs/feo-migration-guide.md` | Frontend Operator migration guide |
| `docs/frontend-routes.md` | Frontend routing configuration |
| `docs/search-index.md` | Search index generation |
| `docs/sso-scopes.md` | SSO scope configuration |
| `docs/support-case-configuration.md` | Support case setup |
| `docs/user-identity.md` | User identity model details |
| `docs/websocket.md` | WebSocket implementation details |
| `docs/intercom-keys.md` | Intercom integration keys |
| `e2e/README.md` | E2E test suite documentation |
| `e2e/KONFLUX.md` | Konflux CI integration guide for E2E tests |

## Common Pitfalls

1. **Dashboard template tests**: `TemplatesWD` must be set relative to the test file's location (usually `../../`) so base layout YAML files can be found
2. **SQLite test cleanup**: Always remove the timestamped `.db` file in test teardown to avoid disk buildup
3. **x-rh-identity header**: All authenticated endpoint tests need a valid base64-encoded identity header
4. **Global DB variable**: `database.DB` is package-level — tests that call `database.Init()` replace it globally, so test packages with DB access should use `TestMain` for setup/teardown
5. **Feature flag gating**: WebSocket functionality requires `chrome-service.websockets.enabled` Unleash flag — it's checked at startup, not per-request
6. **Clowder vs local**: Config loading branches on `clowder.IsClowderEnabled()` — local dev uses env vars, production uses Clowder's `LoadedConfig`
