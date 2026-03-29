package poipoint

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
	servicesPOIPoint "github.com/malikabdulaziz/tmn-backend/services/poipoint"
	"github.com/malikabdulaziz/tmn-backend/web"
	webPOIPoint "github.com/malikabdulaziz/tmn-backend/web/poipoint"
)

type ControllerPOIPointImpl struct {
	service servicesPOIPoint.ServicePOIPointInterface
}

func NewControllerPOIPointImpl(service servicesPOIPoint.ServicePOIPointInterface) ControllerPOIPointInterface {
	return &ControllerPOIPointImpl{service: service}
}

// Create handles POST /poi-points
func (c *ControllerPOIPointImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createPOIPointRequest")).(webPOIPoint.CreatePOIPointRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

// FindAll handles GET /poi-points
func (c *ControllerPOIPointImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webPOIPoint.POIPointRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// FindById handles GET /poi-points/:id
func (c *ControllerPOIPointImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI point id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Update handles PUT /poi-points/:id
func (c *ControllerPOIPointImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("poiPointId")).(int)
	request := r.Context().Value(helpers.ContextKey("updatePOIPointRequest")).(webPOIPoint.UpdatePOIPointRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Delete handles DELETE /poi-points/:id
func (c *ControllerPOIPointImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI point id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "POI point deleted successfully"})
}

// GetPointUsage handles GET /poi-points/:id/usage
func (c *ControllerPOIPointImpl) GetPointUsage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid POI point id"))
	}
	resp := c.service.GetPointUsage(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// FindAllDropdown handles GET /poi-points-dropdown
func (c *ControllerPOIPointImpl) FindAllDropdown(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Reuse FindAll with a large take and no search to get all points for dropdown
	var request webPOIPoint.POIPointRequestFindAll
	request.SetTake(10000)
	request.SetSkip(0)
	request.SetOrderBy("poi_name")
	request.SetOrderDirection("ASC")
	list, _ := c.service.FindAll(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list})
}

// Import handles POST /poi-points-import
func (c *ControllerPOIPointImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseMultipartForm(32 << 20)
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
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: responses})
}

// Export handles GET /poi-points-export
func (c *ControllerPOIPointImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	search := r.URL.Query().Get("search")

	excelBytes, err := c.service.Export(r.Context(), search)
	helpers.PanicIfError(err)

	filename := "POIPoint_Export_" + time.Now().Format("02-01-2006") + ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(excelBytes)
}
