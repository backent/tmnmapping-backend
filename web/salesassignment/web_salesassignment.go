package salesassignment

import "strings"

type CreateSalesAssignmentRequest struct {
	CustomerId       int    `json:"customer_id" validate:"required,gt=0"`
	BrandId          int    `json:"brand_id" validate:"required,gt=0"`
	SalesUserId      int    `json:"sales_user_id" validate:"required,gt=0"`
	Status           string `json:"status" validate:"required,oneof=active inactive"`
	RegistrationDate string `json:"registration_date" validate:"omitempty,datetime=2006-01-02"`
	ExpiryDate       string `json:"expiry_date" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateSalesAssignmentRequest struct {
	CustomerId       int    `json:"customer_id" validate:"required,gt=0"`
	BrandId          int    `json:"brand_id" validate:"required,gt=0"`
	SalesUserId      int    `json:"sales_user_id" validate:"required,gt=0"`
	Status           string `json:"status" validate:"required,oneof=active inactive"`
	RegistrationDate string `json:"registration_date" validate:"omitempty,datetime=2006-01-02"`
	ExpiryDate       string `json:"expiry_date" validate:"omitempty,datetime=2006-01-02"`
}

type SalesAssignmentResponse struct {
	Id               int    `json:"id"`
	CustomerId       int    `json:"customer_id"`
	CustomerCode     string `json:"customer_code"`
	CustomerName     string `json:"customer_name"`
	BrandId          int    `json:"brand_id"`
	BrandCode        string `json:"brand_code"`
	BrandName        string `json:"brand_name"`
	SalesUserId      int    `json:"sales_user_id"`
	SalesUsername    string `json:"sales_username"`
	SalesName        string `json:"sales_name"`
	Status           string `json:"status"`
	RegistrationDate string `json:"registration_date"`
	ExpiryDate       string `json:"expiry_date"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type SalesAssignmentRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *SalesAssignmentRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *SalesAssignmentRequestFindAll) SetTake(take int)          { r.take = take }
func (r *SalesAssignmentRequestFindAll) GetSkip() int              { return r.skip }
func (r *SalesAssignmentRequestFindAll) GetTake() int              { return r.take }
func (r *SalesAssignmentRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *SalesAssignmentRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *SalesAssignmentRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *SalesAssignmentRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *SalesAssignmentRequestFindAll) SetSearch(search string) { r.search = search }
func (r *SalesAssignmentRequestFindAll) GetSearch() string       { return r.search }
