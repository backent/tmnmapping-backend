//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
)

// ── Auth helpers ──────────────────────────────────────────────────────────────

// MustLogin POSTs to /login and returns the auth_token cookie value.
// Use in auth_test.go where the login flow itself is under test.
func MustLogin(t *testing.T, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"password": password,
	})
	resp, err := http.Post(testSuite.server.URL+"/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "MustLogin: expected 200 from POST /login")
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			return c.Value
		}
	}
	t.Fatal("MustLogin: auth_token cookie not set after login")
	return ""
}

// MintToken mints a JWT directly without an HTTP round-trip.
// Use in non-auth tests where a valid token is just a precondition.
func MintToken(t *testing.T, userID int) string {
	t.Helper()
	repo := repositoriesAuth.NewRepositoryAuthJWTImpl()
	token, err := repo.Issue(fmt.Sprintf("%d", userID), time.Hour)
	require.NoError(t, err)
	return token
}

// ── Request helpers ───────────────────────────────────────────────────────────

// NewRequest builds an unauthenticated *http.Request against the test server.
func NewRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	return buildRequest(t, method, path, body, "")
}

// NewAuthRequest builds an *http.Request with a JWT in the Authorization header.
func NewAuthRequest(t *testing.T, method, path string, body interface{}, token string) *http.Request {
	t.Helper()
	return buildRequest(t, method, path, body, token)
}

func buildRequest(t *testing.T, method, path string, body interface{}, token string) *http.Request {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, err := http.NewRequest(method, testSuite.server.URL+path, buf)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	return req
}

// Do executes the request and returns the response.
func Do(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// ── Response helpers ──────────────────────────────────────────────────────────

// WebResponse mirrors the top-level JSON envelope returned by every endpoint.
type WebResponse struct {
	Status string          `json:"status"`
	Code   int             `json:"code"`
	Data   json.RawMessage `json:"data"`
	Extras json.RawMessage `json:"extras"`
}

// DecodeWebResponse decodes the outer WebResponse envelope and closes the body.
func DecodeWebResponse(t *testing.T, resp *http.Response) WebResponse {
	t.Helper()
	defer resp.Body.Close()
	var wr WebResponse
	err := json.NewDecoder(resp.Body).Decode(&wr)
	require.NoError(t, err, "failed to decode WebResponse JSON")
	return wr
}

// DecodeData unmarshals the .data field of a WebResponse into target.
func DecodeData(t *testing.T, wr WebResponse, target interface{}) {
	t.Helper()
	err := json.Unmarshal(wr.Data, target)
	require.NoError(t, err, "failed to decode WebResponse.Data")
}

// ── Database helpers ──────────────────────────────────────────────────────────

// truncateTables truncates the given tables in order, resetting sequences.
// Call at the top of any test that writes rows so tests remain independent.
// The users table is excluded — the seeded admin user must persist across tests.
func truncateTables(t *testing.T, tables ...string) {
	t.Helper()
	for _, tbl := range tables {
		_, err := testSuite.db.Exec(
			fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tbl),
		)
		require.NoError(t, err, "truncate %s", tbl)
	}
}

// insertBuilding inserts a minimal building row via SQL and returns its id.
// Buildings come from ERP sync, so there is no POST /buildings API endpoint.
func insertBuilding(t *testing.T, name string) int {
	t.Helper()
	var id int
	err := testSuite.db.QueryRow(`
		INSERT INTO buildings (name, building_type, grade_resource, latitude, longitude, sellable, connectivity)
		VALUES ($1, 'Office', 'A', -6.2, 106.8, 'sell', 'online')
		RETURNING id`,
		name,
	).Scan(&id)
	require.NoError(t, err, "insertBuilding %q", name)
	return id
}

// seedAdminID returns the id of the admin user seeded by TestMain.
func seedAdminID(t *testing.T) int {
	t.Helper()
	var id int
	err := testSuite.db.QueryRow(
		"SELECT id FROM users WHERE username = $1", "testadmin",
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// ── Convenience shortcut ──────────────────────────────────────────────────────

// doWithAuth mints a token for the seeded admin, builds a request, executes it,
// and returns the response. This is the most common pattern in non-auth tests.
func doWithAuth(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()
	adminID := seedAdminID(t)
	token := MintToken(t, adminID)
	req := NewAuthRequest(t, method, path, body, token)
	return Do(t, req)
}
