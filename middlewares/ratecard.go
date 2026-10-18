package middlewares

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesRateCard "github.com/malikabdulaziz/tmn-backend/repositories/ratecard"
	webRateCard "github.com/malikabdulaziz/tmn-backend/web/ratecard"
)

type RateCardMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesRateCard.RepositoryRateCardInterface
}

func NewRateCardMiddleware(validate *validator.Validate, db *sql.DB, repo repositoriesRateCard.RepositoryRateCardInterface) *RateCardMiddleware {
	return &RateCardMiddleware{Validate: validate, DB: db, RepositoryRateCardInterface: repo}
}

func (m *RateCardMiddleware) ValidateCreateVersion(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webRateCard.CreateVersionRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createRateCardVersionRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *RateCardMiddleware) ValidateUpdateVersion(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webRateCard.UpdateVersionRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		id := parseId(p, "invalid rate card version id")

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		if _, err := m.RepositoryRateCardInterface.FindVersionById(r.Context(), tx, id); err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("rate card version not found"))
		} else {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateRateCardVersionRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("rateCardVersionId"), id)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *RateCardMiddleware) ValidateUpsertBuildingPrice(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webRateCard.UpsertBuildingPriceRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("upsertBuildingPriceRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *RateCardMiddleware) ValidateUpsertPackagePrice(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webRateCard.UpsertPackagePriceRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("upsertPackagePriceRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}
