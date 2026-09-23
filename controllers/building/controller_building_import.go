package building

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
)

// actorOf reads the caller's identity and role, both put on the context by
// RequireAuth. Nothing is written to a building without recording who wrote it.
func actorOf(r *http.Request) servicesBuilding.Actor {
	rawId, ok := r.Context().Value(middlewares.ContextKeyUserId).(string)
	if !ok {
		panic(exceptions.NewUnAuthorized("authorization required"))
	}

	id, err := strconv.Atoi(rawId)
	helpers.PanicIfError(err)

	role, _ := r.Context().Value(middlewares.ContextKeyUserRole).(string)

	return servicesBuilding.Actor{UserId: id, Role: role}
}

// Import previews when dry_run=true. A blank cell CLEARS on this import, so the
// preview reports cleared fields separately -- an operator who reads only
// "120 updated" has not been told the destructive half.
func (controller *ControllerBuildingImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	dryRun := r.URL.Query().Get("dry_run") == "true"

	result := controller.service.Import(r.Context(), fileBytes, ext, dryRun, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: result})
}

func (controller *ControllerBuildingImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := controller.service.Export(r.Context())
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "TMN_Buildings_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (controller *ControllerBuildingImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := controller.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "TMN_Buildings_Template.xlsx", fileBytes)
}

// FindChanges is who changed what, and when. It is the recovery path after a
// destructive upload, so it is readable by anyone who can read buildings.
func (controller *ControllerBuildingImpl) FindChanges(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building id"))
	}

	take, skip := 50, 0
	if raw := r.URL.Query().Get("take"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			take = parsed
		}
	}
	if raw := r.URL.Query().Get("skip"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			skip = parsed
		}
	}

	list, total := controller.service.FindChanges(r.Context(), id, take, skip)
	pagination := web.Pagination{Take: take, Skip: skip, Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}
