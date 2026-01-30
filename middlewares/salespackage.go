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
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
)

type SalesPackageMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesSalesPackage.RepositorySalesPackageInterface
}

func NewSalesPackageMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoSalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface,
) *SalesPackageMiddleware {
	return &SalesPackageMiddleware{
		Validate:                        validate,
		DB:                              db,
		RepositorySalesPackageInterface: repoSalesPackage,
	}
}

// ValidateCreate validates create sales package request
func (m *SalesPackageMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSalesPackage.CreateSalesPackageRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createSalesPackageRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

// ValidateUpdate validates update sales package request and verifies package exists
func (m *SalesPackageMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSalesPackage.UpdateSalesPackageRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid sales package id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositorySalesPackageInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("sales package not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateSalesPackageRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("salesPackageId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
