package poipoint

import (
	"strings"
)

type CreatePOIPointRequest struct {
	POIName     string  `json:"poi_name" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	MotherBrand string  `json:"mother_brand"`
	Branch      string  `json:"branch"`
}

type UpdatePOIPointRequest struct {
	POIName     string  `json:"poi_name" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	MotherBrand string  `json:"mother_brand"`
	Branch      string  `json:"branch"`
}

type POIPointRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *POIPointRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *POIPointRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *POIPointRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *POIPointRequestFindAll) GetTake() int {
	return r.take
}

func (r *POIPointRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *POIPointRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *POIPointRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *POIPointRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *POIPointRequestFindAll) SetSearch(search string) {
	r.search = search
}

func (r *POIPointRequestFindAll) GetSearch() string {
	return r.search
}
