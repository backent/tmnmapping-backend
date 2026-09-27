package models

import "database/sql"

// Selection kinds and modes.
const (
	SelectionKindPlacement = "placement"
	SelectionKindBonus     = "bonus"

	SelectionModeBuilding = "building"
	SelectionModePackage  = "package"
)

// Approval actions recorded on the audit trail.
const (
	ApprovalActionSubmitted   = "submitted"
	ApprovalActionResubmitted = "resubmitted"
	ApprovalActionApproved    = "approved"
	ApprovalActionReturned    = "returned"
)

// Quotation is a priced, discounted proposal for one customer + brand.
//
// It carries two identities: SalesUserId is the commercial owner and drives
// visibility, routing and reporting; CreatedByUserId is whoever entered it and is
// audit only.
type Quotation struct {
	Id          int    `json:"id"`
	QuoteNumber string `json:"quote_number"`

	SalesUserId     int    `json:"sales_user_id"`
	SalesUsername   string `json:"sales_username"`
	SalesName       string `json:"sales_name"`
	CreatedByUserId int    `json:"created_by_user_id"`
	CreatedByName   string `json:"created_by_name"`

	CustomerId   int    `json:"customer_id"`
	CustomerCode string `json:"customer_code"`
	CustomerName string `json:"customer_name"`
	BrandId      int    `json:"brand_id"`
	BrandCode    string `json:"brand_code"`
	BrandName    string `json:"brand_name"`

	RateCardVersionId   int    `json:"rate_card_version_id"`
	RateCardVersionCode string `json:"rate_card_version_code"`

	AttentionTo  string `json:"attention_to"`
	JobTitle     string `json:"job_title"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`

	CampaignYear int    `json:"campaign_year"`
	ValidUntil   string `json:"valid_until"`

	Discount float64 `json:"discount"`
	TaxRate  float64 `json:"tax_rate"`

	Status                 string `json:"status"`
	RequiredApproverUserId int    `json:"required_approver_user_id"`
	RequiredApproverName   string `json:"required_approver_name"`
	Version                int    `json:"version"`

	PlacementGross          int64   `json:"placement_gross"`
	PlacementDiscountAmount int64   `json:"placement_discount_amount"`
	PlacementNet            int64   `json:"placement_net"`
	BonusGross              int64   `json:"bonus_gross"`
	BonusNet                int64   `json:"bonus_net"`
	TotalGross              int64   `json:"total_gross"`
	TotalNet                int64   `json:"total_net"`
	EffectiveDiscountAmount int64   `json:"effective_discount_amount"`
	EffectiveDiscountRate   float64 `json:"effective_discount_rate"`
	Tax                     int64   `json:"tax"`
	TotalIncludingTax       int64   `json:"total_including_tax"`

	Selections []QuotationSelection `json:"selections"`

	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ApprovedAt string `json:"approved_at"`
}

// QuotationSelection is one side of a quotation: Placement or Bonus.
type QuotationSelection struct {
	Id          int    `json:"id"`
	QuotationId int    `json:"quotation_id"`
	Kind        string `json:"kind"`
	Mode        string `json:"mode"`

	SalesPackageId   int    `json:"sales_package_id"`
	SalesPackageName string `json:"sales_package_name"`

	TvcDurationSeconds int `json:"tvc_duration_seconds"`
	Weeks              int `json:"weeks"`
	Spots              int `json:"spots"`

	GrossPrice  int64 `json:"gross_price"`
	Traffic     int64 `json:"traffic"`
	Impressions int64 `json:"impressions"`
	ScreenCount int   `json:"screen_count"`

	Items []QuotationSelectionItem `json:"items"`
}

// QuotationSelectionItem is one building on a selection, snapshotted at submit so a
// later edit to the building cannot rewrite history.
type QuotationSelectionItem struct {
	Id                   int    `json:"id"`
	QuotationSelectionId int    `json:"quotation_selection_id"`
	BuildingId           int    `json:"building_id"`
	BuildingName         string `json:"building_name"`
	BuildingIrisCode     string `json:"building_iris_code"`
	BuildingType         string `json:"building_type"`
	Citytown             string `json:"citytown"`
	UnitPriceIdr         int64  `json:"unit_price_idr"`
	Traffic              int64  `json:"traffic"`
	Impressions          int64  `json:"impressions"`
}

// QuotationApproval is one entry in the audit trail.
type QuotationApproval struct {
	Id          int    `json:"id"`
	QuotationId int    `json:"quotation_id"`
	Version     int    `json:"version"`
	ActorUserId int    `json:"actor_user_id"`
	ActorName   string `json:"actor_name"`
	ActorRole   string `json:"actor_role"`
	Action      string `json:"action"`
	Comment     string `json:"comment"`
	CreatedAt   string `json:"created_at"`
}

var (
	QuotationTable              = "quotations"
	QuotationSelectionTable     = "quotation_selections"
	QuotationSelectionItemTable = "quotation_selection_items"
	QuotationVersionTable       = "quotation_versions"
	QuotationApprovalTable      = "quotation_approvals"
)

type NullAbleQuotation struct {
	Id                      sql.NullInt64
	QuoteNumber             sql.NullString
	SalesUserId             sql.NullInt64
	SalesUsername           sql.NullString
	SalesName               sql.NullString
	CreatedByUserId         sql.NullInt64
	CreatedByName           sql.NullString
	CustomerId              sql.NullInt64
	CustomerCode            sql.NullString
	CustomerName            sql.NullString
	BrandId                 sql.NullInt64
	BrandCode               sql.NullString
	BrandName               sql.NullString
	RateCardVersionId       sql.NullInt64
	RateCardVersionCode     sql.NullString
	AttentionTo             sql.NullString
	JobTitle                sql.NullString
	ContactPhone            sql.NullString
	ContactEmail            sql.NullString
	CampaignYear            sql.NullInt64
	ValidUntil              sql.NullString
	Discount                sql.NullFloat64
	TaxRate                 sql.NullFloat64
	Status                  sql.NullString
	RequiredApproverUserId  sql.NullInt64
	RequiredApproverName    sql.NullString
	Version                 sql.NullInt64
	PlacementGross          sql.NullInt64
	PlacementDiscountAmount sql.NullInt64
	PlacementNet            sql.NullInt64
	BonusGross              sql.NullInt64
	BonusNet                sql.NullInt64
	TotalGross              sql.NullInt64
	TotalNet                sql.NullInt64
	EffectiveDiscountAmount sql.NullInt64
	EffectiveDiscountRate   sql.NullFloat64
	Tax                     sql.NullInt64
	TotalIncludingTax       sql.NullInt64
	CreatedAt               sql.NullString
	UpdatedAt               sql.NullString
	ApprovedAt              sql.NullString
}

func NullAbleQuotationToQuotation(n NullAbleQuotation) Quotation {
	return Quotation{
		Id:                      int(n.Id.Int64),
		QuoteNumber:             n.QuoteNumber.String,
		SalesUserId:             int(n.SalesUserId.Int64),
		SalesUsername:           n.SalesUsername.String,
		SalesName:               n.SalesName.String,
		CreatedByUserId:         int(n.CreatedByUserId.Int64),
		CreatedByName:           n.CreatedByName.String,
		CustomerId:              int(n.CustomerId.Int64),
		CustomerCode:            n.CustomerCode.String,
		CustomerName:            n.CustomerName.String,
		BrandId:                 int(n.BrandId.Int64),
		BrandCode:               n.BrandCode.String,
		BrandName:               n.BrandName.String,
		RateCardVersionId:       int(n.RateCardVersionId.Int64),
		RateCardVersionCode:     n.RateCardVersionCode.String,
		AttentionTo:             n.AttentionTo.String,
		JobTitle:                n.JobTitle.String,
		ContactPhone:            n.ContactPhone.String,
		ContactEmail:            n.ContactEmail.String,
		CampaignYear:            int(n.CampaignYear.Int64),
		ValidUntil:              n.ValidUntil.String,
		Discount:                n.Discount.Float64,
		TaxRate:                 n.TaxRate.Float64,
		Status:                  n.Status.String,
		RequiredApproverUserId:  int(n.RequiredApproverUserId.Int64),
		RequiredApproverName:    n.RequiredApproverName.String,
		Version:                 int(n.Version.Int64),
		PlacementGross:          n.PlacementGross.Int64,
		PlacementDiscountAmount: n.PlacementDiscountAmount.Int64,
		PlacementNet:            n.PlacementNet.Int64,
		BonusGross:              n.BonusGross.Int64,
		BonusNet:                n.BonusNet.Int64,
		TotalGross:              n.TotalGross.Int64,
		TotalNet:                n.TotalNet.Int64,
		EffectiveDiscountAmount: n.EffectiveDiscountAmount.Int64,
		EffectiveDiscountRate:   n.EffectiveDiscountRate.Float64,
		Tax:                     n.Tax.Int64,
		TotalIncludingTax:       n.TotalIncludingTax.Int64,
		Selections:              []QuotationSelection{},
		CreatedAt:               n.CreatedAt.String,
		UpdatedAt:               n.UpdatedAt.String,
		ApprovedAt:              n.ApprovedAt.String,
	}
}
