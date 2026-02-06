package models

import (
	"database/sql"
)

type BuildingRestriction struct {
	Id        int           `json:"id"`
	Name      string        `json:"name"`
	Buildings []BuildingRef `json:"buildings"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// BuildingRestrictionBuilding is a junction row (building_restriction_buildings table)
type BuildingRestrictionBuilding struct {
	Id                    int `json:"id"`
	BuildingRestrictionId int `json:"building_restriction_id"`
	BuildingId            int `json:"building_id"`
}

type NullAbleBuildingRestriction struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

type NullAbleBuildingRestrictionBuilding struct {
	Id                    sql.NullInt64
	BuildingRestrictionId sql.NullInt64
	BuildingId            sql.NullInt64
}

var BuildingRestrictionTable string = "building_restrictions"
var BuildingRestrictionBuildingTable string = "building_restriction_buildings"

func NullAbleBuildingRestrictionToBuildingRestriction(nullable NullAbleBuildingRestriction) BuildingRestriction {
	return BuildingRestriction{
		Id:        int(nullable.Id.Int64),
		Name:      nullable.Name.String,
		Buildings: []BuildingRef{},
		CreatedAt: nullable.CreatedAt.String,
		UpdatedAt: nullable.UpdatedAt.String,
	}
}

func NullAbleBuildingRestrictionBuildingToBuildingRestrictionBuilding(nullable NullAbleBuildingRestrictionBuilding) BuildingRestrictionBuilding {
	return BuildingRestrictionBuilding{
		Id:                    int(nullable.Id.Int64),
		BuildingRestrictionId: int(nullable.BuildingRestrictionId.Int64),
		BuildingId:            int(nullable.BuildingId.Int64),
	}
}
