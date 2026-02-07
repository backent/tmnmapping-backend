package models

import (
	"database/sql"
)

type SavedPolygon struct {
	Id        int                 `json:"id"`
	Name      string              `json:"name"`
	Points    []SavedPolygonPoint `json:"points"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

type SavedPolygonPoint struct {
	Id             int     `json:"id"`
	SavedPolygonId int     `json:"saved_polygon_id"`
	Ord            int     `json:"ord"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	CreatedAt      string  `json:"created_at"`
}

type NullAbleSavedPolygon struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

type NullAbleSavedPolygonPoint struct {
	Id             sql.NullInt64
	SavedPolygonId sql.NullInt64
	Ord            sql.NullInt64
	Lat            sql.NullFloat64
	Lng            sql.NullFloat64
	CreatedAt      sql.NullString
}

var SavedPolygonTable string = "saved_polygons"
var SavedPolygonPointTable string = "saved_polygon_points"

func NullAbleSavedPolygonToSavedPolygon(nullable NullAbleSavedPolygon) SavedPolygon {
	return SavedPolygon{
		Id:        int(nullable.Id.Int64),
		Name:      nullable.Name.String,
		Points:    []SavedPolygonPoint{},
		CreatedAt: nullable.CreatedAt.String,
		UpdatedAt: nullable.UpdatedAt.String,
	}
}

func NullAbleSavedPolygonPointToSavedPolygonPoint(nullable NullAbleSavedPolygonPoint) SavedPolygonPoint {
	return SavedPolygonPoint{
		Id:             int(nullable.Id.Int64),
		SavedPolygonId: int(nullable.SavedPolygonId.Int64),
		Ord:            int(nullable.Ord.Int64),
		Lat:            nullable.Lat.Float64,
		Lng:            nullable.Lng.Float64,
		CreatedAt:      nullable.CreatedAt.String,
	}
}
