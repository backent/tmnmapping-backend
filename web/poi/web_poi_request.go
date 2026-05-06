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
	CategoryId    *int            `json:"category_id"`
	SubCategoryId *int            `json:"sub_category_id"`
	MotherBrandId *int            `json:"mother_brand_id"`
	Points        []POIPointInput `json:"points" validate:"required,min=1,dive"`
}

type UpdatePOIRequest struct {
	Brand         string          `json:"brand" validate:"required"`
	Color         string          `json:"color" validate:"required"`
	CategoryId    *int            `json:"category_id"`
	SubCategoryId *int            `json:"sub_category_id"`
	MotherBrandId *int            `json:"mother_brand_id"`
	Points        []POIPointInput `json:"points" validate:"required,min=1,dive"`
}

type POIRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
	categoryIds    string
	subCategoryIds string
	motherBrandIds string
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

func (r *POIRequestFindAll) SetCategoryIds(ids string) {
	r.categoryIds = ids
}

func (r *POIRequestFindAll) GetCategoryIds() string {
	return r.categoryIds
}

func (r *POIRequestFindAll) SetSubCategoryIds(ids string) {
	r.subCategoryIds = ids
}

func (r *POIRequestFindAll) GetSubCategoryIds() string {
	return r.subCategoryIds
}

func (r *POIRequestFindAll) SetMotherBrandIds(ids string) {
	r.motherBrandIds = ids
}

func (r *POIRequestFindAll) GetMotherBrandIds() string {
	return r.motherBrandIds
}
