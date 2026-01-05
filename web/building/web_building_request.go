package building

import (
	"strings"

	"github.com/malikabdulaziz/tmn-backend/web"
)

type UpdateBuildingRequest struct {
	Sellable     string `json:"sellable" validate:"omitempty,oneof=sell not_sell"`
	Connectivity string `json:"connectivity" validate:"omitempty,oneof=online manual not_yet_checked"`
	ResourceType string `json:"resource_type"`
}

type BuildingRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
}

func (r *BuildingRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *BuildingRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *BuildingRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *BuildingRequestFindAll) GetTake() int {
	return r.take
}

func (r *BuildingRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *BuildingRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BuildingRequestFindAll) GetOrderBy() string {
	// set default order by
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *BuildingRequestFindAll) GetOrderDirection() string {
	// set default order direction
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

var _ web.RequestPagination = (*BuildingRequestFindAll)(nil)
var _ web.RequestOrder = (*BuildingRequestFindAll)(nil)

