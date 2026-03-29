package middlewares

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesPOIPoint "github.com/malikabdulaziz/tmn-backend/repositories/poipoint"
	webPOIPoint "github.com/malikabdulaziz/tmn-backend/web/poipoint"
)

type POIPointMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesPOIPoint.RepositoryPOIPointInterface
}

func NewPOIPointMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoPOIPoint repositoriesPOIPoint.RepositoryPOIPointInterface,
) *POIPointMiddleware {
	return &POIPointMiddleware{
		Validate:                    validate,
		DB:                          db,
		RepositoryPOIPointInterface: repoPOIPoint,
	}
}

// ValidateCreate validates create POI point request
func (m *POIPointMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webPOIPoint.CreatePOIPointRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createPOIPointRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

// ValidateUpdate validates update POI point request and verifies point exists
func (m *POIPointMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webPOIPoint.UpdatePOIPointRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid POI point id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositoryPOIPointInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("POI point not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updatePOIPointRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("poiPointId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
