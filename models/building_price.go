package models

import "database/sql"

// BuildingPrice is what an advertiser pays for one week on one building.
//
// One row per building and no versions -- see migration 021. History is not kept
// here: a submitted quotation stores its own copy of every price it was given, so
// changing a price never rewrites a quotation already sent.
type BuildingPrice struct {
	Id               int    `json:"id"`
	BuildingId       int    `json:"building_id"`
	BuildingName     string `json:"building_name"`
	BuildingIrisCode string `json:"building_iris_code"`
	BuildingType     string `json:"building_type"`
	Citytown         string `json:"citytown"`
	PriceIdrPerWeek  int64  `json:"price_idr_per_week"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type NullAbleBuildingPrice struct {
	Id               sql.NullInt64
	BuildingId       sql.NullInt64
	BuildingName     sql.NullString
	BuildingIrisCode sql.NullString
	BuildingType     sql.NullString
	Citytown         sql.NullString
	PriceIdrPerWeek  sql.NullInt64
	CreatedAt        sql.NullString
	UpdatedAt        sql.NullString
}

var BuildingPriceTable string = "building_prices"

func NullAbleBuildingPriceToBuildingPrice(n NullAbleBuildingPrice) BuildingPrice {
	return BuildingPrice{
		Id:               int(n.Id.Int64),
		BuildingId:       int(n.BuildingId.Int64),
		BuildingName:     n.BuildingName.String,
		BuildingIrisCode: n.BuildingIrisCode.String,
		BuildingType:     n.BuildingType.String,
		Citytown:         n.Citytown.String,
		PriceIdrPerWeek:  n.PriceIdrPerWeek.Int64,
		CreatedAt:        n.CreatedAt.String,
		UpdatedAt:        n.UpdatedAt.String,
	}
}
