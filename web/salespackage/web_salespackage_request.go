package salespackage

import (
	"strings"
)

type CreateSalesPackageRequest struct {
	Name        string `json:"name" validate:"required"`
	BuildingIds []int  `json:"building_ids" validate:"required,min=1"`
}

type UpdateSalesPackageRequest struct {
	Name        string `json:"name" validate:"required"`
	BuildingIds []int  `json:"building_ids" validate:"required,min=1"`
}

type SalesPackageRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
}

func (r *SalesPackageRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *SalesPackageRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *SalesPackageRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *SalesPackageRequestFindAll) GetTake() int {
	return r.take
}

func (r *SalesPackageRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *SalesPackageRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *SalesPackageRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *SalesPackageRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
