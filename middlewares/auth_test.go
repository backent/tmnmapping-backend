package middlewares_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const validToken = "valid-jwt-token"

// authFixture wires a real httprouter with the real panic handler, so the tests
// assert on the HTTP status the client actually receives rather than on a panic value.
type authFixture struct {
	middleware *middlewares.AuthMiddleware
	repoAuth   *mocks.MockRepositoryAuth
	repoUser   *mocks.MockRepositoryUser
	sqlMock    sqlmock.Sqlmock
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	repoAuth := &mocks.MockRepositoryAuth{}
	repoUser := &mocks.MockRepositoryUser{}

	return &authFixture{
		middleware: middlewares.NewAuthMiddleware(libs.NewValidator(), db, repoAuth, repoUser),
		repoAuth:   repoAuth,
		repoUser:   repoUser,
		sqlMock:    sqlMock,
	}
}

// expectUser makes token resolution succeed and return a user with the given role.
func (f *authFixture) expectUser(userId int, role string) {
	f.repoAuth.On("Validate", validToken).Return(userId, true)
	f.sqlMock.ExpectBegin()
	f.sqlMock.ExpectCommit()
	f.repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), userId).
		Return(testutil.NewUser(userId, "tester", "secret123", role), nil)
}

// serve runs handle behind the middleware chain and returns the recorded response.
func (f *authFixture) serve(t *testing.T, handle httprouter.Handle, withToken bool) *httptest.ResponseRecorder {
	t.Helper()

	router := httprouter.New()
	router.PanicHandler = exceptions.RouterPanicHandler
	router.GET("/protected", handle)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if withToken {
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: validToken})
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

// okHandler records what the middleware put on the request context.
func okHandler(seenUserId, seenRole *string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if v, ok := r.Context().Value(middlewares.ContextKeyUserId).(string); ok {
			*seenUserId = v
		}
		if v, ok := r.Context().Value(middlewares.ContextKeyUserRole).(string); ok {
			*seenRole = v
		}
		w.WriteHeader(http.StatusOK)
	}
}

// ---------------------------------------------------------------------------
// RequireAuth
// ---------------------------------------------------------------------------

func TestRequireAuth_NoToken(t *testing.T) {
	f := newAuthFixture(t)

	var userId, role string
	res := f.serve(t, f.middleware.RequireAuth(okHandler(&userId, &role)), false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Empty(t, userId)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	f := newAuthFixture(t)
	f.repoAuth.On("Validate", validToken).Return(0, false)

	var userId, role string
	res := f.serve(t, f.middleware.RequireAuth(okHandler(&userId, &role)), true)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	f.repoAuth.AssertExpectations(t)
}

// A token can outlive the account it points at. The lookup failure must read as
// unauthenticated, not as a 500.
func TestRequireAuth_UserNoLongerExists(t *testing.T) {
	f := newAuthFixture(t)
	f.repoAuth.On("Validate", validToken).Return(7, true)
	f.sqlMock.ExpectBegin()
	f.sqlMock.ExpectRollback()
	f.repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).
		Return(models.User{}, sql.ErrNoRows)

	var userId, role string
	res := f.serve(t, f.middleware.RequireAuth(okHandler(&userId, &role)), true)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.NoError(t, f.sqlMock.ExpectationsWereMet())
}

func TestRequireAuth_PutsUserIdAndRoleOnContext(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(42, models.RoleSales)

	var userId, role string
	res := f.serve(t, f.middleware.RequireAuth(okHandler(&userId, &role)), true)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "42", userId)
	assert.Equal(t, models.RoleSales, role)
	assert.NoError(t, f.sqlMock.ExpectationsWereMet())
}

// The role is read from the database on every request, so a legacy value stored
// before migration 015 still resolves to the current vocabulary.
func TestRequireAuth_NormalizesLegacyRole(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(9, "approver")

	var userId, role string
	res := f.serve(t, f.middleware.RequireAuth(okHandler(&userId, &role)), true)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, models.RoleHeadOfSales, role)
}

func TestRequireAuth_AcceptsAuthorizationHeader(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(3, models.RoleAdmin)

	var userId, role string

	router := httprouter.New()
	router.PanicHandler = exceptions.RouterPanicHandler
	router.GET("/protected", f.middleware.RequireAuth(okHandler(&userId, &role)))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", validToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "3", userId)
}

// ---------------------------------------------------------------------------
// RequirePermission
// ---------------------------------------------------------------------------

func TestRequirePermission_AllowsRoleThatHoldsIt(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(1, models.RoleAdmin)

	var userId, role string
	handle := f.middleware.RequireAuth(
		f.middleware.RequirePermission(models.PermissionPOIsManage)(okHandler(&userId, &role)))

	res := f.serve(t, handle, true)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, models.RoleAdmin, role)
}

func TestRequirePermission_DeniesRoleThatDoesNotWith403(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(1, models.RoleSales)

	var userId, role string
	handle := f.middleware.RequireAuth(
		f.middleware.RequirePermission(models.PermissionPOIsManage)(okHandler(&userId, &role)))

	res := f.serve(t, handle, true)

	// 403, not 401: the caller is authenticated, they just may not do this.
	assert.Equal(t, http.StatusForbidden, res.Code)
	assert.Contains(t, res.Body.String(), "insufficient permission")
	assert.Empty(t, userId, "handler must not run")
}

// A read permission is held by every role, so the same route works for all of them.
func TestRequirePermission_SharedReadAllowsEveryRole(t *testing.T) {
	for _, role := range models.Roles {
		t.Run(role, func(t *testing.T) {
			f := newAuthFixture(t)
			f.expectUser(1, role)

			var seenUserId, seenRole string
			handle := f.middleware.RequireAuth(
				f.middleware.RequirePermission(models.PermissionBuildingsView)(okHandler(&seenUserId, &seenRole)))

			assert.Equal(t, http.StatusOK, f.serve(t, handle, true).Code)
		})
	}
}

// A role the app does not recognise must deny, not fall back to a default.
func TestRequirePermission_DeniesUnknownStoredRole(t *testing.T) {
	f := newAuthFixture(t)
	f.expectUser(1, "wizard")

	var userId, role string
	handle := f.middleware.RequireAuth(
		f.middleware.RequirePermission(models.PermissionBuildingsView)(okHandler(&userId, &role)))

	assert.Equal(t, http.StatusForbidden, f.serve(t, handle, true).Code)
}

// Wiring RequirePermission without RequireAuth in front of it is a programming
// error. It must fail closed rather than let the request through.
func TestRequirePermission_WithoutRequireAuthDenies(t *testing.T) {
	f := newAuthFixture(t)

	var userId, role string
	handle := f.middleware.RequirePermission(models.PermissionBuildingsView)(okHandler(&userId, &role))

	assert.Equal(t, http.StatusUnauthorized, f.serve(t, handle, true).Code)
	assert.Empty(t, userId)
}

// Routes are wired during startup, so a typo'd permission key must crash the
// process immediately rather than quietly 403 every request to that endpoint.
func TestRequirePermission_PanicsAtWiringTimeOnUnknownKey(t *testing.T) {
	f := newAuthFixture(t)

	assert.PanicsWithValue(t, "middlewares: unknown permission buildings.destroy", func() {
		f.middleware.RequirePermission("buildings.destroy")
	})
}
