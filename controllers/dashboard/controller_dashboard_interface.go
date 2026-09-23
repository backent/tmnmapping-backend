package dashboard

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ControllerDashboardInterface interface {
	GetLOIReport(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
