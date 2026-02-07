package savedpolygon

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesSavedPolygon "github.com/malikabdulaziz/tmn-backend/services/savedpolygon"
	"github.com/malikabdulaziz/tmn-backend/web"
	webSavedPolygon "github.com/malikabdulaziz/tmn-backend/web/savedpolygon"
)

type ControllerSavedPolygonImpl struct {
	service servicesSavedPolygon.ServiceSavedPolygonInterface
}

func NewControllerSavedPolygonImpl(service servicesSavedPolygon.ServiceSavedPolygonInterface) ControllerSavedPolygonInterface {
	return &ControllerSavedPolygonImpl{service: service}
}

// Create handles POST /saved-polygons
func (c *ControllerSavedPolygonImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createSavedPolygonRequest")).(webSavedPolygon.CreateSavedPolygonRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

// FindAll handles GET /saved-polygons
func (c *ControllerSavedPolygonImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webSavedPolygon.SavedPolygonRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// FindById handles GET /saved-polygons/:id
func (c *ControllerSavedPolygonImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid saved polygon id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Update handles PUT /saved-polygons/:id
func (c *ControllerSavedPolygonImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("savedPolygonId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateSavedPolygonRequest")).(webSavedPolygon.UpdateSavedPolygonRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Delete handles DELETE /saved-polygons/:id
func (c *ControllerSavedPolygonImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid saved polygon id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Saved polygon deleted successfully"})
}
