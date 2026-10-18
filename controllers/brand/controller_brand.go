package brand

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBrand "github.com/malikabdulaziz/tmn-backend/services/brand"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBrand "github.com/malikabdulaziz/tmn-backend/web/brand"
)

type ControllerBrandInterface interface {
	Create(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Import(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Export(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Template(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerBrandImpl struct {
	service servicesBrand.ServiceBrandInterface
}

func NewControllerBrandImpl(service servicesBrand.ServiceBrandInterface) ControllerBrandInterface {
	return &ControllerBrandImpl{service: service}
}

func (c *ControllerBrandImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createBrandRequest")).(webBrand.CreateBrandRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerBrandImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBrand.BrandRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	// Optional customer filter, used by the brand list when drilled into a customer.
	if raw := r.URL.Query().Get("customer_id"); raw != "" {
		if customerId, err := strconv.Atoi(raw); err == nil {
			request.CustomerId = customerId
		}
	}

	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerBrandImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindById(r.Context(), pathId(p, "invalid brand id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerBrandImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("brandId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateBrandRequest")).(webBrand.UpdateBrandRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerBrandImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p, "invalid brand id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Brand deleted successfully"})
}

func (c *ControllerBrandImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	result := c.service.Import(r.Context(), fileBytes, ext)

	// A rejected import is a client problem, not a server one, and the response body
	// carries the per-row reasons.
	code := http.StatusOK
	status := "OK"
	if !result.Imported {
		code = http.StatusBadRequest
		status = "BAD REQUEST"
	}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: status, Code: code, Data: result})
}

func (c *ControllerBrandImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Export(r.Context(), r.URL.Query().Get("search"))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Brands_Export_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerBrandImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Brands_Template.xlsx", fileBytes)
}

func pathId(p httprouter.Params, message string) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest(message))
	}

	return id
}
