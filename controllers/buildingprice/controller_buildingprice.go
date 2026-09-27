package buildingprice

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	servicesBuildingPrice "github.com/malikabdulaziz/tmn-backend/services/buildingprice"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingPrice "github.com/malikabdulaziz/tmn-backend/web/buildingprice"
)

type ControllerBuildingPriceInterface interface {
	FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Upsert(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Import(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Export(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Template(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerBuildingPriceImpl struct {
	service servicesBuildingPrice.ServiceBuildingPriceInterface
}

func NewControllerBuildingPriceImpl(service servicesBuildingPrice.ServiceBuildingPriceInterface) ControllerBuildingPriceInterface {
	return &ControllerBuildingPriceImpl{service: service}
}

func (c *ControllerBuildingPriceImpl) FindAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var request webBuildingPrice.BuildingPriceRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	list, total := c.service.FindAll(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerBuildingPriceImpl) Upsert(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("buildingPriceUpsertRequest")).(webBuildingPrice.UpsertBuildingPriceRequest)
	resp := c.service.Upsert(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerBuildingPriceImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("buildingId"))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid building id"))
	}

	c.service.Delete(r.Context(), id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Building price removed"})
}

// Import previews when dry_run=true: every row is checked and counted, nothing is
// written. The page shows that summary and applies with a second call.
func (c *ControllerBuildingPriceImpl) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	dryRun := r.URL.Query().Get("dry_run") == "true"
	result := c.service.Import(r.Context(), fileBytes, ext, dryRun)

	// Rejected rows are data, listed in the result -- not a failed request, because
	// the valid rows still apply. A file that cannot be read at all (wrong type,
	// missing columns) panics into a 400 before it gets here.
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: result})
}

func (c *ControllerBuildingPriceImpl) Export(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Export(r.Context())
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Building_Prices_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerBuildingPriceImpl) Template(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.Template()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Building_Prices_Template.xlsx", fileBytes)
}
