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
	Id                  int      `json:"id"`
	POIName             string   `json:"poi_name"`
	Address             string   `json:"address"`
	Latitude            float64  `json:"latitude"`
	Longitude           float64  `json:"longitude"`
	CategoryId          *int     `json:"category_id"`
	SubCategoryId       *int     `json:"sub_category_id"`
	MotherBrandId       *int     `json:"mother_brand_id"`
	BranchId            *int     `json:"branch_id"`
	CategoryName        string   `json:"category_name"`
	SubCategoryName     string   `json:"sub_category_name"`
	MotherBrandName     string   `json:"mother_brand_name"`
	BranchName          string   `json:"branch_name"`
	POIs                []POIRef `json:"pois"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type POIRef struct {
	Id    int    `json:"id"`
	Brand string `json:"brand"`
}

type NullAblePOI struct {
	Id        sql.NullInt64
	Brand     sql.NullString
	Color     sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

type NullAblePOIPoint struct {
	Id              sql.NullInt64
	POIName         sql.NullString
	Address         sql.NullString
	Latitude        sql.NullFloat64
	Longitude       sql.NullFloat64
	CategoryId      sql.NullInt64
	SubCategoryId   sql.NullInt64
	MotherBrandId   sql.NullInt64
	BranchId        sql.NullInt64
	CategoryName    sql.NullString
	SubCategoryName sql.NullString
	MotherBrandName sql.NullString
	BranchName      sql.NullString
	CreatedAt       sql.NullString
	UpdatedAt       sql.NullString
}

var POITable string = "pois"
var POIPointTable string = "poi_points"
var POIPointPOITable string = "poi_point_pois"

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
	p := POIPoint{
		Id:              int(nullable.Id.Int64),
		POIName:         nullable.POIName.String,
		Address:         nullable.Address.String,
		Latitude:        nullable.Latitude.Float64,
		Longitude:       nullable.Longitude.Float64,
		CategoryName:    nullable.CategoryName.String,
		SubCategoryName: nullable.SubCategoryName.String,
		MotherBrandName: nullable.MotherBrandName.String,
		BranchName:      nullable.BranchName.String,
		POIs:            []POIRef{},
		CreatedAt:       nullable.CreatedAt.String,
		UpdatedAt:       nullable.UpdatedAt.String,
	}
	if nullable.CategoryId.Valid {
		id := int(nullable.CategoryId.Int64)
		p.CategoryId = &id
	}
	if nullable.SubCategoryId.Valid {
		id := int(nullable.SubCategoryId.Int64)
		p.SubCategoryId = &id
	}
	if nullable.MotherBrandId.Valid {
		id := int(nullable.MotherBrandId.Int64)
		p.MotherBrandId = &id
	}
	if nullable.BranchId.Valid {
		id := int(nullable.BranchId.Int64)
		p.BranchId = &id
	}
	return p
}
