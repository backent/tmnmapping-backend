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
	repositoriesCategory "github.com/malikabdulaziz/tmn-backend/repositories/category"
	webCategory "github.com/malikabdulaziz/tmn-backend/web/category"
)

type CategoryMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesCategory.RepositoryCategoryInterface
}

func NewCategoryMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoCategory repositoriesCategory.RepositoryCategoryInterface,
) *CategoryMiddleware {
	return &CategoryMiddleware{
		Validate:                    validate,
		DB:                          db,
		RepositoryCategoryInterface: repoCategory,
	}
}

func (m *CategoryMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webCategory.CreateCategoryRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createCategoryRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *CategoryMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webCategory.UpdateCategoryRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid category id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositoryCategoryInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("category not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateCategoryRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("categoryId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
