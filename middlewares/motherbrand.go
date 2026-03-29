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
	repositoriesMotherBrand "github.com/malikabdulaziz/tmn-backend/repositories/motherbrand"
	webMotherBrand "github.com/malikabdulaziz/tmn-backend/web/motherbrand"
)

type MotherBrandMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesMotherBrand.RepositoryMotherBrandInterface
}

func NewMotherBrandMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoMotherBrand repositoriesMotherBrand.RepositoryMotherBrandInterface,
) *MotherBrandMiddleware {
	return &MotherBrandMiddleware{
		Validate:                       validate,
		DB:                             db,
		RepositoryMotherBrandInterface: repoMotherBrand,
	}
}

func (m *MotherBrandMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webMotherBrand.CreateMotherBrandRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createMotherBrandRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *MotherBrandMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webMotherBrand.UpdateMotherBrandRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid mother brand id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositoryMotherBrandInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("mother brand not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateMotherBrandRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("motherBrandId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
