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
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
)

type POIMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesPOI.RepositoryPOIInterface
}

func NewPOIMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repositoryPOI repositoriesPOI.RepositoryPOIInterface,
) *POIMiddleware {
	return &POIMiddleware{
		Validate:                validate,
		DB:                      db,
		RepositoryPOIInterface: repositoryPOI,
	}
}

// ValidateCreate validates create POI request
func (m *POIMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webPOI.CreatePOIRequest
		helpers.DecodeRequest(r, &req)

		// Validate struct
		err := m.Validate.Struct(req)
		helpers.PanicIfError(err)

		// Store in context
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createPOIRequest"), req)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

// ValidateUpdate validates update POI request
func (m *POIMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webPOI.UpdatePOIRequest
		helpers.DecodeRequest(r, &req)

		// Validate struct
		err := m.Validate.Struct(req)
		helpers.PanicIfError(err)

		// Get ID from URL
		poiId, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid POI id"))
		}

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		// Verify POI exists
		_, err = m.RepositoryPOIInterface.FindById(r.Context(), tx, poiId)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("POI not found"))
		}
		helpers.PanicIfError(err)

		// Store in context
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updatePOIRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("poiId"), poiId)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}
