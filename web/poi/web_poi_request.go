package poi

import (
	"strings"
)

type POIPointRequest struct {
	PlaceName string  `json:"place_name" validate:"required"`
	Address   string  `json:"address" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

type CreatePOIRequest struct {
	Name   string           `json:"name" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
}

type UpdatePOIRequest struct {
	Name   string           `json:"name" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
}

type POIRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
}

func (r *POIRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *POIRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *POIRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *POIRequestFindAll) GetTake() int {
	return r.take
}

func (r *POIRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *POIRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *POIRequestFindAll) GetOrderBy() string {
	// set default order by
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *POIRequestFindAll) GetOrderDirection() string {
	// set default order direction
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
