package middlewares

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	webBuildingPrice "github.com/malikabdulaziz/tmn-backend/web/buildingprice"
)

type BuildingPriceMiddleware struct {
	*validator.Validate
}

func NewBuildingPriceMiddleware(validate *validator.Validate) *BuildingPriceMiddleware {
	return &BuildingPriceMiddleware{Validate: validate}
}

func (m *BuildingPriceMiddleware) ValidateUpsert(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBuildingPrice.UpsertBuildingPriceRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("buildingPriceUpsertRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}
