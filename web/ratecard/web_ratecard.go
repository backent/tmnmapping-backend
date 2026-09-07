package ratecard

import "strings"

type CreateVersionRequest struct {
	VersionCode string `json:"version_code" validate:"required,max=50"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Currency    string `json:"currency" validate:"omitempty,len=3"`

	// CopyFromVersionId seeds the new draft with an existing version's prices, so a
	// re-price starts from the previous numbers rather than a blank sheet.
	CopyFromVersionId int `json:"copy_from_version_id"`
}

type UpdateVersionRequest struct {
	VersionCode string `json:"version_code" validate:"required,max=50"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Currency    string `json:"currency" validate:"omitempty,len=3"`
}

type UpsertBuildingPriceRequest struct {
	BuildingId        int   `json:"building_id" validate:"required,gt=0"`
	PriceIdrPer4Weeks int64 `json:"price_idr_per_4_weeks" validate:"gte=0"`
}

type UpsertPackagePriceRequest struct {
	SalesPackageId    int   `json:"sales_package_id" validate:"required,gt=0"`
	PriceIdrPer4Weeks int64 `json:"price_idr_per_4_weeks" validate:"gte=0"`
}

type VersionResponse struct {
	Id                 int    `json:"id"`
	VersionCode        string `json:"version_code"`
	Description        string `json:"description"`
	Currency           string `json:"currency"`
	Status             string `json:"status"`
	IsEditable         bool   `json:"is_editable"`
	PublishedByUserId  int    `json:"published_by_user_id"`
	PublishedByName    string `json:"published_by_name"`
	PublishedAt        string `json:"published_at"`
	BuildingPriceCount int    `json:"building_price_count"`
	PackagePriceCount  int    `json:"package_price_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type BuildingPriceResponse struct {
	Id                int    `json:"id"`
	RateCardVersionId int    `json:"rate_card_version_id"`
	BuildingId        int    `json:"building_id"`
	BuildingName      string `json:"building_name"`
	BuildingIrisCode  string `json:"building_iris_code"`
	BuildingType      string `json:"building_type"`
	Citytown          string `json:"citytown"`
	PriceIdrPer4Weeks int64  `json:"price_idr_per_4_weeks"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type PackagePriceResponse struct {
	Id                int    `json:"id"`
	RateCardVersionId int    `json:"rate_card_version_id"`
	SalesPackageId    int    `json:"sales_package_id"`
	SalesPackageName  string `json:"sales_package_name"`
	PriceIdrPer4Weeks int64  `json:"price_idr_per_4_weeks"`
	BuildingCount     int    `json:"building_count"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type RateCardRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *RateCardRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *RateCardRequestFindAll) SetTake(take int)          { r.take = take }
func (r *RateCardRequestFindAll) GetSkip() int              { return r.skip }
func (r *RateCardRequestFindAll) GetTake() int              { return r.take }
func (r *RateCardRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *RateCardRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *RateCardRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *RateCardRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *RateCardRequestFindAll) SetSearch(search string) { r.search = search }
func (r *RateCardRequestFindAll) GetSearch() string       { return r.search }
