package building

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

type ControllerBuildingImpl struct {
	service servicesBuilding.ServiceBuildingInterface
}

func NewControllerBuildingImpl(service servicesBuilding.ServiceBuildingInterface) ControllerBuildingInterface {
	return &ControllerBuildingImpl{
		service: service,
	}
}

// FindById handles GET /buildings/:id
func (controller *ControllerBuildingImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	buildingId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building id"))
	}

	buildingResponse := controller.service.FindById(r.Context(), buildingId)

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
	web.SetFilters(&request, r)

	buildingResponses, total := controller.service.FindAll(r.Context(), request)

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

	buildingResponse := controller.service.Update(r.Context(), request, buildingId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   buildingResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// SyncManual handles POST /buildings/sync
func (controller *ControllerBuildingImpl) SyncManual(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := controller.service.SyncFromERP(r.Context())
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

// GetFilterOptions handles GET /buildings/filter-options
func (controller *ControllerBuildingImpl) GetFilterOptions(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filterOptions := controller.service.GetFilterOptions(r.Context())

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   filterOptions,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindAllForMapping handles GET /mapping-buildings
func (controller *ControllerBuildingImpl) FindAllForMapping(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuilding.MappingBuildingRequest

	web.SetMappingFilters(&request, r)

	mappingResponse := controller.service.FindAllForMapping(r.Context(), request)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   mappingResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// ExportMappingBuildings handles POST /admin/mapping-building/export (body: filters + map_center, bounds null)
func (controller *ControllerBuildingImpl) ExportMappingBuildings(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body webBuilding.ExportMappingByFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		panic(exceptions.NewBadRequest("invalid request body"))
	}
	request := webBuilding.BuildMappingRequestFromExportBody(&body)
	excelBytes, err := controller.service.ExportForMappingWithFilters(r.Context(), request)
	if err != nil {
		panic(exceptions.NewBadRequest("export failed: " + err.Error()))
	}
	filename := "Target Media Nusantara - Mapping Building List - " + time.Now().Format("02-01-2006") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(excelBytes)
}

// GetLCDPresenceSummary handles GET /dashboard/building-lcd-presence
func (controller *ControllerBuildingImpl) GetLCDPresenceSummary(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	summaryResponse := controller.service.GetLCDPresenceSummary(r.Context())

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   summaryResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}
