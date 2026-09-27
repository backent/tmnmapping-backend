package models

import (
	"database/sql"
)

// SalesPackage is a sellable bundle of buildings, and a priced resource in its own
// right. Its screen count, traffic and impressions are set independently rather than
// summed from its member buildings -- see docs/QUOTATION_DOCUMENT_ANALYSIS.md §4.1.
// The price lives in rate_card_package_prices, versioned like every other price.
type SalesPackage struct {
	Id          int    `json:"id"`
	PackageCode string `json:"package_code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ScreenCount int    `json:"screen_count"`
	Traffic     int    `json:"traffic"`
	Impressions int    `json:"impressions"`

	// What the advertiser pays for one week of the whole package. Set on the
	// package rather than in a rate card -- see migration 020.
	PriceIdrPerWeek int64         `json:"price_idr_per_week"`
	Buildings       []BuildingRef `json:"buildings"`
	CreatedAt       string        `json:"created_at"`
	UpdatedAt       string        `json:"updated_at"`
}

// BuildingRef holds a lightweight subset of building fields used in relation responses
// (sales packages, building restrictions). It carries enough context for the frontend
// selector UI to render rich rows without a separate hydration call.
type BuildingRef struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	ProjectName  string `json:"project_name"`
	Subdistrict  string `json:"subdistrict"`
	Citytown     string `json:"citytown"`
	Province     string `json:"province"`
	BuildingType string `json:"building_type"`
}

// SalesPackageBuilding is a junction row (sales_package_buildings table)
type SalesPackageBuilding struct {
	Id             int `json:"id"`
	SalesPackageId int `json:"sales_package_id"`
	BuildingId     int `json:"building_id"`
}

type NullAbleSalesPackage struct {
	Id              sql.NullInt64
	PackageCode     sql.NullString
	Name            sql.NullString
	Description     sql.NullString
	Status          sql.NullString
	ScreenCount     sql.NullInt64
	Traffic         sql.NullInt64
	Impressions     sql.NullInt64
	PriceIdrPerWeek sql.NullInt64
	CreatedAt       sql.NullString
	UpdatedAt       sql.NullString
}

type NullAbleSalesPackageBuilding struct {
	Id             sql.NullInt64
	SalesPackageId sql.NullInt64
	BuildingId     sql.NullInt64
}

var SalesPackageTable string = "sales_packages"
var SalesPackageBuildingTable string = "sales_package_buildings"

func NullAbleSalesPackageToSalesPackage(nullable NullAbleSalesPackage) SalesPackage {
	return SalesPackage{
		Id:          int(nullable.Id.Int64),
		PackageCode: nullable.PackageCode.String,
		Name:        nullable.Name.String,
		Description: nullable.Description.String,
		Status:      nullable.Status.String,
		ScreenCount: int(nullable.ScreenCount.Int64),
		Traffic:     int(nullable.Traffic.Int64),
		Impressions: int(nullable.Impressions.Int64),

		PriceIdrPerWeek: nullable.PriceIdrPerWeek.Int64,

		Buildings: []BuildingRef{},
		CreatedAt: nullable.CreatedAt.String,
		UpdatedAt: nullable.UpdatedAt.String,
	}
}

func NullAbleSalesPackageBuildingToSalesPackageBuilding(nullable NullAbleSalesPackageBuilding) SalesPackageBuilding {
	return SalesPackageBuilding{
		Id:             int(nullable.Id.Int64),
		SalesPackageId: int(nullable.SalesPackageId.Int64),
		BuildingId:     int(nullable.BuildingId.Int64),
	}
}
