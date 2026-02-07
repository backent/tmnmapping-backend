package savedpolygon

import (
	"strings"
)

type SavedPolygonPointRequest struct {
	Lat float64 `json:"lat" validate:"required"`
	Lng float64 `json:"lng" validate:"required"`
}

type CreateSavedPolygonRequest struct {
	Name   string                    `json:"name" validate:"required"`
	Points []SavedPolygonPointRequest `json:"points" validate:"required,min=3,dive"`
}

type UpdateSavedPolygonRequest struct {
	Name   string                    `json:"name" validate:"required"`
	Points []SavedPolygonPointRequest `json:"points" validate:"required,min=3,dive"`
}

type SavedPolygonRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
}

func (r *SavedPolygonRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *SavedPolygonRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *SavedPolygonRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *SavedPolygonRequestFindAll) GetTake() int {
	return r.take
}

func (r *SavedPolygonRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *SavedPolygonRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *SavedPolygonRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *SavedPolygonRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
