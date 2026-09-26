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
	repositoriesBrand "github.com/malikabdulaziz/tmn-backend/repositories/brand"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	repositoriesSalesAssignment "github.com/malikabdulaziz/tmn-backend/repositories/salesassignment"
	webBrand "github.com/malikabdulaziz/tmn-backend/web/brand"
	webCustomer "github.com/malikabdulaziz/tmn-backend/web/customer"
	webSalesAssignment "github.com/malikabdulaziz/tmn-backend/web/salesassignment"
)

// CustomerMiddleware validates customer requests.
type CustomerMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesCustomer.RepositoryCustomerInterface
}

func NewCustomerMiddleware(validate *validator.Validate, db *sql.DB, repo repositoriesCustomer.RepositoryCustomerInterface) *CustomerMiddleware {
	return &CustomerMiddleware{Validate: validate, DB: db, RepositoryCustomerInterface: repo}
}

func (m *CustomerMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webCustomer.CreateCustomerRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createCustomerRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *CustomerMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webCustomer.UpdateCustomerRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		id := parseId(p, "invalid customer id")

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		if _, err := m.RepositoryCustomerInterface.FindById(r.Context(), tx, id); err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("customer not found"))
		} else {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateCustomerRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("customerId"), id)
		next(w, r.WithContext(ctx), p)
	}
}

// BrandMiddleware validates brand requests.
type BrandMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesBrand.RepositoryBrandInterface
}

func NewBrandMiddleware(validate *validator.Validate, db *sql.DB, repo repositoriesBrand.RepositoryBrandInterface) *BrandMiddleware {
	return &BrandMiddleware{Validate: validate, DB: db, RepositoryBrandInterface: repo}
}

func (m *BrandMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBrand.CreateBrandRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createBrandRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *BrandMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBrand.UpdateBrandRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		id := parseId(p, "invalid brand id")

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		if _, err := m.RepositoryBrandInterface.FindById(r.Context(), tx, id); err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("brand not found"))
		} else {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateBrandRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("brandId"), id)
		next(w, r.WithContext(ctx), p)
	}
}

// SalesAssignmentMiddleware validates sales assignment requests.
type SalesAssignmentMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesSalesAssignment.RepositorySalesAssignmentInterface
}

func NewSalesAssignmentMiddleware(validate *validator.Validate, db *sql.DB, repo repositoriesSalesAssignment.RepositorySalesAssignmentInterface) *SalesAssignmentMiddleware {
	return &SalesAssignmentMiddleware{Validate: validate, DB: db, RepositorySalesAssignmentInterface: repo}
}

func (m *SalesAssignmentMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSalesAssignment.CreateSalesAssignmentRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createSalesAssignmentRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *SalesAssignmentMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSalesAssignment.UpdateSalesAssignmentRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		id := parseId(p, "invalid sales assignment id")

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		if _, err := m.RepositorySalesAssignmentInterface.FindById(r.Context(), tx, id); err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("sales assignment not found"))
		} else {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateSalesAssignmentRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("salesAssignmentId"), id)
		next(w, r.WithContext(ctx), p)
	}
}

func parseId(p httprouter.Params, message string) int {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		panic(exceptions.NewBadRequest(message))
	}

	return id
}
