package salespackage

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
	servicesSalesPackage "github.com/malikabdulaziz/tmn-backend/services/salespackage"
	"github.com/malikabdulaziz/tmn-backend/web"
	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
)

type ControllerSalesPackageImpl struct {
	service servicesSalesPackage.ServiceSalesPackageInterface
}

func NewControllerSalesPackageImpl(service servicesSalesPackage.ServiceSalesPackageInterface) ControllerSalesPackageInterface {
	return &ControllerSalesPackageImpl{service: service}
}

// Create handles POST /sales-packages
func (c *ControllerSalesPackageImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createSalesPackageRequest")).(webSalesPackage.CreateSalesPackageRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

// FindAll handles GET /sales-packages
func (c *ControllerSalesPackageImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webSalesPackage.SalesPackageRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

// FindById handles GET /sales-packages/:id
func (c *ControllerSalesPackageImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid sales package id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Update handles PUT /sales-packages/:id
func (c *ControllerSalesPackageImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("salesPackageId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateSalesPackageRequest")).(webSalesPackage.UpdateSalesPackageRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// Delete handles DELETE /sales-packages/:id
func (c *ControllerSalesPackageImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid sales package id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Sales package deleted successfully"})
}

// Import handles POST /sales-packages-import
func (c *ControllerSalesPackageImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

// Export handles GET /sales-packages-export
func (c *ControllerSalesPackageImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	search := r.URL.Query().Get("search")

	excelBytes, err := c.service.Export(r.Context(), search)
	helpers.PanicIfError(err)

	filename := "SalesPackage_Export_" + time.Now().Format("02-01-2006") + ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(excelBytes)
}
