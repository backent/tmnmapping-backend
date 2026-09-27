package customer

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesCustomer "github.com/malikabdulaziz/tmn-backend/services/customer"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webCustomer "github.com/malikabdulaziz/tmn-backend/web/customer"
)

type ControllerCustomerInterface interface {
	Create(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Import(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Export(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Template(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerCustomerImpl struct {
	service servicesCustomer.ServiceCustomerInterface
}

func NewControllerCustomerImpl(service servicesCustomer.ServiceCustomerInterface) ControllerCustomerInterface {
	return &ControllerCustomerImpl{service: service}
}

func (c *ControllerCustomerImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createCustomerRequest")).(webCustomer.CreateCustomerRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerCustomerImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webCustomer.CustomerRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerCustomerImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindById(r.Context(), pathId(p, "invalid customer id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerCustomerImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("customerId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateCustomerRequest")).(webCustomer.UpdateCustomerRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerCustomerImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p, "invalid customer id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Customer deleted successfully"})
}

func (c *ControllerCustomerImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

func (c *ControllerCustomerImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Export(r.Context(), r.URL.Query().Get("search"))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Customers_Export_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerCustomerImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Customers_Template.xlsx", fileBytes)
}

func pathId(p httprouter.Params, message string) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest(message))
	}

	return id
}
