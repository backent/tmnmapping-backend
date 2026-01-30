package models

import (
	"database/sql"
)

type SalesPackage struct {
	Id        int           `json:"id"`
	Name      string        `json:"name"`
	Buildings []BuildingRef `json:"buildings"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// BuildingRef holds id and name for response (from buildings table)
type BuildingRef struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// SalesPackageBuilding is a junction row (sales_package_buildings table)
type SalesPackageBuilding struct {
	Id              int `json:"id"`
	SalesPackageId  int `json:"sales_package_id"`
	BuildingId      int `json:"building_id"`
}

type NullAbleSalesPackage struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
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
		Id:        int(nullable.Id.Int64),
		Name:      nullable.Name.String,
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
