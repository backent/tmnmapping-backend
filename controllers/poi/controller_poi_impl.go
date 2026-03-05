package poi

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
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	"github.com/malikabdulaziz/tmn-backend/web"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
)

type ControllerPOIImpl struct {
	service servicesPOI.ServicePOIInterface
}

func NewControllerPOIImpl(service servicesPOI.ServicePOIInterface) ControllerPOIInterface {
	return &ControllerPOIImpl{
		service: service,
	}
}

// Create handles POST /pois
func (controller *ControllerPOIImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createPOIRequest")).(webPOI.CreatePOIRequest)

	poiResponse := controller.service.Create(r.Context(), request)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusCreated,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindAll handles GET /pois
func (controller *ControllerPOIImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webPOI.POIRequestFindAll

	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	poiResponses, total := controller.service.FindAll(r.Context(), request)

	pagination := web.Pagination{
		Take:  request.GetTake(),
		Skip:  request.GetSkip(),
		Total: total,
	}

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponses,
		Extras: pagination,
	}

	helpers.ReturnReponseJSON(w, response)
}

// FindById handles GET /pois/:id
func (controller *ControllerPOIImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI id"))
	}

	poiResponse := controller.service.FindById(r.Context(), poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Update handles PUT /pois/:id
func (controller *ControllerPOIImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId := r.Context().Value(helpers.ContextKey("poiId")).(int)
	request := r.Context().Value(helpers.ContextKey("updatePOIRequest")).(webPOI.UpdatePOIRequest)

	poiResponse := controller.service.Update(r.Context(), request, poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   poiResponse,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Delete handles DELETE /pois/:id
func (controller *ControllerPOIImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	poiId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI id"))
	}

	controller.service.Delete(r.Context(), poiId)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusOK,
		Data:   "POI deleted successfully",
	}

	helpers.ReturnReponseJSON(w, response)
}

// Import handles POST /pois/import
func (controller *ControllerPOIImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

	// Determine file type from extension
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if ext != "xlsx" && ext != "csv" {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use .xlsx or .csv files."))
	}

	poiResponses := controller.service.Import(r.Context(), fileBytes, ext)

	response := web.WebResponse{
		Status: "OK",
		Code:   http.StatusCreated,
		Data:   poiResponses,
	}

	helpers.ReturnReponseJSON(w, response)
}

// Export handles GET /pois/export
func (controller *ControllerPOIImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	search := r.URL.Query().Get("search")

	excelBytes, err := controller.service.Export(r.Context(), search)
	helpers.PanicIfError(err)

	filename := "POI_Export_" + time.Now().Format("02-01-2006") + ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(excelBytes)
}
