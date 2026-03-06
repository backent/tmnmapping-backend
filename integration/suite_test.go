//go:build integration

package integration_test

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/malikabdulaziz/tmn-backend/injector"
)

// testSuite holds shared state for all integration tests in this package.
var testSuite struct {
	server *httptest.Server
	db     *sql.DB
}

func TestMain(m *testing.M) {
	// Set env vars required by the application before calling InitializeRouter.
	// DB vars default to the docker-compose.postgres.yml values so devs can run
	// `docker compose -f docker-compose.postgres.yml up -d` and immediately run tests.
	setEnvDefault("POSTGRES_HOST", "localhost")
	setEnvDefault("POSTGRES_PORT", "5432")
	setEnvDefault("POSTGRES_USER", "postgres")
	setEnvDefault("POSTGRES_PASSWORD", "adminlocal")
	setEnvDefault("POSTGRES_DATABASE", "tmn_test")
	setEnvDefault("POSTGRES_SSLMODE", "disable")
	setEnvDefault("APP_SECRET_KEY", "integration-test-secret-key-32chars!!")
	setEnvDefault("APP_TOKEN_EXPIRE_IN_SEC", "3600")
	setEnvDefault("APP_PORT", "0")
	setEnvDefault("ERP_API_BASE_URL", "http://localhost:19999")
	setEnvDefault("ERP_API_KEY", "testkey")
	setEnvDefault("ERP_API_SECRET", "testsecret")
	setEnvDefault("SERVICE_NAME", "tmn-backend-test")
	setEnvDefault("ENVIRONMENT", "test")
	setEnvDefault("LOG_LEVEL", "error")

	// Open a direct DB connection for migrations and test-data management.
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_SSLMODE"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: failed to open test db: %v\n", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr,
			"integration: cannot reach test database at %s:%s/%s — is the PostgreSQL container running?\n"+
				"  Run: docker compose -f docker-compose.postgres.yml up -d\n"+
				"  Error: %v\n",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_DATABASE"),
			err,
		)
		os.Exit(1)
	}
	defer db.Close()
	testSuite.db = db

	// Run all *.up.sql migrations in lexicographic order.
	if err := runMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "integration: migrations failed: %v\n", err)
		os.Exit(1)
	}

	// Seed a test admin user that all tests can use for authenticated requests.
	seedAdminUser(db)

	// Build the full HTTP router via Wire and wrap it in an httptest.Server.
	router := injector.InitializeRouter()
	testSuite.server = httptest.NewServer(router)
	defer testSuite.server.Close()

	os.Exit(m.Run())
}

// setEnvDefault sets the env var only if it is not already set.
func setEnvDefault(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

// runMigrations reads all *.up.sql files under database/migrations/ (relative to
// the repo root, one level up from the integration/ package) and executes them in
// lexicographic order. Idempotent: uses IF NOT EXISTS / IF EXISTS guards in the SQL.
func runMigrations(db *sql.DB) error {
	dir := filepath.Join("..", "database", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // "001_" < "002_" < … guarantees dependency order

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return nil
}

// seedAdminUser inserts a well-known admin account into the users table.
// This user is reused across all tests for authenticated requests.
// ON CONFLICT DO NOTHING makes it safe to run multiple times.
func seedAdminUser(db *sql.DB) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.MinCost)
	if err != nil {
		panic(fmt.Sprintf("integration: bcrypt error: %v", err))
	}
	_, err = db.Exec(`
		INSERT INTO users (username, name, email, password, role)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (username) DO NOTHING`,
		"testadmin", "Test Admin", "admin@test.com", string(hashed), "admin",
	)
	if err != nil {
		panic(fmt.Sprintf("integration: seed admin user: %v", err))
	}
}
