package models

import (
	"database/sql"
	"encoding/json"
)

type BuildingImage struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Building struct {
	Id                 int             `json:"id"`
	ExternalBuildingId string          `json:"external_building_id"`
	IrisCode           string          `json:"iris_code"`
	Name               string          `json:"name"`
	ProjectName        string          `json:"project_name"`
	Audience           int             `json:"audience"`
	Impression         int             `json:"impression"`
	CbdArea            string          `json:"cbd_area"`
	BuildingStatus     string          `json:"building_status"`
	CompetitorLocation bool            `json:"competitor_location"`
	Sellable           string          `json:"sellable"`
	Connectivity       string          `json:"connectivity"`
	ResourceType       string          `json:"resource_type"`
	Subdistrict        string          `json:"subdistrict"`
	Citytown           string          `json:"citytown"`
	Province           string          `json:"province"`
	GradeResource      string          `json:"grade_resource"`
	BuildingType       string          `json:"building_type"`
	CompletionYear     int             `json:"completion_year"`
	Images             []BuildingImage `json:"images"`
	SyncedAt           string          `json:"synced_at"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

type NullAbleBuilding struct {
	Id                 sql.NullInt64
	ExternalBuildingId sql.NullString
	IrisCode           sql.NullString
	Name               sql.NullString
	ProjectName        sql.NullString
	Audience           sql.NullInt64
	Impression         sql.NullInt64
	CbdArea            sql.NullString
	BuildingStatus     sql.NullString
	CompetitorLocation sql.NullBool
	Sellable           sql.NullString
	Connectivity       sql.NullString
	ResourceType       sql.NullString
	Subdistrict        sql.NullString
	Citytown           sql.NullString
	Province           sql.NullString
	GradeResource      sql.NullString
	BuildingType       sql.NullString
	CompletionYear     sql.NullInt64
	Images             sql.NullString
	SyncedAt           sql.NullString
	CreatedAt          sql.NullString
	UpdatedAt          sql.NullString
}

var BuildingTable string = "buildings"

func NullAbleBuildingToBuilding(nullable NullAbleBuilding) Building {
	// Parse images JSON
	var images []BuildingImage
	if nullable.Images.Valid && nullable.Images.String != "" {
		json.Unmarshal([]byte(nullable.Images.String), &images)
	}
	if images == nil {
		images = []BuildingImage{}
	}

	return Building{
		Id:                 int(nullable.Id.Int64),
		ExternalBuildingId: nullable.ExternalBuildingId.String,
		IrisCode:           nullable.IrisCode.String,
		Name:               nullable.Name.String,
		ProjectName:        nullable.ProjectName.String,
		Audience:           int(nullable.Audience.Int64),
		Impression:         int(nullable.Impression.Int64),
		CbdArea:            nullable.CbdArea.String,
		BuildingStatus:     nullable.BuildingStatus.String,
		CompetitorLocation: nullable.CompetitorLocation.Bool,
		Sellable:           nullable.Sellable.String,
		Connectivity:       nullable.Connectivity.String,
		ResourceType:       nullable.ResourceType.String,
		Subdistrict:        nullable.Subdistrict.String,
		Citytown:           nullable.Citytown.String,
		Province:           nullable.Province.String,
		GradeResource:      nullable.GradeResource.String,
		BuildingType:       nullable.BuildingType.String,
		CompletionYear:     int(nullable.CompletionYear.Int64),
		Images:             images,
		SyncedAt:           nullable.SyncedAt.String,
		CreatedAt:          nullable.CreatedAt.String,
		UpdatedAt:          nullable.UpdatedAt.String,
	}
}
