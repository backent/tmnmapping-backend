package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	servicesAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	"github.com/malikabdulaziz/tmn-backend/web"
	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

type ControllerAuthImpl struct {
	*sql.DB
	servicesAuth.ServiceAuthInterface
	repositoriesUser.RepositoryUserInterface
}

func NewControllerAuthImpl(db *sql.DB, servicesAuth servicesAuth.ServiceAuthInterface, repositoriesUser repositoriesUser.RepositoryUserInterface) ControllerAuthInterface {
	return &ControllerAuthImpl{
		DB:                      db,
		ServiceAuthInterface:    servicesAuth,
		RepositoryUserInterface: repositoriesUser,
	}
}

func (implementation *ControllerAuthImpl) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Get validated request from context (set by middleware)
	loginReq := r.Context().Value(helpers.ContextKey("loginRequest")).(webAuth.LoginRequest)

	// Extract client IP address
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	// Call service with validated data
	response, token, maxAge := implementation.ServiceAuthInterface.Login(r.Context(), loginReq.Username, loginReq.Password, ipAddress, loginReq.Remember)

	// Set HTTP-only cookie; MaxAge matches token expiry (2 days if remember, default otherwise)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})

	webResponse := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   response,
	}

	helpers.ReturnReponseJSON(w, webResponse)
}

func (implementation *ControllerAuthImpl) Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Clear the auth_token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Expire immediately
	})

	webResponse := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   "logged out successfully",
	}

	helpers.ReturnReponseJSON(w, webResponse)
}

func (implementation *ControllerAuthImpl) CurrentUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Get userId from context (set by RequireAuth middleware)
	userIdStr := r.Context().Value(helpers.ContextKey("userId")).(string)
	userId, err := strconv.Atoi(userIdStr)
	helpers.PanicIfError(err)

	// Fetch user from database
	tx, err := implementation.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	user, err := implementation.RepositoryUserInterface.FindById(context.Background(), tx, userId)
	helpers.PanicIfError(err)

	// Fetch last login (best-effort — empty string if no log exists yet)
	var lastLogin string
	loginLog, err := implementation.RepositoryUserInterface.FindLastLoginByUserId(context.Background(), tx, userId)
	if err == nil {
		lastLogin = loginLog.LoggedInAt
	}

	// Return user response
	response := webAuth.UserResponse{
		Id:        user.Id,
		Username:  user.Username,
		Name:      user.Name,
		Role:      user.Role,
		LastLogin: lastLogin,
	}

	webResponse := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   response,
	}

	helpers.ReturnReponseJSON(w, webResponse)
}

