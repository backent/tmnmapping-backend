package dashboard

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesDashboard "github.com/malikabdulaziz/tmn-backend/services/dashboard"
	"github.com/malikabdulaziz/tmn-backend/web"
)

type ControllerDashboardImpl struct {
	service servicesDashboard.ServiceDashboardInterface
}

func NewControllerDashboardImpl(service servicesDashboard.ServiceDashboardInterface) ControllerDashboardInterface {
	return &ControllerDashboardImpl{service: service}
}

// GetAcquisitionReport handles GET /dashboard/acquisition?pic=&date_from=&date_to=
func (c *ControllerDashboardImpl) GetAcquisitionReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := strings.Join(r.URL.Query()["pic"], ",")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	report := c.service.GetAcquisitionReport(r.Context(), pic, dateFrom, dateTo)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}

// GetBuildingProposalReport handles GET /dashboard/building-proposal?pic=&date_from=&date_to=
func (c *ControllerDashboardImpl) GetBuildingProposalReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := strings.Join(r.URL.Query()["pic"], ",")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	report := c.service.GetBuildingProposalReport(r.Context(), pic, dateFrom, dateTo)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}

// GetLOIReport handles GET /dashboard/loi?pic=&date_from=&date_to=
func (c *ControllerDashboardImpl) GetLOIReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := strings.Join(r.URL.Query()["pic"], ",")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	report := c.service.GetLOIReport(r.Context(), pic, dateFrom, dateTo)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}
