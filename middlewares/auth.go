package middlewares

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

// ContextKeyUserId holds the authenticated user's id, as a string.
const ContextKeyUserId = helpers.ContextKey("userId")

// ContextKeyUserRole holds the authenticated user's normalized role.
// Always set by RequireAuth; RequirePermission reads it.
const ContextKeyUserRole = helpers.ContextKey("userRole")

type AuthMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesAuth.RepositoryAuthInterface
	repositoriesUser.RepositoryUserInterface
}

func NewAuthMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repositoriesAuth repositoriesAuth.RepositoryAuthInterface,
	repositoriesUser repositoriesUser.RepositoryUserInterface,
) *AuthMiddleware {
	return &AuthMiddleware{
		Validate:                validate,
		DB:                      db,
		RepositoryAuthInterface: repositoriesAuth,
		RepositoryUserInterface: repositoriesUser,
	}
}

// ValidateLogin validates login request before calling controller
func (m *AuthMiddleware) ValidateLogin(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webAuth.LoginRequest
		helpers.DecodeRequest(r, &req)

		// Validate request
		err := m.Validate.Struct(req)
		helpers.PanicIfError(err)

		// Store validated request in context
		ctx := context.WithValue(r.Context(), helpers.ContextKey("loginRequest"), req)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

// RequireAuth validates the JWT token from the cookie and puts the user's id and
// role on the request context. The role is read from the database on every request
// rather than carried in the token, so that a role change takes effect immediately
// and without forcing the user to log in again.
func (m *AuthMiddleware) RequireAuth(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Try to get token from cookie first
		cookie, err := r.Cookie("auth_token")
		var token string

		if err == nil {
			// Cookie found
			token = cookie.Value
		} else {
			// Fallback to Authorization header for backward compatibility
			token = r.Header.Get("Authorization")
		}

		if token == "" {
			panic(exceptions.NewUnAuthorized("authorization required"))
		}

		userId, valid := m.RepositoryAuthInterface.Validate(token)
		if !valid {
			panic(exceptions.NewUnAuthorized("authorization invalid"))
		}

		role := m.resolveRole(r.Context(), userId)

		// Store userId as string in context
		ctx := context.WithValue(r.Context(), ContextKeyUserId, strconv.Itoa(userId))
		ctx = context.WithValue(ctx, ContextKeyUserRole, role)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

// RequirePermission rejects a request whose caller does not hold the permission.
// It must be composed inside RequireAuth, which is what puts the role on the context:
//
//	authMiddleware.RequireAuth(
//	    authMiddleware.RequirePermission(models.PermissionPOIsManage)(controller.Delete))
//
// Routes name an action; models.Permissions decides which roles may perform it.
// Changing that policy is an edit to the map, not to any route.
//
// An unknown permission key panics here rather than at request time. Routes are
// wired in NewRouter during startup, so a typo crashes the process immediately
// instead of quietly returning 403 for every call to that endpoint.
func (m *AuthMiddleware) RequirePermission(permission string) func(httprouter.Handle) httprouter.Handle {
	if !models.IsValidPermission(permission) {
		panic("middlewares: unknown permission " + permission)
	}

	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			role, ok := r.Context().Value(ContextKeyUserRole).(string)
			if !ok {
				// RequirePermission was wired without RequireAuth in front of it.
				panic(exceptions.NewUnAuthorized("authorization required"))
			}

			if !models.RoleCan(role, permission) {
				panic(exceptions.NewForbidden("insufficient permission for this action"))
			}

			next(w, r, p)
		}
	}
}

// resolveRole loads the user's current role, in its own short transaction so the
// connection is released before the handler runs.
func (m *AuthMiddleware) resolveRole(ctx context.Context, userId int) string {
	tx, err := m.DB.Begin()
	helpers.PanicIfError(err)

	user, err := m.RepositoryUserInterface.FindById(ctx, tx, userId)
	if err != nil {
		// The token is valid but the account is gone — treat as unauthenticated.
		// Rollback failure must not mask that cause, so it is deliberately ignored.
		_ = tx.Rollback()
		panic(exceptions.NewUnAuthorized("authorization invalid"))
	}

	helpers.PanicIfError(tx.Commit())

	return models.NormalizeRole(user.Role)
}
