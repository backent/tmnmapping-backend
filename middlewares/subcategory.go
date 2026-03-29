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
	repositoriesSubCategory "github.com/malikabdulaziz/tmn-backend/repositories/subcategory"
	webSubCategory "github.com/malikabdulaziz/tmn-backend/web/subcategory"
)

type SubCategoryMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesSubCategory.RepositorySubCategoryInterface
}

func NewSubCategoryMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoSubCategory repositoriesSubCategory.RepositorySubCategoryInterface,
) *SubCategoryMiddleware {
	return &SubCategoryMiddleware{
		Validate:                       validate,
		DB:                             db,
		RepositorySubCategoryInterface: repoSubCategory,
	}
}

func (m *SubCategoryMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSubCategory.CreateSubCategoryRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createSubCategoryRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *SubCategoryMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webSubCategory.UpdateSubCategoryRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid sub category id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositorySubCategoryInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("sub category not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateSubCategoryRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("subCategoryId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
