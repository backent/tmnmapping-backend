package middlewares

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	webBuildingProject "github.com/malikabdulaziz/tmn-backend/web/buildingproject"
)

type BuildingProjectMiddleware struct {
	*validator.Validate
}

func NewBuildingProjectMiddleware(validate *validator.Validate) *BuildingProjectMiddleware {
	return &BuildingProjectMiddleware{Validate: validate}
}

// ValidateSave serves create and update alike: the form REPLACES the record, so both
// send every column and a blank means null. One shape, one validator.
func (m *BuildingProjectMiddleware) ValidateSave(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var req webBuildingProject.SaveBuildingProjectRequest
		helpers.DecodeRequest(r, &req)
		helpers.PanicIfError(m.Validate.Struct(req))

		ctx := context.WithValue(r.Context(), helpers.ContextKey("buildingProjectSaveRequest"), req)
		next(w, r.WithContext(ctx), p)
	}
}
