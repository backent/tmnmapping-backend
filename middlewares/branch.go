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
	repositoriesBranch "github.com/malikabdulaziz/tmn-backend/repositories/branch"
	webBranch "github.com/malikabdulaziz/tmn-backend/web/branch"
)

type BranchMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesBranch.RepositoryBranchInterface
}

func NewBranchMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoBranch repositoriesBranch.RepositoryBranchInterface,
) *BranchMiddleware {
	return &BranchMiddleware{
		Validate:                  validate,
		DB:                        db,
		RepositoryBranchInterface: repoBranch,
	}
}

func (m *BranchMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBranch.CreateBranchRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createBranchRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *BranchMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBranch.UpdateBranchRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid branch id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositoryBranchInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("branch not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateBranchRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("branchId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
