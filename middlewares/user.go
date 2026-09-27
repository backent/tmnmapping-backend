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
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webUser "github.com/malikabdulaziz/tmn-backend/web/user"
)

type UserMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesUser.RepositoryUserInterface
}

func NewUserMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoUser repositoriesUser.RepositoryUserInterface,
) *UserMiddleware {
	return &UserMiddleware{
		Validate:                validate,
		DB:                      db,
		RepositoryUserInterface: repoUser,
	}
}

func (m *UserMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webUser.CreateUserRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("createUserRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

func (m *UserMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webUser.UpdateUserRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}

		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid user id"))
		}

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		_, err = m.RepositoryUserInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("user not found"))
		}
		helpers.PanicIfError(err)

		// "userIdParam", not "userId": RequireAuth already owns "userId" for the
		// caller's own id, and overwriting it would make an admin editing someone
		// else look like that other person downstream.
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateUserRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("userIdParam"), id)
		next(w, r.WithContext(ctx), p)
	}
}
