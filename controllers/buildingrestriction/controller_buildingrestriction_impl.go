package buildingrestriction

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
)

type ControllerBuildingRestrictionImpl struct {
	service servicesBuildingRestriction.ServiceBuildingRestrictionInterface
}

func NewControllerBuildingRestrictionImpl(service servicesBuildingRestriction.ServiceBuildingRestrictionInterface) ControllerBuildingRestrictionInterface {
	return &ControllerBuildingRestrictionImpl{service: service}
}

// Create handles POST /building-restrictions
func (c *ControllerBuildingRestrictionImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createBuildingRestrictionRequest")).(webBuildingRestriction.CreateBuildingRestrictionRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

// FindAll handles GET /building-restrictions
func (c *ControllerBuildingRestrictionImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuildingRestriction.BuildingRestrictionRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// FindById handles GET /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building restriction id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Update handles PUT /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("buildingRestrictionId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateBuildingRestrictionRequest")).(webBuildingRestriction.UpdateBuildingRestrictionRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Delete handles DELETE /building-restrictions/:id
func (c *ControllerBuildingRestrictionImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building restriction id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Building restriction deleted successfully"})
}

// Import handles POST /building-restrictions-import
func (c *ControllerBuildingRestrictionImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		panic(exceptions.NewBadRequestError("Failed to parse upload. Max file size is 32MB."))
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		panic(exceptions.NewBadRequestError("File is required. Use form field 'file'."))
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	helpers.PanicIfError(err)

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if ext != "xlsx" && ext != "csv" {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use .xlsx or .csv files."))
	}

	responses := c.service.Import(r.Context(), fileBytes, ext)

	helpers.ReturnReponseJSON(w, web.WebResponse{
		Status: "OK",
		Code:   http.StatusCreated,
		Data:   responses,
	})
}

// Export handles GET /building-restrictions-export
func (c *ControllerBuildingRestrictionImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	search := r.URL.Query().Get("search")

	excelBytes, err := c.service.Export(r.Context(), search)
	helpers.PanicIfError(err)

	filename := "BuildingRestriction_Export_" + time.Now().Format("02-01-2006") + ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(excelBytes)
}
