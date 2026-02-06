package buildingrestriction

import (
	"strings"
)

type CreateBuildingRestrictionRequest struct {
	Name        string `json:"name" validate:"required"`
	BuildingIds []int  `json:"building_ids" validate:"required,min=1"`
}

type UpdateBuildingRestrictionRequest struct {
	Name        string `json:"name" validate:"required"`
	BuildingIds []int  `json:"building_ids" validate:"required,min=1"`
}

type BuildingRestrictionRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
}

func (r *BuildingRestrictionRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *BuildingRestrictionRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *BuildingRestrictionRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *BuildingRestrictionRequestFindAll) GetTake() int {
	return r.take
}

func (r *BuildingRestrictionRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *BuildingRestrictionRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BuildingRestrictionRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *BuildingRestrictionRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
