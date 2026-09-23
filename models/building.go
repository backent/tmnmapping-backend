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
	Id                  int             `json:"id"`
	ExternalBuildingId  string          `json:"external_building_id"`
	IrisCode            string          `json:"iris_code"`
	Name                string          `json:"name"`
	ProjectName         string          `json:"project_name"`

	// ProjectId links the building to a building_projects row (migration 024).
	// ProjectIdIris is the project's key, joined rather than stored: the change log
	// and the spreadsheet both speak in codes, because "PRJ-0001 -> PRJ-0002" is
	// readable and "7 -> 8" is not.
	ProjectId     int    `json:"project_id"`
	ProjectIdIris string `json:"project_id_iris"`
	Audience            int             `json:"audience"`
	Impression          int             `json:"impression"`
	CbdArea             string          `json:"cbd_area"`
	BuildingStatus      string          `json:"building_status"`
	CompetitorLocation  bool            `json:"competitor_location"`
	CompetitorExclusive bool            `json:"competitor_exclusive"`
	CompetitorPresence  bool            `json:"competitor_presence"`
	Sellable            string          `json:"sellable"`
	Connectivity        string          `json:"connectivity"`
	ResourceType        string          `json:"resource_type"`
	Subdistrict         string          `json:"subdistrict"`
	Citytown            string          `json:"citytown"`
	Province            string          `json:"province"`
	GradeResource       string          `json:"grade_resource"`
	BuildingType        string          `json:"building_type"`
	CompletionYear      int             `json:"completion_year"`
	Latitude            float64         `json:"latitude"`
	Longitude           float64         `json:"longitude"`
	LcdPresenceStatus   string          `json:"lcd_presence_status"`
	Images              []BuildingImage `json:"images"`
	SyncedAt            string          `json:"synced_at"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}

type NullAbleBuilding struct {
	Id                  sql.NullInt64
	ProjectId           sql.NullInt64
	ProjectIdIris       sql.NullString
	ExternalBuildingId  sql.NullString
	IrisCode            sql.NullString
	Name                sql.NullString
	ProjectName         sql.NullString
	Audience            sql.NullInt64
	Impression          sql.NullInt64
	CbdArea             sql.NullString
	BuildingStatus      sql.NullString
	CompetitorLocation  sql.NullBool
	CompetitorExclusive sql.NullBool
	CompetitorPresence  sql.NullBool
	Sellable            sql.NullString
	Connectivity        sql.NullString
	ResourceType        sql.NullString
	Subdistrict         sql.NullString
	Citytown            sql.NullString
	Province            sql.NullString
	GradeResource       sql.NullString
	BuildingType        sql.NullString
	CompletionYear      sql.NullInt64
	Latitude            sql.NullFloat64
	Longitude           sql.NullFloat64
	LcdPresenceStatus   sql.NullString
	Images              sql.NullString
	SyncedAt            sql.NullString
	CreatedAt           sql.NullString
	UpdatedAt           sql.NullString
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
		CompetitorExclusive: nullable.CompetitorExclusive.Bool,
		CompetitorPresence:  nullable.CompetitorPresence.Bool,
		Sellable:            nullable.Sellable.String,
		Connectivity:        nullable.Connectivity.String,
		ResourceType:        nullable.ResourceType.String,
		Subdistrict:         nullable.Subdistrict.String,
		Citytown:            nullable.Citytown.String,
		Province:            nullable.Province.String,
		GradeResource:       nullable.GradeResource.String,
		BuildingType:        nullable.BuildingType.String,
		CompletionYear:      int(nullable.CompletionYear.Int64),
		Latitude:            nullable.Latitude.Float64,
		Longitude:           nullable.Longitude.Float64,
		LcdPresenceStatus:   nullable.LcdPresenceStatus.String,
		Images:              images,
		SyncedAt:            nullable.SyncedAt.String,
		CreatedAt:           nullable.CreatedAt.String,
		UpdatedAt:           nullable.UpdatedAt.String,
	}
}

// BuildingChange is one field of one building changing, attributed and timestamped.
// Same shape as BuildingProjectChange: one row per FIELD, because knowing a building
// was touched is not useful while knowing audience went from 0 to 4,200 is.
//
// It exists because a blank cell CLEARS on the spreadsheet import, so a careless
// upload can empty columns across thousands of rows. This is the undo trail.
type BuildingChange struct {
	Id int `json:"id"`

	// BuildingId is null once the building is deleted; the external id and name are
	// copied in, so the history stays readable after the row it describes is gone.
	BuildingId         int    `json:"building_id"`
	ExternalBuildingId string `json:"external_building_id"`
	BuildingName       string `json:"building_name"`

	ActorUserId int    `json:"actor_user_id"`
	ActorName   string `json:"actor_name"`
	ActorRole   string `json:"actor_role"`

	Action string `json:"action"`
	// form, import, or sync. ERP still writes photos here after the cutover, and a
	// change nobody made by hand should say so.
	Source  string `json:"source"`
	BatchId string `json:"batch_id"`

	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`

	CreatedAt string `json:"created_at"`
}

type NullAbleBuildingChange struct {
	Id                 sql.NullInt64
	BuildingId         sql.NullInt64
	ExternalBuildingId sql.NullString
	BuildingName       sql.NullString
	ActorUserId        sql.NullInt64
	ActorName          sql.NullString
	ActorRole          sql.NullString
	Action             sql.NullString
	Source             sql.NullString
	BatchId            sql.NullString
	Field              sql.NullString
	OldValue           sql.NullString
	NewValue           sql.NullString
	CreatedAt          sql.NullString
}

func NullAbleBuildingChangeToChange(n NullAbleBuildingChange) BuildingChange {
	return BuildingChange{
		Id:                 int(n.Id.Int64),
		BuildingId:         int(n.BuildingId.Int64),
		ExternalBuildingId: n.ExternalBuildingId.String,
		BuildingName:       n.BuildingName.String,
		ActorUserId:        int(n.ActorUserId.Int64),
		ActorName:          n.ActorName.String,
		ActorRole:          n.ActorRole.String,
		Action:             n.Action.String,
		Source:             n.Source.String,
		BatchId:            n.BatchId.String,
		Field:              n.Field.String,
		OldValue:           n.OldValue.String,
		NewValue:           n.NewValue.String,
		CreatedAt:          n.CreatedAt.String,
	}
}

const (
	BuildingActionCreated = "created"
	BuildingActionUpdated = "updated"
	BuildingActionDeleted = "deleted"

	BuildingSourceForm   = "form"
	BuildingSourceImport = "import"
	BuildingSourceSync   = "sync"
)

var BuildingChangeTable string = "building_changes"
