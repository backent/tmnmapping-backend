package ratecard

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesRateCard "github.com/malikabdulaziz/tmn-backend/services/ratecard"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webRateCard "github.com/malikabdulaziz/tmn-backend/web/ratecard"
)

type ControllerRateCardInterface interface {
	CreateVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindAllVersions(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindVersionById(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindCurrentVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	UpdateVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	DeleteVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	PublishVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params)

	FindBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	UpsertBuildingPrice(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	DeleteBuildingPrice(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ImportBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ExportBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	BuildingPriceTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params)

	FindPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	UpsertPackagePrice(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	DeletePackagePrice(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ImportPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ExportPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	PackagePriceTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerRateCardImpl struct {
	service servicesRateCard.ServiceRateCardInterface
}

func NewControllerRateCardImpl(service servicesRateCard.ServiceRateCardInterface) ControllerRateCardInterface {
	return &ControllerRateCardImpl{service: service}
}

func (c *ControllerRateCardImpl) CreateVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("createRateCardVersionRequest")).(webRateCard.CreateVersionRequest)
	resp := c.service.CreateVersion(r.Context(), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: resp})
}

func (c *ControllerRateCardImpl) FindAllVersions(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := listRequest(r)
	list, total := c.service.FindAllVersions(r.Context(), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerRateCardImpl) FindVersionById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindVersionById(r.Context(), pathId(p, "id", "invalid rate card version id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerRateCardImpl) FindCurrentVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.FindCurrentVersion(r.Context())
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerRateCardImpl) UpdateVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := r.Context().Value(helpers.ContextKey("rateCardVersionId")).(int)
	request := r.Context().Value(helpers.ContextKey("updateRateCardVersionRequest")).(webRateCard.UpdateVersionRequest)
	resp := c.service.UpdateVersion(r.Context(), request, id)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerRateCardImpl) DeleteVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.DeleteVersion(r.Context(), pathId(p, "id", "invalid rate card version id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Rate card version deleted successfully"})
}

func (c *ControllerRateCardImpl) PublishVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resp := c.service.PublishVersion(r.Context(), pathId(p, "id", "invalid rate card version id"), actingUserId(r))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

// ---------------------------------------------------------------------------
// Building prices
// ---------------------------------------------------------------------------

func (c *ControllerRateCardImpl) FindBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := listRequest(r)
	list, total := c.service.FindBuildingPrices(r.Context(), pathId(p, "id", "invalid rate card version id"), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerRateCardImpl) UpsertBuildingPrice(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("upsertBuildingPriceRequest")).(webRateCard.UpsertBuildingPriceRequest)
	resp := c.service.UpsertBuildingPrice(r.Context(), pathId(p, "id", "invalid rate card version id"), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerRateCardImpl) DeleteBuildingPrice(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.DeleteBuildingPrice(r.Context(),
		pathId(p, "id", "invalid rate card version id"),
		pathId(p, "buildingId", "invalid building id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Building price removed"})
}

func (c *ControllerRateCardImpl) ImportBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	result := c.service.ImportBuildingPrices(r.Context(), pathId(p, "id", "invalid rate card version id"), fileBytes, ext)
	writeImportResult(w, result)
}

func (c *ControllerRateCardImpl) ExportBuildingPrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.ExportBuildingPrices(r.Context(), pathId(p, "id", "invalid rate card version id"))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Rate_Card_Building_Prices_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerRateCardImpl) BuildingPriceTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.BuildingPriceTemplate()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Rate_Card_Building_Prices_Template.xlsx", fileBytes)
}

// ---------------------------------------------------------------------------
// Package prices
// ---------------------------------------------------------------------------

func (c *ControllerRateCardImpl) FindPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := listRequest(r)
	list, total := c.service.FindPackagePrices(r.Context(), pathId(p, "id", "invalid rate card version id"), request)
	pagination := web.Pagination{Take: request.GetTake(), Skip: request.GetSkip(), Total: total}
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: list, Extras: pagination})
}

func (c *ControllerRateCardImpl) UpsertPackagePrice(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := r.Context().Value(helpers.ContextKey("upsertPackagePriceRequest")).(webRateCard.UpsertPackagePriceRequest)
	resp := c.service.UpsertPackagePrice(r.Context(), pathId(p, "id", "invalid rate card version id"), request)
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: resp})
}

func (c *ControllerRateCardImpl) DeletePackagePrice(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.DeletePackagePrice(r.Context(),
		pathId(p, "id", "invalid rate card version id"),
		pathId(p, "packageId", "invalid sales package id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Package price removed"})
}

func (c *ControllerRateCardImpl) ImportPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, ext := spreadsheets.ReadUpload(r)
	result := c.service.ImportPackagePrices(r.Context(), pathId(p, "id", "invalid rate card version id"), fileBytes, ext)
	writeImportResult(w, result)
}

func (c *ControllerRateCardImpl) ExportPackagePrices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.ExportPackagePrices(r.Context(), pathId(p, "id", "invalid rate card version id"))
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Rate_Card_Package_Prices_"+time.Now().Format("02-01-2006")+".xlsx", fileBytes)
}

func (c *ControllerRateCardImpl) PackagePriceTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fileBytes, err := c.service.PackagePriceTemplate()
	helpers.PanicIfError(err)
	spreadsheets.WriteXLSX(w, "Rate_Card_Package_Prices_Template.xlsx", fileBytes)
}

// ---------------------------------------------------------------------------

func listRequest(r *http.Request) webRateCard.RateCardRequestFindAll {
	var request webRateCard.RateCardRequestFindAll
	web.SetPagination(&request, r)
	web.SetOrder(&request, r)
	web.SetSearch(&request, r)

	return request
}

// writeImportResult mirrors the advertiser importers: a rejected upload is a client
// error, and the body still carries the per-row reasons.
func writeImportResult(w http.ResponseWriter, result web.ImportResult) {
	code := http.StatusOK
	status := "OK"
	if !result.Imported {
		code = http.StatusBadRequest
		status = "BAD REQUEST"
	}

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: status, Code: code, Data: result})
}

func pathId(p httprouter.Params, name string, message string) int {
	id, err := strconv.Atoi(p.ByName(name))
	if err != nil {
		panic(exceptions.NewBadRequest(message))
	}

	return id
}

func actingUserId(r *http.Request) int {
	raw, ok := r.Context().Value(middlewares.ContextKeyUserId).(string)
	if !ok {
		panic(exceptions.NewUnAuthorized("authorization required"))
	}

	id, err := strconv.Atoi(raw)
	helpers.PanicIfError(err)

	return id
}
