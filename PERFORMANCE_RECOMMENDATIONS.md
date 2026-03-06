# Performance Recommendations for TMN Mapping Backend

This document provides a comprehensive analysis of performance bottlenecks and actionable recommendations for the TMN Mapping Backend (Go / PostgreSQL + PostGIS / httprouter).

---

## Table of Contents

1. [Database & Query Optimization](#1-database--query-optimization)
2. [HTTP Server Configuration](#2-http-server-configuration)
3. [Connection Pool Tuning](#3-connection-pool-tuning)
4. [Caching Strategy](#4-caching-strategy)
5. [ERP Sync Optimization](#5-erp-sync-optimization)
6. [API Design & Response Optimization](#6-api-design--response-optimization)
7. [Concurrency & Resource Management](#7-concurrency--resource-management)
8. [Error Handling & Panic Recovery](#8-error-handling--panic-recovery)
9. [Observability & Monitoring](#9-observability--monitoring)
10. [Deployment & Infrastructure](#10-deployment--infrastructure)
11. [Security Improvements with Performance Impact](#11-security-improvements-with-performance-impact)
12. [Summary & Priority Matrix](#12-summary--priority-matrix)

---

## 1. Database & Query Optimization

### 1.1 Read-Only Transactions Used Everywhere (High Impact)

**Problem:** Every read operation (e.g., `FindAll`, `FindById`, `GetFilterOptions`) starts a full read-write transaction with `service.DB.Begin()`, even though no writes occur. This acquires row-level locks unnecessarily and limits connection pool throughput.

**Current code** (`services/building/service_building_impl.go:106-108`):
```go
tx, err := service.DB.Begin()
helpers.PanicIfError(err)
defer helpers.CommitOrRollback(tx)
```

**Recommendation:** For read-only operations, either:
- Use `service.DB.QueryContext()` directly (no transaction needed for single queries).
- Use `service.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})` for multi-query reads that need consistency.

```go
// Option A: Direct query (preferred for single queries)
building, err := service.RepositoryBuildingInterface.FindById(ctx, service.DB, id)

// Option B: Read-only transaction (for multi-statement reads)
tx, err := service.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
```

**Impact:** Reduces lock contention, allows PostgreSQL to optimize read paths, and frees up connections faster.

---

### 1.2 Duplicate Query in `FindAllForMapping` (High Impact)

**Problem:** The `FindAllForMapping` service method (`services/building/service_building_impl.go:614-832`) executes the **same complex spatial query twice** when map bounds are provided — once with bounds (for viewport data) and once without bounds (for totals). This is the most expensive endpoint since it involves PostGIS spatial operations.

**Current code** (`services/building/service_building_impl.go:736-761`):
```go
// First query: buildings in viewport
buildings, err := service.RepositoryBuildingInterface.FindAllForMapping(ctx, tx, ...)

// Second query: same filters but WITHOUT bounds (for totals)
if minLatPtr != nil && maxLatPtr != nil && minLngPtr != nil && maxLngPtr != nil {
    buildingsForTotals, err = service.RepositoryBuildingInterface.FindAllForMapping(
        ctx, tx, ..., nil, nil, nil, nil, // no bounds
    )
}
```

**Recommendation:** Add a separate `CountByBuildingType` repository method that uses `SELECT building_type, COUNT(*) ... GROUP BY building_type` instead of fetching all rows just to count them. This avoids transferring full row data for the totals.

```go
// New repository method
func (r *RepositoryBuildingImpl) CountByBuildingTypeForMapping(ctx context.Context, tx *sql.Tx,
    /* same filters minus bounds */) (map[string]int, error) {
    SQL := `SELECT COALESCE(LOWER(building_type), 'other'), COUNT(*) FROM buildings b ...
            GROUP BY COALESCE(LOWER(building_type), 'other')`
    // ...
}
```

**Impact:** Can reduce this endpoint's response time by 40-60% for large datasets.

---

### 1.3 `GetFilterOptions` Makes 10 Sequential Queries (Medium Impact)

**Problem:** `GetFilterOptions` (`services/building/service_building_impl.go:591-611`) calls `GetDistinctValues` 10 times in a loop, one for each column. Each is a separate query inside the same transaction.

```go
columns := []string{"building_status", "sellable", "connectivity", ...}
for _, column := range columns {
    values, err := service.RepositoryBuildingInterface.GetDistinctValues(ctx, tx, column)
}
```

**Recommendation:** Combine into a single query or use goroutines to run them concurrently:

```go
// Option A: Single query
SQL := `SELECT 'building_status' AS col, building_status AS val FROM buildings WHERE building_status IS NOT NULL
        UNION ALL
        SELECT 'sellable', sellable FROM buildings WHERE sellable IS NOT NULL
        ...
        ORDER BY col, val`

// Option B: Concurrent queries (use separate connections, not the same tx)
var wg sync.WaitGroup
results := make(map[string][]string)
var mu sync.Mutex
for _, column := range columns {
    wg.Add(1)
    go func(col string) {
        defer wg.Done()
        values, _ := service.RepositoryBuildingInterface.GetDistinctValues(ctx, tx, col)
        mu.Lock()
        results[col] = values
        mu.Unlock()
    }(column)
}
wg.Wait()
```

**Impact:** Reduces 10 round-trips to 1, or runs them in parallel. Response time improvement: ~5-8x for this endpoint.

---

### 1.4 Missing Composite Indexes (Medium Impact)

**Problem:** While individual column indexes exist on the `buildings` table, there are no composite indexes for the common query patterns in `FindAll` and `FindAllForMapping`. PostgreSQL cannot efficiently combine multiple B-tree indexes for AND conditions.

**Recommendation:** Add composite indexes for the most common filter combinations:

```sql
-- For FindAll with common filter + sort patterns
CREATE INDEX idx_buildings_status_name ON buildings(building_status, name);
CREATE INDEX idx_buildings_type_grade ON buildings(building_type, grade_resource);

-- For spatial queries with filters (partial index for non-null locations)
CREATE INDEX idx_buildings_location_type ON buildings USING GIST(location)
    WHERE location IS NOT NULL;

-- For latitude/longitude bounding box queries (used in viewport filtering)
CREATE INDEX idx_buildings_lat_lng ON buildings(latitude, longitude)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- For ILIKE queries (case-insensitive search), use pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_buildings_name_trgm ON buildings USING GIN(name gin_trgm_ops);
CREATE INDEX idx_buildings_citytown_trgm ON buildings USING GIN(citytown gin_trgm_ops);
CREATE INDEX idx_buildings_subdistrict_trgm ON buildings USING GIN(subdistrict gin_trgm_ops);

-- For dashboard queries (DISTINCT ON with ORDER BY)
CREATE INDEX idx_acquisitions_project_modified ON acquisitions(building_project, modified DESC NULLS LAST);
CREATE INDEX idx_building_proposals_project_modified ON building_proposals(building_project, modified DESC NULLS LAST);
CREATE INDEX idx_loi_project_modified ON letters_of_intent(building_project, modified DESC NULLS LAST);
```

**Impact:** Queries using ILIKE filters (`name ILIKE '%search%'`) will be significantly faster with trigram indexes. Dashboard `DISTINCT ON` queries will avoid costly sorts.

---

### 1.5 `LOWER()` Function Calls Prevent Index Usage (Medium Impact)

**Problem:** In `FindAllForMapping`, filters like `LOWER(building_type) IN (...)` and `LOWER(grade_resource) = $N` apply a function to the column, which prevents PostgreSQL from using the B-tree index on those columns.

**Current code** (`repositories/building/repository_building_impl.go:637`):
```go
whereConditions = append(whereConditions, `LOWER(building_type) IN (...)`)
```

**Recommendation:** Either:
- Create functional indexes: `CREATE INDEX idx_buildings_type_lower ON buildings(LOWER(building_type));`
- Or normalize data to lowercase on insert/update and compare directly without `LOWER()`.

---

### 1.6 `building_status` Column Has No Index on `buildings` Table for Dashboard Join (Low Impact)

**Problem:** The dashboard `GetByPersonAndType` query joins `buildings` on `project_name`:
```sql
LEFT JOIN buildings b ON b.project_name = l.building_project
```
There is no index on `buildings.project_name`.

**Recommendation:**
```sql
CREATE INDEX idx_buildings_project_name ON buildings(project_name);
```

---

## 2. HTTP Server Configuration

### 2.1 No Timeouts on HTTP Server (Critical)

**Problem:** The HTTP server in `main.go:63-66` has **no read, write, or idle timeouts** configured:
```go
server := http.Server{
    Addr:    ":" + APP_PORT,
    Handler: router,
}
```

Without timeouts, slow clients or stalled connections can exhaust server resources (goroutines, file descriptors, memory).

**Recommendation:**
```go
server := http.Server{
    Addr:         ":" + APP_PORT,
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
    ReadHeaderTimeout: 5 * time.Second,
    MaxHeaderBytes: 1 << 20, // 1 MB
}
```

**Impact:** Prevents resource exhaustion from slowloris attacks and stalled connections. This is a **production-critical** fix.

---

### 2.2 No Graceful Shutdown (High Impact)

**Problem:** The server starts with `ListenAndServe()` and panics on failure. There is no graceful shutdown handling, meaning in-flight requests (especially ERP syncs) will be abruptly terminated during deployment.

**Recommendation:**
```go
// Start server in goroutine
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.WithField("error", err.Error()).Fatal("Server failed to start")
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

logger.Info("Shutting down server...")
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := server.Shutdown(ctx); err != nil {
    logger.WithField("error", err.Error()).Fatal("Server forced to shutdown")
}
logger.Info("Server exited gracefully")
```

**Impact:** Zero-downtime deployments, no data corruption from interrupted sync operations.

---

### 2.3 No Response Compression (Medium Impact)

**Problem:** There is no gzip/deflate compression for API responses. Building listing endpoints can return large JSON payloads (especially `FindAllForMapping` which returns all buildings for the map view).

**Recommendation:** Add a compression middleware or wrap the handler. A simple approach using the standard library:

```go
import "compress/gzip"

func GzipMiddleware(next httprouter.Handle) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next(w, r, p)
            return
        }
        gz := gzip.NewWriter(w)
        defer gz.Close()
        w.Header().Set("Content-Encoding", "gzip")
        next(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r, p)
    }
}
```

**Impact:** Can reduce response payload size by 70-90% for JSON responses, significantly improving client-perceived latency.

---

## 3. Connection Pool Tuning

### 3.1 Default Pool Size Too Small for Concurrent Sync (Medium Impact)

**Problem:** The default connection pool has `MaxOpenConns=10` and `MaxIdleConns=5` (`libs/db.go`). During ERP sync, the worker pool uses 10 goroutines, each starting its own transaction. This means the sync alone can consume **all** available connections, starving API requests.

**Recommendation:**
- Increase pool size: `DB_MAX_OPEN_CONNECTIONS=25` and `DB_MAX_IDLE_CONNECTIONS=15`
- Or better: use a separate `sql.DB` instance for sync operations with its own pool.
- Add `ConnMaxIdleTime` to clean up stale connections:

```go
DB_CONN_MAX_IDLE_TIME, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_IDLE_TIME_IN_SEC"))
if err != nil {
    DB_CONN_MAX_IDLE_TIME = 60
}
db.SetConnMaxIdleTime(time.Second * time.Duration(DB_CONN_MAX_IDLE_TIME))
```

**Impact:** Prevents connection starvation during sync operations.

---

### 3.2 No `ConnMaxIdleTime` Set (Low Impact)

**Problem:** The pool sets `ConnMaxLifetime` but not `ConnMaxIdleTime`. Idle connections stay alive until their max lifetime, wasting resources when the pool is oversized during low-traffic periods.

**Recommendation:** Add `db.SetConnMaxIdleTime(60 * time.Second)`.

---

## 4. Caching Strategy

### 4.1 No Caching Layer Exists (High Impact)

**Problem:** Every request hits the database directly. Endpoints like `GetFilterOptions` (10 `SELECT DISTINCT` queries) and `FindAllDropdown` (returns all buildings) return data that rarely changes but is queried frequently.

**Recommendation:** Implement an in-memory cache with TTL for frequently accessed, rarely changing data:

```go
import "sync"

type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

type InMemoryCache struct {
    mu      sync.RWMutex
    entries map[string]CacheEntry
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    entry, ok := c.entries[key]
    if !ok || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    return entry.Data, true
}

func (c *InMemoryCache) Set(key string, data interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.entries[key] = CacheEntry{Data: data, ExpiresAt: time.Now().Add(ttl)}
}
```

**Cache candidates:**

| Endpoint | TTL | Reason |
|---|---|---|
| `GET /building-filter-options` | 5 min | Distinct values rarely change |
| `GET /building-dropdown` | 5 min | All building names rarely change |
| `GET /dashboard/*` | 2 min | Dashboard aggregates are expensive |
| `GET /pois` (flat) | 2 min | POI data changes infrequently |

**Invalidation:** Invalidate relevant caches after ERP sync completes and after write operations (POST/PUT/DELETE).

**Impact:** Can reduce database load by 60-80% for read-heavy workloads. The filter options endpoint alone saves 10 queries per call.

---

### 4.2 Consider `ETag`/`Last-Modified` Headers (Low Impact)

**Recommendation:** For endpoints with cached data, return `ETag` or `Last-Modified` headers. When clients send `If-None-Match` or `If-Modified-Since`, return `304 Not Modified` to save bandwidth.

---

## 5. ERP Sync Optimization

### 5.1 One Transaction Per Building (High Impact)

**Problem:** During sync, each building creates its own transaction (`processBuilding` in `services/building/service_building_impl.go:200-407`). For 1000 buildings, this means 1000 separate transactions with 1000 commit round-trips to PostgreSQL.

**Recommendation:** Batch buildings into groups (e.g., 50 per transaction) or use a single transaction for the entire sync:

```go
// Batch approach
const batchSize = 50
for i := 0; i < len(erpBuildings); i += batchSize {
    end := min(i+batchSize, len(erpBuildings))
    batch := erpBuildings[i:end]

    tx, err := service.DB.Begin()
    if err != nil { continue }

    for _, building := range batch {
        service.processBuilding(ctx, tx, building, ...) // pass tx instead of creating one
    }

    tx.Commit()
}
```

**Impact:** Reduces transaction overhead by 20-50x. Fewer fsync operations, less WAL traffic.

---

### 5.2 Sequential ERP API Calls (Medium Impact)

**Problem:** `SyncFromERP` (`services/building/service_building_impl.go:434-588`) fetches buildings, acquisitions, and building proposals **sequentially**:
```go
erpBuildings, err := service.ERPClient.FetchBuildings()      // wait...
erpAcquisitions, err := service.ERPClient.FetchAcquisitions() // wait...
erpBuildingProposals, err := service.ERPClient.FetchBuildingProposals() // wait...
```

**Recommendation:** Fetch all three in parallel using goroutines:
```go
var (
    erpBuildings         []erp.ERPBuilding
    erpAcquisitions      []erp.ERPAcquisition
    erpBuildingProposals []erp.ERPBuildingProposal
    errBuildings, errAcq, errBP error
)

var wg sync.WaitGroup
wg.Add(3)

go func() { defer wg.Done(); erpBuildings, errBuildings = service.ERPClient.FetchBuildings() }()
go func() { defer wg.Done(); erpAcquisitions, errAcq = service.ERPClient.FetchAcquisitions() }()
go func() { defer wg.Done(); erpBuildingProposals, errBP = service.ERPClient.FetchBuildingProposals() }()

wg.Wait()
```

**Impact:** Sync startup time reduced by ~2/3 (3 sequential HTTP calls become 1 parallel batch).

---

### 5.3 No ERP API Retry/Backoff (Medium Impact)

**Problem:** ERP API calls (`services/erp/erp_client.go`) have a 30-second timeout but no retry logic. If the ERP is temporarily unavailable, the entire sync fails immediately.

**Recommendation:** Add exponential backoff retry:
```go
func (c *ERPClient) fetchWithRetry(url string, maxRetries int) (*http.Response, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        resp, err := c.HTTPClient.Do(req)
        if err == nil && resp.StatusCode == http.StatusOK {
            return resp, nil
        }
        lastErr = err
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }
    return nil, lastErr
}
```

---

### 5.4 Sync Uses `FindByExternalId` Per Building (N+1 Problem) (High Impact)

**Problem:** For each building being synced, a separate `FindByExternalId` query checks if the building already exists. With 1000 buildings, that's 1000 lookup queries.

**Recommendation:** Pre-fetch all existing external IDs into a map:
```go
// Before processing buildings:
existingMap := make(map[string]models.Building)
existingBuildings, _ := service.RepositoryBuildingInterface.FindAllByExternalIds(ctx, tx, externalIds)
for _, b := range existingBuildings {
    existingMap[b.ExternalBuildingId] = b
}

// During processing:
if existing, ok := existingMap[erpBuilding.BuildingId]; ok {
    // update
} else {
    // create
}
```

Or use PostgreSQL `UPSERT`:
```sql
INSERT INTO buildings (...) VALUES (...)
ON CONFLICT (external_building_id)
DO UPDATE SET name = EXCLUDED.name, ...
```

**Impact:** Eliminates N lookup queries. For 1000 buildings, saves ~1000 queries.

---

### 5.5 Four Separate Sync Schedulers Run Independently (Low Impact)

**Problem:** In `main.go:49-60`, four separate sync schedulers start independently:
```go
servicesBuilding.StartBuildingSyncScheduler(buildingService, helpers.Logger, syncInterval)
servicesAcquisition.StartAcquisitionSyncScheduler(acquisitionService, helpers.Logger, syncInterval)
servicesBuildingProposal.StartBuildingProposalSyncScheduler(buildingProposalService, helpers.Logger, syncInterval)
servicesLOI.StartLOISyncScheduler(loiService, helpers.Logger, syncInterval)
```

Each scheduler creates its own `sql.DB` instance (via `InitializeBuildingService()`, `InitializeAcquisitionService()`, etc. in `injector/wire.go`), so there are **5 separate connection pools** (1 for the router + 4 for schedulers).

**Recommendation:**
- Share a single `sql.DB` instance across all initializers.
- Or coordinate sync schedulers to run sequentially to reduce peak resource usage.
- Add staggered startup delays to avoid all syncs hitting the ERP API simultaneously.

---

## 6. API Design & Response Optimization

### 6.1 `FindAllForMapping` Returns All Buildings (High Impact)

**Problem:** The mapping endpoint (`GET /mapping-buildings`) fetches **all matching buildings** without pagination. For large datasets, this can return thousands of buildings with full details (images, all fields).

**Recommendation:**
- Add server-side clustering: return cluster markers when zoomed out, individual buildings when zoomed in.
- Return a lightweight response (only `id`, `lat`, `lng`, `type`) for the map view, and load details on-demand when a marker is clicked.
- Implement viewport-only loading with aggressive bounds filtering (already partially done, but the totals query defeats the purpose).

---

### 6.2 `FindAll` and `CountAll` Run Two Separate Queries (Medium Impact)

**Problem:** Every paginated listing runs two queries: one for data and one for count. These share identical WHERE clauses but are built separately with duplicated code (`repositories/building/repository_building_impl.go:244-524`).

**Recommendation:** Use a window function to get the count in the same query:
```sql
SELECT *, COUNT(*) OVER() AS total_count
FROM buildings
WHERE ...
ORDER BY ... LIMIT ... OFFSET ...
```

This eliminates the second round-trip and the duplicated filter logic.

---

### 6.3 No Pagination on Several List Endpoints (Low Impact)

**Problem:** Some endpoints return all records without pagination:
- `FindAllDropdown` — returns all buildings
- `FindAllFlat` (POIs) — returns all POIs with all points
- `GetDistinctPICs` — returns all distinct values

**Recommendation:** These are acceptable if datasets remain small (<1000 records). Monitor growth and add pagination or caching when needed.

---

## 7. Concurrency & Resource Management

### 7.1 Context Not Passed to Database Operations (High Impact)

**Problem:** While `context.Context` is passed through function parameters, `service.DB.Begin()` does not use context-aware `BeginTx`. If a client disconnects mid-request, the database work continues until completion, wasting resources.

**Recommendation:** Use `BeginTx` everywhere:
```go
tx, err := service.DB.BeginTx(ctx, nil) // or &sql.TxOptions{ReadOnly: true}
```

**Impact:** Cancels in-progress queries when clients disconnect, freeing connections.

---

### 7.2 Sync Workers Share Same Context (Low Impact)

**Problem:** All sync workers share a `context.Background()` that is never cancelled. If the application is shutting down, sync operations continue running.

**Recommendation:** Use a cancellable context and cancel it on shutdown:
```go
ctx, cancel := context.WithCancel(context.Background())
// Store cancel function, call on SIGTERM
```

---

## 8. Error Handling & Panic Recovery

### 8.1 `panic`/`recover` Used for Control Flow (Medium Impact)

**Problem:** The application uses `panic` extensively for error handling throughout the codebase via `helpers.PanicIfError()`. Every error triggers a panic, caught by `RouterPanicHandler`. This is idiomatic in some Go web frameworks but has performance implications:
- `panic`/`recover` is significantly slower than returning errors (involves stack unwinding).
- `debug.Stack()` is called on every error, which is expensive.
- Stack traces are captured even for expected errors (404, 400).

**Current code** (`helpers/panic.go`):
```go
func PanicIfError(err interface{}) {
    if err != nil {
        panic(err)
    }
}
```

**Recommendation:**
- Return errors through the normal Go error path for expected errors.
- Only use panics for truly unexpected conditions.
- In the panic handler, avoid `debug.Stack()` for client errors (4xx):

```go
// Only capture stack trace for 5xx errors
if _, ok := i.(error); ok {
    stackTrace = string(debug.Stack())
}
```

**Impact:** Reduces latency for error responses and reduces garbage collection pressure.

---

### 8.2 Stack Trace Allocation in Logger (Low Impact)

**Problem:** `helpers/logger.go:93-94` allocates a 64KB buffer for every stack trace:
```go
buf := make([]byte, 1024*64)
```

**Recommendation:** Use `debug.Stack()` instead, which right-sizes the buffer, or use a `sync.Pool` to reuse buffers.

---

## 9. Observability & Monitoring

### 9.1 No Request Metrics (Medium Impact)

**Problem:** There are no metrics for request latency, throughput, error rates, or database query performance. Without metrics, performance issues are invisible until users report them.

**Recommendation:** Add Prometheus metrics:
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"query_name"},
    )
)
```

**Impact:** Enables data-driven performance tuning and alerting.

---

### 9.2 Logging Creates New Logger Entry Per Call (Low Impact)

**Problem:** `helpers.GetLogger()` creates a new `logrus.Entry` with fields on every call. In the logging middleware (`middlewares/logging.go:60`), `helpers.GetLogger()` is called for every request.

**Recommendation:** Cache the base logger entry at startup:
```go
var baseEntry *logrus.Entry

func InitLogger() {
    // ...
    baseEntry = Logger.WithFields(logrus.Fields{
        "service":     serviceName,
        "environment": environment,
    })
}

func GetLogger() *logrus.Entry {
    return baseEntry
}
```

---

## 10. Deployment & Infrastructure

### 10.1 Dockerfile Missing Build Optimizations (Low Impact)

**Problem:** The build stage in `Dockerfile` doesn't use build flags for optimization:
```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tmn-backend .
```

**Recommendation:** Add linker flags to strip debug info and reduce binary size:
```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tmn-backend .
```

The `-a` flag forces rebuilding all packages; it's unnecessary with `CGO_ENABLED=0`.

---

### 10.2 No Health Check for Database Connection (Medium Impact)

**Problem:** The `/health` endpoint returns `200 OK` unconditionally without checking database connectivity:
```go
router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

**Recommendation:** Add a database ping to the health check:
```go
router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("DB unhealthy"))
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

---

### 10.3 No CORS Configuration (Low Impact)

**Problem:** No CORS headers are set. If the frontend is on a different origin, browsers will block requests or require preflight OPTIONS requests.

**Recommendation:** Add CORS middleware for the appropriate origins to avoid unnecessary preflight round-trips.

---

## 11. Security Improvements with Performance Impact

### 11.1 No Rate Limiting (Medium Impact)

**Problem:** There is no rate limiting on any endpoint. A single client can overwhelm the server with requests, especially on expensive endpoints like `/mapping-buildings` or `/buildings/sync`.

**Recommendation:** Add rate limiting middleware (per-IP or per-user) using a token bucket algorithm. This protects server resources and improves overall throughput under load.

---

### 11.2 SQL String Concatenation for Table Names (Low Impact)

**Problem:** Table names are concatenated into SQL strings using `models.BuildingTable` etc. While not directly a SQL injection risk (values are compile-time constants), this prevents PostgreSQL from caching prepared statement plans.

**Recommendation:** Consider using prepared statements for the most frequent queries, which allows PostgreSQL's query planner to cache execution plans.

---

## 12. Summary & Priority Matrix

| # | Recommendation | Impact | Effort | Priority |
|---|---|---|---|---|
| 2.1 | Add HTTP server timeouts | Critical | Low | **P0** |
| 2.2 | Implement graceful shutdown | High | Low | **P0** |
| 1.1 | Use read-only transactions / direct queries | High | Medium | **P1** |
| 1.2 | Fix duplicate query in `FindAllForMapping` | High | Medium | **P1** |
| 5.4 | Eliminate N+1 in sync (batch lookup or UPSERT) | High | Medium | **P1** |
| 4.1 | Add in-memory cache for filter options & dropdown | High | Medium | **P1** |
| 5.1 | Batch transactions in sync | High | Medium | **P1** |
| 7.1 | Use context-aware `BeginTx` everywhere | High | Low | **P1** |
| 3.1 | Increase connection pool / separate sync pool | Medium | Low | **P2** |
| 5.2 | Parallelize ERP API fetches | Medium | Low | **P2** |
| 1.3 | Parallelize or combine `GetFilterOptions` queries | Medium | Low | **P2** |
| 1.4 | Add composite indexes and trigram indexes | Medium | Low | **P2** |
| 1.5 | Fix `LOWER()` preventing index usage | Medium | Low | **P2** |
| 2.3 | Add response compression (gzip) | Medium | Low | **P2** |
| 6.2 | Use window function for count+data | Medium | Medium | **P2** |
| 8.1 | Reduce panic/recover overhead | Medium | High | **P3** |
| 9.1 | Add Prometheus metrics | Medium | Medium | **P3** |
| 11.1 | Add rate limiting | Medium | Medium | **P3** |
| 10.2 | Add DB health check | Medium | Low | **P3** |
| 5.3 | Add ERP API retry with backoff | Medium | Low | **P3** |
| 6.1 | Add map clustering / lightweight map response | High | High | **P3** |
| 1.6 | Add index on `buildings.project_name` | Low | Low | **P3** |
| 5.5 | Share DB pool across schedulers | Low | Medium | **P3** |
| 9.2 | Cache base logger entry | Low | Low | **P4** |
| 10.1 | Optimize Dockerfile build flags | Low | Low | **P4** |
| 3.2 | Set `ConnMaxIdleTime` | Low | Low | **P4** |
| 4.2 | Add `ETag`/`Last-Modified` headers | Low | Medium | **P4** |
| 10.3 | Add CORS middleware | Low | Low | **P4** |

---

### Quick Wins (< 1 hour each, high value)

1. Add HTTP server timeouts (`main.go` — 5 lines)
2. Use `BeginTx` with `ReadOnly: true` for all read endpoints
3. Parallelize ERP API fetches in `SyncFromERP` (~10 lines)
4. Add `idx_buildings_project_name` index (1 SQL statement)
5. Set `ConnMaxIdleTime` on the connection pool (1 line)
6. Optimize Dockerfile build flags (1 line change)

### Medium Effort (1-3 days, high value)

1. Implement in-memory cache for filter options and dropdown
2. Replace per-building transactions with batch transactions in sync
3. Add `CountByBuildingTypeForMapping` to eliminate duplicate mapping query
4. Add composite and trigram database indexes
5. Implement graceful shutdown

### Larger Efforts (1+ week)

1. Migrate from panic-based to error-return control flow
2. Implement map clustering for the mapping endpoint
3. Add Prometheus metrics and monitoring dashboards
4. Add rate limiting middleware
