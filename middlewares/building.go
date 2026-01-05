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
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

type BuildingMiddleware struct {
	*validator.Validate
	DB *sql.DB
	repositoriesBuilding.RepositoryBuildingInterface
}

func NewBuildingMiddleware(
	validate *validator.Validate,
	db *sql.DB,
	repositoriesBuilding repositoriesBuilding.RepositoryBuildingInterface,
) *BuildingMiddleware {
	return &BuildingMiddleware{
		Validate:                        validate,
		DB:                              db,
		RepositoryBuildingInterface:     repositoriesBuilding,
	}
}

// ValidateUpdate validates update building request
func (m *BuildingMiddleware) ValidateUpdate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBuilding.UpdateBuildingRequest
		helpers.DecodeRequest(r, &req)

		// Validate struct
		err := m.Validate.Struct(req)
		helpers.PanicIfError(err)

		// Get ID from URL
		buildingId, err := strconv.Atoi(p.ByName("id"))
		if err != nil {
			panic(exceptions.NewBadRequest("invalid building id"))
		}

		tx, err := m.DB.Begin()
		helpers.PanicIfError(err)
		defer helpers.CommitOrRollback(tx)

		// Verify building exists
		_, err = m.RepositoryBuildingInterface.FindById(r.Context(), tx, buildingId)
		if err == sql.ErrNoRows {
			panic(exceptions.NewNotFoundError("building not found"))
		}
		helpers.PanicIfError(err)

		// Store in context
		ctx := context.WithValue(r.Context(), helpers.ContextKey("updateBuildingRequest"), req)
		ctx = context.WithValue(ctx, helpers.ContextKey("buildingId"), buildingId)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

