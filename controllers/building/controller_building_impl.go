package building

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

type ControllerBuildingImpl struct {
	servicesBuilding.ServiceBuildingInterface
}

func NewControllerBuildingImpl(servicesBuilding servicesBuilding.ServiceBuildingInterface) ControllerBuildingInterface {
	return &ControllerBuildingImpl{
		ServiceBuildingInterface: servicesBuilding,
	}
}

// FindById handles GET /buildings/:id
func (controller *ControllerBuildingImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	buildingId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building id"))
	}

	buildingResponse := controller.ServiceBuildingInterface.FindById(r.Context(), buildingId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   buildingResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindAll handles GET /buildings
func (controller *ControllerBuildingImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuilding.BuildingRequestFindAll

	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	buildingResponses, total := controller.ServiceBuildingInterface.FindAll(r.Context(), request)

	pagination := web.Pagination{
		Take:  request.GetTake(),
		Skip:  request.GetSkip(),
		Total: total,
	}

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   buildingResponses,
		Extras: pagination,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Update handles PUT /buildings/:id
func (controller *ControllerBuildingImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	buildingId := r.Context().Value(helpers.ContextKey("buildingId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateBuildingRequest")).(webBuilding.UpdateBuildingRequest)

	buildingResponse := controller.ServiceBuildingInterface.Update(r.Context(), request, buildingId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   buildingResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// SyncManual handles POST /buildings/sync
func (controller *ControllerBuildingImpl) SyncManual(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := controller.ServiceBuildingInterface.SyncFromERP(r.Context())
	if err != nil {
		panic(exceptions.NewBadRequest("failed to sync buildings from ERP"))
	}

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   "sync completed successfully",
	}

	helpers.ReturnReponseJSON(w, response)
}

