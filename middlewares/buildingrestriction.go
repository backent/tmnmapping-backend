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
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
)

type BuildingRestrictionMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface
}

func NewBuildingRestrictionMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repoBuildingRestriction repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface,
) *BuildingRestrictionMiddleware {
	return &BuildingRestrictionMiddleware{
		Validate:                              validate,
		DB:                                    db,
		RepositoryBuildingRestrictionInterface: repoBuildingRestriction,
	}
}

// ValidateCreate validates create building restriction request
func (m *BuildingRestrictionMiddleware) ValidateCreate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBuildingRestriction.CreateBuildingRestrictionRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		ctx := context.WithValue(r.Context(), helpers.ContextKey("createBuildingRestrictionRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}

// ValidateUpdate validates update building restriction request and verifies restriction exists
func (m *BuildingRestrictionMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBuildingRestriction.UpdateBuildingRestrictionRequest
		helpers.DecodeRequest(r, &req)
		if err := m.Validate.Struct(req); err != nil {
			helpers.PanicIfError(err)
		}
		id, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid building restriction id"))
		}
		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)
		_, err = m.RepositoryBuildingRestrictionInterface.FindById(r.Context(), tx, id)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("building restriction not found"))
		}
		helpers.PanicIfError(err)
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateBuildingRestrictionRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("buildingRestrictionId"), id)
		next(w, r.WithContext(ctx), p)
	}
}
