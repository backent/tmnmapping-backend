# Integration Tests

Integration tests for the TMN Mapping Backend. They boot the full HTTP server against a real PostgreSQL + PostGIS database and exercise end-to-end request flows.

---

## How It Differs from Unit Tests

| | Unit tests | Integration tests |
|---|---|---|
| Location | `services/*/` | `integration/` |
| Database | sqlmock (no real DB) | Real PostgreSQL + PostGIS |
| Runs on | Any machine, no Docker | Requires Docker |
| Command | `go test ./...` | `go test -tags integration ./integration/...` |

---

## Prerequisites — What You Need to Install

Before running integration tests, make sure the following are installed and ready:

### 1. Go 1.23+
```bash
go version   # should print go1.23 or higher
```
If not installed: https://go.dev/dl/

### 2. Docker
Integration tests require Docker to run the PostgreSQL + PostGIS container.

- **macOS / Windows**: Install [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- **Linux**: Install [Docker Engine](https://docs.docker.com/engine/install/)

Verify Docker is running:
```bash
docker ps    # should list containers (or empty list) without errors
```

---

## Step-by-Step Setup

### Step 1 — Clone the repository and enter the project directory
```bash
git clone <repo-url>
cd tmnmapping-backend
```

### Step 2 — Download Go dependencies
```bash
go mod download
```

### Step 3 — Create a test database in PostgreSQL

Integration tests connect to a database named `tmn_test`. You need to create it in the PostGIS container.

**Option A — Use the existing docker-compose (recommended)**

The repo includes `docker-compose.postgres.yml` which starts a PostGIS container. Adapt it to also expose a `tmn_test` database, or simply create the database after starting it:

```bash
# Start the PostGIS container
docker compose -f docker-compose.postgres.yml up -d

# Wait for it to be healthy (takes ~5 seconds)
docker compose -f docker-compose.postgres.yml ps

# Create the test database
docker exec -it tmnmapping-backend-postgres-1 psql -U postgres -c "CREATE DATABASE tmn_test;"
```

> **Tip:** You can also set `POSTGRES_DATABASE=tmn_backend` to run tests against the main dev database (isolated by truncating tables between tests). However, a separate `tmn_test` database is safer.

**Option B — Use any local PostgreSQL + PostGIS installation**

Point the integration tests at it using environment variables (see below).

### Step 4 — Run the integration tests
```bash
go test -tags integration -timeout 180s -v ./integration/...
```

On the **first run**, the test suite will automatically:
- Connect to the database
- Run all SQL migrations (from `database/migrations/*.up.sql`)
- Seed a test admin user
- Start the HTTP server in-process
- Run all tests

Subsequent runs are faster because migrations are idempotent (`IF NOT EXISTS` / `IF EXISTS` guards).

---

## Environment Variables

The integration tests read the same environment variables as the application. All have sensible defaults matching the docker-compose setup, so **no `.env` file is needed** for integration tests.

| Variable | Default | Description |
|---|---|---|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `postgres` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `adminlocal` | PostgreSQL password |
| `POSTGRES_DATABASE` | `tmn_test` | **Test database name** |
| `POSTGRES_SSLMODE` | `disable` | SSL mode |
| `APP_SECRET_KEY` | `integration-test-secret-key-32chars!!` | JWT signing key |
| `APP_TOKEN_EXPIRE_IN_SEC` | `3600` | Token expiry |

To override, export them before running tests:
```bash
POSTGRES_HOST=myhost POSTGRES_DATABASE=my_test_db \
  go test -tags integration -timeout 180s -v ./integration/...
```

---

## Running Specific Domains

```bash
# Auth tests only
go test -tags integration -run TestLogin      -v ./integration/...
go test -tags integration -run TestCurrent    -v ./integration/...
go test -tags integration -run TestLogout     -v ./integration/...

# POI tests only (exercises real PostGIS SQL)
go test -tags integration -run TestPOI        -v ./integration/...

# Building tests
go test -tags integration -run TestBuilding   -v ./integration/...

# Sales package tests
go test -tags integration -run TestSalesPackage -v ./integration/...

# Building restriction tests
go test -tags integration -run TestBuildingRestriction -v ./integration/...

# Saved polygon tests
go test -tags integration -run TestSavedPolygon -v ./integration/...

# Dashboard tests
go test -tags integration -run TestDashboard  -v ./integration/...
```

---

## Running Unit Tests (No Docker Required)

Unit tests do not need a database or Docker:
```bash
go test ./...
```

---

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `cannot reach test database` | PostgreSQL container not running | Run `docker compose -f docker-compose.postgres.yml up -d` |
| `database "tmn_test" does not exist` | Test DB not created | Run `docker exec ... psql -U postgres -c "CREATE DATABASE tmn_test;"` |
| `context deadline exceeded` | Container takes too long to start | Increase `-timeout` flag or check Docker resources |
| `exec 002_create_buildings_table.up.sql: ERROR: could not open extension control file "postgis.control"` | Container is using plain postgres (no PostGIS) | Ensure you are using the PostGIS-enabled container from `docker-compose.postgres.yml` |
| Port `5432` already in use | Another Postgres is running | Set `POSTGRES_PORT` to another mapped port, or stop the conflicting service |

---

## CI (Jenkins)

The `Jenkinsfile` includes an **Integration Test** stage that:
1. Starts the PostGIS container from `docker-compose.postgres.yml`
2. Creates the `tmn_test` database
3. Runs `go test -tags integration ./integration/...`
4. Always tears down the container afterward

The integration test stage runs **after unit tests** and **before the Docker image is pushed**, so a failing integration test blocks the deploy.
