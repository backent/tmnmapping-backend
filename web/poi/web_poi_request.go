package poi

import (
	"strings"
)

type POIPointInput struct {
	Id        *int    `json:"id,omitempty"`
	POIName   string  `json:"poi_name" validate:"required"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	BranchId  *int    `json:"branch_id,omitempty"`
}

type CreatePOIRequest struct {
	Brand         string          `json:"brand" validate:"required"`
	Color         string          `json:"color" validate:"required"`
	CategoryId    *int            `json:"category_id" validate:"required"`
	SubCategoryId *int            `json:"sub_category_id" validate:"required"`
	MotherBrandId *int            `json:"mother_brand_id" validate:"required"`
	Points        []POIPointInput `json:"points" validate:"required,min=1,dive"`
}

type UpdatePOIRequest struct {
	Brand         string          `json:"brand" validate:"required"`
	Color         string          `json:"color" validate:"required"`
	CategoryId    *int            `json:"category_id" validate:"required"`
	SubCategoryId *int            `json:"sub_category_id" validate:"required"`
	MotherBrandId *int            `json:"mother_brand_id" validate:"required"`
	Points        []POIPointInput `json:"points" validate:"required,min=1,dive"`
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
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *POIRequestFindAll) GetOrderDirection() string {
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
