package poi

import (
	"strings"
)

type POIPointRequest struct {
	POIName     string  `json:"poi_name" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	MotherBrand string  `json:"mother_brand"`
	Branch      string  `json:"branch"`
}

type CreatePOIRequest struct {
	Brand  string           `json:"brand" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
}

type UpdatePOIRequest struct {
	Brand  string           `json:"brand" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
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
