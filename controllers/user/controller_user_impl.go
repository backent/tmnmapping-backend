package user

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesUser "github.com/malikabdulaziz/tmn-backend/services/user"
	"github.com/malikabdulaziz/tmn-backend/web"
	webUser "github.com/malikabdulaziz/tmn-backend/web/user"
)

type ControllerUserImpl struct {
	service servicesUser.ServiceUserInterface
}

func NewControllerUserImpl(service servicesUser.ServiceUserInterface) ControllerUserInterface {
	return &ControllerUserImpl{service: service}
}

func (c *ControllerUserImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createUserRequest")).(webUser.CreateUserRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerUserImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webUser.UserRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerUserImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid user id"))
	}

	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerUserImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("userIdParam")).(int)
	request := r.Context().Value(helpers.ContextKey("updateUserRequest")).(webUser.UpdateUserRequest)

	resp := c.service.Update(r.Context(), request, id, actingUserId(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerUserImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid user id"))
	}

	c.service.Delete(r.Context(), id, actingUserId(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "User deleted successfully"})
}

// actingUserId reads the caller's id, which RequireAuth stores as a string.
// These routes sit behind RequireAuth, so a missing value is a wiring bug.
func actingUserId(r *http.Request) int {
	raw, ok := r.Context().Value(middlewares.ContextKeyUserId).(string)
	if !ok {
		panic(exceptions.NewUnAuthorized("authorization required"))
	}

	id, err := strconv.Atoi(raw)
	helpers.PanicIfError(err)

	return id
}
