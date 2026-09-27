package buildingprice

import "strings"

type UpsertBuildingPriceRequest struct {
	BuildingId      int   `json:"building_id" validate:"required,gt=0"`
	PriceIdrPerWeek int64 `json:"price_idr_per_week" validate:"gte=0"`
}

type BuildingPriceResponse struct {
	Id               int    `json:"id"`
	BuildingId       int    `json:"building_id"`
	BuildingName     string `json:"building_name"`
	BuildingIrisCode string `json:"building_iris_code"`
	BuildingType     string `json:"building_type"`
	Citytown         string `json:"citytown"`
	PriceIdrPerWeek  int64  `json:"price_idr_per_week"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type BuildingPriceRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *BuildingPriceRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *BuildingPriceRequestFindAll) SetTake(take int)          { r.take = take }
func (r *BuildingPriceRequestFindAll) GetSkip() int              { return r.skip }
func (r *BuildingPriceRequestFindAll) GetTake() int              { return r.take }
func (r *BuildingPriceRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *BuildingPriceRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BuildingPriceRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "name"
	}
	return r.orderBy
}

func (r *BuildingPriceRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "ASC"
	}
	return r.orderDirection
}

func (r *BuildingPriceRequestFindAll) SetSearch(search string) { r.search = search }
func (r *BuildingPriceRequestFindAll) GetSearch() string       { return r.search }
