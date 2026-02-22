package building

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ControllerBuildingInterface interface {
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	SyncManual(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetFilterOptions(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAllForMapping(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ExportMappingBuildings(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetLCDPresenceSummary(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

