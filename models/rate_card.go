package models

import "database/sql"

// Rate card version lifecycle.
//
//	draft      -> editable; prices can be uploaded and hand-edited
//	current    -> the one quotations price against; exactly one at a time
//	historical -> superseded, read-only, kept so approved quotations never re-price
const (
	RateCardStatusDraft      = "draft"
	RateCardStatusCurrent    = "current"
	RateCardStatusHistorical = "historical"
)

var RateCardStatuses = []string{RateCardStatusDraft, RateCardStatusCurrent, RateCardStatusHistorical}

func IsValidRateCardStatus(status string) bool {
	for _, known := range RateCardStatuses {
		if known == status {
			return true
		}
	}

	return false
}

// IsEditableRateCardStatus reports whether prices may still be changed.
// Only a draft is editable; publishing freezes the version for good.
func IsEditableRateCardStatus(status string) bool {
	return status == RateCardStatusDraft
}

type RateCardVersion struct {
	Id                 int    `json:"id"`
	VersionCode        string `json:"version_code"`
	Description        string `json:"description"`
	Currency           string `json:"currency"`
	Status             string `json:"status"`
	PublishedByUserId  int    `json:"published_by_user_id"`
	PublishedByName    string `json:"published_by_name"`
	PublishedAt        string `json:"published_at"`
	BuildingPriceCount int    `json:"building_price_count"`
	PackagePriceCount  int    `json:"package_price_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type NullAbleRateCardVersion struct {
	Id                 sql.NullInt64
	VersionCode        sql.NullString
	Description        sql.NullString
	Currency           sql.NullString
	Status             sql.NullString
	PublishedByUserId  sql.NullInt64
	PublishedByName    sql.NullString
	PublishedAt        sql.NullString
	BuildingPriceCount sql.NullInt64
	PackagePriceCount  sql.NullInt64
	CreatedAt          sql.NullString
	UpdatedAt          sql.NullString
}

var RateCardVersionTable string = "rate_card_versions"

func NullAbleRateCardVersionToRateCardVersion(n NullAbleRateCardVersion) RateCardVersion {
	return RateCardVersion{
		Id:                 int(n.Id.Int64),
		VersionCode:        n.VersionCode.String,
		Description:        n.Description.String,
		Currency:           n.Currency.String,
		Status:             n.Status.String,
		PublishedByUserId:  int(n.PublishedByUserId.Int64),
		PublishedByName:    n.PublishedByName.String,
		PublishedAt:        n.PublishedAt.String,
		BuildingPriceCount: int(n.BuildingPriceCount.Int64),
		PackagePriceCount:  int(n.PackagePriceCount.Int64),
		CreatedAt:          n.CreatedAt.String,
		UpdatedAt:          n.UpdatedAt.String,
	}
}

// RateCardBuildingPrice is the weekly rate an advertiser pays for one building.
type RateCardBuildingPrice struct {
	Id                int    `json:"id"`
	RateCardVersionId int    `json:"rate_card_version_id"`
	BuildingId        int    `json:"building_id"`
	BuildingName      string `json:"building_name"`
	BuildingIrisCode  string `json:"building_iris_code"`
	BuildingType      string `json:"building_type"`
	Citytown          string `json:"citytown"`
	PriceIdrPerWeek   int64  `json:"price_idr_per_week"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type NullAbleRateCardBuildingPrice struct {
	Id                sql.NullInt64
	RateCardVersionId sql.NullInt64
	BuildingId        sql.NullInt64
	BuildingName      sql.NullString
	BuildingIrisCode  sql.NullString
	BuildingType      sql.NullString
	Citytown          sql.NullString
	PriceIdrPerWeek   sql.NullInt64
	CreatedAt         sql.NullString
	UpdatedAt         sql.NullString
}

var RateCardBuildingPriceTable string = "rate_card_building_prices"

func NullAbleRateCardBuildingPriceToRateCardBuildingPrice(n NullAbleRateCardBuildingPrice) RateCardBuildingPrice {
	return RateCardBuildingPrice{
		Id:                int(n.Id.Int64),
		RateCardVersionId: int(n.RateCardVersionId.Int64),
		BuildingId:        int(n.BuildingId.Int64),
		BuildingName:      n.BuildingName.String,
		BuildingIrisCode:  n.BuildingIrisCode.String,
		BuildingType:      n.BuildingType.String,
		Citytown:          n.Citytown.String,
		PriceIdrPerWeek:   n.PriceIdrPerWeek.Int64,
		CreatedAt:         n.CreatedAt.String,
		UpdatedAt:         n.UpdatedAt.String,
	}
}

// RateCardPackagePrice is the weekly rate for a whole sales package.
type RateCardPackagePrice struct {
	Id                int    `json:"id"`
	RateCardVersionId int    `json:"rate_card_version_id"`
	SalesPackageId    int    `json:"sales_package_id"`
	SalesPackageName  string `json:"sales_package_name"`
	PriceIdrPerWeek   int64  `json:"price_idr_per_week"`
	BuildingCount     int    `json:"building_count"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type NullAbleRateCardPackagePrice struct {
	Id                sql.NullInt64
	RateCardVersionId sql.NullInt64
	SalesPackageId    sql.NullInt64
	SalesPackageName  sql.NullString
	PriceIdrPerWeek   sql.NullInt64
	BuildingCount     sql.NullInt64
	CreatedAt         sql.NullString
	UpdatedAt         sql.NullString
}

var RateCardPackagePriceTable string = "rate_card_package_prices"

var RateCardPackageBuildingTable string = "rate_card_package_buildings"

func NullAbleRateCardPackagePriceToRateCardPackagePrice(n NullAbleRateCardPackagePrice) RateCardPackagePrice {
	return RateCardPackagePrice{
		Id:                int(n.Id.Int64),
		RateCardVersionId: int(n.RateCardVersionId.Int64),
		SalesPackageId:    int(n.SalesPackageId.Int64),
		SalesPackageName:  n.SalesPackageName.String,
		PriceIdrPerWeek:   n.PriceIdrPerWeek.Int64,
		BuildingCount:     int(n.BuildingCount.Int64),
		CreatedAt:         n.CreatedAt.String,
		UpdatedAt:         n.UpdatedAt.String,
	}
}

// RateCardPackageBuilding freezes which buildings a package contained when the
// version was published.
type RateCardPackageBuilding struct {
	Id                int    `json:"id"`
	RateCardVersionId int    `json:"rate_card_version_id"`
	SalesPackageId    int    `json:"sales_package_id"`
	BuildingId        int    `json:"building_id"`
	BuildingName      string `json:"building_name"`
	CreatedAt         string `json:"created_at"`
}
