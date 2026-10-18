package customer

import "strings"

type CreateCustomerRequest struct {
	Code     string `json:"code" validate:"required,max=50"`
	Name     string `json:"name" validate:"required,max=255"`
	Industry string `json:"industry" validate:"omitempty,max=100"`
	Status   string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateCustomerRequest struct {
	Code     string `json:"code" validate:"required,max=50"`
	Name     string `json:"name" validate:"required,max=255"`
	Industry string `json:"industry" validate:"omitempty,max=100"`
	Status   string `json:"status" validate:"required,oneof=active inactive"`
}

type CustomerResponse struct {
	Id        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Industry  string `json:"industry"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CustomerRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *CustomerRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *CustomerRequestFindAll) SetTake(take int)          { r.take = take }
func (r *CustomerRequestFindAll) GetSkip() int              { return r.skip }
func (r *CustomerRequestFindAll) GetTake() int              { return r.take }
func (r *CustomerRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *CustomerRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *CustomerRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *CustomerRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *CustomerRequestFindAll) SetSearch(search string) { r.search = search }
func (r *CustomerRequestFindAll) GetSearch() string       { return r.search }
