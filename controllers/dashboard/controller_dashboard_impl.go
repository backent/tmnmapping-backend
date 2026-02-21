package dashboard

import (
	"net/http"

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

// GetAcquisitionReport handles GET /dashboard/acquisition?pic=&month=
func (c *ControllerDashboardImpl) GetAcquisitionReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := r.URL.Query().Get("pic")
	month := r.URL.Query().Get("month")

	report := c.service.GetAcquisitionReport(r.Context(), pic, month)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}

// GetBuildingProposalReport handles GET /dashboard/building-proposal?pic=&month=
func (c *ControllerDashboardImpl) GetBuildingProposalReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := r.URL.Query().Get("pic")
	month := r.URL.Query().Get("month")

	report := c.service.GetBuildingProposalReport(r.Context(), pic, month)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}

// GetLOIReport handles GET /dashboard/loi?pic=&month=
func (c *ControllerDashboardImpl) GetLOIReport(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	pic := r.URL.Query().Get("pic")
	month := r.URL.Query().Get("month")

	report := c.service.GetLOIReport(r.Context(), pic, month)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   report,
	})
}
