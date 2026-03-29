package motherbrand

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
	servicesMotherBrand "github.com/malikabdulaziz/tmn-backend/services/motherbrand"
	"github.com/malikabdulaziz/tmn-backend/web"
	webMotherBrand "github.com/malikabdulaziz/tmn-backend/web/motherbrand"
)

type ControllerMotherBrandImpl struct {
	service servicesMotherBrand.ServiceMotherBrandInterface
}

func NewControllerMotherBrandImpl(service servicesMotherBrand.ServiceMotherBrandInterface) ControllerMotherBrandInterface {
	return &ControllerMotherBrandImpl{service: service}
}

func (c *ControllerMotherBrandImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createMotherBrandRequest")).(webMotherBrand.CreateMotherBrandRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerMotherBrandImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webMotherBrand.MotherBrandRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)
	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerMotherBrandImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid mother brand id"))
	}
	resp := c.service.FindById(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerMotherBrandImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("motherBrandId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateMotherBrandRequest")).(webMotherBrand.UpdateMotherBrandRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerMotherBrandImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid mother brand id"))
	}
	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Mother brand deleted successfully"})
}

func (c *ControllerMotherBrandImpl) FindAllDropdown(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webMotherBrand.MotherBrandRequestFindAll
	request.SetTake(100000)
	request.SetSkip(0)
	request.SetOrderBy("name")
	request.SetOrderDirection("ASC")
	list, _ := c.service.FindAll(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list})
}

func (c *ControllerMotherBrandImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

func (c *ControllerMotherBrandImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	search := r.URL.Query().Get("search")

	excelBytes, err := c.service.Export(r.Context(), search)
	helpers.PanicIfError(err)

	filename := "MotherBrand_Export_" + time.Now().Format("02-01-2006") + ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(excelBytes)
}
