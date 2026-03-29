package poi

import (
	"strings"
)

type CreatePOIRequest struct {
	Brand    string `json:"brand" validate:"required"`
	Color    string `json:"color" validate:"required"`
	PointIds []int  `json:"point_ids" validate:"required,min=1"`
}

type UpdatePOIRequest struct {
	Brand    string `json:"brand" validate:"required"`
	Color    string `json:"color" validate:"required"`
	PointIds []int  `json:"point_ids" validate:"required,min=1"`
}

type POIRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
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

func (r *POIRequestFindAll) SetSearch(search string) {
	r.search = search
}

func (r *POIRequestFindAll) GetSearch() string {
	return r.search
}
