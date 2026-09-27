package middlewares

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesQuotation "github.com/malikabdulaziz/tmn-backend/repositories/quotation"
	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
)

type QuotationMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesQuotation.RepositoryQuotationInterface
}

func NewQuotationMiddleware(validate *validator.Validate, db *sql.DB, repo repositoriesQuotation.RepositoryQuotationInterface) *QuotationMiddleware {
	return &QuotationMiddleware{Validate: validate, DB: db, RepositoryQuotationInterface: repo}
}

func (m *QuotationMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webQuotation.CreateQuotationRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))
		validateSelections(m.Validate, req.Placement, req.Bonus)

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createQuotationRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *QuotationMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webQuotation.UpdateQuotationRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))
		validateSelections(m.Validate, req.Placement, req.Bonus)

		id := parseId(p, "invalid quotation id")

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		if _, err := m.RepositoryQuotationInterface.FindById(r.Context(), tx, id); err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("quotation not found"))
		} else {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateQuotationRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("quotationId"), id)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *QuotationMiddleware) ValidateReturn(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webQuotation.ReturnQuotationRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("returnQuotationRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *QuotationMiddleware) ValidatePricingPreview(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webQuotation.PricingPreviewRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))
		validateSelections(m.Validate, req.Placement, req.Bonus)

		ctx := context.WithValue(r.Context(), helpers.ContextKey("pricingPreviewRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

// validateSelections checks the nested selections, which struct validation on the
// parent does not reach through a pointer.
func validateSelections(validate *validator.Validate, selections ...*webQuotation.SelectionRequest) {
	for _, selection := range selections {
		if selection == nil {
			continue
		}
		helpers.PanicIfError(validate.Struct(*selection))
	}
}
