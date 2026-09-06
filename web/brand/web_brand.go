package brand

import "strings"

type CreateBrandRequest struct {
	Code       string `json:"code" validate:"required,max=50"`
	CustomerId int    `json:"customer_id" validate:"required,gt=0"`
	Name       string `json:"name" validate:"required,max=255"`
	Category   string `json:"category" validate:"omitempty,max=100"`
	Status     string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateBrandRequest struct {
	Code       string `json:"code" validate:"required,max=50"`
	CustomerId int    `json:"customer_id" validate:"required,gt=0"`
	Name       string `json:"name" validate:"required,max=255"`
	Category   string `json:"category" validate:"omitempty,max=100"`
	Status     string `json:"status" validate:"required,oneof=active inactive"`
}

type BrandResponse struct {
	Id           int    `json:"id"`
	Code         string `json:"code"`
	CustomerId   int    `json:"customer_id"`
	CustomerCode string `json:"customer_code"`
	CustomerName string `json:"customer_name"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type BrandRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
	CustomerId     int
}

func (r *BrandRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *BrandRequestFindAll) SetTake(take int)          { r.take = take }
func (r *BrandRequestFindAll) GetSkip() int              { return r.skip }
func (r *BrandRequestFindAll) GetTake() int              { return r.take }
func (r *BrandRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *BrandRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BrandRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *BrandRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *BrandRequestFindAll) SetSearch(search string) { r.search = search }
func (r *BrandRequestFindAll) GetSearch() string       { return r.search }
