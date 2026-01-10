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
	take              int
	skip              int
	orderBy           string
	orderDirection    string
	search            string
	buildingStatus    string
	sellable          string
	connectivity      string
	resourceType      string
	competitorLocation *bool
	cbdArea           string
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

func (r *BuildingRequestFindAll) SetSearch(search string) {
	r.search = search
}

func (r *BuildingRequestFindAll) GetSearch() string {
	return r.search
}

func (r *BuildingRequestFindAll) SetBuildingStatus(buildingStatus string) {
	r.buildingStatus = buildingStatus
}

func (r *BuildingRequestFindAll) GetBuildingStatus() string {
	return r.buildingStatus
}

func (r *BuildingRequestFindAll) SetSellable(sellable string) {
	r.sellable = sellable
}

func (r *BuildingRequestFindAll) GetSellable() string {
	return r.sellable
}

func (r *BuildingRequestFindAll) SetConnectivity(connectivity string) {
	r.connectivity = connectivity
}

func (r *BuildingRequestFindAll) GetConnectivity() string {
	return r.connectivity
}

func (r *BuildingRequestFindAll) SetResourceType(resourceType string) {
	r.resourceType = resourceType
}

func (r *BuildingRequestFindAll) GetResourceType() string {
	return r.resourceType
}

func (r *BuildingRequestFindAll) SetCompetitorLocation(competitorLocation *bool) {
	r.competitorLocation = competitorLocation
}

func (r *BuildingRequestFindAll) GetCompetitorLocation() *bool {
	return r.competitorLocation
}

func (r *BuildingRequestFindAll) SetCbdArea(cbdArea string) {
	r.cbdArea = cbdArea
}

func (r *BuildingRequestFindAll) GetCbdArea() string {
	return r.cbdArea
}

var _ web.RequestPagination = (*BuildingRequestFindAll)(nil)
var _ web.RequestOrder = (*BuildingRequestFindAll)(nil)

