//go:build e2e

// Package e2e exercises the real router, middleware and database.
//
// Excluded from `go test ./...` by the build tag, because it needs a live
// database. Run it with:
//
//	go test -tags e2e ./e2e/... -v
//
// It drives injector.InitializeRouter() rather than main(), so the ERP sync
// schedulers — which sync on startup — never run.
//
// It seeds two users of its own, prefixed e2e_, and removes them afterwards.
// No existing row is read or written.
package e2e

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/malikabdulaziz/tmn-backend/injector"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	testPassword    = "e2e-test-password"
	createdPassword = "e2e-created-password"
	adminUsername   = "e2e_admin"
	salesUsername   = "e2e_sales"
)

type webResponse struct {
	Code   int             `json:"code"`
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type userPayload struct {
	Id                  int      `json:"id"`
	Username            string   `json:"username"`
	Role                string   `json:"role"`
	Permissions         []string `json:"permissions"`
	CanCreateQuotations bool     `json:"can_create_quotations"`
}

func setup(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()

	require.NoError(t, godotenv.Load("../.env"), "e2e needs ../.env")

	db := libs.NewDatabase()
	seedUser(t, db, adminUsername, models.RoleAdmin)
	seedUser(t, db, salesUsername, models.RoleSales)

	server := httptest.NewServer(injector.InitializeRouter())
	t.Cleanup(server.Close)

	return server, db
}

func seedUser(t *testing.T, db *sql.DB, username, role string) {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	require.NoError(t, err)

	_, err = db.Exec(`DELETE FROM users WHERE username = $1`, username)
	require.NoError(t, err)

	_, err = db.Exec(
		`INSERT INTO users (username, name, email, password, role) VALUES ($1, $2, $3, $4, $5)`,
		username, "E2E "+role, username+"@e2e.local", string(hashed), role)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE username = $1`, username)
	})
}

// login returns the auth cookie and the user payload from the login response.
func login(t *testing.T, server *httptest.Server, username, password string) (*http.Cookie, userPayload) {
	t.Helper()

	res := do(t, server, nil, http.MethodPost, "/login",
		`{"username":"`+username+`","password":"`+password+`"}`)
	require.Equal(t, http.StatusOK, res.StatusCode, "login as %s", username)

	var body struct {
		Data struct {
			User userPayload `json:"user"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	res.Body.Close()

	for _, cookie := range res.Cookies() {
		if cookie.Name == "auth_token" {
			return cookie, body.Data.User
		}
	}

	t.Fatalf("no auth_token cookie returned for %s", username)

	return nil, userPayload{}
}

func do(t *testing.T, server *httptest.Server, cookie *http.Cookie, method, path, body string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, server.URL+path, reader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	res, err := server.Client().Do(req)
	require.NoError(t, err)

	return res
}

func status(t *testing.T, server *httptest.Server, cookie *http.Cookie, method, path string) int {
	t.Helper()

	res := do(t, server, cookie, method, path, "")
	defer res.Body.Close()

	return res.StatusCode
}

// ---------------------------------------------------------------------------
// The permission list the API issues
// ---------------------------------------------------------------------------

func TestLoginReturnsPermissionsForRole(t *testing.T) {
	server, _ := setup(t)

	_, admin := login(t, server, adminUsername, testPassword)
	assert.Equal(t, models.RoleAdmin, admin.Role)
	assert.Subset(t, admin.Permissions, []string{
		models.PermissionUsersManage,
		models.PermissionMasterDataManage,
		models.PermissionMasterDataScreen,
		models.PermissionPOIsManage,
	})

	_, sales := login(t, server, salesUsername, testPassword)
	assert.Equal(t, models.RoleSales, sales.Role)

	// Reads yes, writes and admin screens no.
	assert.Subset(t, sales.Permissions, []string{
		models.PermissionBuildingsView,
		models.PermissionMappingView,
		models.PermissionSavedPolygonsManage,
	})
	assert.NotContains(t, sales.Permissions, models.PermissionUsersView)
	assert.NotContains(t, sales.Permissions, models.PermissionPOIsManage)
	assert.NotContains(t, sales.Permissions, models.PermissionMasterDataScreen)
}

func TestCurrentUserReturnsPermissions(t *testing.T) {
	server, _ := setup(t)
	cookie, _ := login(t, server, salesUsername, testPassword)

	res := do(t, server, cookie, http.MethodGet, "/current-user", "")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var body struct {
		Data userPayload `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, models.RoleSales, body.Data.Role)
	assert.NotEmpty(t, body.Data.Permissions)
	assert.NotContains(t, body.Data.Permissions, models.PermissionUsersView)
}

// ---------------------------------------------------------------------------
// Enforcement
// ---------------------------------------------------------------------------

func TestUnauthenticatedIsRejected(t *testing.T) {
	server, _ := setup(t)

	assert.Equal(t, http.StatusUnauthorized, status(t, server, nil, http.MethodGet, "/buildings"))
	assert.Equal(t, http.StatusUnauthorized, status(t, server, nil, http.MethodGet, "/users"))
}

// The hole Phase 0 closed: before it, any authenticated user could call these.
func TestSalesIsDeniedWrites(t *testing.T) {
	server, _ := setup(t)
	cookie, _ := login(t, server, salesUsername, testPassword)

	denied := []struct{ method, path string }{
		{http.MethodDelete, "/pois/999999"},
		{http.MethodDelete, "/categories/999999"},
		{http.MethodDelete, "/branches/999999"},
		{http.MethodGet, "/pois-export"},
		{http.MethodGet, "/categories-export"},
		{http.MethodGet, "/users"},
		{http.MethodGet, "/users/1"},
	}

	for _, tc := range denied {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			// 403 not 404: authorization is checked before the handler runs, so a
			// non-existent id never reaches the service.
			assert.Equal(t, http.StatusForbidden, status(t, server, cookie, tc.method, tc.path))
		})
	}
}

func TestSalesIsAllowedReads(t *testing.T) {
	server, _ := setup(t)
	cookie, _ := login(t, server, salesUsername, testPassword)

	// Master data reads stay open: the mapping page loads them for every role.
	allowed := []string{
		"/buildings?take=1&skip=0",
		"/pois?take=1&skip=0",
		"/sales-packages?take=1&skip=0",
		"/categories?take=1&skip=0",
		"/mother-brands-dropdown",
		"/branches-dropdown",
		"/building-filter-options",
		"/saved-polygons?take=1&skip=0",
		"/dashboard/loi",
	}

	for _, path := range allowed {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, http.StatusOK, status(t, server, cookie, http.MethodGet, path))
		})
	}
}

func TestAdminIsAllowedUserManagement(t *testing.T) {
	server, _ := setup(t)
	cookie, _ := login(t, server, adminUsername, testPassword)

	assert.Equal(t, http.StatusOK, status(t, server, cookie, http.MethodGet, "/users"))
	assert.Equal(t, http.StatusOK, status(t, server, cookie, http.MethodGet, "/categories?take=1&skip=0"))
}

// ---------------------------------------------------------------------------
// User management guards, end to end
// ---------------------------------------------------------------------------

func TestUserManagementLifecycleAndGuards(t *testing.T) {
	server, db := setup(t)
	cookie, admin := login(t, server, adminUsername, testPassword)

	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE username = $1`, "e2e_created")
	})

	// Create
	res := do(t, server, cookie, http.MethodPost, "/users",
		`{"username":"e2e_created","name":"E2E Created","email":"created@e2e.local",`+
			`"password":"`+createdPassword+`","role":"sales","can_create_quotations":true,`+
			`"sales_group":"sales_team"}`)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	var created struct {
		Data struct {
			Id       int    `json:"id"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&created))
	res.Body.Close()

	assert.NotZero(t, created.Data.Id)
	assert.Empty(t, created.Data.Password, "the password hash must never leave the API")

	// Duplicate username is rejected with a readable message, not a driver error.
	res = do(t, server, cookie, http.MethodPost, "/users",
		`{"username":"e2e_created","name":"Dup","password":"`+createdPassword+`","role":"sales"}`)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	res.Body.Close()

	// The created user can log in, which proves the password was hashed usably.
	newUserCookie, newUser := login(t, server, "e2e_created", createdPassword)
	assert.True(t, newUser.CanCreateQuotations)
	assert.Equal(t, http.StatusForbidden, status(t, server, newUserCookie, http.MethodGet, "/users"))

	// Editing without a password must not lock them out.
	res = do(t, server, cookie, http.MethodPut, "/users/"+itoa(created.Data.Id),
		`{"username":"e2e_created","name":"E2E Renamed","email":"created@e2e.local",`+
			`"role":"sales","can_create_quotations":false}`)
	require.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()

	// The original password must still work — the edit carried none, so it must
	// have been left alone rather than blanked.
	login(t, server, "e2e_created", createdPassword)

	// Self role change is refused.
	res = do(t, server, cookie, http.MethodPut, "/users/"+itoa(admin.Id),
		`{"username":"`+adminUsername+`","name":"E2E admin","role":"sales"}`)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assertBodyContains(t, res, "cannot change your own role")

	// Self deletion is refused.
	res = do(t, server, cookie, http.MethodDelete, "/users/"+itoa(admin.Id), "")
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assertBodyContains(t, res, "cannot delete your own account")

	// Deleting someone else works.
	assert.Equal(t, http.StatusOK,
		status(t, server, cookie, http.MethodDelete, "/users/"+itoa(created.Data.Id)))
}

func assertBodyContains(t *testing.T, res *http.Response, want string) {
	t.Helper()
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(raw), want)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
