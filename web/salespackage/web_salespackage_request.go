package salespackage

import (
	"strings"
)

type CreateSalesPackageRequest struct {
	PackageCode string `json:"package_code" validate:"required,max=50"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Status      string `json:"status" validate:"required,oneof=active inactive"`

	// Set independently rather than summed from the member buildings: a package is
	// a priced resource in its own right. See docs/QUOTATION_DOCUMENT_ANALYSIS.md §4.1.
	ScreenCount int   `json:"screen_count" validate:"gte=0"`
	Traffic     int   `json:"traffic" validate:"gte=0"`
	Impressions int   `json:"impressions" validate:"gte=0"`
	BuildingIds []int `json:"building_ids" validate:"required,min=1"`
}

type UpdateSalesPackageRequest struct {
	PackageCode string `json:"package_code" validate:"required,max=50"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Status      string `json:"status" validate:"required,oneof=active inactive"`

	// Set independently rather than summed from the member buildings: a package is
	// a priced resource in its own right. See docs/QUOTATION_DOCUMENT_ANALYSIS.md §4.1.
	ScreenCount int   `json:"screen_count" validate:"gte=0"`
	Traffic     int   `json:"traffic" validate:"gte=0"`
	Impressions int   `json:"impressions" validate:"gte=0"`
	BuildingIds []int `json:"building_ids" validate:"required,min=1"`
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
