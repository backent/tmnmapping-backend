package models

import (
	"database/sql"
)

type POI struct {
	Id        int        `json:"id"`
	Brand     string     `json:"brand"`
	Color     string     `json:"color"`
	Points    []POIPoint `json:"points"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

type POIPoint struct {
	Id          int     `json:"id"`
	POIId       int     `json:"poi_id"`
	POIName     string  `json:"poi_name"`
	Address     string  `json:"address"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	MotherBrand string  `json:"mother_brand"`
	Branch      string  `json:"branch"`
	CreatedAt   string  `json:"created_at"`
}

type NullAblePOI struct {
	Id        sql.NullInt64
	Brand     sql.NullString
	Color     sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

type NullAblePOIPoint struct {
	Id          sql.NullInt64
	POIId       sql.NullInt64
	POIName     sql.NullString
	Address     sql.NullString
	Latitude    sql.NullFloat64
	Longitude   sql.NullFloat64
	Category    sql.NullString
	SubCategory sql.NullString
	MotherBrand sql.NullString
	Branch      sql.NullString
	CreatedAt   sql.NullString
}

var POITable string = "pois"
var POIPointTable string = "poi_points"

func NullAblePOIToPOI(nullable NullAblePOI) POI {
	return POI{
		Id:        int(nullable.Id.Int64),
		Brand:     nullable.Brand.String,
		Color:     nullable.Color.String,
		Points:    []POIPoint{},
		CreatedAt: nullable.CreatedAt.String,
		UpdatedAt: nullable.UpdatedAt.String,
	}
}

func NullAblePOIPointToPOIPoint(nullable NullAblePOIPoint) POIPoint {
	return POIPoint{
		Id:          int(nullable.Id.Int64),
		POIId:       int(nullable.POIId.Int64),
		POIName:     nullable.POIName.String,
		Address:     nullable.Address.String,
		Latitude:    nullable.Latitude.Float64,
		Longitude:   nullable.Longitude.Float64,
		Category:    nullable.Category.String,
		SubCategory: nullable.SubCategory.String,
		MotherBrand: nullable.MotherBrand.String,
		Branch:      nullable.Branch.String,
		CreatedAt:   nullable.CreatedAt.String,
	}
}
