package middlewares

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

type AuthMiddleware struct {
	*validator.Validate
	repositoriesAuth.RepositoryAuthInterface
}

func NewAuthMiddleware(validate *validator.Validate, repositoriesAuth repositoriesAuth.RepositoryAuthInterface) *AuthMiddleware {
	return &AuthMiddleware{
		Validate:                validate,
		RepositoryAuthInterface: repositoriesAuth,
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

// RequireAuth validates JWT token from cookie and adds user ID to context
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

		// Store userId as string in context
		ctx := context.WithValue(r.Context(), helpers.ContextKey("userId"), strconv.Itoa(userId))
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

