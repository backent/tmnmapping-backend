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
	repositoriesSavedPolygon "github.com/malikabdulaziz/tmn-backend/repositories/savedpolygon"
	webSavedPolygon "github.com/malikabdulaziz/tmn-backend/web/savedpolygon"
)

type SavedPolygonMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesSavedPolygon.RepositorySavedPolygonInterface
}

func NewSavedPolygonMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoSavedPolygon repositoriesSavedPolygon.RepositorySavedPolygonInterface,
) *SavedPolygonMiddleware {
	return &SavedPolygonMiddleware{
		Validate:                        validate,
		DB:                              db,
		RepositorySavedPolygonInterface: repoSavedPolygon,
	}
}

// ValidateCreate validates create saved polygon request
func (m *SavedPolygonMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSavedPolygon.CreateSavedPolygonRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createSavedPolygonRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

// ValidateUpdate validates update saved polygon request and verifies entity exists
func (m *SavedPolygonMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSavedPolygon.UpdateSavedPolygonRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid saved polygon id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositorySavedPolygonInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("saved polygon not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateSavedPolygonRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("savedPolygonId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
