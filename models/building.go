package models

import "database/sql"

type Building struct {
	Id                  int     `json:"id"`
	ExternalBuildingId  string  `json:"external_building_id"`
	IrisCode            string  `json:"iris_code"`
	Name                string  `json:"name"`
	ProjectName         string  `json:"project_name"`
	Audience            int     `json:"audience"`
	Impression          int     `json:"impression"`
	CbdArea             string  `json:"cbd_area"`
	BuildingStatus      string  `json:"building_status"`
	CompetitorLocation  bool    `json:"competitor_location"`
	Sellable            string  `json:"sellable"`
	Connectivity        string  `json:"connectivity"`
	ResourceType        string  `json:"resource_type"`
	SyncedAt            string  `json:"synced_at"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type NullAbleBuilding struct {
	Id                  sql.NullInt64
	ExternalBuildingId  sql.NullString
	IrisCode            sql.NullString
	Name                sql.NullString
	ProjectName         sql.NullString
	Audience            sql.NullInt64
	Impression          sql.NullInt64
	CbdArea             sql.NullString
	BuildingStatus      sql.NullString
	CompetitorLocation  sql.NullBool
	Sellable            sql.NullString
	Connectivity        sql.NullString
	ResourceType        sql.NullString
	SyncedAt            sql.NullString
	CreatedAt           sql.NullString
	UpdatedAt           sql.NullString
}

var BuildingTable string = "buildings"

func NullAbleBuildingToBuilding(nullable NullAbleBuilding) Building {
	return Building{
		Id:                  int(nullable.Id.Int64),
		ExternalBuildingId:  nullable.ExternalBuildingId.String,
		IrisCode:            nullable.IrisCode.String,
		Name:                nullable.Name.String,
		ProjectName:         nullable.ProjectName.String,
		Audience:            int(nullable.Audience.Int64),
		Impression:          int(nullable.Impression.Int64),
		CbdArea:             nullable.CbdArea.String,
		BuildingStatus:      nullable.BuildingStatus.String,
		CompetitorLocation:  nullable.CompetitorLocation.Bool,
		Sellable:            nullable.Sellable.String,
		Connectivity:        nullable.Connectivity.String,
		ResourceType:        nullable.ResourceType.String,
		SyncedAt:            nullable.SyncedAt.String,
		CreatedAt:           nullable.CreatedAt.String,
		UpdatedAt:           nullable.UpdatedAt.String,
	}
}

