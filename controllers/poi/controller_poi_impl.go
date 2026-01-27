package poi

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	"github.com/malikabdulaziz/tmn-backend/web"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
)

type ControllerPOIImpl struct {
	service servicesPOI.ServicePOIInterface
}

func NewControllerPOIImpl(service servicesPOI.ServicePOIInterface) ControllerPOIInterface {
	return &ControllerPOIImpl{
		service: service,
	}
}

// Create handles POST /pois
func (controller *ControllerPOIImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createPOIRequest")).(webPOI.CreatePOIRequest)

	poiResponse := controller.service.Create(r.Context(), request)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusCreated,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindAll handles GET /pois
func (controller *ControllerPOIImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webPOI.POIRequestFindAll

	web.SetPagination(&request, r)
	web.SetOrder(&request, r)

	poiResponses, total := controller.service.FindAll(r.Context(), request)

	pagination := web.Pagination{
		Take:  request.GetTake(),
		Skip:  request.GetSkip(),
		Total: total,
	}

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponses,
		Extras: pagination,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindById handles GET /pois/:id
func (controller *ControllerPOIImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI id"))
	}

	poiResponse := controller.service.FindById(r.Context(), poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Update handles PUT /pois/:id
func (controller *ControllerPOIImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId := r.Context().Value(helpers.ContextKey("poiId")).(int)
	request := r.Context().Value(helpers.ContextKey("updatePOIRequest")).(webPOI.UpdatePOIRequest)

	poiResponse := controller.service.Update(r.Context(), request, poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Delete handles DELETE /pois/:id
func (controller *ControllerPOIImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI id"))
	}

	controller.service.Delete(r.Context(), poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   "POI deleted successfully",
	}

	helpers.ReturnReponseJSON(w, response)
}
