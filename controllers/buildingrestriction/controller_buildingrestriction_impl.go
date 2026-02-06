package buildingrestriction

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
)

type ControllerBuildingRestrictionImpl struct {
	service servicesBuildingRestriction.ServiceBuildingRestrictionInterface
}

func NewControllerBuildingRestrictionImpl(service servicesBuildingRestriction.ServiceBuildingRestrictionInterface) ControllerBuildingRestrictionInterface {
	return &ControllerBuildingRestrictionImpl{service: service}
}

// Create handles POST /building-restrictions
func (c *ControllerBuildingRestrictionImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createBuildingRestrictionRequest")).(webBuildingRestriction.CreateBuildingRestrictionRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

// FindAll handles GET /building-restrictions
func (c *ControllerBuildingRestrictionImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuildingRestriction.BuildingRestrictionRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// FindById handles GET /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building restriction id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Update handles PUT /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("buildingRestrictionId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateBuildingRestrictionRequest")).(webBuildingRestriction.UpdateBuildingRestrictionRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Delete handles DELETE /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building restriction id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Building restriction deleted successfully"})
}
