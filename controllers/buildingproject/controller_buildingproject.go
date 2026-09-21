package buildingproject

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesBuildingProject "github.com/malikabdulaziz/tmn-backend/services/buildingproject"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingProject "github.com/malikabdulaziz/tmn-backend/web/buildingproject"
)

type ControllerBuildingProjectInterface interface {
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Create(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindChanges(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Import(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Export(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Template(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerBuildingProjectImpl struct {
	service servicesBuildingProject.ServiceBuildingProjectInterface
}

func NewControllerBuildingProjectImpl(service servicesBuildingProject.ServiceBuildingProjectInterface) ControllerBuildingProjectInterface {
	return &ControllerBuildingProjectImpl{service: service}
}

// actorOf reads the caller's identity and role, both put on the context by
// RequireAuth. The role travels with every call because the landlord money is gated
// per column, and because nothing is written without recording who wrote it.
func actorOf(r *http.Request) servicesBuildingProject.Actor {
	rawId, ok := r.Context().Value(middlewares.ContextKeyUserId).(string)
	if !ok {
		panic(exceptions.NewUnAuthorized("authorization required"))
	}

	id, err := strconv.Atoi(rawId)
	helpers.PanicIfError(err)

	role, _ := r.Context().Value(middlewares.ContextKeyUserRole).(string)

	return servicesBuildingProject.Actor{UserId: id, Role: role}
}

func pathId(p httprouter.Params) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid project id"))
	}

	return id
}

func (c *ControllerBuildingProjectImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuildingProject.BuildingProjectRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	request.Status = r.URL.Query().Get("status")
	request.ContractType = r.URL.Query().Get("contract_type")
	request.Pic = r.URL.Query().Get("pic")

	list, total := c.service.FindAll(r.Context(), request, actorOf(r))
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerBuildingProjectImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindById(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerBuildingProjectImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("buildingProjectSaveRequest")).(webBuildingProject.SaveBuildingProjectRequest)
	resp := c.service.Create(r.Context(), request, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerBuildingProjectImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("buildingProjectSaveRequest")).(webBuildingProject.SaveBuildingProjectRequest)
	resp := c.service.Update(r.Context(), request, pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerBuildingProjectImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Building project removed"})
}

func (c *ControllerBuildingProjectImpl) FindChanges(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuildingProject.BuildingProjectRequestFindAll
	web.SetPagination(&request, r)

	list, total := c.service.FindChanges(r.Context(), pathId(p), request.GetTake(), request.GetSkip(), actorOf(r))
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// Import previews when dry_run=true. Blank means CLEAR on this import, so the preview
// reports cleared fields separately -- an operator who reads only "120 updated" has
// not been told the destructive half.
func (c *ControllerBuildingProjectImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	dryRun := r.URL.Query().Get("dry_run") == "true"

	result := c.service.Import(r.Context(), fileBytes, ext, dryRun, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: result})
}

func (c *ControllerBuildingProjectImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Export(r.Context(), actorOf(r))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "TMN_Projects_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerBuildingProjectImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "TMN_Project_Template.xlsx", fileBytes)
}
