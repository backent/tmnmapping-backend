package models

import (
	"database/sql"
)

type POI struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Points    []POIPoint `json:"points"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type POIPoint struct {
	Id        int     `json:"id"`
	POIId     int     `json:"poi_id"`
	PlaceName string  `json:"place_name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	CreatedAt string  `json:"created_at"`
}

type NullAblePOI struct {
	Id        sql.NullInt64
	Name      sql.NullString
	Color     sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

type NullAblePOIPoint struct {
	Id        sql.NullInt64
	POIId     sql.NullInt64
	PlaceName sql.NullString
	Address   sql.NullString
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
	CreatedAt sql.NullString
}

var POITable string = "pois"
var POIPointTable string = "poi_points"

func NullAblePOIToPOI(nullable NullAblePOI) POI {
	return POI{
		Id:        int(nullable.Id.Int64),
		Name:      nullable.Name.String,
		Color:     nullable.Color.String,
		Points:    []POIPoint{},
		CreatedAt: nullable.CreatedAt.String,
		UpdatedAt: nullable.UpdatedAt.String,
	}
}

func NullAblePOIPointToPOIPoint(nullable NullAblePOIPoint) POIPoint {
	return POIPoint{
		Id:        int(nullable.Id.Int64),
		POIId:     int(nullable.POIId.Int64),
		PlaceName: nullable.PlaceName.String,
		Address:   nullable.Address.String,
		Latitude:  nullable.Latitude.Float64,
		Longitude: nullable.Longitude.Float64,
		CreatedAt: nullable.CreatedAt.String,
	}
}
