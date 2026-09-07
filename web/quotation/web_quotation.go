package quotation

import "strings"

// SelectionRequest is one side of a quotation as the client sends it.
//
// The client sends WHAT was selected, never what it costs. Prices are resolved
// server-side from the current rate card, so a tampered or stale client cannot
// change the money.
type SelectionRequest struct {
	Mode string `json:"mode" validate:"required,oneof=building package"`

	// BuildingIds when mode=building. SalesPackageId when mode=package.
	BuildingIds    []int `json:"building_ids"`
	SalesPackageId int   `json:"sales_package_id"`

	TvcDurationSeconds int `json:"tvc_duration_seconds" validate:"required,gt=0"`
	Weeks              int `json:"weeks" validate:"required,gt=0"`
	Spots              int `json:"spots" validate:"required,gt=0"`
}

type CreateQuotationRequest struct {
	// SalesOwnerUserId is optional. It defaults to the caller; a proxy entry user
	// names the owner they are entering on behalf of.
	SalesOwnerUserId int `json:"sales_owner_user_id"`

	CustomerId int `json:"customer_id" validate:"required,gt=0"`
	BrandId    int `json:"brand_id" validate:"required,gt=0"`

	AttentionTo  string `json:"attention_to" validate:"omitempty,max=255"`
	JobTitle     string `json:"job_title" validate:"omitempty,max=255"`
	ContactPhone string `json:"contact_phone" validate:"omitempty,max=50"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email,max=255"`

	CampaignYear int     `json:"campaign_year"`
	ValidUntil   string  `json:"valid_until" validate:"omitempty,datetime=2006-01-02"`
	Discount     float64 `json:"discount" validate:"gte=0,lte=100"`
	TaxRate      float64 `json:"tax_rate" validate:"omitempty,gte=0,lte=1"`

	Placement *SelectionRequest `json:"placement"`
	Bonus     *SelectionRequest `json:"bonus"`
}

type UpdateQuotationRequest = CreateQuotationRequest

type ReturnQuotationRequest struct {
	// Comment is mandatory. Returning a quotation without saying why is the one
	// thing the spec forbids outright.
	Comment string `json:"comment" validate:"required,min=1"`
}

// PricingPreviewRequest powers the wizard's live summary. It runs the same pricing
// the server will apply on submit, so the figure a seller sees is the figure they get.
type PricingPreviewRequest struct {
	Discount  float64           `json:"discount" validate:"gte=0,lte=100"`
	TaxRate   float64           `json:"tax_rate" validate:"omitempty,gte=0,lte=1"`
	Placement *SelectionRequest `json:"placement"`
	Bonus     *SelectionRequest `json:"bonus"`
}

type SelectionItemResponse struct {
	BuildingId       int    `json:"building_id"`
	BuildingName     string `json:"building_name"`
	BuildingIrisCode string `json:"building_iris_code"`
	BuildingType     string `json:"building_type"`
	Citytown         string `json:"citytown"`
	UnitPriceIdr     int64  `json:"unit_price_idr"`
	Traffic          int64  `json:"traffic"`
	Impressions      int64  `json:"impressions"`
}

type SelectionResponse struct {
	Kind               string                  `json:"kind"`
	Mode               string                  `json:"mode"`
	SalesPackageId     int                     `json:"sales_package_id"`
	SalesPackageName   string                  `json:"sales_package_name"`
	TvcDurationSeconds int                     `json:"tvc_duration_seconds"`
	Weeks              int                     `json:"weeks"`
	Spots              int                     `json:"spots"`
	GrossPricePerWeek  int64                   `json:"gross_price_per_week"`
	GrossPrice         int64                   `json:"gross_price"`
	Traffic            int64                   `json:"traffic"`
	Impressions        int64                   `json:"impressions"`
	ScreenCount        int                     `json:"screen_count"`
	Items              []SelectionItemResponse `json:"items"`
}

type PricingResponse struct {
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
}

// ApprovalHint tells the wizard who would approve at the current discount, before
// anything is submitted.
type ApprovalHint struct {
	Band         string `json:"band"`
	ApproverRole string `json:"approver_role"`
	ApproverName string `json:"approver_name"`
	Resolvable   bool   `json:"resolvable"`
	Reason       string `json:"reason"`
}

type PricingPreviewResponse struct {
	Pricing  PricingResponse     `json:"pricing"`
	Approval ApprovalHint        `json:"approval"`
	Sections []SelectionResponse `json:"sections"`
}

type ApprovalResponse struct {
	Version   int    `json:"version"`
	ActorName string `json:"actor_name"`
	ActorRole string `json:"actor_role"`
	Action    string `json:"action"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type QuotationResponse struct {
	Id          int    `json:"id"`
	QuoteNumber string `json:"quote_number"`

	SalesUserId   int    `json:"sales_user_id"`
	SalesName     string `json:"sales_name"`
	CreatedByName string `json:"created_by_name"`
	IsProxyEntry  bool   `json:"is_proxy_entry"`

	CustomerId   int    `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	BrandId      int    `json:"brand_id"`
	BrandName    string `json:"brand_name"`

	RateCardVersionCode string `json:"rate_card_version_code"`

	AttentionTo  string `json:"attention_to"`
	JobTitle     string `json:"job_title"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`

	CampaignYear int     `json:"campaign_year"`
	ValidUntil   string  `json:"valid_until"`
	Discount     float64 `json:"discount"`
	TaxRate      float64 `json:"tax_rate"`

	Status               string `json:"status"`
	IsEditable           bool   `json:"is_editable"`
	RequiredApproverName string `json:"required_approver_name"`
	Version              int    `json:"version"`

	Pricing    PricingResponse     `json:"pricing"`
	Selections []SelectionResponse `json:"selections"`
	Approvals  []ApprovalResponse  `json:"approvals"`

	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ApprovedAt string `json:"approved_at"`
}

type DashboardCounts struct {
	Draft    int `json:"draft"`
	Pending  int `json:"pending"`
	Returned int `json:"returned"`
	Approved int `json:"approved"`
	All      int `json:"all"`
}

type QuotationRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
	Status         string
	Mine           bool
	AwaitingMe     bool
}

func (r *QuotationRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *QuotationRequestFindAll) SetTake(take int)          { r.take = take }
func (r *QuotationRequestFindAll) GetSkip() int              { return r.skip }
func (r *QuotationRequestFindAll) GetTake() int              { return r.take }
func (r *QuotationRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *QuotationRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *QuotationRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *QuotationRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *QuotationRequestFindAll) SetSearch(search string) { r.search = search }
func (r *QuotationRequestFindAll) GetSearch() string       { return r.search }
