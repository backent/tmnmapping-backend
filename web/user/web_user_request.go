package user

import "strings"

type CreateUserRequest struct {
	Username            string `json:"username" validate:"required,min=3,max=50"`
	Name                string `json:"name" validate:"required,max=50"`
	Email               string `json:"email" validate:"omitempty,email,max=50"`
	Password            string `json:"password" validate:"required,min=8,max=72"`
	Role                string `json:"role" validate:"required,oneof=admin sales head_of_sales head_of_business_control ceo"`
	CanCreateQuotations bool   `json:"can_create_quotations"`
	SalesGroup          string `json:"sales_group" validate:"omitempty,oneof=sales_team everyone_sales freelancer"`
}

// UpdateUserRequest omits the password on purpose: an empty Password means
// "leave it alone", so editing a profile cannot silently reset someone's login.
// max=72 matches bcrypt's input limit.
type UpdateUserRequest struct {
	Username            string `json:"username" validate:"required,min=3,max=50"`
	Name                string `json:"name" validate:"required,max=50"`
	Email               string `json:"email" validate:"omitempty,email,max=50"`
	Password            string `json:"password" validate:"omitempty,min=8,max=72"`
	Role                string `json:"role" validate:"required,oneof=admin sales head_of_sales head_of_business_control ceo"`
	CanCreateQuotations bool   `json:"can_create_quotations"`
	SalesGroup          string `json:"sales_group" validate:"omitempty,oneof=sales_team everyone_sales freelancer"`
}

type UserRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *UserRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *UserRequestFindAll) SetTake(take int)          { r.take = take }
func (r *UserRequestFindAll) GetSkip() int              { return r.skip }
func (r *UserRequestFindAll) GetTake() int              { return r.take }
func (r *UserRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *UserRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *UserRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *UserRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *UserRequestFindAll) SetSearch(search string) { r.search = search }
func (r *UserRequestFindAll) GetSearch() string       { return r.search }
