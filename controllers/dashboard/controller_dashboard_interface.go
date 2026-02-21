package dashboard

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ControllerDashboardInterface interface {
	GetAcquisitionReport(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetBuildingProposalReport(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetLOIReport(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
