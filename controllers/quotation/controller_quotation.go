package quotation

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesQuotation "github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/malikabdulaziz/tmn-backend/web"
	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
)

type ControllerQuotationInterface interface {
	Create(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Update(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Submit(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Approve(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Return(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	PreviewPricing(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	DashboardCounts(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerQuotationImpl struct {
	service servicesQuotation.ServiceQuotationInterface
}

func NewControllerQuotationImpl(service servicesQuotation.ServiceQuotationInterface) ControllerQuotationInterface {
	return &ControllerQuotationImpl{service: service}
}

func (c *ControllerQuotationImpl) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createQuotationRequest")).(webQuotation.CreateQuotationRequest)
	resp := c.service.Create(r.Context(), request, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerQuotationImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webQuotation.QuotationRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	request.Status = r.URL.Query().Get("status")
	request.Mine = r.URL.Query().Get("mine") == "true"
	request.AwaitingMe = r.URL.Query().Get("awaiting_me") == "true"

	list, total := c.service.FindAll(r.Context(), request, actorOf(r))
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerQuotationImpl) FindById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindById(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("quotationId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateQuotationRequest")).(webQuotation.UpdateQuotationRequest)
	resp := c.service.Update(r.Context(), request, id, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Quotation deleted successfully"})
}

func (c *ControllerQuotationImpl) Submit(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.Submit(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) Approve(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.Approve(r.Context(), pathId(p), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) Return(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("returnQuotationRequest")).(webQuotation.ReturnQuotationRequest)
	resp := c.service.Return(r.Context(), pathId(p), request.Comment, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) PreviewPricing(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("pricingPreviewRequest")).(webQuotation.PricingPreviewRequest)
	resp := c.service.PreviewPricing(r.Context(), request, actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerQuotationImpl) DashboardCounts(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.DashboardCounts(r.Context(), actorOf(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// actorOf reads the caller's identity and role, both put on the context by
// RequireAuth. Quotation rules depend on both, so they travel together.
func actorOf(r *http.Request) servicesQuotation.Actor {
	rawId, ok := r.Context().Value(middlewares.ContextKeyUserId).(string)
	if !ok {
		panic(exceptions.NewUnAuthorized("authorization required"))
	}

	id, err := strconv.Atoi(rawId)
	helpers.PanicIfError(err)

	role, _ := r.Context().Value(middlewares.ContextKeyUserRole).(string)

	return servicesQuotation.Actor{UserId: id, Role: role}
}

func pathId(p httprouter.Params) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid quotation id"))
	}

	return id
}
