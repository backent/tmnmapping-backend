package salesassignment

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesSalesAssignment "github.com/malikabdulaziz/tmn-backend/services/salesassignment"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webSalesAssignment "github.com/malikabdulaziz/tmn-backend/web/salesassignment"
)

type ControllerSalesAssignmentInterface interface {
	Create(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Import(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Export(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Template(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerSalesAssignmentImpl struct {
	service servicesSalesAssignment.ServiceSalesAssignmentInterface
}

func NewControllerSalesAssignmentImpl(service servicesSalesAssignment.ServiceSalesAssignmentInterface) ControllerSalesAssignmentInterface {
	return &ControllerSalesAssignmentImpl{service: service}
}

func (c *ControllerSalesAssignmentImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createSalesAssignmentRequest")).(webSalesAssignment.CreateSalesAssignmentRequest)
	resp := c.service.Create(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerSalesAssignmentImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webSalesAssignment.SalesAssignmentRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerSalesAssignmentImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindById(r.Context(), pathId(p, "invalid salesassignment id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerSalesAssignmentImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("salesAssignmentId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateSalesAssignmentRequest")).(webSalesAssignment.UpdateSalesAssignmentRequest)
	resp := c.service.Update(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerSalesAssignmentImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p, "invalid salesassignment id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "SalesAssignment deleted successfully"})
}

func (c *ControllerSalesAssignmentImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

func (c *ControllerSalesAssignmentImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Export(r.Context(), r.URL.Query().Get("search"))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Sales_Assignments_Export_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerSalesAssignmentImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Sales_Assignments_Template.xlsx", fileBytes)
}

func pathId(p httprouter.Params, message string) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest(message))
	}

	return id
}
