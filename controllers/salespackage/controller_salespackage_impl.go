package salespackage

import (
	"net/http"
	"strconv"

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
