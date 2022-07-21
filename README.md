# chrome-service-backend
Source code repository for chrome backend

## Setup steps for xander

1. Initialize new go project using go.mode
2. Use the latest version of GO. You can use [gvm](https://github.com/moovweb/gvm) for version management
3. Create a simple app useing the [chi](https://github.com/go-chi/chi) router
4. Create two GET endpoints
  1. /health - will be eventually used for readiness/liveliness probe
  2. /api/chrome/v1/hello-world - make sure you use sub router for the `/api/chrome/v1/` part
5. Prapare a PostgreSQL 14 DB connection. You can use [quickstarts](https://github.com/RedHatInsights/quickstarts/blob/main/pkg/database/db.go) as an example. Note: Do not create any tables in the database
6. Install and setup [GORM](https://gorm.io/index.html)
7. Add a readme guide to install and start the API
