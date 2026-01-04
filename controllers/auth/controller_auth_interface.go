package auth

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ControllerAuthInterface interface {
	Login(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CurrentUser(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

